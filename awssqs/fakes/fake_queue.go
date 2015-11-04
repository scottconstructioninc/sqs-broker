package fakes

import (
	"github.com/cf-platform-eng/sqs-broker/awssqs"
)

type FakeQueue struct {
	DescribeCalled       bool
	DescribeQueueName    string
	DescribeQueueDetails awssqs.QueueDetails
	DescribeError        error

	CreateCalled       bool
	CreateQueueName    string
	CreateQueueDetails awssqs.QueueDetails
	CreateQueueURL     string
	CreateError        error

	ModifyCalled       bool
	ModifyQueueName    string
	ModifyQueueDetails awssqs.QueueDetails
	ModifyError        error

	DeleteCalled    bool
	DeleteQueueName string
	DeleteError     error
}

func (f *FakeQueue) Describe(queueName string) (awssqs.QueueDetails, error) {
	f.DescribeCalled = true
	f.DescribeQueueName = queueName

	return f.DescribeQueueDetails, f.DescribeError
}

func (f *FakeQueue) Create(queueName string, queueDetails awssqs.QueueDetails) (string, error) {
	f.CreateCalled = true
	f.CreateQueueName = queueName
	f.CreateQueueDetails = queueDetails

	return f.CreateQueueURL, f.CreateError
}

func (f *FakeQueue) Modify(queueName string, queueDetails awssqs.QueueDetails) error {
	f.ModifyCalled = true
	f.ModifyQueueName = queueName
	f.ModifyQueueDetails = queueDetails

	return f.ModifyError
}

func (f *FakeQueue) Delete(queueName string) error {
	f.DeleteCalled = true
	f.DeleteQueueName = queueName

	return f.DeleteError
}
