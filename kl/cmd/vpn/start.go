package vpn

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"runtime"
)

const (
	wireguardImageName = "ghcr.io/kloudlite/hub/wireguard:latest"
	CONT_MARK_KEY      = "kl.container"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn",
	Long:  `start vpn`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := startVPN(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startVPN() error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	wgConfig, err := fc.GetHostWgConfig()
	if err != nil {
		return fn.NewE(err)
	}

	if runtime.GOOS != "linux" {
		fn.Log(text.Green("add below config to your wireguard client and start vpn"))
		fn.Log(wgConfig)
		return nil
	}

	//if err := fc.SetWGConfig(wgConfig); err != nil {
	//	return fn.NewE(err)
	//}

	//current, err := user.Current()
	//if err != nil {
	//	return fn.NewE(err)
	//}
	//if current.Uid != "0" {
	//	return fn.Errorf("root permission required to start vpn")
	//}

	//var errBuf strings.Builder
	//cmd := exec.Command("wg-quick", "up", "kl")
	//cmd.Stderr = &errBuf
	//
	//err = cmd.Run()
	//if err != nil {
	//	return fn.Errorf(errBuf.String())
	//}

	if err := startWireguard(wgConfig, false); err != nil {
		return err
	}

	fn.Log(text.Green("kloudlite vpn has been started"))

	return nil
}

func startWireguard(wgConfig string, stopWg bool) error {
	k3sClient, err := k3s.NewClient()
	if err != nil {
		return err
	}

	if err := k3sClient.EnsureImage(wireguardImageName); err != nil {
		return err
	}

	dockerClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	existingContainers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-wg", "true")),
		),
	})
	if err != nil {
		return err
	}

	if len(existingContainers) > 0 && existingContainers[0].Labels["kl-wg"] == "true" {
		execConfig := container.ExecOptions{
			Cmd: []string{"wg-quick", "down", "kl-host-wg"},
		}

		execID, err := dockerClient.ContainerExecCreate(context.Background(), existingContainers[0].ID, execConfig)
		if err != nil {
			return err
		}

		if err = dockerClient.ContainerExecStart(context.Background(), execID.ID, container.ExecStartOptions{}); err != nil {
			return err
		}

		if err = dockerClient.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{
			Signal: "SIGKILL",
		}); err != nil {
			return err
		}
		if err = dockerClient.ContainerRemove(context.Background(), existingContainers[0].ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
		if stopWg {
			return nil
		}
	}

	createdContainer, err := dockerClient.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"kl-wg":       "true",
		},
		Image: wireguardImageName,
		Cmd: []string{
			"sh",
			"-c",
			`
wg-quick down kl-host-wg || echo 'starting wireguard'

# Create the WireGuard config file
cat <<EOF > /etc/wireguard/kl-host-wg.conf
` + wgConfig + `
EOF

# Start WireGuard with the new config
wg-quick up kl-host-wg

echo "wireguard server is running"
tail -f /dev/null &
pid=$!
trap "kill $pid" SIGINT SIGTERM
wait $pid
`,
		},
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			"kl-wg-cache:/var/lib/rancher/k3s",
		},
	}, &network.NetworkingConfig{}, nil, "")

	if err != nil {
		return fn.NewE(err, "failed to create container")
	}

	if err := dockerClient.ContainerStart(context.Background(), createdContainer.ID, container.StartOptions{}); err != nil {
		return fn.NewE(err, "failed to start container")
	}

	return nil
}
