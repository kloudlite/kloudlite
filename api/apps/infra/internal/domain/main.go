package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/messaging"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
	UpdateCluster(action UpdateClusterAction) error
}

type domain struct {
	tf              TF
	messageProducer messaging.Producer[messaging.Json]
	messageTopic    string
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	err := d.tf.CreateKubernetes(action)
	err = d.tf.SetupCSI(action.ClusterID, action.Provider)
	d.messageProducer.SendMessage(d.messageTopic, action.ClusterID, messaging.Json{
		"cluster_id": action.ClusterID,
		"status":     "live",
	})
	return err
}

func (d *domain) UpdateCluster(action UpdateClusterAction) error {
	err := d.tf.UpdateKubernetes(action)
	return err
}

func makeDomain(tf TF, messageProducer messaging.Producer[messaging.Json], messageTopic string) Domain {
	return &domain{
		tf:              tf,
		messageProducer: messageProducer,
		messageTopic:    messageTopic,
	}
}

type TF interface {
	CreateKubernetes(action SetupClusterAction) error
	UpdateKubernetes(action UpdateClusterAction) (e error)
	SetupCSI(clusterId string, provider string) error
	SetupOperator(clusterId string) error
	SetupMonitoring(clusterId string) error
	SetupIngress(clusterId string) error
	SetupWireguard(clusterId string) error
}

var Module = fx.Module("domain", fx.Provide(makeDomain))

// fx.Invoke(
// 	func(d Domain) {
// 		d.CreateCluster(SetupClusterAction{
// 			ClusterID:    "sample-121",
// 			Region:       "blr1",
// 			Provider:     "do",
// 			MastersCount: 1,
// 			NodesCount:   2,
// 		})
// 	}),
