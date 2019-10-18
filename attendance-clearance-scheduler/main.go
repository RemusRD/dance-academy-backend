package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/iot"
	"log"
	"os"
	"strconv"
	"strings"
)

type AttendanceClearanceEvent struct {
	ClassId string `json:"classId"`
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION"))},
	)
	if err != nil {
		log.Fatal(err)
	}
	cloudWatchEventsClient := cloudwatchevents.New(sess)

	svc := dynamodb.New(sess)

	keyCond := expression.Key("type").Equal(expression.Value("CLASS"))
	proj := expression.NamesList(expression.Name("id"), expression.Name("dayOfWeek"), expression.Name("hourOfDay"))

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithProjection(proj).
		Build()
	if err != nil {
		fmt.Println(err)
	}
	result, err := svc.Query(
		&dynamodb.QueryInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			KeyConditionExpression:    expr.KeyCondition(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(os.Getenv("TABLE_NAME")),
		})

	if err != nil {
		fmt.Println(err)
	}

	var items []Class

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &items)

	parsedClasses := MapToCronFormat(items)
	for _, class := range parsedClasses {

		ruleName := class.Id + "-attendance-clearance"

		_, err = cloudWatchEventsClient.PutRule(
			&cloudwatchevents.PutRuleInput{
				Name:               aws.String(ruleName),
				RoleArn:            aws.String(os.Getenv("ATTENDANCE_CLEARANCE_ROLE_ARN")),
				ScheduleExpression: aws.String("cron(0 " + class.HourOfDay + " ? * " + class.DayOfWeek + " *)"),
			},
		)

		if err != nil {
			log.Fatal(err)
		}
		attendanceClearanceEvent := AttendanceClearanceEvent{
			ClassId: class.Id,
		}
		attendanceClearanceEventJson, err := json.Marshal(attendanceClearanceEvent)
		_, err = cloudWatchEventsClient.PutTargets(&cloudwatchevents.PutTargetsInput{
			Rule: aws.String(ruleName),
			Targets: []*cloudwatchevents.Target{
				{
					Arn:   aws.String(os.Getenv("LAMBDA_ARN")),
					Id:    aws.String("attendance-scheluder-target"),
					Input: aws.String(string(attendanceClearanceEventJson)),
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Class struct {
	Id        string `dynamodbav:"id"`
	HourOfDay string `dynamodbav:"hourOfDay"`
	DayOfWeek string `dynamodbav:"dayOfWeek"`
}

func MapToCronFormat(vs []Class) []Class {
	vsm := make([]Class, len(vs))
	for i, v := range vs {
		vsm[i] = Class{
			Id:        v.Id,
			HourOfDay: MapHourOfDayToUTC(v.HourOfDay),
			DayOfWeek: MapDayOfWeekToCronDayString(v.DayOfWeek),
		}
	}
	return vsm
}
func MapDayOfWeekToCronDayString(day string) string {
	cronDay := day
	if day == "LUNES" {
		cronDay = iot.DayOfWeekMon
	} else if day == "MARTES" {
		cronDay = iot.DayOfWeekTue
	} else if day == "MIERCOLES" {
		cronDay = iot.DayOfWeekWed
	} else if day == "JUEVES" {
		cronDay = iot.DayOfWeekThu
	} else if day == "VIERNES" {
		cronDay = iot.DayOfWeekFri
	} else if day == "SABADO" {
		cronDay = iot.DayOfWeekSat
	} else {
		log.Fatal("Unrecognized day ", day)
	}
	return cronDay
}

func MapHourOfDayToUTC(time string) string {
	timeSlice := strings.Split(time, ":")
	hour := timeSlice[0]
	parsedHour, _ := strconv.Atoi(hour)
	if parsedHour-16 > 0 {
		return strconv.Itoa((parsedHour - 2) - 14)
	} else {
		return "00"
	}
}
