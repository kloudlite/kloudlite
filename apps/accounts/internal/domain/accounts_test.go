package domain_test

import (
	"context"
	"github.com/kloudlite/operator/pkg/kubectl"
	"google.golang.org/grpc"
	"kloudlite.io/apps/accounts/internal/domain"
	"kloudlite.io/apps/accounts/internal/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	authMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth/mocks"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	iamMock "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam/mocks"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
	"reflect"
	"testing"
)

type fields struct {
	authClient              auth.AuthClient
	iamClient               iam.IAMClient
	consoleClient           console.ConsoleClient
	containerRegistryClient container_registry.ContainerRegistryClient
	commsClient             comms.CommsClient
	accountRepo             repos.DbRepo[*entities.Account]
	invitationRepo          repos.DbRepo[*entities.Invitation]
	k8sYamlClient           *kubectl.YAMLClient
	k8sExtendedClient       k8s.ExtendedK8sClient
	logger                  logging.Logger
}

func getDomain(f fields) domain.Domain {
	return domain.NewDomain(
		f.iamClient,
		f.consoleClient,
		//f.containerRegistryClient,
		f.authClient,
		f.commsClient,
		f.k8sYamlClient,
		f.k8sExtendedClient,

		f.accountRepo,
		f.invitationRepo,

		f.logger,
	)
}

func Test_domain_ActivateAccount(t *testing.T) {
	type args struct {
		ctx  domain.UserContext
		name string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "activating account which does not exist",
			fields: fields{
				authClient: func() auth.AuthClient {
					cli := authMock.NewAuthClient()
					return cli
				}(),
				iamClient: func() iam.IAMClient {
					cli := iamMock.NewIAMClient()
					cli.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
						return &iam.CanOut{}, nil
					}
					return cli
				}(),
				//accountRepo: repos.NewInMemoryRepo[*entities.Account]("accounts", "account", map[repos.ID]*entities.Account{
				//	"test": {
				//		Account:  crdsv1.Account{},
				//		IsActive: fn.New(false),
				//	},
				//}),
			},
			args:    args{},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.ActivateAccount(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ActivateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ActivateAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_CreateAccount(t *testing.T) {
	type args struct {
		ctx     domain.UserContext
		account entities.Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entities.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.CreateAccount(tt.args.ctx, tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_DeactivateAccount(t *testing.T) {
	type args struct {
		ctx  domain.UserContext
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.DeactivateAccount(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeactivateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeactivateAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_DeleteAccount(t *testing.T) {
	type args struct {
		ctx  domain.UserContext
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.DeleteAccount(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeleteAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_GetAccount(t *testing.T) {
	type args struct {
		ctx  domain.UserContext
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entities.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.GetAccount(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_ListAccounts(t *testing.T) {
	type args struct {
		ctx         domain.UserContext
		accountName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*entities.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			got, err := d.ListAccounts(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAccounts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_domain_ResyncAccount(t *testing.T) {
	type args struct {
		ctx  domain.UserContext
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := getDomain(tt.fields)
			if err := d.ResyncAccount(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ResyncAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func Test_domain_UpdateAccount(t *testing.T) {
//	type args struct {
//		ctx     domain.UserContext
//		account *entities.Account
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *entities.Account
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := getDomain(tt.fields)
//			got, err := d.UpdateAccount(tt.args.ctx, tt.args.account)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("UpdateAccount() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("UpdateAccount() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
