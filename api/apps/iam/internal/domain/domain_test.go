package domain_test

import (
	"context"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kloudlite.io/apps/iam/internal/domain"
	"kloudlite.io/apps/iam/internal/entities"
	t "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	reposMock "kloudlite.io/pkg/repos/mocks"
)

var _ = Describe("domain.AddRoleBinding() says", func() {
	var rbRepo *reposMock.DbRepo[*entities.RoleBinding]
	//var roleBindingMap map[t.Action][]t.Role

	BeforeEach(func() {
		rbRepo = reposMock.NewDbRepo[*entities.RoleBinding]()
	})

	When("role binding already exists", func() {
		BeforeEach(func() {
			rbRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.RoleBinding, error) {
				return &entities.RoleBinding{}, nil
			}
		})

		It("should return error", func() {
			d := domain.NewDomain(rbRepo, nil)
			_, err := d.AddRoleBinding(context.Background(), entities.RoleBinding{
				UserId:       "sample",
				ResourceType: "sample",
				ResourceRef:  "sample",
				Role:         "sample",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already exists"))
		})
	})

	When("role binding does not exist", func() {
		BeforeEach(func() {
			rbRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.RoleBinding, error) {
				return nil, nil
			}
		})

		It("creating an empty role binding should throw validation error", func() {
			d := domain.NewDomain(rbRepo, nil)
			_, err := d.AddRoleBinding(context.Background(), entities.RoleBinding{})
			Expect(err).To(HaveOccurred())
			Expect(errors.As(err, &common.ValidationError{})).To(BeTrue())
		})

		It("should create a new role binding", func() {
			rbRepo.MockCreate = func(ctx context.Context, data *entities.RoleBinding) (*entities.RoleBinding, error) {
				return data, nil
			}
			d := domain.NewDomain(rbRepo, nil)
			_, err := d.AddRoleBinding(context.Background(), entities.RoleBinding{
				UserId:       "sample",
				ResourceType: "sample",
				ResourceRef:  "sample",
				Role:         "sample",
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
var _ = Describe("domain.Can() says", func() {
	var rbRepo *reposMock.DbRepo[*entities.RoleBinding]

	BeforeEach(func() {
		rbRepo = reposMock.NewDbRepo[*entities.RoleBinding]()
	})

	When("no action-role-binding map exists", func() {
		It("should fail with error", func() {
			d := domain.NewDomain(rbRepo, nil)
			_, err := d.Can(context.Background(), "sample", []string{"sample"}, "sample")
			Expect(err).To(HaveOccurred())
			Expect(errors.As(err, &domain.UnAuthorizedError{})).To(BeTrue())
		})
	})

	Context("when action-role-binging map exists", func() {
		When("no role bindings exist in db", func() {
			It("should fail with error", func() {
				rbRepo.MockFind = func(ctx context.Context, query repos.Query) ([]*entities.RoleBinding, error) {
					return nil, nil
				}
				d := domain.NewDomain(rbRepo, map[t.Action][]t.Role{"sample": {"sample"}})
				_, err := d.Can(context.Background(), "sample", []string{"sample"}, "sample")
				Expect(err).To(HaveOccurred())
				Expect(errors.As(err, &domain.UnAuthorizedError{})).To(BeTrue())
			})
		})

		When("action is not permitted for the role", func() {
			It("1. if allowed roles is empty, should fail with error", func() {
				rbRepo.MockFind = func(ctx context.Context, query repos.Query) ([]*entities.RoleBinding, error) {
					return []*entities.RoleBinding{
						{
							UserId:       "sample",
							ResourceType: "sample-resource",
							ResourceRef:  "sample",
							Role:         "sample",
						},
					}, nil
				}

				actionRoleBindings := map[t.Action][]t.Role{
					"sample-action": {},
				}

				d := domain.NewDomain(rbRepo, actionRoleBindings)
				_, err := d.Can(context.Background(), "sample", []string{"sample"}, "sample-action")
				Expect(err).To(HaveOccurred())
				Expect(errors.As(err, &domain.UnAuthorizedError{})).To(BeTrue())
			})

			It("2. if allowed roles does not contain current role, should fail with error", func() {
				rbRepo.MockFind = func(ctx context.Context, query repos.Query) ([]*entities.RoleBinding, error) {
					return []*entities.RoleBinding{
						{
							UserId:       "sample-userid",
							ResourceType: "sample-resource",
							ResourceRef:  "sample-resourceRef",
							Role:         "sample-role",
						},
					}, nil
				}

				actionRoleBindings := map[t.Action][]t.Role{
					"sample-action": {"example-role"},
				}

				d := domain.NewDomain(rbRepo, actionRoleBindings)
				_, err := d.Can(context.Background(), "sample", []string{"sample"}, "sample-action")
				Expect(err).To(HaveOccurred())
				Expect(errors.As(err, &domain.UnAuthorizedError{})).To(BeTrue())
			})
		})

		When("action is permitted for the role", func() {
			It("should return true", func() {
				rbRepo.MockFind = func(ctx context.Context, query repos.Query) ([]*entities.RoleBinding, error) {
					return []*entities.RoleBinding{
						{
							UserId:       "sample-userid",
							ResourceType: "sample-resource",
							ResourceRef:  "sample-resourceRef",
							Role:         "sample-role",
						},
					}, nil
				}

				actionRoleBindings := map[t.Action][]t.Role{
					"sample-action": {"sample-role"},
				}

				d := domain.NewDomain(rbRepo, actionRoleBindings)
				_, err := d.Can(context.Background(), "sample", []string{"sample"}, "sample-action")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when system is internally accessing resources", func() {
		It("should allow", func() {
			d := domain.NewDomain(rbRepo, map[t.Action][]t.Role{"sample-action": {"sample"}})
			can, err := d.Can(context.Background(), "sys-user", []string{"sample"}, "sample-action")
			Expect(err).NotTo(HaveOccurred())
			Expect(can).To(BeTrue())
		})
	})
})
