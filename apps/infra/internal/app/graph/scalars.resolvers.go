package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	fn "kloudlite.io/pkg/functions"
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

func (r *patchResolver) Value(ctx context.Context, obj *json_patch.PatchOperation) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.Value, nil
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

// MetadataIn returns generated.MetadataInResolver implementation.
func (r *Resolver) MetadataIn() generated.MetadataInResolver { return &metadataInResolver{r} }

// PatchIn returns generated.PatchInResolver implementation.
func (r *Resolver) PatchIn() generated.PatchInResolver { return &patchInResolver{r} }

type metadataResolver struct{ *Resolver }
type patchResolver struct{ *Resolver }
type metadataInResolver struct{ *Resolver }
type patchInResolver struct{ *Resolver }
