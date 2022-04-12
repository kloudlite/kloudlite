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
	SendUpdateClusterResponse(action UpdateClusterAction)
	SendDeleteClusterResponse(action DeleteClusterAction)
	SendAddPeerResponse(action AddPeerAction)
	SendDeletePeerResponse(action DeletePeerAction)
}
