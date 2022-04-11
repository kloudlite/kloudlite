package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
	UpdateCluster(action UpdateClusterAction) error
}

type domain struct {
	infraCli InfraClient
	//messageProducer messaging.Producer[messaging.Json]
	messageTopic string
	jobResponder InfraJobResponder
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	publicIp, publicKey, err := d.infraCli.CreateCluster(action)
	if err != nil {
		d.jobResponder.SendCreateClusterResponse(SetupClusterResponse{
			ClusterID: action.ClusterID,
			PublicIp:  publicIp,
			PublicKey: publicKey,
			Done:      false,
			Message:   err.Error(),
		})
	}
	d.jobResponder.SendCreateClusterResponse(SetupClusterResponse{
		ClusterID: action.ClusterID,
		PublicIp:  publicIp,
		PublicKey: publicKey,
		Done:      true,
	})
	return err
}

func (d *domain) UpdateCluster(action UpdateClusterAction) error {
	err := d.infraCli.UpdateCluster(action)
	return err
}
func makeDomain(
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
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(makeDomain),
	fx.Invoke(func(d Domain) {
		d.CreateCluster(SetupClusterAction{
			ClusterID:  "cluster-test-new",
			Region:     "blr1",
			Provider:   "do",
			NodesCount: 3,
		})
	}),
)
