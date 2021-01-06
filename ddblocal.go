package ddblocal

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// StringGenerator generates (random) strings.
type StringGenerator interface {
	Generate() (string, error)
}

// ExecutorTerminator executes and terminates the DynamoDB Local process.
type ExecutorTerminator interface {
	Execute(name string, arg ...string) error
	Terminate() error
}

// PresenceChecker checks if DynamoDB local process is present/running.
type PresenceChecker interface {
	IsPresent(port int) bool
}

// ClientInitializer initializes DynamoDB client for use with the emulator.
type ClientInitializer interface {
	InitClient(port int) (dynamodbiface.DynamoDBAPI, error)
}

// Emulator wraps the DynamoDB process for the purposes of programmatic control
// in tests.
type Emulator struct {
	client  dynamodbiface.DynamoDBAPI
	tng     StringGenerator
	et      ExecutorTerminator
	pc      PresenceChecker
	ci      ClientInitializer
	port    int
	libPath string
	jarPath string
}

// Client returns an instance of DynamoDB client configured for the emulator.
func (e *Emulator) Client() dynamodbiface.DynamoDBAPI {
	return e.client
}

// Runner runs the test against a randomly named table, so that each test can
// be run in parallel and in isolation from other tests. TableName in the
// supplied tableDef will be overriden by a random name.
func (e *Emulator) Runner(t testing.TB, tableDef *dynamodb.CreateTableInput, f func(client dynamodbiface.DynamoDBAPI, tableName string)) {
	tableName, err := e.tng.Generate()
	if err != nil {
		t.Fatalf("failed to generate table name: %v", err)
	}

	tableDef.TableName = aws.String(tableName)
	if tableDef.ProvisionedThroughput == nil {
		tableDef.ProvisionedThroughput = &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		}
	}
	_, err = e.client.CreateTable(tableDef)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		if _, err := e.client.DeleteTable(&dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		}); err != nil {
			t.Fatalf("failed to delete table: %v", err)
		}
	})

	f(e.client, tableName)
}

// Close cleans up an instance of DynamoDB local server if it was started.
func (e *Emulator) Close() error {
	if err := e.et.Terminate(); err != nil {
		return err
	}
	return nil
}

func (e *Emulator) start() error {
	if e.pc.IsPresent(e.port) {
		return nil
	}
	if err := e.et.Execute(
		"java",
		fmt.Sprintf("-Djava.library.path=%s", e.libPath),
		"-jar",
		e.jarPath,
		"-port",
		strconv.Itoa(e.port),
		"-sharedDb",
		"-inMemory",
	); err != nil {
		return err
	}
	return nil
}

func (e *Emulator) initClient() error {
	client, err := e.ci.InitClient(e.port)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

// EmulatorOption is a unit of Emulator configuration.
type EmulatorOption func(*Emulator)

// CustomStringGenerator makes it possible to provide an alternative
// implementations of the tableNameGenerator.
func CustomStringGenerator(gen StringGenerator) EmulatorOption {
	return func(e *Emulator) {
		e.tng = gen
	}
}

// CustomExecutorTerminator makes it possible to provide an alternative
// implementaion of the ExecutorTerminator.
func CustomExecutorTerminator(execTerm ExecutorTerminator) EmulatorOption {
	return func(e *Emulator) {
		e.et = execTerm
	}
}

// CustomPresenceChecker makes it possible to provide an alternative
// implementation of the PresenceChecker to the emulator.
func CustomPresenceChecker(presCheck PresenceChecker) EmulatorOption {
	return func(e *Emulator) {
		e.pc = presCheck
	}
}

// CustomClientInitializer makes it possible to provide an alternative
// implementaion of the ClientInitializer to the Emulator.
func CustomClientInitializer(initClient ClientInitializer) EmulatorOption {
	return func(e *Emulator) {
		e.ci = initClient
	}
}

// CustomPort makes it possible to override the default port configuration of
// the emulator.
func CustomPort(port int) EmulatorOption {
	return func(e *Emulator) {
		e.port = port
	}
}

// CustomLibPath makes it possible to override the default library path
// configuration of the emulator.
func CustomLibPath(libPath string) EmulatorOption {
	return func(e *Emulator) {
		e.libPath = libPath
	}
}

// CustomJarPath makes it possible to override the default jar path
// configuration of the emulator.
func CustomJarPath(jarPath string) EmulatorOption {
	return func(e *Emulator) {
		e.jarPath = jarPath
	}
}

func New(options ...EmulatorOption) (*Emulator, error) {
	// default Emulator configuration
	ddb := &Emulator{
		pc:      NewPresenceChecker(),
		et:      NewExecutorTerminator(),
		tng:     NewStringGenerator(),
		ci:      NewClientInitialier(),
		port:    8000,
		libPath: os.Getenv("DDBLOCAL_LIB"),
		jarPath: os.Getenv("DDBLOCAL_JAR"),
	}

	// apply option overrides
	for _, option := range options {
		option(ddb)
	}

	// run an instance of the DynamoDB local server if not running already
	if err := ddb.start(); err != nil {
		return nil, err
	}

	// init DynamoDB client
	if err := ddb.initClient(); err != nil {
		return nil, err
	}

	return ddb, nil
}
