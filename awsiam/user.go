package awsiam

import (
	"errors"
)

type User interface {
	Describe(userName string) (UserDetails, error)
	Create(userName string) error
	Delete(userName string) error
	ListAccessKeys(userName string) ([]string, error)
	CreateAccessKey(userName string) (string, string, error)
	DeleteAccessKey(userName string, accessKeyID string) error
}

type UserDetails struct {
	UserName string
	ARN      string
	UserID   string
}

var (
	ErrUserDoesNotExist = errors.New("iam user does not exist")
)
