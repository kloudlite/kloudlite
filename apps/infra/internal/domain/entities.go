package domain

func getMasterNodesCount(totalNodes int) int {
	return 3
}

type SetupClusterAction struct {
	ClusterID  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
}

func (s *SetupClusterAction) MasterNodesCount() int {
	return getMasterNodesCount(s.NodesCount)
}

func (s *SetupClusterAction) AgentNodesCount() int {
	return s.NodesCount - getMasterNodesCount(s.NodesCount)
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
	PeerIp    string `json:"peer_ip"`
}

type AddPeerResponse struct {
	ClusterID string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
	Message   string `json:"message"`
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

func (s *UpdateClusterAction) MasterNodesCount() int {
	return getMasterNodesCount(s.NodesCount)
}

func (s *UpdateClusterAction) AgentNodesCount() int {
	return s.NodesCount
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
