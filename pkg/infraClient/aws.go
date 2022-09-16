package infraclient

type doprovider struct {
	provider string
}

func NewAWSProvider() *doprovider {
	return &doprovider{
		provider: "",
	}
}
