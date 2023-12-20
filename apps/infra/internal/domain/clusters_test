package domain

import (
	"context"
	"errors"
	"testing"

	domainMocks "github.com/kloudlite/api/apps/infra/internal/domain/mocks"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	iamMock "github.com/kloudlite/api/mocks/grpc-interfaces/kloudlite.io/rpc/iam"
	msgOfficeInternalMock "github.com/kloudlite/api/mocks/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	k8sMock "github.com/kloudlite/api/mocks/pkg/k8s"

	reposMock "github.com/kloudlite/api/mocks/pkg/repos"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCreateCluster(t *testing.T) {
	type domainArgs struct {
		env *env.Env

		byocClusterRepo reposMock.DbRepo[*entities.BYOCCluster]
		clusterRepo     reposMock.DbRepo[*entities.Cluster]
		nodeRepo        reposMock.DbRepo[*entities.Node]
		nodePoolRepo    reposMock.DbRepo[*entities.NodePool]
		domainEntryRepo reposMock.DbRepo[*entities.DomainEntry]
		secretRepo      reposMock.DbRepo[*entities.CloudProviderSecret]
		vpnDeviceRepo   reposMock.DbRepo[*entities.VPNDevice]
		pvcRepo         reposMock.DbRepo[*entities.PersistentVolumeClaim]
		buildRunRepo    reposMock.DbRepo[*entities.BuildRun]

		k8sClient k8sMock.Client


		iamClient                   iamMock.IAMClient
		accountsSvc                 domainMocks.AccountsSvc
		messageOfficeInternalClient msgOfficeInternalMock.MessageOfficeInternalClient
	}

	type args struct {
		ctx     InfraContext
		cluster entities.Cluster
	}

	type want struct {
		cluster   *entities.Cluster
		errorLike func(t *testing.T, gotErr error)
	}

	type test struct {
		name        string
		buildDomain func(d *domainArgs)
		args        args
		want        want
	}

	logerr := func(t *testing.T, gotErr error, wantErr error) {
		t.Errorf("CreateCluster() errored, got error = %v, want error = %v", gotErr, wantErr)
	}

	tests := []test{
		{
			name: "1. when iam grpc service is unreachable/down, creation should fail",
			buildDomain: func(d *domainArgs) {
				d.iamClient = iamMock.IAMClient{}
				d.iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return nil, errors.New("iam grpc service is unreachable/down")
				}
			},
			args: args{
				ctx:     InfraContext{},
				cluster: entities.Cluster{},
			},
			want: want{
				cluster: nil,
				errorLike: func(t *testing.T, gotErr error) {
					var werr ErrGRPCCall
					if !errors.As(gotErr, &werr) {
						logerr(t, gotErr, werr)
					}
				},
			},
		},
		{
			name: "2. when user is not allowed to create cluster, creation should fail",
			buildDomain: func(d *domainArgs) {
				d.iamClient = iamMock.IAMClient{}
				d.iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{
						Status: false,
					}, nil
				}
			},
			args: args{
				ctx:     InfraContext{},
				cluster: entities.Cluster{},
			},
			want: want{
				cluster: nil,
				errorLike: func(t *testing.T, gotErr error) {
					var werr ErrIAMUnauthorized
					if !errors.As(gotErr, &werr) {
						logerr(t, gotErr, werr)
					}
				},
			},
		},
		{
			name: "3. when account does not exist, creation should fail",
			buildDomain: func(d *domainArgs) {
				d.iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{
						Status: true,
					}, nil
				}

				d.accountsSvc.MockGetAccount = func(ctx context.Context, userId, accountName string) (*accounts.GetAccountOut, error) {
					return nil, errors.New("account does not exist")
				}
			},
			args: args{
				ctx: InfraContext{},
				cluster: entities.Cluster{
					Cluster: clustersv1.Cluster{
						ObjectMeta: metav1.ObjectMeta{},
					},
				},
			},
			want: want{
				cluster: nil,
				errorLike: func(t *testing.T, gotErr error) {
					werr := errors.New("account does not exist")
					if gotErr.Error() != werr.Error() {
						logerr(t, gotErr, werr)
					}
				},
			},
		},
		{
			name: "4. when cluster already exists, creation should fail",
			buildDomain: func(d *domainArgs) {
				d.iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
					return &iam.CanOut{
						Status: true,
					}, nil
				}

				d.accountsSvc.MockGetAccount = func(ctx context.Context, userId, accountName string) (*accounts.GetAccountOut, error) {
					return &accounts.GetAccountOut{
						IsActive:        true,
						TargetNamespace: "sample",
						AccountId:       "sample",
					}, nil
				}

				d.clusterRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Cluster, error) {
					return &entities.Cluster{
						BaseEntity: repos.BaseEntity{},
						Cluster: clustersv1.Cluster{
							ObjectMeta: metav1.ObjectMeta{
								Name: "sample",
							},
						},
						ResourceMetadata: common.ResourceMetadata{},
						AccountName:      "",
						SyncStatus:       types.SyncStatus{},
					}, nil
				}

				d.secretRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.CloudProviderSecret, error) {
					return &entities.CloudProviderSecret{}, nil
				}

				d.messageOfficeInternalClient.MockGenerateClusterToken = func(ctx context.Context, in *message_office_internal.GenerateClusterTokenIn, opts ...grpc.CallOption) (*message_office_internal.GenerateClusterTokenOut, error) {
					return &message_office_internal.GenerateClusterTokenOut{
						ClusterToken: "sample",
					}, nil
				}

				d.k8sClient.MockValidateObject = func(ctx context.Context, obj client.Object) error {
					return nil
				}

				d.k8sClient.MockApplyYAML = func(ctx context.Context, yamls ...[]byte) error {
					return nil
				}

				d.clusterRepo.MockCreate = func(ctx context.Context, data *entities.Cluster) (*entities.Cluster, error) {
					return &entities.Cluster{}, nil
				}
			},
			args: args{
				ctx: InfraContext{},
				cluster: entities.Cluster{
					Cluster: clustersv1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "sample",
						},
					},
				},
			},
			want: want{
				cluster: nil,
				errorLike: func(t *testing.T, gotErr error) {
					var werr ErrClusterAlreadyExists
					if !errors.As(gotErr, &werr) {
						logerr(t, gotErr, werr)
					}
				},
			},
		},
		// {
		// 	name: "5. when cluster already exists",
		// 	buildDomain: func(d *domainArgs) {
		// 		d.iamClient = iamMock.IAMClient{}
		// 		d.iamClient.MockCan = func(ctx context.Context, in *iam.CanIn, opts ...grpc.CallOption) (*iam.CanOut, error) {
		// 			return &iam.CanOut{
		// 				Status: true,
		// 			}, nil
		// 		}
		//
		// 		d.accountsSvc.MockGetAccount = func(ctx context.Context, userId, accountName string) (*accounts.GetAccountOut, error) {
		// 			return &accounts.GetAccountOut{
		// 				IsActive:        true,
		// 				TargetNamespace: "sample",
		// 				AccountId:       "sample",
		// 			}, nil
		// 		}
		//
		// 		d.clusterRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.Cluster, error) {
		// 			return &entities.Cluster{
		// 				BaseEntity: repos.BaseEntity{},
		// 				Cluster: clustersv1.Cluster{
		// 					ObjectMeta: metav1.ObjectMeta{
		// 						Name: "sample",
		// 					},
		// 				},
		// 				ResourceMetadata: common.ResourceMetadata{},
		// 				AccountName:      "",
		// 				SyncStatus:       types.SyncStatus{},
		// 			}, nil
		// 		}
		//
		// 		d.secretRepo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.CloudProviderSecret, error) {
		// 			return &entities.CloudProviderSecret{}, nil
		// 		}
		//
		// 		d.messageOfficeInternalClient.MockGenerateClusterToken = func(ctx context.Context, in *message_office_internal.GenerateClusterTokenIn, opts ...grpc.CallOption) (*message_office_internal.GenerateClusterTokenOut, error) {
		// 			return &message_office_internal.GenerateClusterTokenOut{
		// 				ClusterToken: "sample",
		// 			}, nil
		// 		}
		//
		// 		d.k8sClient.MockValidateObject = func(ctx context.Context, obj client.Object) error {
		// 			return nil
		// 		}
		//
		// 		d.k8sClient.MockApplyYAML = func(ctx context.Context, yamls ...[]byte) error {
		// 			return nil
		// 		}
		//
		// 		d.clusterRepo.MockCreate = func(ctx context.Context, data *entities.Cluster) (*entities.Cluster, error) {
		// 			return &entities.Cluster{}, nil
		// 		}
		// 	},
		// 	args: args{
		// 		ctx:     InfraContext{},
		// 		cluster: entities.Cluster{},
		// 	},
		// 	want: want{
		// 		cluster: nil,
		// 		errorLike: func(t *testing.T, gotErr error) {
		// 			var werr ErrIAMUnauthorized
		// 			if !errors.As(gotErr, &werr) {
		// 				logerr(t, gotErr, werr)
		// 			}
		// 		},
		// 	},
		// },
	}

	logger, _ := logging.New(&logging.Options{Name: "test"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dargs := domainArgs{}
			tt.buildDomain(&dargs)
			d := domain{
				logger:                      logger,
				env:                         &env.Env{},
				byocClusterRepo:             &dargs.byocClusterRepo,
				clusterRepo:                 &dargs.clusterRepo,
				nodeRepo:                    &dargs.nodeRepo,
				nodePoolRepo:                &dargs.nodePoolRepo,
				domainEntryRepo:             &dargs.domainEntryRepo,
				secretRepo:                  &dargs.secretRepo,
				vpnDeviceRepo:               &dargs.vpnDeviceRepo,
				pvcRepo:                     &dargs.pvcRepo,
				buildRunRepo:                &dargs.buildRunRepo,
				k8sClient:                   &dargs.k8sClient,
				iamClient:                   &dargs.iamClient,
				accountsSvc:                 &dargs.accountsSvc,
				messageOfficeInternalClient: &dargs.messageOfficeInternalClient,
			}
			got, err := d.CreateCluster(tt.args.ctx, tt.args.cluster)
			if err != nil && tt.want.errorLike != nil {
				tt.want.errorLike(t, err)
			}
			if got != tt.want.cluster {
				t.Errorf("CreateCluster() returned, got cluster = %v, want cluster = %v", got, tt.want)
			}
		})
	}
}
