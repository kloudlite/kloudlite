package domain

type SetupClusterAction struct {
	ClusterID    string `json:"cluster_id"`
	Region       string `json:"region"`
	Provider     string `json:"provider"`
	// MastersCount int    `json:"masters_count"`
	NodesCount   int    `json:"nodes_count"`
}

type UpdateClusterAction struct {
	ClusterID    string `json:"cluster_id"`
	Region       string `json:"region"`
	Provider     string `json:"provider"`
	MastersCount int    `json:"masters_count"`
	NodesCount   int    `json:"nodes_count"`
}
