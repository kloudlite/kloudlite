package domain

import "kloudlite.io/pkg/repos"

func getMasterNodesCount(totalNodes int) int {
	return 3
}

type SetupClusterAction struct {
	ClusterId  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
}

type SetupClusterResponse struct {
	ClusterId string `json:"cluster_id"`
	PublicIp  string `json:"public_ip"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
	Message   string `json:"message"`
}

type SetupAccountResponse struct {
	ClusterId   string `json:"cluster_id"`
	AccountId   string `json:"account_id"`
	Message     string `json:"message"`
	Done        bool   `json:"done"`
	WgPublicKey string `json:"wg_public_key"`
	WgPort      string `json:"wg_port"`
}

func (s *SetupClusterAction) MasterNodesCount() int {
	return getMasterNodesCount(s.NodesCount)
}

func (s *SetupClusterAction) AgentNodesCount() int {
	return s.NodesCount - getMasterNodesCount(s.NodesCount)
}

type AddPeerAction struct {
	ClusterId string `json:"cluster_id"`
	AccountId string `json:"account_id"`
	PublicKey string `json:"public_key"`
	PeerIp    string `json:"peer_ip"`
}

type AddPeerResponse struct {
	ClusterId     string `json:"cluster_id"`
	AccountId     string `json:"account_id"`
	PublicKey     string `json:"public_key"`
	Message       string `json:"message"`
	WireguardPort string `json:"wireguard_port"`
	Done          bool   `json:"done"`
}

type AddAccountAction struct {
	ClusterId repos.ID `json:"cluster_id"`
	Region    string   `json:"region"`
	Provider  string   `json:"provider"`
	AccountId string   `json:"account_id"`
	AccountIp string   `json:"account_ip"`
}

type AddAccountResponse struct {
	ClusterId string `json:"cluster_id"`
	AccountId string `json:"account_id"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
}

type DeletePeerAction struct {
	ClusterId string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
}

type DeletePeerResponse struct {
	ClusterId string `json:"cluster_id"`
	PublicKey string `json:"public_key"`
	Done      bool   `json:"done"`
}

type UpdateClusterAction struct {
	ClusterId  string `json:"cluster_id"`
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
	ClusterId  string `json:"cluster_id"`
	Region     string `json:"region"`
	Provider   string `json:"provider"`
	NodesCount int    `json:"nodes_count"`
	Done       bool   `json:"done"`
}

type DeleteClusterAction struct {
	ClusterId string `json:"cluster_id"`
	Provider  string `json:"provider"`
}

type DeleteClusterResponse struct {
	ClusterId string `json:"cluster_id"`
	Provider  string `json:"provider"`
	Done      bool   `json:"done"`
}
