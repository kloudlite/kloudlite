package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *secretResolver) Type(ctx context.Context, obj *entities.Secret) (*string, error) {
	s := string(obj.Type)
	return &s, nil
}

func (r *secretResolver) Data(ctx context.Context, obj *entities.Secret) (map[string]interface{}, error) {
	if obj == nil || obj.Data == nil {
		return nil, nil
	}
	m := make(map[string]any, len(obj.Data))
	if err := fn.JsonConversion(obj.Data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *secretResolver) StringData(ctx context.Context, obj *entities.Secret) (map[string]interface{}, error) {
	if obj == nil || obj.StringData == nil {
		return nil, nil
	}
	m := make(map[string]any, len(obj.StringData))
	if err := fn.JsonConversion(obj.StringData, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *secretInResolver) Type(ctx context.Context, obj *entities.Secret, data *string) error {
	if data == nil {
		return nil
	}
	obj.Type = corev1.SecretType(*data)
	return nil
}

func (r *secretInResolver) Data(ctx context.Context, obj *entities.Secret, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}

	if obj.Data == nil {
		obj.Data = make(map[string][]byte, len(data))
	}
	return fn.JsonConversion(data, &obj.Data)
}

func (r *secretInResolver) StringData(ctx context.Context, obj *entities.Secret, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	if obj.StringData == nil {
		obj.StringData = make(map[string]string, len(data))
	}
	return fn.JsonConversion(data, &obj.StringData)
}

// Secret returns generated.SecretResolver implementation.
func (r *Resolver) Secret() generated.SecretResolver { return &secretResolver{r} }

// SecretIn returns generated.SecretInResolver implementation.
func (r *Resolver) SecretIn() generated.SecretInResolver { return &secretInResolver{r} }

type secretResolver struct{ *Resolver }
type secretInResolver struct{ *Resolver }
