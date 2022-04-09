package domain

type InfraClient interface {
	CreateKubernetes(action SetupClusterAction) error
	UpdateKubernetes(action UpdateClusterAction) (e error)
}
