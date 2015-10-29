package fakes

import (
	"github.com/cf-platform-eng/sqs-broker/awsiam"
)

type FakeUser struct {
	DescribeCalled      bool
	DescribeUserName    string
	DescribeUserDetails awsiam.UserDetails
	DescribeError       error

	CreateCalled   bool
	CreateUserName string
	CreateError    error

	DeleteCalled   bool
	DeleteUserName string
	DeleteError    error

	CreateAccessKeyCalled          bool
	CreateAccessKeyUserName        string
	CreateAccessKeyAccessKeyID     string
	CreateAccessKeySecretAccessKey string
	CreateAccessKeyError           error

	DeleteAccessKeyCalled      bool
	DeleteAccessKeyUserName    string
	DeleteAccessKeyAccessKeyID string
	DeleteAccessKeyError       error
}

func (f *FakeUser) Describe(userName string) (awsiam.UserDetails, error) {
	f.DescribeCalled = true
	f.DescribeUserName = userName

	return f.DescribeUserDetails, f.DescribeError
}

func (f *FakeUser) Create(userName string) error {
	f.CreateCalled = true
	f.CreateUserName = userName

	return f.CreateError
}

func (f *FakeUser) Delete(userName string) error {
	f.DeleteCalled = true
	f.DeleteUserName = userName

	return f.DeleteError
}

func (f *FakeUser) CreateAccessKey(userName string) (string, string, error) {
	f.CreateAccessKeyCalled = true
	f.CreateAccessKeyUserName = userName

	return f.CreateAccessKeyAccessKeyID, f.CreateAccessKeySecretAccessKey, f.CreateAccessKeyError
}

func (f *FakeUser) DeleteAccessKey(userName string, accessKeyID string) error {
	f.DeleteAccessKeyCalled = true
	f.DeleteAccessKeyUserName = userName
	f.DeleteAccessKeyAccessKeyID = accessKeyID

	return f.DeleteAccessKeyError
}
