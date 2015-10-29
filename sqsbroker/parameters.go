package sqsbroker

type ProvisionParameters struct {
	DelaySeconds                  string `mapstructure:"delay_seconds"`
	MaximumMessageSize            string `mapstructure:"maximum_message_size"`
	MessageRetentionPeriod        string `mapstructure:"message_retention_period"`
	ReceiveMessageWaitTimeSeconds string `mapstructure:"receive_message_wait_time_seconds"`
	VisibilityTimeout             string `mapstructure:"visibility_timeout"`
}

type UpdateParameters struct {
	DelaySeconds                  string `mapstructure:"delay_seconds"`
	MaximumMessageSize            string `mapstructure:"maximum_message_size"`
	MessageRetentionPeriod        string `mapstructure:"message_retention_period"`
	ReceiveMessageWaitTimeSeconds string `mapstructure:"receive_message_wait_time_seconds"`
	VisibilityTimeout             string `mapstructure:"visibility_timeout"`
}
