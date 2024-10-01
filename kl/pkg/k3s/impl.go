package k3s

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"io"
	"os"
	"strings"
	"text/template"
)

const (
	CONT_MARK_KEY = "kl.container"
)

//go:embed scripts/startup-script.sh.tmpl
var startupScript string

func (c *client) CreateClustersAccounts(accountName string) error {
	if err := c.EnsureImage(constants.GetK3SImageName()); err != nil {
		return fn.NewE(err)
	}
	existingContainers, err := c.c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}
	stopAll := false
	if (existingContainers != nil) && (len(existingContainers) > 0) {
		for _, ec := range existingContainers {
			if ec.Labels["kl-account"] != accountName && ec.Labels["kl-k3s"] == "true" {
				fn.Log(text.Yellow(fmt.Sprintf("[#] another cluster is running for another account. do you want to stop it and start cluster for account %s? [Y/n] ", accountName)))
				if !fn.Confirm("Y", "Y") {
					return nil
				}
				stopAll = true
				break
			}
		}
		if stopAll {
			for _, ec := range existingContainers {
				if err := c.c.ContainerStop(context.Background(), ec.ID, container.StopOptions{}); err != nil {
					return fn.Error("failed to stop container")
				}
				if err := c.c.ContainerRemove(context.Background(), ec.ID, container.RemoveOptions{}); err != nil {
					return fn.Error("failed to remove container")
				}
			}
		}
	}

	existingContainers, err = c.c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	if existingContainers != nil && len(existingContainers) > 0 {
		if existingContainers[0].State != "running" {
			if err := c.c.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return fn.Error("failed to start container")
			}
		}
		return nil
	}

	if err := c.EnsureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	clusterConfig, err := c.apic.GetClusterConfig(accountName)
	if err != nil {
		return fn.NewE(err)
	}

	createdConatiner, err := c.c.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"kl-k3s":      "true",
			"kl-account":  accountName,
		},
		Image: constants.GetK3SImageName(),
		Cmd: []string{
			"server",
			"--disable", "traefik",
			"--node-name", clusterConfig.ClusterName,
		},
		ExposedPorts: nat.PortSet{
			"51820/udp": struct{}{},
			"6443/tcp":  struct{}{},
		},
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: "kloudlite",
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", clusterConfig.ClusterName),
		},
		PortBindings: map[nat.Port][]nat.PortBinding{
			"6443/tcp": {
				{
					HostPort: "6443",
				},
			},
			"51820/udp": {
				{
					HostPort: "51820",
				},
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"kloudlite": {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: constants.HostIp,
				},
			},
		},
	}, nil, "")

	if err != nil {
		return fn.Error("failed to create container")
	}

	if err := c.c.ContainerStart(context.Background(), createdConatiner.ID, container.StartOptions{}); err != nil {
		return fn.Error("failed to start container")
	}

	script, err := generateConnectionScript(clusterConfig)
	if err != nil {
		return fn.Error("failed to generate connection script")
	}
	execConfig := container.ExecOptions{
		Cmd: []string{"sh", "-c", script},
	}

	resp, err := c.c.ContainerExecCreate(context.Background(), createdConatiner.ID, execConfig)
	if err != nil {
		return fn.Error("failed to create exec")
	}

	err = c.c.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return fn.Error("failed to start exec")
	}

	//if err := c.ensureK3sServerIsReady(); err != nil {
	//	return fn.NewE(err, "failed to ensure k3s server is ready")
	//}

	return nil
}

func generateConnectionScript(clusterConfig *fileclient.AccountClusterConfig) (string, error) {
	t := template.New("connectionScript")

	p, err := t.Parse(startupScript)
	if err != nil {
		return "", fn.NewE(err)
	}
	b := new(bytes.Buffer)
	err = p.Execute(b, clusterConfig)
	if err != nil {
		return "", fn.NewE(err)
	}
	return b.String(), nil
}

