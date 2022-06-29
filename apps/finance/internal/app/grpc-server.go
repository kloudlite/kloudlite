package app

import (
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
)

type financeServerImpl struct {
	finance.UnimplementedFinanceServer
	d domain.Domain
}

func fxFinanceGrpcServer(domain domain.Domain) finance.FinanceServer {
	return &financeServerImpl{
		d: domain,
	}
}
