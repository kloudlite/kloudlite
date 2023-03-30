package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kloudlite/cluster-operator/lib/operator"
	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/types"
)

func (r *metadataResolver) Labels(ctx context.Context, obj *v1.ObjectMeta) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	var m map[string]any
	if err := fn.JsonConversion(obj.Labels, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *metadataResolver) CreationTimestamp(ctx context.Context, obj *v1.ObjectMeta) (string, error) {
	return obj.CreationTimestamp.Format(time.RFC3339), nil
}

func (r *metadataResolver) DeletionTimestamp(ctx context.Context, obj *v1.ObjectMeta) (*string, error) {
	d := obj.GetDeletionTimestamp()
	if d == nil {
		return nil, nil
	}
	return fn.New(d.Format(time.RFC3339)), nil
}

func (r *patchResolver) Value(ctx context.Context, obj *json_patch.PatchOperation) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.Value, nil
}

func (r *statusResolver) Checks(ctx context.Context, obj *operator.Status) (map[string]interface{}, error) {
	m := make(map[string]any, len(obj.Checks))
	if err := fn.JsonConversion(obj.Checks, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *statusResolver) DisplayVars(ctx context.Context, obj *operator.Status) (map[string]interface{}, error) {
	var m map[string]any
	b, err := obj.DisplayVars.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *syncStatusResolver) SyncScheduledAt(ctx context.Context, obj *types.SyncStatus) (string, error) {
	return obj.SyncScheduledAt.Format(time.RFC3339), nil
}

func (r *syncStatusResolver) LastSyncedAt(ctx context.Context, obj *types.SyncStatus) (*string, error) {
	return fn.New(obj.LastSyncedAt.Format(time.RFC3339)), nil
}

func (r *syncStatusResolver) Action(ctx context.Context, obj *types.SyncStatus) (model.SyncAction, error) {
	return model.SyncAction(obj.Action), nil
}

func (r *syncStatusResolver) State(ctx context.Context, obj *types.SyncStatus) (model.SyncState, error) {
	return model.SyncState(obj.State), nil
}

func (r *metadataInResolver) Labels(ctx context.Context, obj *v1.ObjectMeta, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Labels)
}

func (r *patchInResolver) Value(ctx context.Context, obj *json_patch.PatchOperation, data interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(&data, &obj.Value)
}

// Metadata returns generated.MetadataResolver implementation.
func (r *Resolver) Metadata() generated.MetadataResolver { return &metadataResolver{r} }

// Patch returns generated.PatchResolver implementation.
func (r *Resolver) Patch() generated.PatchResolver { return &patchResolver{r} }

// Status returns generated.StatusResolver implementation.
func (r *Resolver) Status() generated.StatusResolver { return &statusResolver{r} }

// SyncStatus returns generated.SyncStatusResolver implementation.
func (r *Resolver) SyncStatus() generated.SyncStatusResolver { return &syncStatusResolver{r} }

// MetadataIn returns generated.MetadataInResolver implementation.
func (r *Resolver) MetadataIn() generated.MetadataInResolver { return &metadataInResolver{r} }

// PatchIn returns generated.PatchInResolver implementation.
func (r *Resolver) PatchIn() generated.PatchInResolver { return &patchInResolver{r} }

type metadataResolver struct{ *Resolver }
type patchResolver struct{ *Resolver }
type statusResolver struct{ *Resolver }
type syncStatusResolver struct{ *Resolver }
type metadataInResolver struct{ *Resolver }
type patchInResolver struct{ *Resolver }
