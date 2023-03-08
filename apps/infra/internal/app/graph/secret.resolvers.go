package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *secretResolver) Status(ctx context.Context, obj *entities.Secret) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}

	checks := make(map[string]any, len(obj.Status.Checks))
	for k, v := range obj.Status.Checks {
		checks[k] = v
	}

	var dvars map[string]any
	if err := fn.JsonConversion(obj.Status.DisplayVars, &dvars); err != nil {
		return nil, err
	}

	return &model.Status{
		IsReady:     obj.Status.IsReady,
		Checks:      checks,
		DisplayVars: dvars,
	}, nil
}

func (r *secretResolver) StringData(ctx context.Context, obj *entities.Secret) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	var m map[string]any
	if err := fn.JsonConversion(obj.StringData, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *secretResolver) Data(ctx context.Context, obj *entities.Secret) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	var m map[string]any
	if err := fn.JsonConversion(obj.Data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *secretResolver) Type(ctx context.Context, obj *entities.Secret) (*string, error) {
	if obj == nil {
		return fn.New(""), nil
	}
	return fn.New(string(obj.Type)), nil
}

func (r *secretInResolver) StringData(ctx context.Context, obj *entities.Secret, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.StringData)
}

func (r *secretInResolver) Data(ctx context.Context, obj *entities.Secret, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Data)
}

func (r *secretInResolver) Type(ctx context.Context, obj *entities.Secret, data *string) error {
	if obj == nil {
		return nil
	}
	obj.Type = corev1.SecretType(*data)
	return nil
}

// Secret returns generated.SecretResolver implementation.
func (r *Resolver) Secret() generated.SecretResolver { return &secretResolver{r} }

// SecretIn returns generated.SecretInResolver implementation.
func (r *Resolver) SecretIn() generated.SecretInResolver { return &secretInResolver{r} }

type secretResolver struct{ *Resolver }
type secretInResolver struct{ *Resolver }
