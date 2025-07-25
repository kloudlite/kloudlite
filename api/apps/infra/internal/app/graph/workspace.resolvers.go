package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
	"time"

	"github.com/kloudlite/api/apps/infra/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/infra/internal/app/graph/model"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreationTime is the resolver for the creationTime field.
func (r *workspaceResolver) CreationTime(ctx context.Context, obj *entities.Workspace) (string, error) {
	if obj == nil {
		return "", errors.Newf("workspace obj is nil")
	}
	return obj.CreationTime.Format(time.RFC3339), nil
}

// DispatchAddr is the resolver for the dispatchAddr field.
func (r *workspaceResolver) DispatchAddr(ctx context.Context, obj *entities.Workspace) (*model.GithubComKloudliteAPIAppsInfraInternalEntitiesDispatchAddr, error) {
	panic(fmt.Errorf("not implemented: DispatchAddr - dispatchAddr"))
}

// ID is the resolver for the id field.
func (r *workspaceResolver) ID(ctx context.Context, obj *entities.Workspace) (repos.ID, error) {
	if obj == nil {
		return "", errors.Newf("workspace obj is nil")
	}
	return obj.Id, nil
}

// Spec is the resolver for the spec field.
func (r *workspaceResolver) Spec(ctx context.Context, obj *entities.Workspace) (*model.GithubComKloudliteOperatorApisCrdsV1WorkspaceSpec, error) {
	var m model.GithubComKloudliteOperatorApisCrdsV1WorkspaceSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, errors.NewE(err)
	}
	return &m, nil
}

// UpdateTime is the resolver for the updateTime field.
func (r *workspaceResolver) UpdateTime(ctx context.Context, obj *entities.Workspace) (string, error) {
	if obj == nil || obj.UpdateTime.IsZero() {
		return "", errors.Newf("workspace is nil")
	}
	return obj.UpdateTime.Format(time.RFC3339), nil
}

// WorkmachineName is the resolver for the workmachineName field.
func (r *workspaceResolver) WorkmachineName(ctx context.Context, obj *entities.Workspace) (string, error) {
	panic(fmt.Errorf("not implemented: WorkmachineName - workmachineName"))
}

// Metadata is the resolver for the metadata field.
func (r *workspaceInResolver) Metadata(ctx context.Context, obj *entities.Workspace, data *v1.ObjectMeta) error {
	if obj == nil {
		return errors.Newf("workspace is nil")
	}
	return fn.JsonConversion(data, &obj.ObjectMeta)
}

// Spec is the resolver for the spec field.
func (r *workspaceInResolver) Spec(ctx context.Context, obj *entities.Workspace, data *model.GithubComKloudliteOperatorApisCrdsV1WorkspaceSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// Status is the resolver for the status field.
func (r *workspaceInResolver) Status(ctx context.Context, obj *entities.Workspace, data *model.GithubComKloudliteOperatorToolkitReconcilerStatusIn) error {
	if obj == nil {
		return errors.Newf("workspace is nil")
	}
	return fn.JsonConversion(data, &obj.Status)
}

// Workspace returns generated.WorkspaceResolver implementation.
func (r *Resolver) Workspace() generated.WorkspaceResolver { return &workspaceResolver{r} }

// WorkspaceIn returns generated.WorkspaceInResolver implementation.
func (r *Resolver) WorkspaceIn() generated.WorkspaceInResolver { return &workspaceInResolver{r} }

type workspaceResolver struct{ *Resolver }
type workspaceInResolver struct{ *Resolver }
