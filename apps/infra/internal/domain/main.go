package domain

import (
	"fmt"
	"go.uber.org/fx"
)

type Domain interface {
	CreateCluster(action SetupClusterAction) error
}

type domain struct {
	tf TF
}

func (d *domain) CreateCluster(action SetupClusterAction) error {
	err := d.tf.CreateKubernetes(action)
	return err
}

func makeDomain(tf TF) Domain {
	return &domain{tf}
}

type TF interface {
	CreateKubernetes(action SetupClusterAction) error
	SetupCSI(clusterId string) error
	SetupOperator(clusterId string) error
	SetupMonitoring(clusterId string) error
	SetupIngress(clusterId string) error
	SetupWireguard(clusterId string) error
}

var Module = fx.Module("domain", fx.Provide(makeDomain), fx.Invoke(
	func(d Domain) {
		fmt.Println("domain created")
		d.CreateCluster(SetupClusterAction{
			ClusterID:    "sample-123",
			Name:         "mine",
			Region:       "blr1",
			Provider:     "do",
			MastersCount: 1,
			NodesCount:   2,
		})
	}),
)
