package k3s

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

const (
	CONT_MARK_KEY = "kl.container"
)

//go:embed scripts/startup-script.sh.tmpl
var startupScript string

func (c *client) CreateClustersTeams(teamName string) error {
	defer spinner.Client.UpdateMessage("setting up cluster")()
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
		return fn.NewE(err, "failed to list containers")
	}
	stopAll := false
	if (existingContainers != nil) && (len(existingContainers) > 0) {
		for _, ec := range existingContainers {
			if ec.Labels["kl-team"] != teamName && ec.Labels["kl-k3s"] == "true" {
				fn.Log(text.Yellow(fmt.Sprintf("[#] another cluster is running for another team. do you want to stop it and start cluster for team %s? [Y/n] ", teamName)))
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
					return fn.NewE(err, "failed to stop container")
				}
				if err := c.c.ContainerRemove(context.Background(), ec.ID, container.RemoveOptions{}); err != nil {
					return fn.NewE(err, "failed to remove container")
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
		return fn.NewE(err, "failed to list containers")
	}

	if existingContainers != nil && len(existingContainers) > 0 {
		if existingContainers[0].State != "running" {
			if err := c.c.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return fn.NewE(err, "failed to start container")
			}
		}
		return nil
	}

	if err := c.EnsureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	clusterConfig, err := c.apic.GetClusterConfig(teamName)
	if err != nil {
		return fn.NewE(err)
	}

	configFolder, err := fileclient.GetConfigFolder()
	if err != nil {
		return fn.NewE(err)
	}

	createdConatiner := container.CreateResponse{}
	if flags.IsDev() {
		createdConatiner, err = c.c.ContainerCreate(context.Background(), &container.Config{
			Labels: map[string]string{
				CONT_MARK_KEY: "true",
				"kl-k3s":      "true",
				"kl-team":     teamName,
			},
			Image: constants.GetK3SImageName(),
			Cmd: []string{
				"server",
				"--disable", "traefik",
				"--node-name", clusterConfig.ClusterName,
			},
			ExposedPorts: nat.PortSet{
				"33820/udp": struct{}{},
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
				fmt.Sprintf("%s:/.cache/kl", configFolder),
			},
			PortBindings: map[nat.Port][]nat.PortBinding{
				"6443/tcp": {
					{
						HostPort: "6443",
					},
				},
				"33820/udp": {
					{
						HostPort: "33820",
					},
				},
			},
		}, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"kloudlite": {
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: constants.K3sServerIp,
					},
				},
			},
		}, nil, "")
		if err != nil {
			return fn.NewE(err, "failed to create container")
		}
	} else {
		createdConatiner, err = c.c.ContainerCreate(context.Background(), &container.Config{
			Labels: map[string]string{
				CONT_MARK_KEY: "true",
				"kl-k3s":      "true",
				"kl-team":     teamName,
			},
			Image: constants.GetK3SImageName(),
			Cmd: []string{
				"server",
				"--disable", "traefik",
				"--node-name", clusterConfig.ClusterName,
			},
			ExposedPorts: nat.PortSet{
				"33820/udp": struct{}{},
			},
		}, &container.HostConfig{
			Privileged:  true,
			NetworkMode: "kloudlite",
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
			Binds: []string{
				fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", clusterConfig.ClusterName),
				fmt.Sprintf("%s:/.cache/kl", configFolder),
			},
			PortBindings: map[nat.Port][]nat.PortBinding{
				"33820/udp": {
					{
						HostPort: "33820",
					},
				},
			},
		}, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"kloudlite": {
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: constants.K3sServerIp,
					},
				},
			},
		}, nil, "")
		if err != nil {
			return fn.NewE(err, "failed to create container")
		}
	}

	if err := c.c.ContainerStart(context.Background(), createdConatiner.ID, container.StartOptions{}); err != nil {
		return fn.NewE(err, "failed to start container")
	}

	script, err := c.generateConnectionScript(clusterConfig)
	if err != nil {
		return fn.NewE(err, "failed to generate connection script")
	}

	if err = c.runScriptInContainer(script); err != nil {
		return fn.NewE(err, "failed to run script")
	}

	start := time.Now()
	defer func() {
		if flags.IsVerbose {
			fn.Log(text.Yellow(fmt.Sprintf("Time taken to create cluster: %.2fs", time.Since(start).Seconds())))
		}
	}()
	return c.EnsureK3sServerIsReady()

}

