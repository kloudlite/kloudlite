package domain

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(cxt context.Context, action SetupClusterAction) error
	UpdateCluster(cxt context.Context, action UpdateClusterAction) error
	DeleteCluster(cxt context.Context, action DeleteClusterAction) error
	AddPeerToAccount(cxt context.Context, action AddPeerAction) error
	DeletePeerFromCluster(cxt context.Context, action DeletePeerAction) error
	GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) (map[string]string, error)
	AddAccountToCluster(ctx context.Context, payload AddAccountAction) error
}

type domain struct {
	infraCli     InfraClient
	messageTopic string
	jobResponder InfraJobResponder
}

func (d *domain) AddPeerToAccount(cxt context.Context, action AddPeerAction) error {
	err := d.infraCli.AddPeer(cxt, action)
	fmt.Println("AddPeerToAccount", err)
	if err != nil {
		d.jobResponder.SendAddPeerResponse(cxt, AddPeerResponse{
			ClusterId: action.ClusterId,
			AccountId: action.AccountId,
			PublicKey: action.PublicKey,
			Message:   err.Error(),
			Done:      false,
		})
		return err
	}
	d.jobResponder.SendAddPeerResponse(cxt, AddPeerResponse{
		ClusterId: action.ClusterId,
		AccountId: action.AccountId,
		PublicKey: action.PublicKey,
		Message:   "Peer added",
		Done:      true,
	})
	return err
}

func (d *domain) AddAccountToCluster(ctx context.Context, action AddAccountAction) error {
	port, publicKey, err := d.infraCli.AddAccount(ctx, action)
	if err != nil {
		d.jobResponder.SendSetupAccountResponse(ctx, SetupAccountResponse{
			ClusterId: string(action.ClusterId),
			AccountId: action.AccountId,
			Message:   err.Error(),
			Done:      false,
		})
		return err
	}
	d.jobResponder.SendSetupAccountResponse(ctx, SetupAccountResponse{
		ClusterId:   string(action.ClusterId),
		AccountId:   action.AccountId,
		WgPublicKey: publicKey,
		WgPort:      port,
		Message:     "Peer added",
		Done:        true,
	})
	return err
}

func (d *domain) GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) (map[string]string, error) {
	return d.infraCli.GetResourceOutput(ctx, clusterId, resName, namespace)
}

// DeletePeerFromCluster implements Domain
func (d *domain) DeletePeerFromCluster(cxt context.Context, action DeletePeerAction) error {
	return d.infraCli.DeletePeer(cxt, action)
}

// DeleteCluster implements Domain
func (d *domain) DeleteCluster(cxt context.Context, action DeleteClusterAction) error {
	return d.infraCli.DeleteCluster(cxt, action)
}

func (d *domain) CreateCluster(cxt context.Context, action SetupClusterAction) error {
	publicIp, publicKey, err := d.infraCli.CreateCluster(cxt, action)
	if err != nil {
		d.jobResponder.SendCreateClusterResponse(cxt, SetupClusterResponse{
			ClusterId: action.ClusterId,
			PublicIp:  publicIp,
			PublicKey: publicKey,
			Done:      false,
			Message:   err.Error(),
		})
	}
	d.jobResponder.SendCreateClusterResponse(cxt, SetupClusterResponse{
		ClusterId: action.ClusterId,
		PublicIp:  publicIp,
		PublicKey: publicKey,
		Done:      true,
	})
	return err
}

func (d *domain) UpdateCluster(cxt context.Context, action UpdateClusterAction) error {
	err := d.infraCli.UpdateCluster(cxt, action)
	fmt.Println("UpdateCluster", err)
	if err != nil {
		return d.jobResponder.SendUpdateClusterResponse(cxt, UpdateClusterResponse{
			ClusterId:  action.ClusterId,
			Region:     action.Region,
			Provider:   action.Provider,
			NodesCount: action.NodesCount,
			Done:       false,
		})
	}
	return d.jobResponder.SendUpdateClusterResponse(cxt, UpdateClusterResponse{
		ClusterId:  action.ClusterId,
		Region:     action.Region,
		Provider:   action.Provider,
		NodesCount: action.NodesCount,
		Done:       true,
	})
}
func fxDomain(
	env *Env,
	infraCli InfraClient,
	infraJobResp InfraJobResponder,
) Domain {
	return &domain{
		infraCli:     infraCli,
		jobResponder: infraJobResp,
		messageTopic: env.KafkaInfraActionResulTopic,
	}
}

