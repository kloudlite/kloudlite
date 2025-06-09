package ports

import (
	"context"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
)

type IAMService interface {
	AddMembership(ctx context.Context, in *iam.AddMembershipIn) (*iam.AddMembershipOut, error)
	// Can(ctx context.Context, in *CanIn, opts ...grpc.CallOption) (*CanOut, error)
	// ListMembershipsForResource(ctx context.Context, in *MembershipsForResourceIn, opts ...grpc.CallOption) (*ListMembershipsOut, error)
	// ListMembershipsForUser(ctx context.Context, in *MembershipsForUserIn, opts ...grpc.CallOption) (*ListMembershipsOut, error)
	// GetMembership(ctx context.Context, in *GetMembershipIn, opts ...grpc.CallOption) (*GetMembershipOut, error)
	// // Mutation
	// AddMembership(ctx context.Context, in *AddMembershipIn, opts ...grpc.CallOption) (*AddMembershipOut, error)
	// UpdateMembership(ctx context.Context, in *UpdateMembershipIn, opts ...grpc.CallOption) (*UpdateMembershipOut, error)
	// RemoveMembership(ctx context.Context, in *RemoveMembershipIn, opts ...grpc.CallOption) (*RemoveMembershipOut, error)
	// RemoveResource(ctx context.Context, in *RemoveResourceIn, opts ...grpc.CallOption) (*RemoveResourceOut, error)
}
