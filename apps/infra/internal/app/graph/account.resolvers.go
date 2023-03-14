package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kloudlite/cluster-operator/lib/operator"
	v1 "github.com/kloudlite/wg-operator/apis/wg/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	fn "kloudlite.io/pkg/functions"
)

func (r *accountResolver) Status(ctx context.Context, obj *v1.Account) (*operator.Status, error) {
	if obj != nil {
		return nil, nil
	}

	var m operator.Status
	if err := fn.JsonConversion(obj.Status, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *accountResolver) Spec(ctx context.Context, obj *v1.Account) (*model.AccountSpec, error) {
	if obj == nil {
		return nil, nil
	}
	var m model.AccountSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *accountInResolver) Spec(ctx context.Context, obj *v1.Account, data *model.AccountSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// AccountIn returns generated.AccountInResolver implementation.
func (r *Resolver) AccountIn() generated.AccountInResolver { return &accountInResolver{r} }

type accountResolver struct{ *Resolver }
type accountInResolver struct{ *Resolver }