type Env struct {
	KafkaInfraActionResulTopic string `env:"KAFKA_INFRA_ACTION_RESULT_TOPIC", required:"true"`
}

var Module = fx.Module("domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
	fx.Invoke(func(ij InfraJobResponder, d Domain, p messaging.Producer[messaging.Json], lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				//ij.SendCreateClusterResponse(ctx, SetupClusterResponse{
				//	ClusterId: "clus-le8xeokcvycsn8uwutsmuzimk5up",
				//	PublicIp:  "1234",
				//	PublicKey: "12345",
				//	Done:      true,
				//})
				go func() {
					//d.CreateCluster(ctx, SetupClusterAction{
					//	ClusterId:  "test-to-dlete",
					//	Region:     "blr1",
					//	Provider:   "do",
					//	NodesCount: 0,
					//})
					//d.AddPeerToAccount(ctx, AddPeerAction{
					//	ClusterId: "test-to-dlete",
					//	AccountId: "test-account",
					//	PublicKey: "BeNQy7XnndbfgtiCuhv3wMzW/ennPk2Ee4FZyRJkISg=",
					//	PeerIp:    "10.12.1.2",
					//})
					//d.AddAccountToCluster(ctx, AddAccountAction{
					//	ClusterId: "test-to-dlete",
					//	Provider:  "do",
					//	AccountId: "test-account",
					//	AccountIp: "10.12.1.1",
					//})
					//d.UpdateCluster(ctx, UpdateClusterAction{
					//	ClusterId:  "test-to-dlete",
					//	Region:     "blr1",
					//	Provider:   "do",
					//	NodesCount: 1,
					//})
					//key1, _ := wgtypes.GenerateKey()
					//fmt.Println(key1.String())
					//d.AddPeerToAccount(ctx, AddPeerAction{
					//	ClusterId: "hotspot-dev",
					//	PublicKey: key1.PublicKey().String(),
					//	PeerIp:    "10.13.13.101",
					//})

					//key2, _ := wgtypes.GenerateKey()
					//fmt.Println(key2.String())
					//d.AddPeerToAccount(ctx, AddPeerAction{
					//	ClusterId: "hotspot-dev",
					//	PublicKey: key2.PublicKey().String(),
					//	PeerIp:    "10.13.13.102",
					//})

					//key3, _ := wgtypes.GenerateKey()
					//fmt.Println(key3.String())
					//d.AddPeerToAccount(ctx, AddPeerAction{
					//	ClusterId: "hotspot-dev",
					//	PublicKey: key3.PublicKey().String(),
					//	PeerIp:    "10.13.13.103",
					//})

				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		})
		//ClusterId  string `json:"cluster_id"`
		//Region     string `json:"region"`
		//Provider   string `json:"provider"`
		//NodesCount int    `json:"nodes_count"`

		//d.DeleteCluster(DeleteClusterAction{
		//	ClusterId: "hotspot-dev",
		//	Provider:  "do",
		//})

		// if err != nil {
		// 	panic(err)
		// }

		//d.CreateCluster(SetupClusterAction{
		//	ClusterId:  "hotspot-dev-2",
		//	Region:     "blr1",
		//	Provider:   "do",
		//	NodesCount: 4,
		//})

		//d.UpdateCluster(UpdateClusterAction{
		//	ClusterId:  "hotspot-dev",
		//	Region:     "blr1",
		//	Provider:   "do",
		//	NodesCount: 2,
		//})

		//key, _ := wgtypes.GenerateKey()
		//fmt.Println(key.String())
		//d.AddPeerToAccount(AddPeerAction{
		//	ClusterId: "hotspot-dev-2",
		//	PublicKey: key.PublicKey().String(),
		//	PeerIp:    "10.13.13.104",
		//})

		//d.DeletePeerFromCluster(DeletePeerAction{
		//	ClusterId: "hotspot-dev-2",
		//	PublicKey: "1uBcGZvNsNh7wlNzawDXiAExIfbgyFgbJqTwGRTmdiY=",
		//})
	}),
)
