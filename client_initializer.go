package ddblocal

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type clientInitializer func(port int) (dynamodbiface.DynamoDBAPI, error)

func (c clientInitializer) InitClient(port int) (dynamodbiface.DynamoDBAPI, error) {
	return c(port)
}

func initClient(port int) (dynamodbiface.DynamoDBAPI, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:    aws.String(fmt.Sprintf("http://localhost:%d", port)),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),
			Region:      aws.String("test"),
		},
	})
	if err != nil {
		return nil, err
	}
	client := dynamodb.New(sess)
	return client, nil
}

// NewClientInitialier returns a new instance of ClientInitializer with default
// configuration.
func NewClientInitialier() ClientInitializer {
	return clientInitializer(initClient)
}
