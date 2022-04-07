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
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	err := d.infraCli.CreateKubernetes(action)
	//err = d.infraCli.SetupCSI(action.ClusterID, action.Provider)
	//err = d.infraCli.SetupIngress(action.ClusterID)
	//fmt.Println(err)
	return err
}

func (d *domain) UpdateCluster(action UpdateClusterAction) error {
	err := d.infraCli.UpdateKubernetes(action)
	return err
}
func makeDomain(
	env *Env,
	infraCli InfraClient,
	//messageProducer messaging.Producer[messaging.Json],
) Domain {
	return &domain{
		infraCli: infraCli,
		//messageProducer: messageProducer,
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
			ClusterID:    "cluster-ltujdwouztzgfeg",
			Region:       "blr1",
			Provider:     "do",
			MastersCount: 1, NodesCount: 2,
		})
	}),
)
