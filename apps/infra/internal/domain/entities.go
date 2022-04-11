package domain

type SetupClusterAction struct {
	ClusterID  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
}

type SetupClusterResponse struct {
	ClusterID string `json:"cluster_id"`
	PublicIp  string `json:"public_ip"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
	Message   string `json:"message"`
}

type AddPeerAction struct {
	ClusterID string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
}

type AddPeerResponse struct {
	ClusterID string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
}

type DeletePeerAction struct {
	ClusterID string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
}

type DeletePeerResponse struct {
	ClusterID string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
}

type UpdateClusterAction struct {
	ClusterID  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
}

type UpdateClusterResponse struct {
	ClusterID  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
	Done       bool   `json:"done"`
}

type DeleteClusterAction struct {
	ClusterID string `json:"cluster_id"`
	Provider  string `json:"provider"`
}

type DeleteClusterResponse struct {
	ClusterID string `json:"cluster_id"`
	Provider  string `json:"provider"`
	Done      bool   `json:"done"`
}