func (c *client) generateConnectionScript(clusterConfig *fileclient.TeamClusterConfig) (string, error) {
	defer spinner.Client.UpdateMessage("generating connection script")()
	t := template.New("connectionScript")

	p, err := t.Parse(startupScript)
	if err != nil {
		return "", fn.NewE(err)
	}

	clusterConfig.Version = flags.Version
	if clusterConfig.Version == "" || clusterConfig.Version == "v1.0.0-nightly" {
		clusterConfig.Version = "v1.0.8-nightly"
	}

	teamName, err := c.fc.CurrentTeamName()
	if err != nil {
		return "", err
	}

	vpnTeamConfig, err := c.fc.GetVpnTeamConfig(teamName)
	if err != nil {
		return "", err
	}

	cc := struct {
		ClusterToken   string `json:"clusterToken"`
		ClusterName    string `json:"cluster"`
		InstallCommand fileclient.InstallCommand
		Installed      bool
		WGConfig       fileclient.WGConfig
		Version        string
		GatewayIP      string
		ClusterCIDR    string
		IpAddress      string
		ImageTag       string
		ImageBase      string
	}{
		ClusterToken:   clusterConfig.ClusterToken,
		ClusterName:    clusterConfig.ClusterName,
		InstallCommand: clusterConfig.InstallCommand,
		Installed:      clusterConfig.Installed,
		WGConfig:       clusterConfig.WGConfig,
		Version:        clusterConfig.Version,
		GatewayIP:      clusterConfig.GatewayIP,
		ClusterCIDR:    clusterConfig.ClusterCIDR,
		IpAddress:      vpnTeamConfig.IpAddress,
		ImageTag:       clusterConfig.Version,
		ImageBase:      flags.ImageBase,
	}

	b := new(bytes.Buffer)
	err = p.Execute(b, cc)
	if err != nil {
		return "", fn.NewE(err)
	}
	return b.String(), nil
}

func (c *client) DeletePods() error {
	defer spinner.Client.UpdateMessage("deleting pods")()
	script := `
kubectl taint nodes --all shutdown=true:NoExecute	
kubectl delete pods -n kloudlite --all --force --grace-period=0
kubectl delete pods -n kl-gateway --all --force --grace-period=0
`
	return c.runScriptInContainer(script)
}

