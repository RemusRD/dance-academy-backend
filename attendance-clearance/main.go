package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
)

type BatchCreateUsersEvent struct {
	ClassId string `json:"classId"`
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event BatchCreateUsersEvent) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION"))},
	)
	if err != nil {
		log.Fatal(err)
	}

	svc := dynamodb.New(sess)

	updateExpression := "SET #a = :emptyMap"

	mapAttribute := "attendance"
	output, err := svc.UpdateItem(
		&dynamodb.UpdateItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"type": {
					S: aws.String("CLASS"),
				},
				"id": {
					S: aws.String(event.ClassId),
				},
			},
			TableName:                aws.String(os.Getenv("TABLE_NAME")),
			UpdateExpression:         aws.String(updateExpression),
			ExpressionAttributeNames: map[string]*string{"#a": &mapAttribute},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":emptyMap": {M: make(map[string]*dynamodb.AttributeValue)},
			},
		},
	)

	println(output)

	if err != nil {
		log.Fatal(err)
	}
	return "", nil
}
