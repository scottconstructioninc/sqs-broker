package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/sqs-broker/awsiam"
	"github.com/cf-platform-eng/sqs-broker/awssqs"
	"github.com/cf-platform-eng/sqs-broker/sqsbroker"

	"github.com/cf-platform-eng/sqs-broker/awsrds"
	"github.com/cf-platform-eng/sqs-broker/rdsbroker"
	"github.com/cf-platform-eng/sqs-broker/sqlengine"
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
	flag.StringVar(&configFilePath, "config", "config.json", "Location of the config file")
	flag.StringVar(&port, "port", "3000", "Listen port")
}

func buildLogger(logLevel string) lager.Logger {
	laggerLogLevel, ok := logLevels[strings.ToUpper(logLevel)]
	if !ok {
		log.Fatal("Invalid log level: ", logLevel)
	}

	logger := lager.NewLogger("aws-broker")
	// TODO: Proivde log destination as a config/env option
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, laggerLogLevel))

	return logger
}

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	switch html.EscapeString(r.URL.Path) {
	case "/v2/catalog":
		fmt.Println("Render full catalog")
	default:
		fmt.Println("Parse and route to correct broker")
	}
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}

	logger := buildLogger(config.LogLevel)

	awsConfig := aws.NewConfig().WithRegion(config.Region)
	awsSession := session.New(awsConfig)

	iamsvc := iam.New(awsSession)
	rdssvc := rds.New(awsSession)
	sqssvc := sqs.New(awsSession)

	queue := awssqs.NewSQSQueue(sqssvc, logger)
	user := awsiam.NewIAMUser(iamsvc, logger)

	dbInstance := awsrds.NewRDSDBInstance(config.RDSConfig.Region, iamsvc, rdssvc, logger)
	dbCluster := awsrds.NewRDSDBCluster(config.RDSConfig.Region, iamsvc, rdssvc, logger)
	sqlProvider := sqlengine.NewProviderService(logger)

	SQSServiceBroker := sqsbroker.New(config.SQSConfig, queue, user, logger)
	RDSServiceBroker := rdsbroker.New(config.RDSConfig, dbInstance, dbCluster, sqlProvider, logger)

	credentials := brokerapi.BrokerCredentials{
		Username: config.Username,
		Password: config.Password,
	}

	SQSBrokerAPI := brokerapi.New(SQSServiceBroker, logger, credentials)
	fmt.Printf("%v", SQSBrokerAPI)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if html.EscapeString(r.URL.Path) == "/v2/catalog" {

			// TODO: Make this not suck

			sqs := SQSServiceBroker.ServicesAsString()
			sqs = sqs[:len(sqs)-1]

			rds := RDSServiceBroker.ServicesAsString()
			rds = rds[1:len(rds)]

			services := fmt.Sprintf("{ \"services\": %v, %v }", sqs, rds)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(services))

		} else {
			fmt.Println("Parse and route to correct broker")
		}
	})

	fmt.Printf("AWS Service Broker started on %q using config %q...\n", port, configFilePath)
	http.ListenAndServe(":"+port, nil)
}
