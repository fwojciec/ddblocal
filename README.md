# ddblocal

`ddblocal` is a wrapper around a local instance of [DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html) for use in testing.

It will start a local DynamoDB server at the start of a test run and tear it down at the end of the run or it will reuse a running instance of DynamoDB.

It provides a `Runner` higher order function which can be used to run test logic against a DynamoDB table in isolation by forcing random table names.

## Example use

```go
var ddb *ddblocal.Emulator

func TestMain(m *testing.M) {
	exitCode, err := runTests(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}

func runTests(m *testing.M) (int, error) {
	var err error
	ddb, err = ddblocal.New()
	if err != nil {
		return 0, fmt.Errorf("failed to start the DynamoDB local emulator")
	}
	defer ddb.Close()
	exitCode := m.Run()
	return exitCode, nil
}

func TestExampleIntegration(t *testing.T) {
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
```
