package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kloudlite/cluster-operator/lib/operator"
	"github.com/kloudlite/wg-operator/apis/wg/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/pkg/types"
)

func (r *accountResolver) SyncStatus(ctx context.Context, obj *v1.Account) (*types.SyncStatus, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *accountResolver) Spec(ctx context.Context, obj *v1.Account) (*model.AccountSpec, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *accountResolver) Status(ctx context.Context, obj *v1.Account) (*operator.Status, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *accountInResolver) Spec(ctx context.Context, obj *v1.Account, data *model.AccountSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// AccountIn returns generated.AccountInResolver implementation.
func (r *Resolver) AccountIn() generated.AccountInResolver { return &accountInResolver{r} }

type accountResolver struct{ *Resolver }
type accountInResolver struct{ *Resolver }
