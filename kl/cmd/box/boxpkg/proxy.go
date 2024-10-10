package boxpkg

import (
	"crypto/md5"
	"fmt"

	"github.com/kloudlite/kl/pkg/egob"
)

type ProxyConfig struct {
	// TargetContainerId   string
	TargetContainerPath string
	ExposedPorts        []int
}

func (c *ProxyConfig) GetHash() string {
	b, err := egob.Marshal(c)
	if err != nil {
		return ""
	}

	hash := md5.New()
	hash.Write(b)

	return fmt.Sprintf("%x", hash.Sum(nil))
}

//func (c *client) SyncProxy(config ProxyConfig) error {
//	defer spinner.Client.UpdateMessage("updating port configuration")()
//
//	if err := c.ensureImage(constants.SocatImage); err != nil {
//		return functions.NewE(err, "failed to pull image")
//	}
//
//	existingProxies, err := c.cli.ContainerList(context.Background(), container.ListOptions{
//		Filters: filters.NewArgs(
//			dockerLabelFilter(CONT_MARK_KEY, "true"),
//			dockerLabelFilter("proxy", "true"),
//		),
//	})
//	if err != nil {
//		return functions.NewE(err, "failed to list containers")
//	}
//
//	if len(existingProxies) > 0 {
//		portHash := existingProxies[0].Labels["port-hash"]
//		if portHash == config.GetHash() && existingProxies[0].State == "running" {
//			return nil
//		}
//	}
//
//	targetContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
//		Filters: filters.NewArgs(
//			dockerLabelFilter(CONT_MARK_KEY, "true"),
//			dockerLabelFilter(CONT_PATH_KEY, config.TargetContainerPath),
//		),
//	})
//	if err != nil {
//		return functions.NewE(err, "failed to list containers")
//	}
//
//	if len(targetContainers) == 0 {
//		return nil
//	}
//
//	if len(existingProxies) > 0 {
//		if err := c.cli.ContainerKill(context.Background(), existingProxies[0].ID, "SIGKILL"); err != nil {
//			return functions.NewE(err, "failed to stop container")
//		}
//
//		if err = c.cli.ContainerRemove(context.Background(), existingProxies[0].ID, container.RemoveOptions{
//			Force: true,
//		}); err != nil {
//			return functions.NewE(err, "failed to remove container")
//		}
//	}
//
//	if len(config.ExposedPorts) == 0 {
//		return nil
//	}
//
//	targetContainer := targetContainers[0]
//	targetIpAddress := targetContainer.NetworkSettings.Networks["kloudlite"].IPAddress
//	socatCommand := ""
//	for _, port := range config.ExposedPorts {
//		socatCommand += fmt.Sprintf(`socat TCP-LISTEN:%d,fork TCP:%s:%d & `, port, targetIpAddress, port)
//		socatCommand += fmt.Sprintf(`socat UDP-RECVFROM:%d,fork UDP-SENDTO:%s:%d & `, port, targetIpAddress, port)
//	}
//	socatCommand += "tail -f /dev/null"
//
//	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
//		Image: constants.SocatImage,
//		Labels: map[string]string{
//			CONT_MARK_KEY: "true",
//			"proxy":       "true",
//			"port-hash":   config.GetHash(),
//		},
//		ExposedPorts: func() nat.PortSet {
//			ports := nat.PortSet{}
//			for _, port := range config.ExposedPorts {
//				ports[nat.Port(fmt.Sprintf("%d/tcp", port))] = struct{}{}
//				ports[nat.Port(fmt.Sprintf("%d/udp", port))] = struct{}{}
//			}
//			return ports
//		}(),
//		Entrypoint: []string{"sh", "-c", socatCommand},
//	}, &container.HostConfig{
//		PortBindings: func() nat.PortMap {
//			portBindings := nat.PortMap{}
//			for _, port := range config.ExposedPorts {
//				portBindings[nat.Port(fmt.Sprintf("%d/tcp", port))] = []nat.PortBinding{
//					{
//						HostPort: fmt.Sprintf("%d", port),
//					},
//				}
//				portBindings[nat.Port(fmt.Sprintf("%d/udp", port))] = []nat.PortBinding{
//					{
//						HostPort: fmt.Sprintf("%d", port),
//					},
//				}
//			}
//			return portBindings
//		}(),
//	}, &network.NetworkingConfig{
//		EndpointsConfig: map[string]*network.EndpointSettings{
//			"kloudlite": {},
//		},
//	}, nil, "")
//
//	if err != nil {
//		return functions.NewE(err, "failed to create container")
//	}
//
//	if err := c.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
//		return functions.NewE(err, "failed to start container")
//	}
//	return nil
//}
