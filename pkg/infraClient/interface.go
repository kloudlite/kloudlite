package infraclient

type ProviderClient interface {
	NewNode() error
	DeleteNode() error
	UpdateNode() error
	AttachNode() error
}
