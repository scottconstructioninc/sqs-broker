package sqsbroker

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/frodenas/brokerapi"
	"github.com/mitchellh/mapstructure"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/sqs-broker/awsiam"
	"github.com/cf-platform-eng/sqs-broker/awssqs"
)

const instanceIDLogKey = "instance-id"
const bindingIDLogKey = "binding-id"
const detailsLogKey = "details"
const acceptsIncompleteLogKey = "acceptsIncomplete"

type SQSBroker struct {
	sqsPrefix                    string
	allowUserProvisionParameters bool
	allowUserUpdateParameters    bool
	catalog                      Catalog
	queue                        awssqs.Queue
	user                         awsiam.User
	logger                       lager.Logger
}

func New(
	config Config,
	queue awssqs.Queue,
	user awsiam.User,
	logger lager.Logger,
) *SQSBroker {
	return &SQSBroker{
		sqsPrefix:                    config.SQSPrefix,
		allowUserProvisionParameters: config.AllowUserProvisionParameters,
		allowUserUpdateParameters:    config.AllowUserUpdateParameters,
		catalog:                      config.Catalog,
		queue:                        queue,
		user:                         user,
		logger:                       logger.Session("broker"),
	}
}

func (b *SQSBroker) Services() brokerapi.CatalogResponse {
	catalogResponse := brokerapi.CatalogResponse{}

	brokerCatalog, err := json.Marshal(b.catalog)
	if err != nil {
		b.logger.Error("marshal-error", err)
		return catalogResponse
	}

	apiCatalog := brokerapi.Catalog{}
	if err = json.Unmarshal(brokerCatalog, &apiCatalog); err != nil {
		b.logger.Error("unmarshal-error", err)
		return catalogResponse
	}

	catalogResponse.Services = apiCatalog.Services

	return catalogResponse
}

