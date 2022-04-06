package domain

import (
	"go.uber.org/fx"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
	UpdateCluster(action UpdateClusterAction) error
}

type domain struct {
	tf TF
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	err := d.tf.CreateKubernetes(action)
	err = d.tf.SetupCSI(action.ClusterID, action.Provider)
	return err
}

func (d *domain) UpdateCluster(action UpdateClusterAction) error {
	err := d.tf.UpdateKubernetes(action)
	return err
}

func makeDomain(tf TF) Domain {
	return &domain{tf}
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

var Module = fx.Module("domain", fx.Provide(makeDomain), fx.Invoke(
	func(d Domain) {
		d.CreateCluster(SetupClusterAction{
			ClusterID:    "sample-121",
			Region:       "blr1",
			Provider:     "do",
			MastersCount: 1,
			NodesCount:   2,
		})
	}),
)
