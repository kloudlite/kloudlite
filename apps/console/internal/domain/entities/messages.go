package entities

import "kloudlite.io/pkg/repos"

type SetupClusterAction struct {
	ClusterID  repos.ID `json:"cluster_id"`
	Region     string   `json:"region"`
	Provider   string   `json:"provider"`
	NodesCount int      `json:"nodes_count"`
}

type SetupClusterResponse struct {
	ClusterID repos.ID `json:"cluster_id"`
	PublicIp  string   `json:"public_ip"`
	PublicKey string   `json:"public_key"`
	Done      bool     `json:"done"`
	Message   string   `json:"message"`
}

type AddPeerAction struct {
	ClusterID repos.ID `json:"cluster_id"`
	PublicKey string   `json:"public_key"`
	PeerIp    string   `json:"peer_ip"`
}

type AddPeerResponse struct {
	ClusterID repos.ID `json:"cluster_id"`
	DeviceID  repos.ID `json:"device_id"`
	PublicKey string   `json:"public_key"`
	Message   string   `json:"message"`
	Done      bool     `json:"done"`
}

type DeletePeerAction struct {
	ClusterID repos.ID `json:"cluster_id"`
	DeviceID  repos.ID `json:"device_id"`
	PublicKey string   `json:"public_key"`
}

type DeletePeerResponse struct {
	ClusterID repos.ID `json:"cluster_id"`
	DeviceID  repos.ID `json:"device_id"`
	PublicKey string   `json:"public_key"`
	Done      bool     `json:"done"`
}

type UpdateClusterAction struct {
	ClusterID  repos.ID `json:"cluster_id"`
	Region     string   `json:"region"`
	Provider   string   `json:"provider"`
	NodesCount int      `json:"nodes_count"`
}

type UpdateClusterResponse struct {
	ClusterID  repos.ID `json:"cluster_id"`
	Region     string   `json:"region"`
	Provider   string   `json:"provider"`
	NodesCount int      `json:"nodes_count"`
	Done       bool     `json:"done"`
}

type DeleteClusterAction struct {
	ClusterID repos.ID `json:"cluster_id"`
	Provider  string   `json:"provider"`
}

type DeleteClusterResponse struct {
	ClusterID repos.ID `json:"cluster_id"`
	Provider  string   `json:"provider"`
	Done      bool     `json:"done"`
}
