package domain

type InfraClient interface {
	CreateCluster(action SetupClusterAction) (publicIp string, publicKey string, err error)
	UpdateCluster(action UpdateClusterAction) (e error)
	DeleteCluster(action DeleteClusterAction) (e error)
	AddPeer(action AddPeerAction) (e error)
	DeletePeer(action DeletePeerAction) (e error)
}

type InfraJobResponder interface {
	SendCreateClusterResponse(action SetupClusterResponse) error
	SendUpdateClusterResponse(action UpdateClusterResponse) error
	SendDeleteClusterResponse(action DeleteClusterResponse) error
	SendAddPeerResponse(action AddPeerResponse) error
	SendDeletePeerResponse(action DeletePeerResponse) error
}
