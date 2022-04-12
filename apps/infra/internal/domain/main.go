package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
	UpdateCluster(action UpdateClusterAction) error
	DeleteCluster(action DeleteClusterAction) error
	AddPeerToCluster(action AddPeerAction) error
	DeletePeerFromCluster(action DeletePeerAction) error
}

type domain struct {
	infraCli InfraClient
	//messageProducer messaging.Producer[messaging.Json]
	messageTopic string
	jobResponder InfraJobResponder
}

// AddPeerToCluster implements Domain
func (d *domain) AddPeerToCluster(action AddPeerAction) error {

	return d.infraCli.AddPeer(action)
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
	// infraJobResp InfraJobResponder,
) Domain {
	return &domain{
		infraCli: infraCli,
		// jobResponder: infraJobResp,
		messageTopic: env.KafkaInfraActionResulTopic,
	}
}

type Env struct {
	KafkaInfraActionResulTopic string `env:"KAFKA_INFRA_ACTION_RESULT_TOPIC", required:"true"`
}

var Module = fx.Module("domain",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(makeDomain),
	fx.Invoke(func(d Domain) {
		err := d.DeleteCluster(DeleteClusterAction{
			ClusterID: "cluster-test-new",
			Provider:  "do",
		})

		if err != nil {
			panic(err)
		}

		// d.CreateCluster(SetupClusterAction{
		// 	ClusterID:  "cluster-test-new",
		// 	Region:     "blr1",
		// 	Provider:   "do",
		// 	NodesCount: 5,
		// })

		// d.UpdateCluster(UpdateClusterAction{
		// 	ClusterID:  "cluster-test-new",
		// 	Region:     "blr1",
		// 	Provider:   "do",
		// 	NodesCount: 2,
		// })

		// key, _ := wgtypes.GenerateKey()
		// d.AddPeerToCluster(AddPeerAction{
		// 	ClusterID: "cluster-test-new",
		// 	PublicKey: key.PublicKey().String(),
		// 	PeerIp:    "10.13.13.101",
		// })

		// d.DeletePeerFromCluster(DeletePeerAction{
		// 	ClusterID: "cluster-test-new",
		// 	PublicKey: "BmQvaNhCzW5CC7DuU7StkI5Z7/Ko+DMb/EQF9E3/2SE=",
		// })
	}),
)
