package batch_user_creation_lambda

import (
	"context"
	"encoding/csv"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io"
	"log"
	"os"
	"strings"
)

type BatchCreateUsersEvent struct {
	Csv string `json:"csv"`
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event BatchCreateUsersEvent) (string, error) {

	r := csv.NewReader(strings.NewReader(event.Csv))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		schoolId := record[0]
		username := record[1]
		firstName := record[1]

		if len(record[2]) > 0 {
			firstSurnameFirstLetter := string(record[2][0])
			username += firstSurnameFirstLetter
		}
		if len(record[3]) > 0 {
			secondSurnameFirstLetter := string(record[3][0])
			username += secondSurnameFirstLetter
		}
		username = strings.ToLower(strings.TrimSpace(username));
		var role string
		if record[4] == "M" {
			role = "GIRL"
		} else {
			role = "BOY"
		}
		phone_number := record[5]
		mail := record[6]

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION"))},
			)
		if err != nil {
			log.Fatal(err)
		}

		cognitoClient := cognitoidentityprovider.New(sess)

		user, err := cognitoClient.AdminCreateUser(
			&cognitoidentityprovider.AdminCreateUserInput{
				DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
				ForceAliasCreation:     nil,
				MessageAction:          nil,
				TemporaryPassword:      nil,
				UserAttributes: []*cognitoidentityprovider.AttributeType{
					{
						Name:  aws.String("email"),
						Value: aws.String(mail),
					},
					{
						Name:  aws.String("email_verified"),
						Value: aws.String("true"),
					},
				},
				UserPoolId:             aws.String(os.Getenv("USER_POOL_ID")),
				Username:               aws.String(username),
				ValidationData:         nil,
			},

			)
		if err != nil {
			log.Fatal(err)
		}

		svc := dynamodb.New(sess)

		type Class struct {
			Type string `dynamodbav:"type"`
			Id string `dynamodbav:"id"`
			Name string `dynamodbav:"name"`
			Phone_Number string `dynamodbav:"phone_number"`
			Role string `dynamodbav:"role"`
			Classes []string `dynamodbav:"classes"`
			SchoolId string `dynamodbav:"schoolId"`
		}

		userInfo := Class {
			Type:         "USER",
			Id:           *user.User.Attributes[0].Value,
			Name:         strings.Title(strings.ToLower(firstName)),
			Phone_Number: "+34 " + phone_number,
			Role:         role,
			Classes:      []string{"b2aad21f-e7d9-46ff-8be9-3333599f7d84"},
			SchoolId:     schoolId,
		}

		av, err := dynamodbattribute.MarshalMap(userInfo)

		input := &dynamodb.PutItemInput{
			Item:                        av,
			TableName:                   aws.String(os.Getenv("TABLE_NAME")),
		}
		_,err = svc.PutItem(input)
		if err != nil {
			log.Fatal(err)
		}

	}
	return "OK", nil
}



