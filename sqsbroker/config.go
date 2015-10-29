package sqsbroker

import (
	"errors"
	"fmt"
)

type Config struct {
	Region                       string  `json:"region"`
	QueuePrefix                  string  `json:"queue_prefix"`
	AllowUserProvisionParameters bool    `json:"allow_user_provision_parameters"`
	AllowUserUpdateParameters    bool    `json:"allow_user_update_parameters"`
	Catalog                      Catalog `json:"catalog"`
}

func (c Config) Validate() error {
	if c.Region == "" {
		return errors.New("Must provide a non-empty Region")
	}

	if c.QueuePrefix == "" {
		return errors.New("Must provide a non-empty QueuePrefix")
	}

	if err := c.Catalog.Validate(); err != nil {
		return fmt.Errorf("Validating Catalog configuration: %s", err)
	}

	return nil
}
