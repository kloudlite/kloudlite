package infraclient

type azureProvider struct {
	provider string
}

func NewAzureProvider() *azureProvider {
	return &azureProvider{
		provider: "",
	}
}
