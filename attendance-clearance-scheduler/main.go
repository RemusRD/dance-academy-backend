package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"log"
	"os"
)

func main() {
	lambda.Start(HandleRequest)
}

type AttendanceClearanceEvent struct {
	ClassId string `json:"classId"`
}

func HandleRequest(ctx context.Context) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION"))},
	)
	if err != nil {
		log.Fatal(err)
	}
	cloudWatchEventsClient := cloudwatchevents.New(sess)

	ruleName := "512195ea-159c-4d6f-9086-0906ba8024a1-attendance-clearance"

	_, err = cloudWatchEventsClient.PutRule(
		&cloudwatchevents.PutRuleInput{
			Name:               aws.String(ruleName),
			RoleArn:            aws.String(os.Getenv("ATTENDANCE_CLEARANCE_ROLE_ARN")),
			ScheduleExpression: aws.String("0 6 ? * TUE *"),
		},
	)

	if err != nil {
		log.Fatal(err)
	}
	attendanceClearanceEvent := AttendanceClearanceEvent{
		ClassId: "512195ea-159c-4d6f-9086-0906ba8024a1",
	}
	attendanceClearanceEventJson, err := json.Marshal(attendanceClearanceEvent)
	_, err = cloudWatchEventsClient.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []*cloudwatchevents.Target{
			{
				Arn:   aws.String(os.Getenv("LAMBDA_ARN")),
				Id:    aws.String("myCloudWatchEventsTarget"),
				Input: aws.String(string(attendanceClearanceEventJson)),
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return "", nil
}
