package awssqs

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pivotal-golang/lager"
)

type QueuePolicy struct {
	Version   string
	Id        string
	Statement []QueuePolicyStatement
}

type QueuePolicyStatement struct {
	Sid       string
	Effect    string
	Principal map[string]string
	Action    string
	Resource  string
}

type SQSQueue struct {
	sqssvc *sqs.SQS
	logger lager.Logger
}

func NewSQSQueue(
	sqssvc *sqs.SQS,
	logger lager.Logger,
) *SQSQueue {
	return &SQSQueue{
		sqssvc: sqssvc,
		logger: logger.Session("sqs-queue"),
	}
}

func (s *SQSQueue) Describe(queueName string) (QueueDetails, error) {
	queueDetails := QueueDetails{}

	queueURL, err := s.getQueueURL(queueName)
	if err != nil {
		return queueDetails, err
	}

	queueAttributes, err := s.getQueueAttributes(queueURL)
	if err != nil {
		return queueDetails, err
	}

	return s.buildQueueDetails(queueURL, queueAttributes), nil
}

func (s *SQSQueue) Create(queueName string, queueDetails QueueDetails) error {
	createQueueInput := s.buildCreateQueueInput(queueName, queueDetails)
	s.logger.Debug("create-queue", lager.Data{"input": createQueueInput})

	createQueueOutput, err := s.sqssvc.CreateQueue(createQueueInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("create-queue", lager.Data{"output": createQueueOutput})

	return nil
}

func (s *SQSQueue) Modify(queueName string, queueDetails QueueDetails) error {
	queueURL, err := s.getQueueURL(queueName)
	if err != nil {
		return err
	}

	err = s.setQueueAttributes(queueURL, queueDetails)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQSQueue) Delete(queueName string) error {
	queueURL, err := s.getQueueURL(queueName)
	if err != nil {
		return err
	}

	deleteQueueInput := &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	}
	s.logger.Debug("delete-queue", lager.Data{"input": deleteQueueInput})

	deleteQueueOutput, err := s.sqssvc.DeleteQueue(deleteQueueInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return ErrQueueDoesNotExist
				}
			}
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("delete-queue", lager.Data{"output": deleteQueueOutput})

	return nil
}

func (s *SQSQueue) AddPermission(queueName string, label string, userARN string, action string) error {
	queueDetails, err := s.Describe(queueName)
	if err != nil {
		return err
	}

	queuePolicyStatement := s.buildQueuePolicyStatement(label, userARN, action, queueDetails.QueueArn)
	if queueDetails.Policy != "" {
		policy := QueuePolicy{}
		if err = json.Unmarshal([]byte(queueDetails.Policy), &policy); err != nil {
			return err
		}

		policy.Statement = append(policy.Statement, queuePolicyStatement)
		queueDetails.Policy, err = s.buildQueuePolicy(policy.Statement)
		if err != nil {
			return err
		}
	} else {
		queueDetails.Policy, err = s.buildQueuePolicy([]QueuePolicyStatement{queuePolicyStatement})
		if err != nil {
			return err
		}
	}

	err = s.setQueueAttributes(queueDetails.QueueURL, queueDetails)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQSQueue) RemovePermission(queueName string, label string) error {
	queueURL, err := s.getQueueURL(queueName)
	if err != nil {
		return err
	}

	removePermissionInput := &sqs.RemovePermissionInput{
		QueueUrl: aws.String(queueURL),
		Label:    aws.String(label),
	}
	s.logger.Debug("remove-permission", lager.Data{"input": removePermissionInput})

	removePermissionOutput, err := s.sqssvc.RemovePermission(removePermissionInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return ErrQueueDoesNotExist
				}
			}
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("remove-permission", lager.Data{"output": removePermissionOutput})

	return nil
}

func (s *SQSQueue) getQueueURL(queueName string) (string, error) {
	getQueueURLInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}
	s.logger.Debug("get-queue-url", lager.Data{"input": getQueueURLInput})

	getQueueURLOutput, err := s.sqssvc.GetQueueUrl(getQueueURLInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return "", ErrQueueDoesNotExist
				}
			}
			return "", errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return "", err
	}
	s.logger.Debug("get-queue-url", lager.Data{"output": getQueueURLOutput})

	return aws.StringValue(getQueueURLOutput.QueueUrl), nil
}

func (s *SQSQueue) getQueueAttributes(queueURL string) (map[string]string, error) {
	getQueueAttributesInput := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: aws.StringSlice([]string{"All"}),
	}
	s.logger.Debug("get-queue-attributes", lager.Data{"input": getQueueAttributesInput})

	getQueueAttributesOutput, err := s.sqssvc.GetQueueAttributes(getQueueAttributesInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return nil, ErrQueueDoesNotExist
				}
			}
			return nil, errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return nil, err
	}
	s.logger.Debug("get-queue-attributes", lager.Data{"output": getQueueAttributesOutput})

	return aws.StringValueMap(getQueueAttributesOutput.Attributes), nil
}

