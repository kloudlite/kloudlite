package domain

type SetupClusterAction struct {
	ClusterID    string
	Name         string
	Region       string
	Provider     string
	MastersCount int
	NodesCount   int
}
