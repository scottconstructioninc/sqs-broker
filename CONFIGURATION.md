# Configuration

A sample configuration can be found at [config-sample.json](https://github.com/cf-platform-eng/sqs-broker/blob/master/config-sample.json).

## General Configuration

| Option     | Required | Type   | Description
|:-----------|:--------:|:------ |:-----------
| log_level  | Y        | String | Broker Log Level (DEBUG, INFO, ERROR, FATAL)
| username   | Y        | String | Broker Auth Username
| password   | Y        | String | Broker Auth Password
| sqs_config | Y        | Hash   | [SQS Broker configuration](https://github.com/cf-platform-eng/sqs-broker/blob/master/CONFIGURATION.md#sqs-broker-configuration)

## SQS Broker Configuration

| Option                         | Required | Type    | Description
|:-------------------------------|:--------:|:------- |:-----------
| region                         | Y        | String  | SQS Region
| sqs_prefix                     | Y        | String  | Prefix to add to SQS Queue Names
| allow_user_provision_parameters| N        | Boolean | Allow users to send arbitrary parameters on provision calls (defaults to `false`)
| allow_user_update_parameters   | N        | Boolean | Allow users to send arbitrary parameters on update calls (defaults to `false`)
| catalog                        | Y        | Hash    | [SQS Broker catalog](https://github.com/cf-platform-eng/sqs-broker/blob/master/CONFIGURATION.md#sqs-broker-catalog)

## SQS Broker catalog

Please refer to the [Catalog Documentation](https://docs.cloudfoundry.org/services/api.html#catalog-mgmt) for more details about these properties.

### Catalog

| Option   | Required | Type      | Description
|:---------|:--------:|:--------- |:-----------
| services | N        | []Service | A list of [Services](https://github.com/cf-platform-eng/sqs-broker/blob/master/CONFIGURATION.md#service)

### Service

| Option                        | Required | Type          | Description
|:------------------------------|:--------:|:------------- |:-----------
| id                            | Y        | String        | An identifier used to correlate this service in future requests to the catalog
| name                          | Y        | String        | The CLI-friendly name of the service that will appear in the catalog. All lowercase, no spaces
| description                   | Y        | String        | A short description of the service that will appear in the catalog
| bindable                      | N        | Boolean       | Whether the service can be bound to applications
| tags                          | N        | []String      | A list of service tags
| metadata.displayName          | N        | String        | The name of the service to be displayed in graphical clients
| metadata.imageUrl             | N        | String        | The URL to an image
| metadata.longDescription      | N        | String        | Long description
| metadata.providerDisplayName  | N        | String        | The name of the upstream entity providing the actual service
| metadata.documentationUrl     | N        | String        | Link to documentation page for service
| metadata.supportUrl           | N        | String        | Link to support for the service
| requires                      | N        | []String      | A list of permissions that the user would have to give the service, if they provision it (only `syslog_drain` is supported)
| plan_updateable               | N        | Boolean       | Whether the service supports upgrade/downgrade for some plans
| plans                         | N        | []ServicePlan | A list of [Plans](https://github.com/cf-platform-eng/sqs-broker/blob/master/CONFIGURATION.md#service-plan) for this service
| dashboard_client.id           | N        | String        | The id of the Oauth2 client that the service intends to use
| dashboard_client.secret       | N        | String        | A secret for the dashboard client
| dashboard_client.redirect_uri | N        | String        | A domain for the service dashboard that will be whitelisted by the UAA to enable SSO

### Service Plan

| Option               | Required | Type          | Description
|:---------------------|:--------:|:------------- |:-----------
| id                   | Y        | String        | An identifier used to correlate this plan in future requests to the catalog
| name                 | Y        | String        | The CLI-friendly name of the plan that will appear in the catalog. All lowercase, no spaces
| description          | Y        | String        | A short description of the plan that will appear in the catalog
| metadata.bullets     | N        | []String      | Features of this plan, to be displayed in a bulleted-list
| metadata.costs       | N        | Cost Object   | An array-of-objects that describes the costs of a service, in what currency, and the unit of measure
| metadata.displayName | N        | String        | Name of the plan to be display in graphical clients
| free                 | N        | Boolean       | This field allows the plan to be limited by the non_basic_services_allowed field in a Cloud Foundry Quota
| sqs_properties       | Y        | SQSProperties | [SQS Properties](https://github.com/cf-platform-eng/sqs-broker/blob/master/CONFIGURATION.md#sqs-properties)

## SQS Properties

Please refer to the [Amazon Simple Queue Service Documentation](https://aws.amazon.com/documentation/sqs/) for more details about these properties.

| Option                            | Required | Type   | Description
|:----------------------------------|:--------:|:------ |:-----------
| delay_seconds                     | N        | String | The time in seconds that the delivery of all messages in the queue will be delayed
| maximum_message_size              | N        | String | The limit of how many bytes a message can contain before Amazon SQS rejects it
| message_retention_period          | N        | String | The number of seconds Amazon SQS retains a message
| policy                            | N        | String | TThe queue's policy
| receive_message_wait_time_seconds | N        | String | The time for which a ReceiveMessage call will wait for a message to arrive
| visibility_timeout                | N        | String | The visibility timeout for the queue


