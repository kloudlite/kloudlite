package domain

type InfraClient interface {
	CreateCluster(action SetupClusterAction) (publicIp string, publicKey string, err error)
	UpdateCluster(action UpdateClusterAction) (e error)
	DeleteCluster(action DeleteClusterAction) (e error)
	AddPeer(action AddPeerAction) (e error)
	DeletePeer(action DeletePeerAction) (e error)
}

type InfraJobResponder interface {
	SendCreateClusterResponse(action SetupClusterResponse)
	SendUpdateClusterResponse(action UpdateClusterResponse)
	SendDeleteClusterResponse(action DeleteClusterResponse)
	SendAddPeerResponse(action AddPeerResponse)
	SendDeletePeerResponse(action DeletePeerResponse)
}
