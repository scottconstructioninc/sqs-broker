package awssqs

import (
	"errors"
)

type Queue interface {
	Describe(queueName string) (QueueDetails, error)
	Create(queueName string, queueDetails QueueDetails) error
	Modify(queueName string, queueDetails QueueDetails) error
	Delete(queueName string) error
	AddPermission(queueName string, label string, userARN string, action string) error
	RemovePermission(queueName string, label string) error
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
