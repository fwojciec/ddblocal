package ddblocal_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func TestEmulatorIntegration(t *testing.T) {
	if !*integration {
		t.SkipNow()
	}

	tableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Key"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Key"),
				KeyType:       aws.String("HASH"),
			},
		},
	}

	ddb.Runner(t, tableInput, func(client dynamodbiface.DynamoDBAPI, tableName string) {
		_, err := client.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"Key": {S: aws.String("Hello")},
			},
		})
		ok(t, err)

		res, err := client.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"Key": {S: aws.String("Hello")},
			},
		})
		ok(t, err)

		key := res.Item["Key"].S
		assert(t, key != nil, "key should not be nil")
		equals(t, "Hello", *key)
	})
}
