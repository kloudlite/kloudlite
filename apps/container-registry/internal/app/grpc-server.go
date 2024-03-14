package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/container_registry"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/repos"
)

type ContainerRegistryGRPCServer grpc.Server

type grpcServer struct {
	container_registry.UnimplementedContainerRegistryServer
	d  domain.Domain
	ev *env.Env
}

// CreateReadOnlyCredentials implements container_registry.ContainerRegistryServer.
func (g *grpcServer) CreateReadOnlyCredential(ctx context.Context, in *container_registry.CreateReadOnlyCredentialIn) (*container_registry.CreateReadOnlyCredentialOut, error) {
	regctx := domain.RegistryContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserName:    "created-by-kloudlite",
		AccountName: in.AccountName,
		UserEmail:   "",
	}

	creds, err := g.d.CreateAdminCredential(regctx, entities.Credential{
		AccountName: in.AccountName,
		Access:      entities.RepoAccessReadOnly,
		Expiration:  entities.Expiration{Unit: entities.ExpirationUnitYear, Value: 17},
	})

	dockerConfigJson, err := json.Marshal(map[string]any{
		"auths": map[string]any{
			g.ev.RegistryHost: map[string]any{
				"auth": base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", creds.UserName, creds.TokenKey))),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &container_registry.CreateReadOnlyCredentialOut{
		DockerConfigJson: dockerConfigJson,
	}, nil
}

func InitializeGrpcServer(server ContainerRegistryGRPCServer, d domain.Domain, ev *env.Env) {
	gs := grpcServer{d: d, ev: ev}
	container_registry.RegisterContainerRegistryServer(server, &gs)
}
