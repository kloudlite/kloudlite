package app

import (
	"context"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/repos"
)

type financeServerImpl struct {
	finance.UnimplementedFinanceServer
	d domain.Domain
}

func (f *financeServerImpl) StartBillable(ctx context.Context, in *finance.StartBillableIn) (*finance.StartBillableOut, error) {
	billable, err := f.d.StartBillable(ctx, repos.ID(in.AccountId), in.BillableType, in.Quantity)
	if err != nil {
		return nil, err
	}
	return &finance.StartBillableOut{BillingId: string(billable.Id)}, err
}

func (f *financeServerImpl) StopBillable(ctx context.Context, in *finance.StopBillableIn) (*finance.StopBillableOut, error) {
	err := f.d.StopBillable(ctx, repos.ID(in.BillableId))
	if err != nil {
		return nil, err
	}
	return &finance.StopBillableOut{}, err
}

func fxFinanceGrpcServer(domain domain.Domain) finance.FinanceServer {
	return &financeServerImpl{
		d: domain,
	}
}
