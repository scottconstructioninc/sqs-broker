package awsiam_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/sqs-broker/awsiam"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("IAM User", func() {
	var (
		region   string
		userName string

		iamsvc  *iam.IAM
		iamCall func(r *request.Request)

		testSink *lagertest.TestSink
		logger   lager.Logger

		user User
	)

	BeforeEach(func() {
		region = "iam-region"
		userName = "iam-user"
	})

	JustBeforeEach(func() {
		iamsvc = iam.New(nil)

		logger = lager.NewLogger("iamuser_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		user = NewIAMUser(region, iamsvc, logger)
	})

	var _ = Describe("Describe", func() {
		var (
			properUserDetails UserDetails

			getUser       *iam.User
			getUserInput  *iam.GetUserInput
			getUserOutput *iam.GetUserOutput
			getUserError  error
		)

		BeforeEach(func() {
			properUserDetails = UserDetails{
				UserName: userName,
				ARN:      "user-arn",
				UserID:   "user-id",
			}

			getUser = &iam.User{
				Arn:    aws.String("user-arn"),
				UserId: aws.String("user-id"),
			}
			getUserInput = &iam.GetUserInput{
				UserName: aws.String(userName),
			}
			getUserOutput = &iam.GetUserOutput{}
			getUserError = nil
		})

		JustBeforeEach(func() {
			iamsvc.Handlers.Clear()

			iamCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("GetUser"))
				Expect(r.Params).To(BeAssignableToTypeOf(&iam.GetUserInput{}))
				Expect(r.Params).To(Equal(getUserInput))
				data := r.Data.(*iam.GetUserOutput)
				data.User = getUser
				r.Error = getUserError
			}
			iamsvc.Handlers.Send.PushBack(iamCall)
		})

		It("returns the proper User Details", func() {
			userDetails, err := user.Describe(userName)
			Expect(err).ToNot(HaveOccurred())
			Expect(userDetails).To(Equal(properUserDetails))
		})

		Context("when getting the User fails", func() {
			BeforeEach(func() {
				getUserError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, err := user.Describe(userName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					getUserError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					_, err := user.Describe(userName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("Create", func() {
		var (
			createUserInput *iam.CreateUserInput
			createUserError error
		)

		BeforeEach(func() {
			createUserInput = &iam.CreateUserInput{
				UserName: aws.String(userName),
			}
			createUserError = nil
		})

		JustBeforeEach(func() {
			iamsvc.Handlers.Clear()

			iamCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("CreateUser"))
				Expect(r.Params).To(BeAssignableToTypeOf(&iam.CreateUserInput{}))
				Expect(r.Params).To(Equal(createUserInput))
				r.Error = createUserError
			}
			iamsvc.Handlers.Send.PushBack(iamCall)
		})

		It("creates the User", func() {
			err := user.Create(userName)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when creating the User fails", func() {
			BeforeEach(func() {
				createUserError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := user.Create(userName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					createUserError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := user.Create(userName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("Delete", func() {
		var (
			deleteUserInput *iam.DeleteUserInput
			deleteUserError error
		)

		BeforeEach(func() {
			deleteUserInput = &iam.DeleteUserInput{
				UserName: aws.String(userName),
			}
			deleteUserError = nil
		})

		JustBeforeEach(func() {
			iamsvc.Handlers.Clear()

			iamCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("DeleteUser"))
				Expect(r.Params).To(BeAssignableToTypeOf(&iam.DeleteUserInput{}))
				Expect(r.Params).To(Equal(deleteUserInput))
				r.Error = deleteUserError
			}
			iamsvc.Handlers.Send.PushBack(iamCall)
		})

		It("deletes the User", func() {
			err := user.Delete(userName)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deleting the User fails", func() {
			BeforeEach(func() {
				deleteUserError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := user.Delete(userName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					deleteUserError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := user.Delete(userName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("CreateAccessKey", func() {
		var (
			createAccessKey *iam.AccessKey

			createAccessKeyInput *iam.CreateAccessKeyInput
			createAccessKeyError error
		)

		BeforeEach(func() {
			createAccessKey = &iam.AccessKey{
				UserName:        aws.String(userName),
				AccessKeyId:     aws.String("access-key-id"),
				SecretAccessKey: aws.String("secret-access-key"),
			}

			createAccessKeyInput = &iam.CreateAccessKeyInput{
				UserName: aws.String(userName),
			}
			createAccessKeyError = nil
		})

		JustBeforeEach(func() {
			iamsvc.Handlers.Clear()

			iamCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("CreateAccessKey"))
				Expect(r.Params).To(BeAssignableToTypeOf(&iam.CreateAccessKeyInput{}))
				Expect(r.Params).To(Equal(createAccessKeyInput))
				data := r.Data.(*iam.CreateAccessKeyOutput)
				data.AccessKey = createAccessKey
				r.Error = createAccessKeyError
			}
			iamsvc.Handlers.Send.PushBack(iamCall)
		})

		It("creates the Access Key", func() {
			accessKeyID, secretAccessKey, err := user.CreateAccessKey(userName)
			Expect(err).ToNot(HaveOccurred())
			Expect(accessKeyID).To(Equal("access-key-id"))
			Expect(secretAccessKey).To(Equal("secret-access-key"))
		})

		Context("when creating the Access Key fails", func() {
			BeforeEach(func() {
				createAccessKeyError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				_, _, err := user.CreateAccessKey(userName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					createAccessKeyError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					_, _, err := user.CreateAccessKey(userName)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})

	var _ = Describe("DeleteAccessKey", func() {
		var (
			accessKeyID string

			deleteAccessKeyInput *iam.DeleteAccessKeyInput
			deleteAccessKeyError error
		)

		BeforeEach(func() {
			accessKeyID = "access-key-id"

			deleteAccessKeyInput = &iam.DeleteAccessKeyInput{
				UserName:    aws.String(userName),
				AccessKeyId: aws.String(accessKeyID),
			}
			deleteAccessKeyError = nil
		})

		JustBeforeEach(func() {
			iamsvc.Handlers.Clear()

			iamCall = func(r *request.Request) {
				Expect(r.Operation.Name).To(Equal("DeleteAccessKey"))
				Expect(r.Params).To(BeAssignableToTypeOf(&iam.DeleteAccessKeyInput{}))
				Expect(r.Params).To(Equal(deleteAccessKeyInput))
				r.Error = deleteAccessKeyError
			}
			iamsvc.Handlers.Send.PushBack(iamCall)
		})

		It("deletes the Access Key", func() {
			err := user.DeleteAccessKey(userName, accessKeyID)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when deleting the Access Key fails", func() {
			BeforeEach(func() {
				deleteAccessKeyError = errors.New("operation failed")
			})

			It("returns the proper error", func() {
				err := user.DeleteAccessKey(userName, accessKeyID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("operation failed"))
			})

			Context("and it is an AWS error", func() {
				BeforeEach(func() {
					deleteAccessKeyError = awserr.New("code", "message", errors.New("operation failed"))
				})

				It("returns the proper error", func() {
					err := user.DeleteAccessKey(userName, accessKeyID)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("code: message"))
				})
			})
		})
	})
})
