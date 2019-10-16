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

	ruleName := "ec34978a-5c23-40ab-bca1-bf8bb4dfaafe-attendance-clearance"

	cloudWatchEventsClient.PutRule(
		&cloudwatchevents.PutRuleInput{
			Description:        nil,
			EventBusName:       nil,
			EventPattern:       nil,
			Name:               aws.String(ruleName),
			RoleArn:            aws.String(os.Getenv("ATTENDANCE_CLEARANCE_ROLE_ARN")),
			ScheduleExpression: nil,
			State:              nil,
			Tags:               nil,
		},
	)
	attendanceClearanceEvent := AttendanceClearanceEvent{
		ClassId: "ec34978a-5c23-40ab-bca1-bf8bb4dfaafe",
	}
	attendanceClearanceEventJson, err := json.Marshal(attendanceClearanceEvent)
	cloudWatchEventsClient.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []*cloudwatchevents.Target{
			&cloudwatchevents.Target{
				Arn:   aws.String(os.Getenv("LAMBDA_ARN")),
				Id:    aws.String("myCloudWatchEventsTarget"),
				Input: aws.String(string(attendanceClearanceEventJson)),
			},
		},
	})

	return "", nil
}
