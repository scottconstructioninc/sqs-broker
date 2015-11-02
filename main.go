package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/sqs-broker/awsiam"
	"github.com/cf-platform-eng/sqs-broker/awssqs"
	"github.com/cf-platform-eng/sqs-broker/sqsbroker"
)

var (
	configFilePath string
	port           string

	logLevels = map[string]lager.LogLevel{
		"DEBUG": lager.DEBUG,
		"INFO":  lager.INFO,
		"ERROR": lager.ERROR,
		"FATAL": lager.FATAL,
	}
)

func init() {
	flag.StringVar(&configFilePath, "config", "", "Location of the config file")
	flag.StringVar(&port, "port", "3000", "Listen port")
}

func buildLogger(logLevel string) lager.Logger {
	laggerLogLevel, ok := logLevels[strings.ToUpper(logLevel)]
	if !ok {
		log.Fatal("Invalid log level: ", logLevel)
	}

	logger := lager.NewLogger("sqs-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, laggerLogLevel))

	return logger
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}

	logger := buildLogger(config.LogLevel)

	awsConfig := aws.NewConfig().WithRegion(config.SQSConfig.Region)
	awsSession := session.New(awsConfig)

	sqssvc := sqs.New(awsSession)
	queue := awssqs.NewSQSQueue(sqssvc, logger)

	iamsvc := iam.New(awsSession)
	user := awsiam.NewIAMUser(iamsvc, logger)

	serviceBroker := sqsbroker.New(config.SQSConfig, queue, user, logger)

	credentials := brokerapi.BrokerCredentials{
		Username: config.Username,
		Password: config.Password,
	}

	brokerAPI := brokerapi.New(serviceBroker, logger, credentials)
	http.Handle("/", brokerAPI)

	fmt.Println("SQS Service Broker started on port " + port + "...")
	http.ListenAndServe(":"+port, nil)
}
