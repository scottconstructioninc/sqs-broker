package awssqs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAWSSQS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS SQS Suite")
}