func (c *client) EnsureK3sServerIsReady() error {
	defer spinner.Client.UpdateMessage("attaching your device to the team")()

	pingScript := `
	cat > /tmp/ping.sh <<EOF
	echo "Checking if 100.64.0.1 is reachable from kl-gateway pod..."
	while true; do
	  if timeout 1 kubectl exec -n kl-gateway deploy/default -c ip-manager -- ping -c 1 100.64.0.1 &> /dev/null; then
	    echo "100.64.0.1 is reachable!"
	    break
	  else
	    echo "waiting for cluster to be ready..."
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
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking image %s", imageName))()
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

func (c *client) StartAppInterceptService(ports []apiclient.AppPort, toStart bool) error {
	defer spinner.Client.UpdateMessage("starting intercept service")()
	if err := c.EnsureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	k3sTracker, err := c.fc.GetK3sTracker()
	if err != nil {
		return fn.Error("k3s server is not ready, please wait")
	}

	lastCheckedAt, err := time.Parse(time.RFC3339, k3sTracker.LastCheckedAt)
	if err != nil {
		return fn.Error("k3s server is not ready, please wait")
	}

	if time.Since(lastCheckedAt) > 3*time.Second {
		return fn.Error("k3s server is not ready, please wait")
	}

	spm := make(map[int]struct{}, len(k3sTracker.DeviceRouter.Service.Spec.Ports))
	for _, v := range k3sTracker.DeviceRouter.Service.Spec.Ports {
		spm[v.Port] = struct{}{}
	}

	newPorts := make([]fileclient.Port, 0, len(ports)+len(k3sTracker.DeviceRouter.Service.Spec.Ports))

	processed := make(map[int]struct{})

	for _, p := range ports {
		processed[p.DevicePort] = struct{}{}
		if _, ok := spm[p.DevicePort]; ok {

			if toStart {
				return fmt.Errorf("port, already occupied by another intercept")
			}
			continue
		}

		newPorts = append(newPorts,
			fileclient.Port{
				Name:       fmt.Sprintf("udp-%d", p.AppPort),
				Port:       p.DevicePort,
				Protocol:   "UDP",
				TargetPort: p.DevicePort,
			},
			fileclient.Port{
				Name:       fmt.Sprintf("tcp-%d", p.AppPort),
				Port:       p.DevicePort,
				Protocol:   "TCP",
				TargetPort: p.DevicePort,
			},
		)
	}

	for _, p := range k3sTracker.DeviceRouter.Service.Spec.Ports {
		if _, ok := processed[p.Port]; !ok && p.Name != "not-in-use" {
			newPorts = append(newPorts, fileclient.Port{
				Name:       p.Name,
				Port:       p.Port,
				Protocol:   p.Protocol,
				TargetPort: p.TargetPort,
			})
		}
	}

	if newPorts == nil || len(newPorts) == 0 {
		newPorts = append(newPorts, fileclient.Port{
			Name:       "not-in-use",
			Port:       59595,
			Protocol:   "TCP",
			TargetPort: 59595,
		})
	}

	portsJson, err := json.Marshal(newPorts)
	if err != nil {
		return err
	}

	tmpl := `
cat > /tmp/service-device-router.patch.json <<EOF
[
  {
		"op": "replace",
		"path": "/spec/ports",
		"value": %s
  }
]
EOF

kubectl patch svc/kl-device-router -n kl-local --type=json --patch-file /tmp/service-device-router.patch.json
`

	return c.runScriptInContainer(fmt.Sprintf(tmpl, portsJson))
}

func (c *client) RemoveAllIntercepts() error {
	defer spinner.Client.UpdateMessage("Cleaning up intercepts...")()
	script := `
cat > /tmp/service-device-router.patch.json <<EOF
[
	{
		"op": "replace",
		"path": "/spec/ports",
		"value": {"name": "not-in-use","port": 59595,"protocol": "TCP","targetPort": 59595}
	}
]
EOF

kubectl patch svc/kl-device-router -n kl-local --type=json --patch-file /tmp/service-device-router.patch.json
`
	return c.runScriptInContainer(script)
}

func (c *client) RestartWgProxyContainer() error {
	defer spinner.Client.UpdateMessage("restarting kloudlite-gateway")()
	script := `
kubectl delete pod/$(kubectl get pods -n kl-gateway | tail -n +2 | awk '{print $1}') -n kl-gateway --grace-period 0
`
	if err := c.runScriptInContainer(script); err != nil {
		return err
	}
	fn.Log(text.Yellow("It will usually take a minute for the cluster to come online"))

	return nil
}

func (c *client) runScriptInContainer(script string) error {
	defer spinner.Client.UpdateMessage("setting up cluster, this may take a while")()

	f := spinner.Client.UpdateMessage("ensuring k3s server is ready")
	existingContainers, err := c.c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})

	f()

	if err != nil {
		return fn.Errorf("failed to list containers: %w", err)
	}

	if len(existingContainers) == 0 {
		return fn.Errorf("no k3s container running locally")
	}

	f = spinner.Client.UpdateMessage("setting up cluster resources, please wait")
	execID, err := c.c.ContainerExecCreate(context.Background(), existingContainers[0].ID, container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", script},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fn.NewE(err, "failed to create exec instance for container")
	}

	resp, err := c.c.ContainerExecAttach(context.Background(), execID.ID, container.ExecStartOptions{
		Detach: false,
	})
	f()

	if err != nil {

		res, err2 := c.c.ContainerLogs(context.Background(), existingContainers[0].ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
		})

		if err2 != nil {
			return fn.NewE(err2, "failed to attach to exec instance")
		}

		defer res.Close()

		scanner := bufio.NewScanner(res)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 8 {
				line = line[8:]
			}
			fn.Log(text.Yellow("[stderr]"), line)
		}

		return fn.NewE(err, "failed to attach to exec instance")
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Reader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 {
			line = line[8:]
		}
		if flags.IsVerbose {
			fn.Log(text.Blue("[kube]"), line)
		}
	}

	//else {
	//output := new(strings.Builder)
	//if _, err := io.Copy(output, resp.Reader); err != nil {
	//	return fn.Errorf("failed to read exec output: %w", err)
	//}
	//}

	execInspect, err := c.c.ContainerExecInspect(context.Background(), execID.ID)
	if err != nil {
		return fn.NewE(err, "failed to inspect exec")
	}

	if execInspect.ExitCode != 0 {
		return fn.Errorf("script exited with error, exit code: %d", execInspect.ExitCode)
	}

	return nil
}

func (c *client) CheckK3sRunningLocally() (bool, error) {
	defer spinner.Client.UpdateMessage("checking k3s server")()
	existingContainers, err := c.c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})

	if err != nil {
		return false, fn.Errorf("failed to list containers: %w", err)
	}

	if len(existingContainers) == 0 {
		return false, fn.Errorf("no k3s container running locally")
	}
	return true, nil
}
