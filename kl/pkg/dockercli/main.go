package dockercli

import "github.com/docker/docker/client"

func GetClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return cli, nil
}
