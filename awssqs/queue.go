package awssqs

import (
	"errors"
)

type Queue interface {
	Describe(queueName string) (QueueDetails, error)
	Create(queueName string, queueDetails QueueDetails) (string, error)
	Modify(queueName string, queueDetails QueueDetails) error
	Delete(queueName string) error
}

type QueueDetails struct {
	QueueURL                      string
	QueueArn                      string
	DelaySeconds                  string
	MaximumMessageSize            string
	MessageRetentionPeriod        string
	Policy                        string
	ReceiveMessageWaitTimeSeconds string
	VisibilityTimeout             string
}

var (
	ErrQueueDoesNotExist = errors.New("sqs queue does not exist")
)
