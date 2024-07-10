package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	common_types "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	operator "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type grpcServer struct {
	d domain.Domain
	console.UnimplementedConsoleServer
	kcli k8s.Client
}

func (g *grpcServer) ArchiveEnvironmentsForCluster(ctx context.Context, in *console.ArchiveEnvironmentsForClusterIn) (*console.ArchiveEnvironmentsForClusterOut, error) {
	consoleCtx := domain.ConsoleContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserName:    in.UserName,
		UserEmail:   in.UserEmail,
		AccountName: in.AccountName,
	}

	archiveStatus, err := g.d.ArchiveEnvironmentsForCluster(consoleCtx, in.ClusterName)
	if err != nil {
		return &console.ArchiveEnvironmentsForClusterOut{Archived: false}, err
	}

	return &console.ArchiveEnvironmentsForClusterOut{Archived: archiveStatus}, nil
}

// CreateManagedResource implements console.ConsoleServer.
func (g *grpcServer) CreateManagedResource(ctx context.Context, in *console.CreateManagedResourceIn) (*console.CreateManagedResourceOut, error) {
	consoleCtx := domain.ConsoleContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserName:    in.UserName,
		UserEmail:   in.UserEmail,
		AccountName: in.AccountName,
	}

	// domain.ManagedResourceContext{
	// 	ConsoleContext:     ctx,
	// 	ManagedServiceName: new(string),
	// 	EnvironmentName:    new(string),
	// }

	var outputSecret corev1.Secret
	if err := json.Unmarshal(in.OutputSecret, &outputSecret); err != nil {
		return nil, err
	}

	createdBy := common.CreatedOrUpdatedBy{
		UserId:    repos.ID(in.UserId),
		UserName:  in.UserName,
		UserEmail: in.UserEmail,
	}

	_, err := g.d.CreateRootManagedResource(consoleCtx, in.AccountNamespace, &entities.ManagedResource{
		ManagedResource: crdsv1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      in.MresName,
				Namespace: in.MsvcTargetNamespace,
			},
			Spec: crdsv1.ManagedResourceSpec{
				ResourceTemplate: crdsv1.MresResourceTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "RootCredentials",
						APIVersion: in.MsvcApiVersion,
					},
					MsvcRef: common_types.MsvcRef{
						Name:      in.MsvcName,
						Namespace: "",
					},
					Spec: map[string]apiextensionsv1.JSON{},
				},
			},
			Status: operator.Status{
				IsReady: true,
			},
			Output: common_types.ManagedResourceOutput{
				CredentialsRef: common_types.LocalObjectReference{
					Name: outputSecret.Name,
				},
			},
		},
		AccountName:           in.AccountName,
		ManagedServiceName:    in.MsvcName,
		ClusterName:           in.ClusterName,
		SyncedOutputSecretRef: &outputSecret,
		ResourceMetadata: common.ResourceMetadata{
			DisplayName:   fmt.Sprintf("%s/%s", in.MsvcName, in.MresName),
			CreatedBy:     createdBy,
			LastUpdatedBy: createdBy,
		},
		// SyncStatus:            types.SyncStatus{},
	})
	if err != nil {
		return nil, err
	}
	return &console.CreateManagedResourceOut{Ok: true}, nil
}

func newConsoleGrpcServer(d domain.Domain, kcli k8s.Client) console.ConsoleServer {
	return &grpcServer{
		d:    d,
		kcli: kcli,
	}
}
