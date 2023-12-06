package domain_test

import (
	"context"
	"fmt"

	"kloudlite.io/common"

	"github.com/kloudlite/operator/pkg/kubectl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"kloudlite.io/apps/accounts/internal/domain"
	"kloudlite.io/apps/accounts/internal/entities"
	authMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth/mocks"
	commsMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms/mocks"
	consoleMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console/mocks"
	_ "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	iamMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam/mocks"
	k8sMock "kloudlite.io/mocks/pkg/k8s"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// "kloudlite.io/pkg/repos"
	reposMock "kloudlite.io/mocks/pkg/repos"
)

var _ = Describe("domain.ActivateAccount says", func() {
	// Given("an account", func() {})
	var authClient *authMock.AuthClient
	var iamClient *iamMock.IAMClient
	var consoleClient *consoleMock.ConsoleClient
	// var containerRegistryClient container_registry.ContainerRegistryClient
	var commsClient *commsMock.CommsClient
	var accountRepo *reposMock.DbRepo[*entities.Account]
	var invitationRepo *reposMock.DbRepo[*entities.Invitation]
	var k8sYamlClient kubectl.YAMLClient
	var k8sExtendedClient *k8sMock.ExtendedK8sClient

	logger, err := logging.New(&logging.Options{Name: "test"})
	if err != nil {
		panic(err)
	}

	BeforeEach(func() {
		authClient = authMock.NewAuthClient()
		iamClient = iamMock.NewIAMClient()
		consoleClient = consoleMock.NewConsoleClient()
		// containerRegistryClient = container_registry.NewContainerRegistryClient()
		commsClient = commsMock.NewCommsClient()
		accountRepo = reposMock.NewDbRepo[*entities.Account]()
		invitationRepo = reposMock.NewDbRepo[*entities.Invitation]()
		// k8sYamlClient = kubectl.NewYAMLClient()
		k8sExtendedClient = k8sMock.NewExtendedK8sClient()
	})

	getDomain := func() domain.Domain {
		return domain.NewDomain(
			iamClient,
			consoleClient,
			// f.containerRegistryClient,
			authClient,
			commsClient,
			k8sYamlClient,
			k8sExtendedClient,

			accountRepo,
			invitationRepo,

			logger,
		)
	}

	When("user has no IAM permission to activate account", func() {
		It("account activation should fail", func() {
			d := getDomain()

			iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
				return &iam.CanOut{Status: false}, nil
			}

			accountRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Account, error) {
				return nil, fmt.Errorf("not found")
			}

			_, err := d.ActivateAccount(domain.UserContext{}, "sample")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unauthorized"))
		})
	})

	When("user has IAM permission to activate account", func() {
		Context("but account does not exist", func() {
			It("it should fail", func() {
				d := getDomain()

				iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{Status: true}, nil
				}

				accountRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Account, error) {
					return nil, fmt.Errorf("mock: account not found")
				}

				_, err := d.ActivateAccount(domain.UserContext{}, "sample")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("mock: account not found"))
			})
		})

		Context("but account is already active", func() {
			It("it should fail", func() {
				d := getDomain()

				iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{Status: true}, nil
				}

				accountRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Account, error) {
					return &entities.Account{IsActive: fn.New(true)}, nil
				}

				_, err := d.ActivateAccount(domain.UserContext{}, "sample")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already active"))
			})
		})

		Context("and account exists and it is currently disabled", func() {
			It("then, account should get activated", func() {
				d := getDomain()

				iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{Status: true}, nil
				}

				accountRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Account, error) {
					return &entities.Account{IsActive: fn.New(false)}, nil
				}

				accountRepo.MockUpdateById = func(ctx context.Context, id repos.ID, updatedData *entities.Account, opts ...repos.UpdateOpts) (*entities.Account, error) {
					return &entities.Account{IsActive: fn.New(true)}, nil
				}
				_, err := d.ActivateAccount(domain.UserContext{}, "sample")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

var _ = Describe("domain.CreateAccount() says", func() {
	// Given("an account", func() {})
	var authClient *authMock.AuthClient
	var iamClient *iamMock.IAMClient
	var consoleClient *consoleMock.ConsoleClient
	// var containerRegistryClient container_registry.ContainerRegistryClient
	var commsClient *commsMock.CommsClient
	var accountRepo *reposMock.DbRepo[*entities.Account]
	var invitationRepo *reposMock.DbRepo[*entities.Invitation]
	var k8sYamlClient kubectl.YAMLClient
	var k8sExtendedClient *k8sMock.ExtendedK8sClient

	logger, err := logging.New(&logging.Options{Name: "test"})
	if err != nil {
		panic(err)
	}

	BeforeEach(func() {
		authClient = authMock.NewAuthClient()
		iamClient = iamMock.NewIAMClient()
		consoleClient = consoleMock.NewConsoleClient()
		// containerRegistryClient = container_registry.NewContainerRegistryClient()
		commsClient = commsMock.NewCommsClient()
		accountRepo = reposMock.NewDbRepo[*entities.Account]()
		invitationRepo = reposMock.NewDbRepo[*entities.Invitation]()
		// k8sYamlClient = kubectl.NewYAMLClient()
		k8sExtendedClient = k8sMock.NewExtendedK8sClient()
	})

	getDomain := func() domain.Domain {
		return domain.NewDomain(
			iamClient,
			consoleClient,
			// f.containerRegistryClient,
			authClient,
			commsClient,
			k8sYamlClient,
			k8sExtendedClient,

			accountRepo,
			invitationRepo,

			logger,
		)
	}

	When("account already exists", func() {
		It("fails", func() {
			d := getDomain()

			// iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) { }

			accountRepo.MockCreate = func(ctx context.Context, data *entities.Account) (*entities.Account, error) {
				return nil, fmt.Errorf("account already exists")
			}

			k8sExtendedClient.MockValidateStruct = func(ctx context.Context, obj client.Object) error {
				return nil
			}

			_, err := d.CreateAccount(domain.UserContext{}, entities.Account{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("account already exists"))
		})
	})

	When("account does not exist", func() {
		Context("but account is not valid", func() {
			BeforeEach(func() {
				k8sExtendedClient.MockValidateStruct = func(ctx context.Context, obj client.Object) error {
					return fmt.Errorf("invalid account data")
				}
			})

			It("fails", func() {
				d := getDomain()

				accountRepo.MockCreate = func(ctx context.Context, data *entities.Account) (*entities.Account, error) {
					return nil, fmt.Errorf("account already exists")
				}

				_, err := d.CreateAccount(domain.UserContext{}, entities.Account{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid account data"))
			})
		})

		Context("account is valid", func() {
			BeforeEach(func() {
				k8sExtendedClient.MockValidateStruct = func(ctx context.Context, obj client.Object) error {
					return nil
				}
			})

			It("succeeds", func() {
				d := getDomain()

				accountRepo.MockCreate = func(ctx context.Context, data *entities.Account) (*entities.Account, error) {
					return data, nil
				}

				iamClient.MockAddMembership = func(ctx context.Context, in *iam.AddMembershipIn, opts ...grpc.CallOption) (*iam.AddMembershipOut, error) {
					return &iam.AddMembershipOut{Result: true}, nil
				}

				acc, err := d.CreateAccount(domain.UserContext{
					Context:   context.TODO(),
					UserId:    "sample-userid",
					UserName:  "sample-username",
					UserEmail: "sample-useremail",
				}, entities.Account{})

				Expect(err).To(BeNil())

				Expect(acc.CreatedBy).To(BeEquivalentTo(common.CreatedOrUpdatedBy{
					UserId:    "sample-userid",
					UserName:  "sample-username",
					UserEmail: "sample-useremail",
				}))
				Expect(acc.LastUpdatedBy).To(BeEquivalentTo(common.CreatedOrUpdatedBy{
					UserId:    "sample-userid",
					UserName:  "sample-username",
					UserEmail: "sample-useremail",
				}))
			})
		})
	})
})
