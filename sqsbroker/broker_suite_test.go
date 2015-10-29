package sqsbroker_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSQSBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SQS Broker Suite")
}
