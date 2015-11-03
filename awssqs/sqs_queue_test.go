package awssqs_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/sqs-broker/awssqs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("SQS Queue", func() {
	var (
		queueName string
		queueURL  string

		awsSession *session.Session
		sqssvc     *sqs.SQS
		sqsCall    func(r *request.Request)

		testSink *lagertest.TestSink
		logger   lager.Logger

		queue Queue
	)

	BeforeEach(func() {
		queueName = "sqs-queue"
		queueURL = "sqs-queue-url"
	})

	JustBeforeEach(func() {
		awsSession = session.New(nil)
		sqssvc = sqs.New(awsSession)

		logger = lager.NewLogger("sqsqueue_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		queue = NewSQSQueue(sqssvc, logger)
	})

	var _ = Describe("Describe", func() {
		var (
			properQueueDetails QueueDetails

			getQueueURLInput *sqs.GetQueueUrlInput
			getQueueURLError error

			getQueueAttributes      map[string]*string
			getQueueAttributesInput *sqs.GetQueueAttributesInput
			getQueueAttributesError error
		)

		BeforeEach(func() {
			properQueueDetails = QueueDetails{
				QueueURL:               queueURL,
				QueueArn:               "test-queue-arn",
				DelaySeconds:           "test-delay-seconds",
				MaximumMessageSize:     "test-maximum-message-size",
				MessageRetentionPeriod: "test-message-retention-period",
				Policy:                 "test-policy",
				ReceiveMessageWaitTimeSeconds: "test-receive-message-wait-time-seconds",
				VisibilityTimeout:             "test-visibility-timeout",
			}

			getQueueURLInput = &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			}
			getQueueURLError = nil

			getQueueAttributes = map[string]*string{
				"QueueArn":                      aws.String("test-queue-arn"),
				"DelaySeconds":                  aws.String("test-delay-seconds"),
				"MaximumMessageSize":            aws.String("test-maximum-message-size"),
				"MessageRetentionPeriod":        aws.String("test-message-retention-period"),
				"Policy":                        aws.String("test-policy"),
				"ReceiveMessageWaitTimeSeconds": aws.String("test-receive-message-wait-time-seconds"),
				"VisibilityTimeout":             aws.String("test-visibility-timeout"),
			}
			getQueueAttributesInput = &sqs.GetQueueAttributesInput{
				QueueUrl:       aws.String(queueURL),
				AttributeNames: aws.StringSlice([]string{"All"}),
			}
			getQueueAttributesError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("GetQueueUrl|GetQueueAttributes"))
				switch r.Operation.Name {
				case "GetQueueUrl":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueUrlInput{}))
					Expect(r.Params).To(Equal(getQueueURLInput))
					data := r.Data.(*sqs.GetQueueUrlOutput)
					data.QueueUrl = aws.String(queueURL)
					r.Error = getQueueURLError
				case "GetQueueAttributes":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueAttributesInput{}))
					Expect(r.Params).To(Equal(getQueueAttributesInput))
					data := r.Data.(*sqs.GetQueueAttributesOutput)
					data.Attributes = getQueueAttributes
					r.Error = getQueueAttributesError
				}
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("gets the Queue Attibutes", func() {
			queueDetails, err := queue.Describe(queueName)
			Expect(queueDetails).To(Equal(properQueueDetails))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when getting the Queue URL fails", func() {
			BeforeEach(func() {
				getQueueURLError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := queue.Describe(queueName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueURLError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					_, err := queue.Describe(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueURLError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					_, err := queue.Describe(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})

		Context("when getting the Queue Attibutes fails", func() {
			BeforeEach(func() {
				getQueueAttributesError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := queue.Describe(queueName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueAttributesError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					_, err := queue.Describe(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueAttributesError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					_, err := queue.Describe(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Create", func() {
		var (
			queueDetails QueueDetails

			createQueueInput *sqs.CreateQueueInput
			createQueueError error
		)

		BeforeEach(func() {
			queueDetails = QueueDetails{}

			createQueueInput = &sqs.CreateQueueInput{
				QueueName:  aws.String(queueName),
				Attributes: map[string]*string{},
			}
			createQueueError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("CreateQueue"))
				Expect(r.Params).To(BeAssignableToTypeOf(&sqs.CreateQueueInput{}))
				Expect(r.Params).To(Equal(createQueueInput))
				r.Error = createQueueError
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("creates the Queue", func() {
			err := queue.Create(queueName, queueDetails)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has DelaySeconds", func() {
			BeforeEach(func() {
				queueDetails.DelaySeconds = "test-delay-seconds"
				createQueueInput.Attributes["DelaySeconds"] = aws.String("test-delay-seconds")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has MaximumMessageSize", func() {
			BeforeEach(func() {
				queueDetails.MaximumMessageSize = "test-maximum-message-size"
				createQueueInput.Attributes["MaximumMessageSize"] = aws.String("test-maximum-message-size")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has MessageRetentionPeriod", func() {
			BeforeEach(func() {
				queueDetails.MessageRetentionPeriod = "test-message-retention-period"
				createQueueInput.Attributes["MessageRetentionPeriod"] = aws.String("test-message-retention-period")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Policy", func() {
			BeforeEach(func() {
				queueDetails.Policy = "test-policy"
				createQueueInput.Attributes["Policy"] = aws.String("test-policy")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ReceiveMessageWaitTimeSeconds", func() {
			BeforeEach(func() {
				queueDetails.ReceiveMessageWaitTimeSeconds = "test-receive-message-wait-time-seconds"
				createQueueInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String("test-receive-message-wait-time-seconds")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has VisibilityTimeout", func() {
			BeforeEach(func() {
				queueDetails.VisibilityTimeout = "test-visibility-timeout"
				createQueueInput.Attributes["VisibilityTimeout"] = aws.String("test-visibility-timeout")
			})

			It("does not return error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when creating the Queue fails", func() {
			BeforeEach(func() {
				createQueueError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.Create(queueName, queueDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					createQueueError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.Create(queueName, queueDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("Modify", func() {
		var (
			queueDetails QueueDetails

			getQueueURLInput *sqs.GetQueueUrlInput
			getQueueURLError error

			setQueueAttributesInput *sqs.SetQueueAttributesInput
			setQueueAttributesError error
		)

		BeforeEach(func() {
			queueDetails = QueueDetails{}

			getQueueURLInput = &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			}
			getQueueURLError = nil

			setQueueAttributesInput = &sqs.SetQueueAttributesInput{
				QueueUrl:   aws.String(queueURL),
				Attributes: map[string]*string{},
			}
			setQueueAttributesError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("GetQueueUrl|SetQueueAttributes"))
				switch r.Operation.Name {
				case "GetQueueUrl":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueUrlInput{}))
					Expect(r.Params).To(Equal(getQueueURLInput))
					data := r.Data.(*sqs.GetQueueUrlOutput)
					data.QueueUrl = aws.String(queueURL)
					r.Error = getQueueURLError
				case "SetQueueAttributes":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.SetQueueAttributesInput{}))
					Expect(r.Params).To(Equal(setQueueAttributesInput))
					r.Error = setQueueAttributesError
				}
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("sets the Queue Attibutes", func() {
			err := queue.Modify(queueName, queueDetails)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when has DelaySeconds", func() {
			BeforeEach(func() {
				queueDetails.DelaySeconds = "test-delay-seconds"
				setQueueAttributesInput.Attributes["DelaySeconds"] = aws.String("test-delay-seconds")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has MaximumMessageSize", func() {
			BeforeEach(func() {
				queueDetails.MaximumMessageSize = "test-maximum-message-size"
				setQueueAttributesInput.Attributes["MaximumMessageSize"] = aws.String("test-maximum-message-size")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has MessageRetentionPeriod", func() {
			BeforeEach(func() {
				queueDetails.MessageRetentionPeriod = "test-message-retention-period"
				setQueueAttributesInput.Attributes["MessageRetentionPeriod"] = aws.String("test-message-retention-period")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has Policy", func() {
			BeforeEach(func() {
				queueDetails.Policy = "test-policy"
				setQueueAttributesInput.Attributes["Policy"] = aws.String("test-policy")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has ReceiveMessageWaitTimeSeconds", func() {
			BeforeEach(func() {
				queueDetails.ReceiveMessageWaitTimeSeconds = "test-receive-message-wait-time-seconds"
				setQueueAttributesInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String("test-receive-message-wait-time-seconds")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when has VisibilityTimeout", func() {
			BeforeEach(func() {
				queueDetails.VisibilityTimeout = "test-visibility-timeout"
				setQueueAttributesInput.Attributes["VisibilityTimeout"] = aws.String("test-visibility-timeout")
			})

			It("does not return error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when getting the Queue URL fails", func() {
			BeforeEach(func() {
				getQueueURLError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueURLError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.Modify(queueName, queueDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueURLError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.Modify(queueName, queueDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})

		Context("when setting the Queue Attibutes fails", func() {
			BeforeEach(func() {
				setQueueAttributesError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.Modify(queueName, queueDetails)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					setQueueAttributesError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.Modify(queueName, queueDetails)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					setQueueAttributesError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.Modify(queueName, queueDetails)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("Delete", func() {
		var (
			getQueueURLInput *sqs.GetQueueUrlInput
			getQueueURLError error

			deleteQueueInput *sqs.DeleteQueueInput
			deleteQueueError error
		)

		BeforeEach(func() {
			getQueueURLInput = &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			}
			getQueueURLError = nil

			deleteQueueInput = &sqs.DeleteQueueInput{
				QueueUrl: aws.String(queueURL),
			}
			deleteQueueError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("GetQueueUrl|DeleteQueue"))
				switch r.Operation.Name {
				case "GetQueueUrl":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueUrlInput{}))
					Expect(r.Params).To(Equal(getQueueURLInput))
					data := r.Data.(*sqs.GetQueueUrlOutput)
					data.QueueUrl = aws.String(queueURL)
					r.Error = getQueueURLError
				case "DeleteQueue":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.DeleteQueueInput{}))
					Expect(r.Params).To(Equal(deleteQueueInput))
					r.Error = deleteQueueError
				}
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("deletes the Queue", func() {
			err := queue.Delete(queueName)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when getting the Queue URL fails", func() {
			BeforeEach(func() {
				getQueueURLError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.Delete(queueName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueURLError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.Delete(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueURLError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.Delete(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})

		Context("when deleting the Queue fails", func() {
			BeforeEach(func() {
				deleteQueueError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.Delete(queueName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					deleteQueueError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.Delete(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					deleteQueueError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.Delete(queueName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("AddPermission", func() {
		var (
			label      string
			accountIds []string
			actions    []string

			getQueueURLInput *sqs.GetQueueUrlInput
			getQueueURLError error

			addPermissionInput *sqs.AddPermissionInput
			addPermissionError error
		)

		BeforeEach(func() {
			label = "test-label"
			accountIds = []string{"principal"}
			actions = []string{"*"}

			getQueueURLInput = &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			}
			getQueueURLError = nil

			addPermissionInput = &sqs.AddPermissionInput{
				QueueUrl:      aws.String(queueURL),
				Label:         aws.String(label),
				AWSAccountIds: aws.StringSlice(accountIds),
				Actions:       aws.StringSlice(actions),
			}
			addPermissionError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("GetQueueUrl|AddPermission"))
				switch r.Operation.Name {
				case "GetQueueUrl":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueUrlInput{}))
					Expect(r.Params).To(Equal(getQueueURLInput))
					data := r.Data.(*sqs.GetQueueUrlOutput)
					data.QueueUrl = aws.String(queueURL)
					r.Error = getQueueURLError
				case "AddPermission":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.AddPermissionInput{}))
					Expect(r.Params).To(Equal(addPermissionInput))
					r.Error = addPermissionError
				}
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("adds Permissions to the Queue", func() {
			err := queue.AddPermission(queueName, label, accountIds, actions)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when getting the Queue URL fails", func() {
			BeforeEach(func() {
				getQueueURLError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.AddPermission(queueName, label, accountIds, actions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueURLError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.AddPermission(queueName, label, accountIds, actions)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueURLError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.AddPermission(queueName, label, accountIds, actions)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})

		Context("when adding Permissions fails", func() {
			BeforeEach(func() {
				addPermissionError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.AddPermission(queueName, label, accountIds, actions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					addPermissionError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.AddPermission(queueName, label, accountIds, actions)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					addPermissionError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.AddPermission(queueName, label, accountIds, actions)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})
	})

	var _ = Describe("RemovePermission", func() {
		var (
			label string

			getQueueURLInput *sqs.GetQueueUrlInput
			getQueueURLError error

			removePermissionInput *sqs.RemovePermissionInput
			removePermissionError error
		)

		BeforeEach(func() {
			label = "test-label"

			getQueueURLInput = &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			}
			getQueueURLError = nil

			removePermissionInput = &sqs.RemovePermissionInput{
				QueueUrl: aws.String(queueURL),
				Label:    aws.String(label),
			}
			removePermissionError = nil
		})

		JustBeforeEach(func() {
			sqssvc.Handlers.Clear()

			sqsCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(MatchRegexp("GetQueueUrl|RemovePermission"))
				switch r.Operation.Name {
				case "GetQueueUrl":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.GetQueueUrlInput{}))
					Expect(r.Params).To(Equal(getQueueURLInput))
					data := r.Data.(*sqs.GetQueueUrlOutput)
					data.QueueUrl = aws.String(queueURL)
					r.Error = getQueueURLError
				case "RemovePermission":
					Expect(r.Params).To(BeAssignableToTypeOf(&sqs.RemovePermissionInput{}))
					Expect(r.Params).To(Equal(removePermissionInput))
					r.Error = removePermissionError
				}
			}
			sqssvc.Handlers.Send.PushBack(sqsCall)
		})

		It("removes Permissions from the Queue", func() {
			err := queue.RemovePermission(queueName, label)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when getting the Queue URL fails", func() {
			BeforeEach(func() {
				getQueueURLError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.RemovePermission(queueName, label)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getQueueURLError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.RemovePermission(queueName, label)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					getQueueURLError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.RemovePermission(queueName, label)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})

		Context("when removing Permissions fails", func() {
			BeforeEach(func() {
				removePermissionError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := queue.RemovePermission(queueName, label)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					removePermissionError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := queue.RemovePermission(queueName, label)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})

			Context("and it is a 404 error", func() {
				BeforeEach(func() {
					awsError := awserr.New("code", "message", errors.New("operation failed"))
					removePermissionError = awserr.NewRequestFailure(awsError, 404, "request-id")
				})

				It("returns the proper error", func() {
					err := queue.RemovePermission(queueName, label)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrQueueDoesNotExist))
				})
			})
		})
	})
})
