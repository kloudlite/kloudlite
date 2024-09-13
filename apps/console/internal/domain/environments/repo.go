package environments

// import (
// 	"fmt"
// 	"log/slog"
// 	"strings"
//
// 	"github.com/kloudlite/api/apps/console/internal/domain/ports"
// 	"github.com/kloudlite/api/apps/console/internal/domain/types"
// 	"github.com/kloudlite/api/apps/console/internal/entities"
// 	"github.com/kloudlite/api/common"
// 	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
// 	"github.com/kloudlite/api/pkg/errors"
// 	"github.com/kloudlite/api/pkg/k8s"
// 	"github.com/kloudlite/api/pkg/repos"
//
// 	iamT "github.com/kloudlite/api/apps/iam/types"
// 	t "github.com/kloudlite/api/pkg/types"
// 	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
// )
//
// type Repo struct {
// 	logger *slog.Logger
//
// 	k8sClient       k8s.Client
// 	environmentRepo repos.DbRepo[*entities.Environment]
//
// 	resourceEventPublisher types.ResourceEventPublisher
//
// 	iamSvc ports.IAMService
// }
//
// func getEnvironmentTargetNamespace(envName string) string {
// 	return fmt.Sprintf("env-%s", envName)
// }
//
// // ArchiveEnvironmentsForCluster implements Domain.
// func (r *Repo) ArchiveEnvironmentsForCluster(ctx types.ConsoleContext, clusterName string) (bool, error) {
// 	panic("unimplemented")
// }
//
// // CloneEnvironment implements Domain.
// func (r *Repo) CloneEnvironment(ctx types.ConsoleContext, args CloneEnvironmentArgs) (*entities.Environment, error) {
// 	panic("unimplemented")
// }
//
// // CreateEnvironment implements Domain.
// func (r *Repo) CreateEnvironment(ctx types.ConsoleContext, env entities.Environment) (*entities.Environment, error) {
// 	if strings.TrimSpace(env.ClusterName) == "" {
// 		return nil, errors.New("clustername must be set while creating environments")
// 	}
//
// 	env.EnsureGVK()
// 	if err := r.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	env.IncrementRecordVersion()
//
// 	if env.Spec.TargetNamespace == "" {
// 		env.Spec.TargetNamespace = getEnvironmentTargetNamespace(env.Name)
// 	}
//
// 	if env.Spec.Routing == nil {
// 		env.Spec.Routing = &crdsv1.EnvironmentRouting{
// 			Mode: crdsv1.EnvironmentRoutingModePrivate,
// 		}
// 	}
//
// 	env.CreatedBy = common.CreatedOrUpdatedBy{
// 		UserId:    ctx.UserId,
// 		UserName:  ctx.UserName,
// 		UserEmail: ctx.UserEmail,
// 	}
// 	env.LastUpdatedBy = env.CreatedBy
//
// 	env.AccountName = ctx.AccountName
// 	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)
//
// 	nenv, err := r.environmentRepo.Create(ctx, &env)
// 	if err != nil {
// 		if r.environmentRepo.ErrAlreadyExists(err) {
// 			// TODO: better insights into error, when it is being caused by duplicated indexes
// 			return nil, errors.NewE(err)
// 		}
// 		return nil, errors.NewE(err)
// 	}
//
// 	// FIXME: this should be setup when actually dispatching resource
// 	// if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, EnvironmentName: env.Name}, &env); err != nil {
// 	// 	return nil, errors.NewE(err)
// 	// }
//
// 	r.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, nenv.Name, types.PublishAdd)
//
// 	if _, err := r.iamSvc.AddMembership(ctx, &iam.AddMembershipIn{
// 		UserId:       string(ctx.UserId),
// 		ResourceType: string(iamT.ResourceEnvironment),
// 		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, nenv.Name),
// 		Role:         string(iamT.RoleResourceOwner),
// 	}); err != nil {
// 		r.logger.Error("while adding membership, got", "err", err)
// 	}
//
// 	if err := d.applyEnvironmentTargetNamespace(ctx, nenv); err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if err := d.applyK8sResource(ctx, nenv.Name, &nenv.Environment, nenv.RecordVersion); err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if err := d.syncImagePullSecretsToEnvironment(ctx, nenv.Name); err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	return nenv, nil
// }
//
// // func (d *domain) CreateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
// // 	if strings.TrimSpace(env.ClusterName) == "" {
// // 		return nil, fmt.Errorf("clustername must be set while creating environments")
// // 	}
// //
// // 	env.EnsureGVK()
// // 	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	env.IncrementRecordVersion()
// //
// // 	if env.Spec.TargetNamespace == "" {
// // 		env.Spec.TargetNamespace = d.getEnvironmentTargetNamespace(env.Name)
// // 	}
// //
// // 	if env.Spec.Routing == nil {
// // 		env.Spec.Routing = &crdsv1.EnvironmentRouting{
// // 			Mode: crdsv1.EnvironmentRoutingModePrivate,
// // 		}
// // 	}
// //
// // 	env.CreatedBy = common.CreatedOrUpdatedBy{
// // 		UserId:    ctx.UserId,
// // 		UserName:  ctx.UserName,
// // 		UserEmail: ctx.UserEmail,
// // 	}
// // 	env.LastUpdatedBy = env.CreatedBy
// //
// // 	env.AccountName = ctx.AccountName
// // 	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)
// //
// // 	nenv, err := d.environmentRepo.Create(ctx, &env)
// // 	if err != nil {
// // 		if d.environmentRepo.ErrAlreadyExists(err) {
// // 			// TODO: better insights into error, when it is being caused by duplicated indexes
// // 			return nil, errors.NewE(err)
// // 		}
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, EnvironmentName: env.Name}, &env); err != nil {
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, nenv.Name, PublishAdd)
// //
// // 	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
// // 		UserId:       string(ctx.UserId),
// // 		ResourceType: string(iamT.ResourceEnvironment),
// // 		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, nenv.Name),
// // 		Role:         string(iamT.RoleResourceOwner),
// // 	}); err != nil {
// // 		d.logger.Errorf(err, "error while adding membership")
// // 	}
// //
// // 	if err := d.applyEnvironmentTargetNamespace(ctx, nenv); err != nil {
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	if err := d.applyK8sResource(ctx, nenv.Name, &nenv.Environment, nenv.RecordVersion); err != nil {
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	if err := d.syncImagePullSecretsToEnvironment(ctx, nenv.Name); err != nil {
// // 		return nil, errors.NewE(err)
// // 	}
// //
// // 	return nenv, nil
// // }
//
// // DeleteEnvironment implements Domain.
// func (r *Repo) DeleteEnvironment(ctx types.ConsoleContext, name string) error {
// 	panic("unimplemented")
// }
//
// // GetEnvironment implements Domain.
// func (r *Repo) GetEnvironment(ctx types.ConsoleContext, name string) (*entities.Environment, error) {
// 	panic("unimplemented")
// }
//
// // ListEnvironments implements Domain.
// func (r *Repo) ListEnvironments(ctx types.ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error) {
// 	panic("unimplemented")
// }
//
// // UpdateEnvironment implements Domain.
// func (r *Repo) UpdateEnvironment(ctx types.ConsoleContext, env entities.Environment) (*entities.Environment, error) {
// 	panic("unimplemented")
// }
//
// var _ Domain = (*Repo)(nil)
