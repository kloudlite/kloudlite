package domain

type InfraClient interface {
	CreateKubernetes(action SetupClusterAction) error
	UpdateKubernetes(action UpdateClusterAction) (e error)
	SetupCSI(clusterId string, provider string) error
	SetupOperator(clusterId string) error
	SetupMonitoring(clusterId string) error
	SetupIngress(clusterId string) error
	SetupWireguard(clusterId string) error
}
