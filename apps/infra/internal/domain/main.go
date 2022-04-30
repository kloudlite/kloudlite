package domain

import (
	"context"
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(cxt context.Context, action SetupClusterAction) error
	UpdateCluster(cxt context.Context, action UpdateClusterAction) error
	DeleteCluster(cxt context.Context, action DeleteClusterAction) error
	AddPeerToCluster(cxt context.Context, action AddPeerAction) error
	DeletePeerFromCluster(cxt context.Context, action DeletePeerAction) error
	GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) ([]byte, error)
}

type domain struct {
	infraCli     InfraClient
	messageTopic string
	jobResponder InfraJobResponder
}

func (d *domain) GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) ([]byte, error) {
	return d.infraCli.GetResourceOutput(ctx, clusterId, resName, namespace)
}

// AddPeerToCluster implements Domain
func (d *domain) AddPeerToCluster(cxt context.Context, action AddPeerAction) error {
	err := d.infraCli.AddPeer(cxt, action)
	if err != nil {
		d.jobResponder.SendAddPeerResponse(cxt, AddPeerResponse{
			ClusterID: action.ClusterID,
			PublicKey: action.PublicKey,
			Message:   err.Error(),
			Done:      false,
		})
		return err
	}
	d.jobResponder.SendAddPeerResponse(cxt, AddPeerResponse{
		ClusterID: action.ClusterID,
		PublicKey: action.PublicKey,
		Message:   "Peer added",
		Done:      true,
	})
	return err
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
			ClusterID: action.ClusterID,
			PublicIp:  publicIp,
			PublicKey: publicKey,
			Done:      false,
			Message:   err.Error(),
		})
	}
	d.jobResponder.SendCreateClusterResponse(cxt, SetupClusterResponse{
		ClusterID: action.ClusterID,
		PublicIp:  publicIp,
		PublicKey: publicKey,
		Done:      true,
	})
	return err
}

func (d *domain) UpdateCluster(cxt context.Context, action UpdateClusterAction) error {
	_, _, err := d.infraCli.CreateCluster(cxt, SetupClusterAction{
		ClusterID:  action.ClusterID,
		Region:     action.Region,
		Provider:   action.Provider,
		NodesCount: action.NodesCount,
	})
	return err
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
				//	ClusterID: "clus-le8xeokcvycsn8uwutsmuzimk5up",
				//	PublicIp:  "1234",
				//	PublicKey: "12345",
				//	Done:      true,
				//})
				go func() {
					d.CreateCluster(ctx, SetupClusterAction{
						ClusterID:  "hotspot-dev-k3s",
						Region:     "blr1",
						Provider:   "do",
						NodesCount: 3,
					})
					//d.UpdateCluster(ctx, UpdateClusterAction{
					//	ClusterID:  "hotspot-dev",
					//	Region:     "blr1",
					//	Provider:   "do",
					//	NodesCount: 3,
					//})
					//key1, _ := wgtypes.GenerateKey()
					//fmt.Println(key1.String())
					//d.AddPeerToCluster(ctx, AddPeerAction{
					//	ClusterID: "hotspot-dev",
					//	PublicKey: key1.PublicKey().String(),
					//	PeerIp:    "10.13.13.101",
					//})

					//key2, _ := wgtypes.GenerateKey()
					//fmt.Println(key2.String())
					//d.AddPeerToCluster(ctx, AddPeerAction{
					//	ClusterID: "hotspot-dev",
					//	PublicKey: key2.PublicKey().String(),
					//	PeerIp:    "10.13.13.102",
					//})

					//key3, _ := wgtypes.GenerateKey()
					//fmt.Println(key3.String())
					//d.AddPeerToCluster(ctx, AddPeerAction{
					//	ClusterID: "hotspot-dev",
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
		//ClusterID  string `json:"cluster_id"`
		//Region     string `json:"region"`
		//Provider   string `json:"provider"`
		//NodesCount int    `json:"nodes_count"`

		//d.DeleteCluster(DeleteClusterAction{
		//	ClusterID: "hotspot-dev",
		//	Provider:  "do",
		//})

		// if err != nil {
		// 	panic(err)
		// }

		//d.CreateCluster(SetupClusterAction{
		//	ClusterID:  "hotspot-dev-2",
		//	Region:     "blr1",
		//	Provider:   "do",
		//	NodesCount: 4,
		//})

		//d.UpdateCluster(UpdateClusterAction{
		//	ClusterID:  "hotspot-dev",
		//	Region:     "blr1",
		//	Provider:   "do",
		//	NodesCount: 2,
		//})

		//key, _ := wgtypes.GenerateKey()
		//fmt.Println(key.String())
		//d.AddPeerToCluster(AddPeerAction{
		//	ClusterID: "hotspot-dev-2",
		//	PublicKey: key.PublicKey().String(),
		//	PeerIp:    "10.13.13.104",
		//})

		//d.DeletePeerFromCluster(DeletePeerAction{
		//	ClusterID: "hotspot-dev-2",
		//	PublicKey: "1uBcGZvNsNh7wlNzawDXiAExIfbgyFgbJqTwGRTmdiY=",
		//})
	}),
)