func (b *SQSBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, acceptsIncomplete bool) (brokerapi.ProvisioningResponse, bool, error) {
	b.logger.Debug("provision", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	provisioningResponse := brokerapi.ProvisioningResponse{}

	provisionParameters := ProvisionParameters{}
	if b.allowUserProvisionParameters {
		if err := mapstructure.Decode(details.Parameters, &provisionParameters); err != nil {
			return provisioningResponse, false, err
		}
	}

	servicePlan, ok := b.catalog.FindServicePlan(details.PlanID)
	if !ok {
		return provisioningResponse, false, fmt.Errorf("Service Plan '%s' not found", details.PlanID)
	}

	createQueueDetails := b.createQueueDetails(instanceID, servicePlan, provisionParameters, details)
	if err := b.queue.Create(b.queueName(instanceID), *createQueueDetails); err != nil {
		return provisioningResponse, false, err
	}

	return provisioningResponse, false, nil
}

func (b *SQSBroker) Update(instanceID string, details brokerapi.UpdateDetails, acceptsIncomplete bool) (bool, error) {
	b.logger.Debug("update", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	updateParameters := UpdateParameters{}
	if b.allowUserUpdateParameters {
		if err := mapstructure.Decode(details.Parameters, &updateParameters); err != nil {
			return false, err
		}
	}

	service, ok := b.catalog.FindService(details.ServiceID)
	if !ok {
		return false, fmt.Errorf("Service '%s' not found", details.ServiceID)
	}

	if !service.PlanUpdateable {
		return false, brokerapi.ErrInstanceNotUpdateable
	}

	servicePlan, ok := b.catalog.FindServicePlan(details.PlanID)
	if !ok {
		return false, fmt.Errorf("Service Plan '%s' not found", details.PlanID)
	}

	modifyQueueDetails := b.modifyQueueDetails(instanceID, servicePlan, updateParameters, details)
	if err := b.queue.Modify(b.queueName(instanceID), *modifyQueueDetails); err != nil {
		if err == awssqs.ErrQueueDoesNotExist {
			return false, brokerapi.ErrInstanceDoesNotExist
		}
		return false, err
	}

	return false, nil
}

func (b *SQSBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, acceptsIncomplete bool) (bool, error) {
	b.logger.Debug("deprovision", lager.Data{
		instanceIDLogKey:        instanceID,
		detailsLogKey:           details,
		acceptsIncompleteLogKey: acceptsIncomplete,
	})

	if err := b.queue.Delete(b.queueName(instanceID)); err != nil {
		if err == awssqs.ErrQueueDoesNotExist {
			return false, brokerapi.ErrInstanceDoesNotExist
		}
		return false, err
	}

	return false, nil
}

func (b *SQSBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.BindingResponse, error) {
	var err error
	var accessKeyID, secretAccessKey string

	b.logger.Debug("bind", lager.Data{
		instanceIDLogKey: instanceID,
		bindingIDLogKey:  bindingID,
		detailsLogKey:    details,
	})

	bindingResponse := brokerapi.BindingResponse{}

	service, ok := b.catalog.FindService(details.ServiceID)
	if !ok {
		return bindingResponse, fmt.Errorf("Service '%s' not found", details.ServiceID)
	}

	if !service.Bindable {
		return bindingResponse, brokerapi.ErrInstanceNotBindable
	}

	queueDetails, err := b.queue.Describe(b.queueName(instanceID))
	if err != nil {
		if err == awssqs.ErrQueueDoesNotExist {
			return bindingResponse, brokerapi.ErrInstanceDoesNotExist
		}
		return bindingResponse, err
	}

	if err = b.user.Create(b.userName(bindingID)); err != nil {
		return bindingResponse, err
	}
	defer func() {
		if err != nil {
			if accessKeyID != "" {
				b.user.DeleteAccessKey(b.userName(bindingID), accessKeyID)
			}
			b.user.Delete(b.userName(bindingID))
		}
	}()

	accessKeyID, secretAccessKey, err = b.user.CreateAccessKey(b.userName(bindingID))
	if err != nil {
		return bindingResponse, err
	}

	var userDetails awsiam.UserDetails
	userDetails, err = b.user.Describe(b.userName(bindingID))
	if err != nil {
		return bindingResponse, err
	}

	if err = b.queue.AddPermission(b.queueName(instanceID), b.queueLabel(bindingID), userDetails.ARN, "*"); err != nil {
		if err == awssqs.ErrQueueDoesNotExist {
			return bindingResponse, brokerapi.ErrInstanceDoesNotExist
		}
		return bindingResponse, err
	}

	bindingResponse.Credentials = &brokerapi.CredentialsHash{
		Username: accessKeyID,
		Password: secretAccessKey,
		URI:      queueDetails.QueueURL,
	}

	return bindingResponse, nil
}

func (b *SQSBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	b.logger.Debug("unbind", lager.Data{
		instanceIDLogKey: instanceID,
		bindingIDLogKey:  bindingID,
		detailsLogKey:    details,
	})

	if err := b.queue.RemovePermission(b.queueName(instanceID), b.queueLabel(bindingID)); err != nil {
		if err == awssqs.ErrQueueDoesNotExist {
			return brokerapi.ErrInstanceDoesNotExist
		}
		return err
	}

	accessKeys, err := b.user.ListAccessKeys(b.userName(bindingID))
	if err != nil {
		return err
	}

	for _, accessKey := range accessKeys {
		if err := b.user.DeleteAccessKey(b.userName(bindingID), accessKey); err != nil {
			return err
		}
	}

	if err := b.user.Delete(b.userName(bindingID)); err != nil {
		return err
	}

	return nil
}

func (b *SQSBroker) LastOperation(instanceID string) (brokerapi.LastOperationResponse, error) {
	b.logger.Debug("last-operation", lager.Data{
		instanceIDLogKey: instanceID,
	})

	return brokerapi.LastOperationResponse{}, errors.New("This broker does not support LastOperation")
}

func (b *SQSBroker) queueName(instanceID string) string {
	return fmt.Sprintf("%s-%s", b.sqsPrefix, instanceID)
}

func (b *SQSBroker) queueLabel(bindingID string) string {
	return fmt.Sprintf("%s-%s", b.sqsPrefix, bindingID)
}

func (b *SQSBroker) userName(bindingID string) string {
	return fmt.Sprintf("%s-%s", b.sqsPrefix, bindingID)
}

func (b *SQSBroker) createQueueDetails(instanceID string, servicePlan ServicePlan, provisionParameters ProvisionParameters, details brokerapi.ProvisionDetails) *awssqs.QueueDetails {
	queueDetails := b.queueDetailsFromPlan(servicePlan)

	if provisionParameters.DelaySeconds != "" {
		queueDetails.DelaySeconds = provisionParameters.DelaySeconds
	}

	if provisionParameters.MaximumMessageSize != "" {
		queueDetails.MaximumMessageSize = provisionParameters.MaximumMessageSize
	}

	if provisionParameters.MessageRetentionPeriod != "" {
		queueDetails.MessageRetentionPeriod = provisionParameters.MessageRetentionPeriod
	}

	if provisionParameters.ReceiveMessageWaitTimeSeconds != "" {
		queueDetails.ReceiveMessageWaitTimeSeconds = provisionParameters.ReceiveMessageWaitTimeSeconds
	}

	if provisionParameters.VisibilityTimeout != "" {
		queueDetails.VisibilityTimeout = provisionParameters.VisibilityTimeout
	}

	return queueDetails
}

func (b *SQSBroker) modifyQueueDetails(instanceID string, servicePlan ServicePlan, updateParameters UpdateParameters, details brokerapi.UpdateDetails) *awssqs.QueueDetails {
	queueDetails := b.queueDetailsFromPlan(servicePlan)

	if updateParameters.DelaySeconds != "" {
		queueDetails.DelaySeconds = updateParameters.DelaySeconds
	}

	if updateParameters.MaximumMessageSize != "" {
		queueDetails.MaximumMessageSize = updateParameters.MaximumMessageSize
	}

	if updateParameters.MessageRetentionPeriod != "" {
		queueDetails.MessageRetentionPeriod = updateParameters.MessageRetentionPeriod
	}

	if updateParameters.ReceiveMessageWaitTimeSeconds != "" {
		queueDetails.ReceiveMessageWaitTimeSeconds = updateParameters.ReceiveMessageWaitTimeSeconds
	}

	if updateParameters.VisibilityTimeout != "" {
		queueDetails.VisibilityTimeout = updateParameters.VisibilityTimeout
	}

	return queueDetails
}

func (b *SQSBroker) queueDetailsFromPlan(servicePlan ServicePlan) *awssqs.QueueDetails {
	queueDetails := &awssqs.QueueDetails{
		DelaySeconds:                  servicePlan.SQSProperties.DelaySeconds,
		MaximumMessageSize:            servicePlan.SQSProperties.MaximumMessageSize,
		MessageRetentionPeriod:        servicePlan.SQSProperties.MessageRetentionPeriod,
		ReceiveMessageWaitTimeSeconds: servicePlan.SQSProperties.ReceiveMessageWaitTimeSeconds,
		VisibilityTimeout:             servicePlan.SQSProperties.VisibilityTimeout,
	}

	return queueDetails
}
