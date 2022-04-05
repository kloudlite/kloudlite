package domain

type SetupClusterAction struct {
	ClusterID    string
	Region       string
	Provider     string
	MastersCount int
	NodesCount   int
}

type UpdateClusterAction struct {
	ClusterID    string
	Region       string
	Provider     string
	MastersCount int
	NodesCount   int
}