func (c *client) ensureK3sServerIsReady() error {
	defer spinner.Client.UpdateMessage("ensuring k3s server is ready")()

	pingScript := `
	cat > /tmp/ping.sh <<EOF
	echo "Checking if 100.64.0.1 is reachable from wg-proxy pod..."
	while true; do
	  if timeout 1 kubectl exec -n kl-gateway deploy/default -c ip-manager -- ping -c 1 100.64.0.1 &> /dev/null; then
	    echo "100.64.0.1 is reachable!"
	    break
	  else
	    echo "Cannot reach 100.64.0.1 from $POD_NAME, retrying in 3 seconds..."
	    sleep 0.5
	  fi
	done
EOF
chmod +x /tmp/ping.sh
/tmp/ping.sh
	`

	err := c.runScriptInContainer(pingScript)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) imageExists(imageName string) (bool, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageName)
	images, err := c.c.ImageList(context.Background(), image.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return false, err
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *client) EnsureImage(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking image %s", i))()

	imageExists, err := c.imageExists(i)
	if err == nil && imageExists {
		return nil
	}

	out, err := c.c.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return fn.NewE(err, fmt.Sprintf("failed to pull image %s", i))
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}

func (c *client) EnsureKloudliteNetwork() error {
	defer spinner.Client.UpdateMessage("ensuring kloudlite network")()

	networks, err := c.c.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", "kloudlite", "true")),
		),
	})
	if err != nil {
		return fn.NewE(err)
	}

	if len(networks) == 0 {
		_, err := c.c.NetworkCreate(context.Background(), "kloudlite", network.CreateOptions{
			Driver: "bridge",
			Labels: map[string]string{
				"kloudlite": "true",
			},
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet: "172.18.0.0/16",
					},
				},
			},

			Internal: false,
		})
		if err != nil {
			return fn.NewE(err)
		}
	}

	return nil
}

func (c *client) StartAppInterceptService(ports []apiclient.AppPort) error {
	defer spinner.Client.UpdateMessage("starting intercept service")()
	if err := c.EnsureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	if len(ports) == 0 {
		script := `
kubectl get svc/device-router -n wg-proxy
exit_code=$?

if [ $exit_code -eq 0 ]; then
  kubectl delete svc device-router -n wg-proxy
fi
`
		return c.runScriptInContainer(script)
	}

	tmpl := `
cat > /tmp/service-device-router.yml <<EOF
apiVersion: v1
kind: Service
metadata:
  name: device-router
  namespace: wg-proxy
  annotations:
    kloudlite.io/networking.proxy.to: "172.18.0.3"
spec:
  type: LoadBalancer
  ports:
  {{range .Ports}}
    - protocol: UDP
      port: {{.AppPort}}
      targetPort: {{if eq .DevicePort 0}}{{.AppPort}}{{else}}{{.DevicePort}}{{end}}
  {{end}}
EOF

kubectl apply -f /tmp/service-device-router.yml
`

	t, err := template.New("script").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Ports []apiclient.AppPort
	}{
		Ports: ports,
	}

	var script bytes.Buffer
	if err := t.Execute(&script, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return c.runScriptInContainer(script.String())
}

func (c *client) RestartWgProxyContainer() error {
	defer spinner.Client.UpdateMessage("restarting kloudlite-gateway")()
	script := `
kubectl delete pod $(kubectl get pods -n kl-gateway | grep -i default- | awk '{print $1}') -n kl-gateway
kubectl delete pod $(kubectl get pods -n wg-proxy | grep -i default- | awk '{print $1}') -n wg-proxy
`
	return c.runScriptInContainer(script)
}

func (c *client) runScriptInContainer(script string) error {
	existingContainers, err := c.c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(existingContainers) == 0 {
		return fmt.Errorf("no k3s container found")
	}

	execID, err := c.c.ContainerExecCreate(context.Background(), existingContainers[0].ID, container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", script},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec instance for container: %w", err)
	}

	resp, err := c.c.ContainerExecAttach(context.Background(), execID.ID, container.ExecStartOptions{
		Detach: false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	output := new(strings.Builder)
	if _, err := io.Copy(output, resp.Reader); err != nil {
		return fmt.Errorf("failed to read exec output: %w", err)
	}

	execInspect, err := c.c.ContainerExecInspect(context.Background(), execID.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec instance: %w", err)
	}

	if execInspect.ExitCode != 0 {
		return fmt.Errorf("script exited with error, exit code: %d", execInspect.ExitCode)
	}

	return nil
}
