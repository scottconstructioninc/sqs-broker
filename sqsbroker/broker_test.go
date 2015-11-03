package sqsbroker_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/sqs-broker/sqsbroker"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/cf-platform-eng/sqs-broker/awsiam"
	iamfake "github.com/cf-platform-eng/sqs-broker/awsiam/fakes"
	"github.com/cf-platform-eng/sqs-broker/awssqs"
	sqsfake "github.com/cf-platform-eng/sqs-broker/awssqs/fakes"
)

var _ = Describe("SQS Broker", func() {
	var (
		sqsProperties1 SQSProperties
		sqsProperties2 SQSProperties
		plan1          ServicePlan
		plan2          ServicePlan
		service1       Service
		service2       Service
		catalog        Catalog

		config Config

		queue *sqsfake.FakeQueue
		user  *iamfake.FakeUser

		testSink *lagertest.TestSink
		logger   lager.Logger

		sqsBroker *SQSBroker

		allowUserProvisionParameters bool
		allowUserUpdateParameters    bool
		serviceBindable              bool
		planUpdateable               bool

		instanceID = "instance-id"
		bindingID  = "binding-id"
		queueName  = "cf-instance-id"
		queueLabel = "cf-binding-id"
		userName   = "cf-binding-id"
	)

	BeforeEach(func() {
		allowUserProvisionParameters = true
		allowUserUpdateParameters = true
		serviceBindable = true
		planUpdateable = true

		queue = &sqsfake.FakeQueue{}
		user = &iamfake.FakeUser{}

		sqsProperties1 = SQSProperties{}
		sqsProperties2 = SQSProperties{}
	})

	JustBeforeEach(func() {
		plan1 = ServicePlan{
			ID:            "Plan-1",
			Name:          "Plan 1",
			Description:   "This is the Plan 1",
			SQSProperties: sqsProperties1,
		}
		plan2 = ServicePlan{
			ID:            "Plan-2",
			Name:          "Plan 2",
			Description:   "This is the Plan 2",
			SQSProperties: sqsProperties2,
		}

		service1 = Service{
			ID:             "Service-1",
			Name:           "Service 1",
			Description:    "This is the Service 1",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []ServicePlan{plan1},
		}
		service2 = Service{
			ID:             "Service-2",
			Name:           "Service 2",
			Description:    "This is the Service 2",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []ServicePlan{plan2},
		}

		catalog = Catalog{
			Services: []Service{service1, service2},
		}

		config = Config{
			Region:                       "sqs-region",
			SQSPrefix:                    "cf",
			AllowUserProvisionParameters: allowUserProvisionParameters,
			AllowUserUpdateParameters:    allowUserUpdateParameters,
			Catalog:                      catalog,
		}

		logger = lager.NewLogger("sqsbroker_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		sqsBroker = New(config, queue, user, logger)
	})

	var _ = Describe("Services", func() {
		var (
			properCatalogResponse brokerapi.CatalogResponse
		)

		BeforeEach(func() {
			properCatalogResponse = brokerapi.CatalogResponse{
				Services: []brokerapi.Service{
					brokerapi.Service{
						ID:             "Service-1",
						Name:           "Service 1",
						Description:    "This is the Service 1",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-1",
								Name:        "Plan 1",
								Description: "This is the Plan 1",
							},
						},
					},
					brokerapi.Service{
						ID:             "Service-2",
						Name:           "Service 2",
						Description:    "This is the Service 2",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-2",
								Name:        "Plan 2",
								Description: "This is the Plan 2",
							},
						},
					},
				},
			}
		})

		It("returns the proper CatalogResponse", func() {
			brokerCatalog := sqsBroker.Services()
			Expect(brokerCatalog).To(Equal(properCatalogResponse))
		})

	})

	var _ = Describe("Provision", func() {
		var (
			provisionDetails  brokerapi.ProvisionDetails
			acceptsIncomplete bool

			properProvisioningResponse brokerapi.ProvisioningResponse
		)

		BeforeEach(func() {
			provisionDetails = brokerapi.ProvisionDetails{
				OrganizationGUID: "organization-id",
				PlanID:           "Plan-1",
				ServiceID:        "Service-1",
				SpaceGUID:        "space-id",
				Parameters:       map[string]interface{}{},
			}
			acceptsIncomplete = false

			properProvisioningResponse = brokerapi.ProvisioningResponse{}
		})

		It("returns the proper response", func() {
			provisioningResponse, asynch, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
			Expect(provisioningResponse).To(Equal(properProvisioningResponse))
			Expect(asynch).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
			Expect(queue.CreateCalled).To(BeTrue())
			Expect(queue.CreateQueueName).To(Equal(queueName))
			Expect(queue.CreateQueueDetails.DelaySeconds).To(Equal(""))
			Expect(queue.CreateQueueDetails.MaximumMessageSize).To(Equal(""))
			Expect(queue.CreateQueueDetails.MessageRetentionPeriod).To(Equal(""))
			Expect(queue.CreateQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal(""))
			Expect(queue.CreateQueueDetails.VisibilityTimeout).To(Equal(""))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has DelaySeconds", func() {
			BeforeEach(func() {
				sqsProperties1.DelaySeconds = "test-delay-seconds"
			})

			It("makes the proper calls", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(queue.CreateQueueDetails.DelaySeconds).To(Equal("test-delay-seconds"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has DelaySeconds Parameter", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"delay_seconds": "test-delay-seconds-parameter"}
				})

				It("makes the proper calls", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(queue.CreateQueueDetails.DelaySeconds).To(Equal("test-delay-seconds-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("makes the proper calls", func() {
						_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(queue.CreateQueueDetails.DelaySeconds).To(Equal("test-delay-seconds"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has MaximumMessageSize", func() {
			BeforeEach(func() {
				sqsProperties1.MaximumMessageSize = "test-maximum-message-size"
			})

			It("makes the proper calls", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(queue.CreateQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has MaximumMessageSize Parameter", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"maximum_message_size": "test-maximum-message-size-parameter"}
				})

				It("makes the proper calls", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(queue.CreateQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("makes the proper calls", func() {
						_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(queue.CreateQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has MessageRetentionPeriod", func() {
			BeforeEach(func() {
				sqsProperties1.MessageRetentionPeriod = "test-message-retention-period"
			})

			It("makes the proper calls", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(queue.CreateQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has MessageRetentionPeriod Parameter", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"message_retention_period": "test-message-retention-period-parameter"}
				})

				It("makes the proper calls", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(queue.CreateQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("makes the proper calls", func() {
						_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(queue.CreateQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has ReceiveMessageWaitTimeSeconds", func() {
			BeforeEach(func() {
				sqsProperties1.ReceiveMessageWaitTimeSeconds = "test-receive-message-wait-time-seconds"
			})

			It("makes the proper calls", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(queue.CreateQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has ReceiveMessageWaitTimeSeconds Parameter", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"receive_message_wait_time_seconds": "test-receive-message-wait-time-seconds-parameter"}
				})

				It("makes the proper calls", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(queue.CreateQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("makes the proper calls", func() {
						_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(queue.CreateQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has VisibilityTimeout", func() {
			BeforeEach(func() {
				sqsProperties1.VisibilityTimeout = "test-visibility-timeout"
			})

			It("makes the proper calls", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(queue.CreateQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has VisibilityTimeout Parameter", func() {
				BeforeEach(func() {
					provisionDetails.Parameters = map[string]interface{}{"visibility_timeout": "test-visibility-timeout-parameter"}
				})

				It("makes the proper calls", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(queue.CreateQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user provision parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserProvisionParameters = false
					})

					It("makes the proper calls", func() {
						_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
						Expect(queue.CreateQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when Parameters are not valid", func() {
			BeforeEach(func() {
				provisionDetails.Parameters = map[string]interface{}{"delay_seconds": true}
			})

			It("returns the proper error", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'delay_seconds' expected type 'string', got unconvertible type 'bool'"))
			})

			Context("but user provision parameters are not allowed", func() {
				BeforeEach(func() {
					allowUserProvisionParameters = false
				})

				It("does not return an error", func() {
					_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when Service Plan is not found", func() {
			BeforeEach(func() {
				provisionDetails.PlanID = "unknown"
			})

			It("returns the proper error", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service Plan 'unknown' not found"))
			})
		})

		Context("when creating the Queue fails", func() {
			BeforeEach(func() {
				queue.CreateError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, _, err := sqsBroker.Provision(instanceID, provisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})
	})

	var _ = Describe("Update", func() {
		var (
			updateDetails     brokerapi.UpdateDetails
			acceptsIncomplete bool
		)

		BeforeEach(func() {
			updateDetails = brokerapi.UpdateDetails{
				ServiceID:  "Service-2",
				PlanID:     "Plan-2",
				Parameters: map[string]interface{}{},
				PreviousValues: brokerapi.PreviousValues{
					PlanID:         "Plan-1",
					ServiceID:      "Service-1",
					OrganizationID: "organization-id",
					SpaceID:        "space-id",
				},
			}
			acceptsIncomplete = false
		})

		It("returns the proper response", func() {
			asynch, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
			Expect(asynch).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
			Expect(queue.ModifyCalled).To(BeTrue())
			Expect(queue.ModifyQueueName).To(Equal(queueName))
			Expect(queue.ModifyQueueDetails.DelaySeconds).To(Equal(""))
			Expect(queue.ModifyQueueDetails.MaximumMessageSize).To(Equal(""))
			Expect(queue.ModifyQueueDetails.MessageRetentionPeriod).To(Equal(""))
			Expect(queue.ModifyQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal(""))
			Expect(queue.ModifyQueueDetails.VisibilityTimeout).To(Equal(""))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has DelaySeconds", func() {
			BeforeEach(func() {
				sqsProperties2.DelaySeconds = "test-delay-seconds"
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(queue.ModifyQueueDetails.DelaySeconds).To(Equal("test-delay-seconds"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has DelaySeconds Parameter", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"delay_seconds": "test-delay-seconds-parameter"}
				})

				It("makes the proper calls", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(queue.ModifyQueueDetails.DelaySeconds).To(Equal("test-delay-seconds-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user update parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("makes the proper calls", func() {
						_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(queue.ModifyQueueDetails.DelaySeconds).To(Equal("test-delay-seconds"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has MaximumMessageSize", func() {
			BeforeEach(func() {
				sqsProperties2.MaximumMessageSize = "test-maximum-message-size"
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(queue.ModifyQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has MaximumMessageSize Parameter", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"maximum_message_size": "test-maximum-message-size-parameter"}
				})

				It("makes the proper calls", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(queue.ModifyQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user update parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("makes the proper calls", func() {
						_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(queue.ModifyQueueDetails.MaximumMessageSize).To(Equal("test-maximum-message-size"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has MessageRetentionPeriod", func() {
			BeforeEach(func() {
				sqsProperties2.MessageRetentionPeriod = "test-message-retention-period"
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(queue.ModifyQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has MessageRetentionPeriod Parameter", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"message_retention_period": "test-message-retention-period-parameter"}
				})

				It("makes the proper calls", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(queue.ModifyQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user update parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("makes the proper calls", func() {
						_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(queue.ModifyQueueDetails.MessageRetentionPeriod).To(Equal("test-message-retention-period"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has ReceiveMessageWaitTimeSeconds", func() {
			BeforeEach(func() {
				sqsProperties2.ReceiveMessageWaitTimeSeconds = "test-receive-message-wait-time-seconds"
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(queue.ModifyQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has ReceiveMessageWaitTimeSeconds Parameter", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"receive_message_wait_time_seconds": "test-receive-message-wait-time-seconds-parameter"}
				})

				It("makes the proper calls", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(queue.ModifyQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user update parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("makes the proper calls", func() {
						_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(queue.ModifyQueueDetails.ReceiveMessageWaitTimeSeconds).To(Equal("test-receive-message-wait-time-seconds"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when has VisibilityTimeout", func() {
			BeforeEach(func() {
				sqsProperties2.VisibilityTimeout = "test-visibility-timeout"
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(queue.ModifyQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("and has VisibilityTimeout Parameter", func() {
				BeforeEach(func() {
					updateDetails.Parameters = map[string]interface{}{"visibility_timeout": "test-visibility-timeout-parameter"}
				})

				It("makes the proper calls", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(queue.ModifyQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout-parameter"))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("but user update parameters are not allowed", func() {
					BeforeEach(func() {
						allowUserUpdateParameters = false
					})

					It("makes the proper calls", func() {
						_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
						Expect(queue.ModifyQueueDetails.VisibilityTimeout).To(Equal("test-visibility-timeout"))
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("when Parameters are not valid", func() {
			BeforeEach(func() {
				updateDetails.Parameters = map[string]interface{}{"delay_seconds": true}
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'delay_seconds' expected type 'string', got unconvertible type 'bool'"))
			})

			Context("but user update parameters are not allowed", func() {
				BeforeEach(func() {
					allowUserUpdateParameters = false
				})

				It("does not return an error", func() {
					_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when Service is not found", func() {
			BeforeEach(func() {
				updateDetails.ServiceID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service 'unknown' not found"))
			})
		})

		Context("when Plans is not updateable", func() {
			BeforeEach(func() {
				planUpdateable = false
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceNotUpdateable))
			})
		})

		Context("when Service Plan is not found", func() {
			BeforeEach(func() {
				updateDetails.PlanID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service Plan 'unknown' not found"))
			})
		})

		Context("when modifying the Queue fails", func() {
			BeforeEach(func() {
				queue.ModifyError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Update(instanceID, updateDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})
	})

	var _ = Describe("Deprovision", func() {
		var (
			deprovisionDetails brokerapi.DeprovisionDetails
			acceptsIncomplete  bool
		)

		BeforeEach(func() {
			deprovisionDetails = brokerapi.DeprovisionDetails{
				ServiceID: "Service-1",
				PlanID:    "Plan-1",
			}
			acceptsIncomplete = false
		})

		It("returns the proper response", func() {
			asynch, err := sqsBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
			Expect(asynch).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := sqsBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
			Expect(queue.DeleteCalled).To(BeTrue())
			Expect(queue.DeleteQueueName).To(Equal(queueName))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deleting the Queue fails", func() {
			BeforeEach(func() {
				queue.DeleteError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Queue does not exists", func() {
				BeforeEach(func() {
					queue.DeleteError = awssqs.ErrQueueDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := sqsBroker.Deprovision(instanceID, deprovisionDetails, acceptsIncomplete)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Bind", func() {
		var (
			bindDetails brokerapi.BindDetails
		)

		BeforeEach(func() {
			bindDetails = brokerapi.BindDetails{
				ServiceID:  "Service-1",
				PlanID:     "Plan-1",
				AppGUID:    "Application-1",
				Parameters: map[string]interface{}{},
			}

			queue.DescribeQueueDetails = awssqs.QueueDetails{
				QueueURL: "queue-url",
			}

			user.CreateAccessKeyAccessKeyID = "user-access-key-id"
			user.CreateAccessKeySecretAccessKey = "user-secret-access-key"

			user.DescribeUserDetails = awsiam.UserDetails{
				UserName: userName,
				ARN:      "user-arn",
			}
		})

		It("returns the proper response", func() {
			bindingResponse, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
			credentials := bindingResponse.Credentials.(*brokerapi.CredentialsHash)
			Expect(bindingResponse.SyslogDrainURL).To(BeEmpty())
			Expect(credentials.Host).To(Equal(""))
			Expect(credentials.Port).To(Equal(int64(0)))
			Expect(credentials.Name).To(Equal(""))
			Expect(credentials.Username).To(Equal("user-access-key-id"))
			Expect(credentials.Password).To(Equal("user-secret-access-key"))
			Expect(credentials.URI).To(Equal("queue-url"))
			Expect(credentials.JDBCURI).To(Equal(""))
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes the proper calls", func() {
			_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
			Expect(user.CreateAccessKeyCalled).To(BeTrue())
			Expect(user.CreateAccessKeyUserName).To(Equal(userName))
			Expect(user.DescribeCalled).To(BeTrue())
			Expect(user.DescribeUserName).To(Equal(userName))
			Expect(queue.AddPermissionCalled).To(BeTrue())
			Expect(queue.AddPermissionQueueName).To(Equal(queueName))
			Expect(queue.AddPermissionQueueLabel).To(Equal(queueLabel))
			Expect(queue.AddPermissionAccountIds).To(Equal([]string{"user-arn"}))
			Expect(queue.AddPermissionActions).To(Equal([]string{"*"}))
			Expect(user.DeleteCalled).To(BeFalse())
			Expect(user.DeleteAccessKeyCalled).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when Service is not found", func() {
			BeforeEach(func() {
				bindDetails.ServiceID = "unknown"
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Service 'unknown' not found"))
			})
		})

		Context("when Service is not bindable", func() {
			BeforeEach(func() {
				serviceBindable = false
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceNotBindable))
			})
		})

		Context("when describing the Queue fails", func() {
			BeforeEach(func() {
				queue.DescribeError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})

		Context("when creating User fails", func() {
			BeforeEach(func() {
				user.CreateError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})

		Context("when creating User Access Keys fails", func() {
			BeforeEach(func() {
				user.CreateAccessKeyError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(user.DeleteCalled).To(BeTrue())
				Expect(user.DeleteUserName).To(Equal(userName))
			})
		})

		Context("when describing User fails", func() {
			BeforeEach(func() {
				user.DescribeError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(user.DeleteAccessKeyCalled).To(BeTrue())
				Expect(user.DeleteAccessKeyUserName).To(Equal(userName))
				Expect(user.DeleteAccessKeyAccessKeyID).To(Equal("user-access-key-id"))
				Expect(user.DeleteCalled).To(BeTrue())
				Expect(user.DeleteUserName).To(Equal(userName))
			})
		})

		Context("when Adding Permissions fails", func() {
			BeforeEach(func() {
				queue.AddPermissionError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			It("makes the proper calls", func() {
				_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
				Expect(err).To(HaveOccurred())
				Expect(user.DeleteAccessKeyCalled).To(BeTrue())
				Expect(user.DeleteAccessKeyUserName).To(Equal(userName))
				Expect(user.DeleteAccessKeyAccessKeyID).To(Equal("user-access-key-id"))
				Expect(user.DeleteCalled).To(BeTrue())
				Expect(user.DeleteUserName).To(Equal(userName))
			})

			Context("when the Queue does not exists", func() {
				BeforeEach(func() {
					queue.AddPermissionError = awssqs.ErrQueueDoesNotExist
				})

				It("returns the proper error", func() {
					_, err := sqsBroker.Bind(instanceID, bindingID, bindDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Unbind", func() {
		var (
			unbindDetails brokerapi.UnbindDetails
		)

		BeforeEach(func() {
			unbindDetails = brokerapi.UnbindDetails{
				ServiceID: "Service-1",
				PlanID:    "Plan-1",
			}
		})

		It("makes the proper calls", func() {
			err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
			Expect(queue.RemovePermissionCalled).To(BeTrue())
			Expect(queue.RemovePermissionQueueName).To(Equal(queueName))
			Expect(queue.RemovePermissionQueueLabel).To(Equal(queueLabel))
			Expect(user.ListAccessKeysCalled).To(BeTrue())
			Expect(user.ListAccessKeysUserName).To(Equal(userName))
			Expect(user.DeleteCalled).To(BeTrue())
			Expect(user.DeleteUserName).To(Equal(userName))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when removing Permissions fails", func() {
			BeforeEach(func() {
				queue.RemovePermissionError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("when the Queue does not exists", func() {
				BeforeEach(func() {
					queue.RemovePermissionError = awssqs.ErrQueueDoesNotExist
				})

				It("returns the proper error", func() {
					err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
				})
			})
		})

		Context("when listing the User Access Keys fails", func() {
			BeforeEach(func() {
				user.ListAccessKeysError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})

		Context("when User has Access Keys", func() {
			BeforeEach(func() {
				user.ListAccessKeysAccessKeys = []string{"access-key-id-1"}
			})

			It("makes the proper calls", func() {
				err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
				Expect(user.DeleteAccessKeyCalled).To(BeTrue())
				Expect(user.DeleteAccessKeyUserName).To(Equal(userName))
				Expect(user.DeleteAccessKeyAccessKeyID).To(Equal("access-key-id-1"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when deleting the User Access Keys fails", func() {
				BeforeEach(func() {
					user.DeleteAccessKeyError = errors.New("operation failed")
				})

				It("returns the proper error", func() {
					err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("operation failed"))
				})
			})
		})

		Context("when deleting the User fails", func() {
			BeforeEach(func() {
				user.DeleteError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := sqsBroker.Unbind(instanceID, bindingID, unbindDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})
		})
	})

	var _ = Describe("LastOperation", func() {
		It("returns the proper error", func() {
			_, err := sqsBroker.LastOperation(instanceID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("This broker does not support LastOperation"))
		})
	})
})
