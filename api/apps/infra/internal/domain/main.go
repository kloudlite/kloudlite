package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/messaging"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
	UpdateCluster(action UpdateClusterAction) error
	DeleteCluster(action DeleteClusterAction) error
	AddPeerToCluster(action AddPeerAction) error
	DeletePeerFromCluster(action DeletePeerAction) error
}

type domain struct {
	infraCli     InfraClient
	messageTopic string
	//jobResponder InfraJobResponder
}

// AddPeerToCluster implements Domain
func (d *domain) AddPeerToCluster(action AddPeerAction) error {
	err := d.infraCli.AddPeer(action)
	if err != nil {
		//d.jobResponder.SendAddPeerResponse(AddPeerResponse{
		//	ClusterID: action.ClusterID,
		//	PublicKey: action.PublicKey,
		//	Message:   err.Error(),
		//	Done:      false,
		//})
		return err
	}
	//d.jobResponder.SendAddPeerResponse(AddPeerResponse{
	//	ClusterID: action.ClusterID,
	//	PublicKey: action.PublicKey,
	//	Message:   "Peer added",
	//	Done:      true,
	//})
	return err
}

// DeletePeerFromCluster implements Domain
func (d *domain) DeletePeerFromCluster(action DeletePeerAction) error {
	return d.infraCli.DeletePeer(action)
}

// DeleteCluster implements Domain
func (d *domain) DeleteCluster(action DeleteClusterAction) error {
	return d.infraCli.DeleteCluster(action)
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	_, _, err := d.infraCli.CreateCluster(action)
	// publicIp, publicKey, err := d.infraCli.CreateCluster(action)
	// if err != nil {
	// 	d.jobResponder.SendCreateClusterResponse(SetupClusterResponse{
	// 		ClusterID: action.ClusterID,
	// 		PublicIp:  publicIp,
	// 		PublicKey: publicKey,
	// 		Done:      false,
	// 		Message:   err.Error(),
	// 	})
	// }
	// d.jobResponder.SendCreateClusterResponse(SetupClusterResponse{
	// 	ClusterID: action.ClusterID,
	// 	PublicIp:  publicIp,
	// 	PublicKey: publicKey,
	// 	Done:      true,
	// })
	return err
}

func (d *domain) UpdateCluster(action UpdateClusterAction) error {
	_, _, err := d.infraCli.CreateCluster(SetupClusterAction{
		ClusterID:  action.ClusterID,
		Region:     action.Region,
		Provider:   action.Provider,
		NodesCount: action.NodesCount,
	})
	return err
}
func makeDomain(
	env *Env,
	infraCli InfraClient,
	//infraJobResp InfraJobResponder,
) Domain {
	return &domain{
		infraCli: infraCli,
		//jobResponder: infraJobResp,
		messageTopic: env.KafkaInfraActionResulTopic,
	}
}

type Env struct {
	KafkaInfraActionResulTopic string `env:"KAFKA_INFRA_ACTION_RESULT_TOPIC", required:"true"`
}

var Module = fx.Module("domain",
	fx.Provide(config.LoadEnv[*Env]()),
	fx.Provide(makeDomain),
	fx.Invoke(func(d Domain, p messaging.Producer[messaging.Json], lifecycle fx.Lifecycle) {

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
		//	PeerIp:    "10.13.13.103",
		//})

		//d.DeletePeerFromCluster(DeletePeerAction{
		//	ClusterID: "hotspot-dev-2",
		//	PublicKey: "1uBcGZvNsNh7wlNzawDXiAExIfbgyFgbJqTwGRTmdiY=",
		//})
	}),
)