func (s *SQSQueue) setQueueAttributes(queueURL string, queueDetails QueueDetails) error {
	setQueueAttributesInput := s.buildSetQueueAttributesInput(queueURL, queueDetails)
	s.logger.Debug("set-queue-attributes", lager.Data{"input": setQueueAttributesInput})

	setQueueAttributesOutput, err := s.sqssvc.SetQueueAttributes(setQueueAttributesInput)
	if err != nil {
		s.logger.Error("aws-sqs-error", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return ErrQueueDoesNotExist
				}
			}
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}
	s.logger.Debug("set-queue-attributes", lager.Data{"output": setQueueAttributesOutput})

	return nil
}

func (s *SQSQueue) buildQueueDetails(queueURL string, attributes map[string]string) QueueDetails {
	queueDetails := QueueDetails{
		QueueURL:               queueURL,
		QueueArn:               attributes["QueueArn"],
		DelaySeconds:           attributes["DelaySeconds"],
		MaximumMessageSize:     attributes["MaximumMessageSize"],
		MessageRetentionPeriod: attributes["MessageRetentionPeriod"],
		Policy:                 attributes["Policy"],
		ReceiveMessageWaitTimeSeconds: attributes["ReceiveMessageWaitTimeSeconds"],
		VisibilityTimeout:             attributes["VisibilityTimeout"],
	}

	return queueDetails
}

func (s *SQSQueue) buildCreateQueueInput(queueName string, queueDetails QueueDetails) *sqs.CreateQueueInput {
	createQueueInput := &sqs.CreateQueueInput{
		QueueName:  aws.String(queueName),
		Attributes: map[string]*string{},
	}

	if queueDetails.DelaySeconds != "" {
		createQueueInput.Attributes["DelaySeconds"] = aws.String(queueDetails.DelaySeconds)
	}

	if queueDetails.MaximumMessageSize != "" {
		createQueueInput.Attributes["MaximumMessageSize"] = aws.String(queueDetails.MaximumMessageSize)
	}

	if queueDetails.MessageRetentionPeriod != "" {
		createQueueInput.Attributes["MessageRetentionPeriod"] = aws.String(queueDetails.MessageRetentionPeriod)
	}

	if queueDetails.Policy != "" {
		createQueueInput.Attributes["Policy"] = aws.String(queueDetails.Policy)
	}

	if queueDetails.ReceiveMessageWaitTimeSeconds != "" {
		createQueueInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String(queueDetails.ReceiveMessageWaitTimeSeconds)
	}

	if queueDetails.VisibilityTimeout != "" {
		createQueueInput.Attributes["VisibilityTimeout"] = aws.String(queueDetails.VisibilityTimeout)
	}

	return createQueueInput
}

func (s *SQSQueue) buildSetQueueAttributesInput(queueURL string, queueDetails QueueDetails) *sqs.SetQueueAttributesInput {
	setQueueAttributesInput := &sqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(queueURL),
		Attributes: map[string]*string{},
	}

	if queueDetails.DelaySeconds != "" {
		setQueueAttributesInput.Attributes["DelaySeconds"] = aws.String(queueDetails.DelaySeconds)
	}

	if queueDetails.MaximumMessageSize != "" {
		setQueueAttributesInput.Attributes["MaximumMessageSize"] = aws.String(queueDetails.MaximumMessageSize)
	}

	if queueDetails.MessageRetentionPeriod != "" {
		setQueueAttributesInput.Attributes["MessageRetentionPeriod"] = aws.String(queueDetails.MessageRetentionPeriod)
	}

	if queueDetails.Policy != "" {
		setQueueAttributesInput.Attributes["Policy"] = aws.String(queueDetails.Policy)
	}

	if queueDetails.ReceiveMessageWaitTimeSeconds != "" {
		setQueueAttributesInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String(queueDetails.ReceiveMessageWaitTimeSeconds)
	}

	if queueDetails.VisibilityTimeout != "" {
		setQueueAttributesInput.Attributes["VisibilityTimeout"] = aws.String(queueDetails.VisibilityTimeout)
	}

	return setQueueAttributesInput
}

func (s *SQSQueue) buildQueuePolicy(queuePolicyStatements []QueuePolicyStatement) (string, error) {
	queuePolicy := QueuePolicy{
		Version:   "2012-10-17",
		Id:        "SQSQueuePolicy",
		Statement: queuePolicyStatements,
	}

	policy, err := json.Marshal(queuePolicy)
	if err != nil {
		return "", err
	}

	return string(policy), nil
}

func (s *SQSQueue) buildQueuePolicyStatement(policyID string, userARN string, action string, queueARN string) QueuePolicyStatement {
	queuePolicyStatement := QueuePolicyStatement{
		Sid:    policyID,
		Effect: "Allow",
		Principal: map[string]string{
			"AWS": userARN,
		},
		Action:   "sqs:" + action,
		Resource: queueARN,
	}

	return queuePolicyStatement
}
