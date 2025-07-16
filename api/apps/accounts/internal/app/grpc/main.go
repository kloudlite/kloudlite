package grpc

import (
	"context"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/repos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/common"
	"google.golang.org/grpc/metadata"
)

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
	d domain.Domain
}

func (a *accountsGrpcServer) CheckAccountNameAvailability(ctx context.Context, request *accounts.CheckAccountNameAvailabilityRequest) (*accounts.CheckAccountNameAvailabilityResponse, error) {
	availability, err := a.d.CheckNameAvailability(ctx, request.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check account name availability: %v", err)
	}
	if availability == nil {
		return nil, status.Errorf(codes.Internal, "check account name availability returned nil")
	}
	return &accounts.CheckAccountNameAvailabilityResponse{
		Result:         availability.Result,
		SuggestedNames: availability.SuggestedNames,
	}, nil
}

func getUserContext(ctx context.Context) (*domain.UserContext, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unable to extract metadata from context")
	}
	userId := incomingContext.Get("userId")[0]
	if len(userId) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in metadata")
	}
	userName := incomingContext.Get("userName")[0]
	userEmail := incomingContext.Get("userEmail")[0]
	return &domain.UserContext{
		UserId:    repos.ID(userId),
		UserName:  userName,
		UserEmail: userEmail,
	}, nil
}

func (a *accountsGrpcServer) CreateAccount(ctx context.Context, req *accounts.CreateAccountRequest) (*accounts.CreateAccountResponse, error) {
	userContext, err := getUserContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user context: %v", err)
	}
	accountActive := func() *bool {
		active := true
		return &active
	}()
	newAccount := entities.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: req.DisplayName,
		},
		IsActive:               accountActive,
		ContactEmail:           userContext.UserEmail,
		KloudliteGatewayRegion: req.Name,
	}
	account, err := a.d.CreateAccount(*userContext, newAccount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}
	if account == nil {
		return nil, status.Errorf(codes.Internal, "account creation returned nil")
	}
	return &accounts.CreateAccountResponse{
		AccountId: string(account.Id),
	}, nil
}

func (a *accountsGrpcServer) DeleteAccount(context.Context, *accounts.DeleteAccountRequest) (*accounts.DeleteAccountResponse, error) {
	panic("unimplemented")
}

// DisableAccount implements v2.AccountsV2Server.
func (a *accountsGrpcServer) DisableAccount(context.Context, *accounts.DisableAccountRequest) (*accounts.DisableAccountResponse, error) {
	panic("unimplemented")
}

// EnableAccount implements v2.AccountsV2Server.
func (a *accountsGrpcServer) EnableAccount(context.Context, *accounts.EnableAccountRequest) (*accounts.EnableAccountResponse, error) {
	panic("unimplemented")
}

// GetAccountDetails implements v2.AccountsV2Server.
func (a *accountsGrpcServer) GetAccountDetails(context.Context, *accounts.GetAccountDetailsRequest) (*accounts.GetAccountDetailsResponse, error) {
	panic("unimplemented")
}

// ListAccounts implements v2.AccountsV2Server.
func (a *accountsGrpcServer) ListAccounts(context.Context, *accounts.ListAccountsRequest) (*accounts.ListAccountsResponse, error) {
	panic("unimplemented")
}

func NewServer(d domain.Domain) accounts.AccountsServer {
	return &accountsGrpcServer{
		d: d,
	}
}
