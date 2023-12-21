package domain

import (
	"fmt"
	"github.com/kloudlite/api/pkg/errors"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
)

type ErrIAMUnauthorized struct {
	UserId   string
	Resource string
	Action   string
}

func (e ErrIAMUnauthorized) Error() string {
	return fmt.Sprintf("user (%q) is unauthorized to perform action (%q) on resource (%q)", e.UserId, e.Action, e.Resource)
}

type ErrGRPCCall struct {
	Err error
}

func (e ErrGRPCCall) Error() string {
	return fmt.Sprintf("grpc call failed with error: %v", errors.NewE(e.Err))
}

func (d *domain) canPerformActionInAccount(ctx InfraContext, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(action),
	})
	if err != nil {
		return ErrGRPCCall{Err: err}
	}
	if !co.Status {
		return ErrIAMUnauthorized{
			UserId:   string(ctx.UserId),
			Resource: fmt.Sprintf("account: %s", ctx.AccountName),
			Action:   string(action),
		}
	}
	return nil
}
