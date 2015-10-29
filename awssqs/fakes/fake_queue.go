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
	CreateError        error

	ModifyCalled       bool
	ModifyQueueName    string
	ModifyQueueDetails awssqs.QueueDetails
	ModifyError        error

	DeleteCalled    bool
	DeleteQueueName string
	DeleteError     error

	AddPermissionCalled     bool
	AddPermissionQueueName  string
	AddPermissionQueueLabel string
	AddPermissionAccountIds []string
	AddPermissionActions    []string
	AddPermissionError      error

	RemovePermissionCalled     bool
	RemovePermissionQueueName  string
	RemovePermissionQueueLabel string
	RemovePermissionError      error
}

func (f *FakeQueue) Describe(queueName string) (awssqs.QueueDetails, error) {
	f.DescribeCalled = true
	f.DescribeQueueName = queueName

	return f.DescribeQueueDetails, f.DescribeError
}

func (f *FakeQueue) Create(queueName string, queueDetails awssqs.QueueDetails) error {
	f.CreateCalled = true
	f.CreateQueueName = queueName
	f.CreateQueueDetails = queueDetails

	return f.CreateError
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

func (f *FakeQueue) AddPermission(queueName string, label string, accountIds []string, actions []string) error {
	f.AddPermissionCalled = true
	f.AddPermissionQueueName = queueName
	f.AddPermissionQueueLabel = label
	f.AddPermissionAccountIds = accountIds
	f.AddPermissionActions = actions

	return f.AddPermissionError
}

func (f *FakeQueue) RemovePermission(queueName string, label string) error {
	f.RemovePermissionCalled = true
	f.RemovePermissionQueueName = queueName
	f.RemovePermissionQueueLabel = label

	return f.RemovePermissionError
}
