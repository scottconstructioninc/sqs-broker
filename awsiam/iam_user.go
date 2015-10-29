package awsiam

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pivotal-golang/lager"
)

type IAMUser struct {
	region string
	iamsvc *iam.IAM
	logger lager.Logger
}

func NewIAMUser(
	region string,
	iamsvc *iam.IAM,
	logger lager.Logger,
) *IAMUser {
	return &IAMUser{
		region: region,
		iamsvc: iamsvc,
		logger: logger.Session("iam-user"),
	}
}

func (i *IAMUser) Describe(userName string) (UserDetails, error) {
	userDetails := UserDetails{
		UserName: userName,
	}

	getUserInput := &iam.GetUserInput{
		UserName: aws.String(userName),
	}
	i.logger.Debug("describe-user", lager.Data{"input": getUserInput})

	getUserOutput, err := i.iamsvc.GetUser(getUserInput)
	if err != nil {
		i.logger.Error("aws-iam-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return userDetails, errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return userDetails, err
	}
	i.logger.Debug("describe-user", lager.Data{"output": getUserOutput})

	userDetails.ARN = aws.StringValue(getUserOutput.User.Arn)
	userDetails.UserID = aws.StringValue(getUserOutput.User.UserId)

	return userDetails, nil
}

func (i *IAMUser) Create(userName string) error {
	createUserInput := &iam.CreateUserInput{
		UserName: aws.String(userName),
	}
	i.logger.Debug("create-user", lager.Data{"input": createUserInput})

	createUserOutput, err := i.iamsvc.CreateUser(createUserInput)
	if err != nil {
		i.logger.Error("aws-iam-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	i.logger.Debug("create-user", lager.Data{"output": createUserOutput})

	return nil
}

func (i *IAMUser) Delete(userName string) error {
	deleteUserInput := &iam.DeleteUserInput{
		UserName: aws.String(userName),
	}
	i.logger.Debug("delete-user", lager.Data{"input": deleteUserInput})

	deleteUserOutput, err := i.iamsvc.DeleteUser(deleteUserInput)
	if err != nil {
		i.logger.Error("aws-iam-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	i.logger.Debug("delete-user", lager.Data{"output": deleteUserOutput})

	return nil
}

func (i *IAMUser) CreateAccessKey(userName string) (string, string, error) {
	createAccessKeyInput := &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	}
	i.logger.Debug("create-access-key", lager.Data{"input": createAccessKeyInput})

	createAccessKeyOutput, err := i.iamsvc.CreateAccessKey(createAccessKeyInput)
	if err != nil {
		i.logger.Error("aws-iam-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return "", "", errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return "", "", err
	}
	i.logger.Debug("create-access-key", lager.Data{"output": createAccessKeyOutput})

	return aws.StringValue(createAccessKeyOutput.AccessKey.AccessKeyId), aws.StringValue(createAccessKeyOutput.AccessKey.SecretAccessKey), nil
}

func (i *IAMUser) DeleteAccessKey(userName string, accessKeyID string) error {
	deleteAccessKeyInput := &iam.DeleteAccessKeyInput{
		UserName:    aws.String(userName),
		AccessKeyId: aws.String(accessKeyID),
	}
	i.logger.Debug("delete-access-key", lager.Data{"input": deleteAccessKeyInput})

	deleteAccessKeyOutput, err := i.iamsvc.DeleteAccessKey(deleteAccessKeyInput)
	if err != nil {
		i.logger.Error("aws-iam-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	i.logger.Debug("delete-access-key", lager.Data{"output": deleteAccessKeyOutput})

	return nil
}
