package ddblocal_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/fwojciec/ddblocal"
	"github.com/fwojciec/ddblocal/mocks"
)

var (
	ddb         *ddblocal.Emulator
	integration = flag.Bool("integration", false, "also perform integration tests")
)

func TestMain(m *testing.M) {
	flag.Parse()
	exitCode, err := runTests(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}

func runTests(m *testing.M) (int, error) {
	if !*integration {
		return m.Run(), nil
	}
	var err error
	ddb, err = ddblocal.New()
	if err != nil {
		return 0, fmt.Errorf("failed to start the DynamoDB local emulator")
	}
	defer ddb.Close()
	exitCode := m.Run()
	return exitCode, nil
}

func TestInitDefaults(t *testing.T) {
	t.Parallel()

	os.Setenv("DDBLOCAL_LIB", "test_lib_path")
	os.Setenv("DDBLOCAL_JAR", "test_jar_path")
	t.Cleanup(func() {
		os.Unsetenv("DDBLOCAL_LIB")
		os.Unsetenv("DDBLOCAL_JAR")
	})

	var recName string
	var recArgs []string
	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc: func(name string, arg ...string) error {
			recName = name
			recArgs = arg
			return nil
		},
	}

	// need custom presence checker because during integration test the
	// server is present which prevents the executorMock from being called
	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	_, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
	)
	ok(t, err)

	equals(t, "java", recName)
	equals(
		t,
		[]string{
			"-Djava.library.path=test_lib_path",
			"-jar",
			"test_jar_path",
			"-port",
			"8000",
			"-sharedDb",
			"-inMemory",
		},
		recArgs,
	)
}

func TestInitCustomPort(t *testing.T) {
	t.Parallel()

	var recArgs []string
	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc: func(_ string, arg ...string) error {
			recArgs = arg
			return nil
		},
	}

	var recPortPC int
	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool {
			recPortPC = port
			return false
		},
	}

	var recPortCI int
	cim := &mocks.ClientInitializerMock{
		InitClientFunc: func(port int) (dynamodbiface.DynamoDBAPI, error) {
			recPortCI = port
			return nil, nil
		},
	}

	_, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
		ddblocal.CustomClientInitializer(cim),
		ddblocal.CustomPort(8888),
	)
	ok(t, err)

	equals(t, 8888, recPortPC)
	equals(t, 8888, recPortCI)
	assert(t, contains([]string{"-port", "8888"}, recArgs), "should pass the correct port value in args")
}

func TestInitCustomLibPath(t *testing.T) {
	t.Parallel()

	var recArgs []string
	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc: func(_ string, arg ...string) error {
			recArgs = arg
			return nil
		},
	}

	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	_, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
		ddblocal.CustomLibPath("test_lib_path"),
	)
	ok(t, err)

	assert(t, contains([]string{"-Djava.library.path=test_lib_path"}, recArgs), "should pass the correct lib path value in args")
}

func TestInitCustomJarPath(t *testing.T) {
	t.Parallel()

	var recArgs []string
	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc: func(_ string, arg ...string) error {
			recArgs = arg
			return nil
		},
	}

	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	_, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
		ddblocal.CustomJarPath("test_jar_path"),
	)
	ok(t, err)

	assert(t, contains([]string{"-jar", "test_jar_path"}, recArgs), "should pass the correct lib path value in args")
}

func TestClose(t *testing.T) {
	t.Parallel()

	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc:   func(name string, arg ...string) error { return nil },
		TerminateFunc: func() error { return nil },
	}

	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	ddb, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
	)
	ok(t, err)

	err = ddb.Close()
	ok(t, err)

	equals(t, 1, len(etm.TerminateCalls()))
}

func TestRunnerCreatesTableCorrectly(t *testing.T) {
	t.Parallel()

	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc:   func(name string, arg ...string) error { return nil },
		TerminateFunc: func() error { return nil },
	}

	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	sgm := &mocks.StringGeneratorMock{
		GenerateFunc: func() (string, error) {
			return "test_name", nil
		},
	}

	var recCreateTableInput *dynamodb.CreateTableInput
	ddbcm := &mocks.DynamoDBAPIMock{
		CreateTableFunc: func(in1 *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error) {
			recCreateTableInput = in1
			return nil, nil
		},
		DeleteTableFunc: func(in1 *dynamodb.DeleteTableInput) (*dynamodb.DeleteTableOutput, error) {
			return nil, nil
		},
	}
	cim := &mocks.ClientInitializerMock{
		InitClientFunc: func(port int) (dynamodbiface.DynamoDBAPI, error) {
			return ddbcm, nil
		},
	}

	ddb, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
		ddblocal.CustomStringGenerator(sgm),
		ddblocal.CustomClientInitializer(cim),
	)
	ok(t, err)

	var recTableName string
	tableInput := &dynamodb.CreateTableInput{
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{IndexName: aws.String("GSI")},
		},
	}
	ddb.Runner(t, tableInput, func(_ dynamodbiface.DynamoDBAPI, tableName string) {
		recTableName = tableName
	})
	equals(t, "test_name", recTableName)

	expCreateTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String("test_name"),
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI"),
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(1),
					WriteCapacityUnits: aws.Int64(1),
				},
			},
		},
	}
	equals(t, expCreateTableInput, recCreateTableInput)
}

func TestRunnerDeletesTableCorrectly(t *testing.T) {
	t.Parallel()

	etm := &mocks.ExecutorTerminatorMock{
		ExecuteFunc:   func(name string, arg ...string) error { return nil },
		TerminateFunc: func() error { return nil },
	}

	pcm := &mocks.PresenceCheckerMock{
		IsPresentFunc: func(port int) bool { return false },
	}

	sgm := &mocks.StringGeneratorMock{
		GenerateFunc: func() (string, error) {
			return "test_name", nil
		},
	}

	var recDeleteTableInput *dynamodb.DeleteTableInput
	ddbcm := &mocks.DynamoDBAPIMock{
		CreateTableFunc: func(in1 *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error) {
			return nil, nil
		},
		DeleteTableFunc: func(in1 *dynamodb.DeleteTableInput) (*dynamodb.DeleteTableOutput, error) {
			recDeleteTableInput = in1
			return nil, nil
		},
	}
	cim := &mocks.ClientInitializerMock{
		InitClientFunc: func(port int) (dynamodbiface.DynamoDBAPI, error) {
			return ddbcm, nil
		},
	}

	ddb, err := ddblocal.New(
		ddblocal.CustomExecutorTerminator(etm),
		ddblocal.CustomPresenceChecker(pcm),
		ddblocal.CustomStringGenerator(sgm),
		ddblocal.CustomClientInitializer(cim),
	)
	ok(t, err)

	t.Run("subtest", func(t *testing.T) {
		// this needs to run in a subtest so that the Runner's
		// t.Cleanup routine has a chance to run
		var testTableInput = &dynamodb.CreateTableInput{}
		var recTableName string
		ddb.Runner(t, testTableInput, func(_ dynamodbiface.DynamoDBAPI, tableName string) {
			recTableName = tableName
		})
		equals(t, "test_name", recTableName)
	})

	expDeleteTableInput := &dynamodb.DeleteTableInput{
		TableName: aws.String("test_name"),
	}
	equals(t, expDeleteTableInput, recDeleteTableInput)
}

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

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

func contains(needle []string, hay []string) bool {
	hayLine := strings.Join(hay, " ")
	needleLine := strings.Join(needle, " ")
	return strings.Contains(hayLine, needleLine)
}
