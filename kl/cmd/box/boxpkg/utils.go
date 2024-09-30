package boxpkg

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"text/template"
	"time"

	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func dockerLabelFilter(key, value string) filters.KeyValuePair {
	return filters.Arg("label", fmt.Sprintf("%s=%s", key, value))
}

const (
	NO_RUNNING_CONTAINERS = "no container running"
)

var UserCanceled = fmt.Errorf("user canceled")

type Container struct {
	Name string
	Path string
}

func (c *client) SetCwd(cwd string) {
	c.cwd = cwd
}

func (c *client) ensurePublicKey() error {
	sshPath := path.Join(xdg.Home, ".ssh")
	if _, err := os.Stat(path.Join(sshPath, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path.Join(sshPath, "id_rsa"), "-N", "")
		err := cmd.Run()
		if err != nil {
			return fn.NewE(err)
		}
	}

	return nil
}
func (c *client) ensureCacheExist() error {

	caches := []string{"kl-nix-store", "kl-home-cache", "kl-k3s-cache"}

	for _, cache := range caches {
		vlist, err := c.cli.VolumeList(c.cmd.Context(), volume.ListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("%s=true", cache),
			}),
		})
		if err != nil {
			return fn.NewE(err)
		}

		if len(vlist.Volumes) == 0 {
			if _, err := c.cli.VolumeCreate(c.cmd.Context(), volume.CreateOptions{
				Labels: map[string]string{
					cache: "true",
				},
				Name: cache,
			}); err != nil {
				return fn.NewE(err)
			}
		}

	}

	return nil
}

func GetDockerHostIp() (string, error) {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fn.NewE(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.To4().String(), nil
}

func (c *client) imageExists(imageName string) (bool, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageName)
	images, err := c.cli.ImageList(context.Background(), image.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *client) ensureImage(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking image %s", i))()

	if imageExists, err := c.imageExists(i); err == nil && imageExists {
		return nil
	}

	out, err := c.cli.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return fn.NewE(err, fmt.Sprintf("failed to pull image %s", i))
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}

func (c *client) restartContainer(path string) error {
	defer spinner.Client.UpdateMessage("restart container")()

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
			dockerLabelFilter(CONT_PATH_KEY, path),
		),
		All: true,
	})
	if len(existingContainers) == 0 {
		return nil
	}

	if err != nil {
		return fn.NewE(err, "failed to list containers")
	}

	timeOut := 0
	if err := c.cli.ContainerRestart(context.Background(), existingContainers[0].ID, container.StopOptions{
		Signal:  "SIGKILL",
		Timeout: &timeOut,
	}); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (c *client) startContainer(klconfHash string) (string, error) {

	err := c.stopOtherContainers()
	if err != nil {
		return "", fn.NewE(err)
	}

	if err := c.ensurePublicKey(); err != nil {
		return "", fn.NewE(err)
	}

	if err := c.ensureCacheExist(); err != nil {
		return "", fn.NewE(err)
	}

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
			dockerLabelFilter(CONT_PATH_KEY, c.cwd),
		),
	})

	if err != nil {
		return "", fn.Error("failed to list containers")
	}

	if len(existingContainers) > 0 {
		if existingContainers[0].State != "running" {
			if err := c.cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return "", fn.NewE(err)
			}

			sshPortStr, ok := existingContainers[0].Labels[SSH_PORT_KEY]
			if !ok {
				return "", fn.Error("failed to get ssh port")
			}

			sshPort, err := strconv.Atoi(sshPortStr)
			if err != nil {
				return "", fn.NewE(err)
			}

			if err := c.waithForSshReady(sshPort, existingContainers[0].ID); err != nil {
				return "", fn.NewE(err)
			}
		}

		return existingContainers[0].ID, nil
	}

	sshPort, err := c.getFreePort()
	if err != nil {
		return "", fn.Error("failed to get free port")
	}

	vmounts, err := c.generateMounts()
	if err != nil {
		return "", fn.NewE(err)
	}

	boxhashFileName, err := hashctrl.BoxHashFileName(c.cwd)
	if err != nil {
		return "", fn.NewE(err)
	}

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Image: GetImageName(),
		Labels: map[string]string{
			CONT_MARK_KEY:           "true",
			CONT_WORKSPACE_MARK_KEY: "true",
			CONT_PATH_KEY:           c.cwd,
			SSH_PORT_KEY:            fmt.Sprintf("%d", sshPort),
			KLCONFIG_HASH_KEY:       klconfHash,
		},
		Env: []string{
			fmt.Sprintf("KL_HASH_FILE=/.cache/kl/box-hash/%s", boxhashFileName),
			fmt.Sprintf("SSH_PORT=%d", sshPort),
			fmt.Sprintf("KL_WORKSPACE=%s", c.cwd),
			"KLCONFIG_PATH=/workspace/kl.yml",
			fmt.Sprintf("KL_DNS=%s", constants.KLDNS),
			fmt.Sprintf("KL_BASE_URL=%s", constants.BaseURL),
		},
		Hostname:     "box",
		ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", sshPort)): {}},
	}, &container.HostConfig{
		ExtraHosts: []string{
			fmt.Sprintf("k3s-cluster.local:%s", constants.HostIp),
		},
		Privileged:  true,
		NetworkMode: "kloudlite",
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", sshPort)): []nat.PortBinding{
				{
					HostPort: fmt.Sprintf("%d", sshPort),
				},
			},
		},
		Binds: func() []string {
			binds := make([]string, 0, len(vmounts))
			for _, m := range vmounts {
				binds = append(binds, fmt.Sprintf("%s:%s:z", m.Source, m.Target))
			}
			binds = append(binds, fmt.Sprintf("%s:/workspace:z", c.cwd))
			return binds
		}(),
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"kloudlite": {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: constants.InterceptWorkspaceServiceIp,
				},
			},
		},
	}, nil, fmt.Sprintf("kl-%s", boxhashFileName[len(boxhashFileName)-8:]))
	if err != nil {
		return "", fn.NewE(err, "failed to create container")
	}

	if err := c.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", fn.NewE(err, "failed to start container")
	}

	if err := c.waithForSshReady(sshPort, resp.ID); err != nil {
		return "", fn.NewE(err)
	}

	return resp.ID, nil
}

func (c *client) stopOtherContainers() error {
	defer spinner.Client.UpdateMessage("stopping other containers")()

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
		),
	})
	if err != nil {
		return fn.NewE(err)
	}

	for _, d := range existingContainers {
		if d.Labels[CONT_PATH_KEY] != c.cwd {
			spinner.Client.Stop()
			fn.Logf(text.Yellow(fmt.Sprintf("[#] another workspace is active and running at %s. this action will stop that workspace and terminate all the processes running in the that container. do you want to proceed? [Y/n]", d.Labels[CONT_PATH_KEY])))
			if !fn.Confirm("y", "y") {
				return fn.NewE(fn.NewE(UserCanceled))
			}

			if err := c.stopContainer(d.Labels[CONT_PATH_KEY]); err != nil {
				return fn.NewE(err)
			}
		} else {
			if d.State != "running" {
				if err := c.stopContainer(d.Labels[CONT_PATH_KEY]); err != nil {
					return fn.NewE(err)
				}
			}
		}
	}

	return nil
}

func (c *client) stopContainer(_ string) error {
	defer spinner.Client.UpdateMessage("stopping container")()

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			// dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
			// dockerLabelFilter(CONT_PATH_KEY, path),
		),
		All: true,
	})
	if err != nil {
		return fn.NewE(err, "failed to list containers")
	}

	for _, c2 := range existingContainers {
		timeOut := 0
		if err := c.cli.ContainerStop(context.Background(), c2.ID, container.StopOptions{
			Timeout: &timeOut,
		}); err != nil {
			return fn.NewE(err)
		}
		if c2.Labels["kl-k3s"] == "true" {
			continue
		}
		if err := c.cli.ContainerRemove(context.Background(), c2.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return fn.NewE(err)
		}

	}

	return nil
}

func (c *client) getFreePort() (int, error) {

	if c.env.SSHPort != 0 {
		return c.env.SSHPort, nil
	}

	var resp int
	data, err := fileclient.GetExtraData()
	if err != nil {
		return 0, fn.NewE(err)
	}
	for {
		port := rand.Intn(65535-1024) + 1025
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			resp = port
			for _, v := range data.SelectedEnvs {
				if v.SSHPort == resp {
					continue
				}
			}
			c.env.SSHPort = resp
			break
		}
	}
	return resp, nil
}

func (c *client) generateMounts() ([]mount.Mount, error) {
	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return nil, fn.NewE(err)
	}

	if err := userOwn(td); err != nil {
		return nil, fn.NewE(err)
	}

	userHomeDir, err := fileclient.GetUserHomeDir()
	if err != nil {
		return nil, fn.NewE(err)
	}

	sshPath := path.Join(userHomeDir, ".ssh", "id_rsa.pub")
	rsaPath := path.Join(userHomeDir, ".ssh", "id_rsa")

	akByte, err := os.ReadFile(sshPath)
	if err != nil {
		return nil, fn.NewE(err)
	}

	ak := string(akByte)

	akTmpPath := path.Join(td, "authorized_keys")

	//gitConfigPath := path.Join(userHomeDir, ".gitconfig")

	akByte, err = os.ReadFile(path.Join(userHomeDir, ".ssh", "authorized_keys"))
	if err == nil {
		ak += fmt.Sprint("\n", string(akByte))
	}

	// for wsl
	if err := func() error {
		if runtime.GOOS != constants.RuntimeLinux {
			return nil
		}

		usersPath := "/mnt/c/Users"
		_, err := os.Stat(usersPath)
		if err != nil {
			return nil
		}

		de, err := os.ReadDir(usersPath)
		if err != nil {
			return fn.NewE(err)
		}

		for _, de2 := range de {
			pth := path.Join(usersPath, de2.Name(), ".ssh", "id_rsa.pub")
			if _, err := os.Stat(pth); err != nil {
				continue
			}

			b, err := os.ReadFile(pth)
			if err != nil {
				return fn.NewE(err)
			}

			ak += fmt.Sprint("\n", string(b))
		}

		return nil
	}(); err != nil {
		return nil, fn.NewE(err)
	}

	if err := writeOnUserScope(akTmpPath, []byte(ak)); err != nil {
		return nil, fn.NewE(err)
	}

	configFolder, err := fileclient.GetConfigFolder()
	if err != nil {
		return nil, fn.NewE(err)
	}

	volumes := []mount.Mount{
		{Type: mount.TypeBind, Source: akTmpPath, Target: "/tmp/ssh2/authorized_keys", ReadOnly: true},
		{Type: mount.TypeBind, Source: sshPath, Target: "/tmp/ssh2/id_rsa.pub", ReadOnly: true},
		{Type: mount.TypeBind, Source: rsaPath, Target: "/tmp/ssh2/id_rsa", ReadOnly: true},
		{Type: mount.TypeVolume, Source: "kl-home-cache", Target: "/home"},
		//{Type: mount.TypeBind, Source: gitConfigPath, Target: "/tmp/gitconfig/.gitconfig", ReadOnly: true},
		{Type: mount.TypeVolume, Source: "kl-nix-store", Target: "/nix"},
		{Type: mount.TypeBind, Source: configFolder, Target: "/.cache/kl"},
	}
	//_, err = os.Stat(gitConfigPath)
	//if err == nil {
	//	volumes = append(volumes, mount.Mount{Type: mount.TypeBind, Source: gitConfigPath, Target: "/tmp/gitconfig/.gitconfig", ReadOnly: true})
	//}

	dockerSock := "/var/run/docker.sock"
	// if runtime.GOOS == constants.RuntimeWindows {
	// 	dockerSock = "\\\\.\\pipe\\docker_engine"
	// }

	volumes = append(volumes,
		mount.Mount{Type: mount.TypeVolume, Source: dockerSock, Target: "/var/run/docker.sock"},
	)

	return volumes, nil
}

func (c *client) SyncVpn(wg string) error {
	defer spinner.Client.UpdateMessage("validating vpn configuration")()

	err := c.ensureImage(constants.GetWireguardImageName())
	if err != nil {
		return err
	}

	existingVPN, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter("wg", "true"),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}
	md5sum := md5.Sum([]byte(wg))
	if len(existingVPN) > 0 {
		if existingVPN[0].Labels["wgsum"] == fmt.Sprintf("%x", md5sum[:]) {
			if existingVPN[0].State != "running" {
				err := c.cli.ContainerStart(context.Background(), existingVPN[0].ID, container.StartOptions{})
				if err != nil {
					return fn.Error("failed to start container")
				}
			}
			return nil
		}
		err := c.cli.ContainerStop(context.Background(), existingVPN[0].ID, container.StopOptions{
			Signal: "SIGKILL",
		})
		if err != nil {
			return fn.Error("failed to stop container")
		}
		err = c.cli.ContainerRemove(context.Background(), existingVPN[0].ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			return fn.Error("failed to remove container")
		}
	}
	script := fmt.Sprintf("echo %s | base64 -d > /etc/wireguard/wg0.conf && (wg-quick down wg0 || echo done) && wg-quick up wg0 && tail -f /dev/null", wg)

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"wg":          "true",
			"wgsum":       fmt.Sprintf("%x", md5sum[:]),
		},
		Image: constants.GetWireguardImageName(),
		Cmd:   []string{"sh", "-c", script},
	}, &container.HostConfig{
		CapAdd:      []string{"NET_ADMIN"},
		NetworkMode: "host",
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return fn.Error("failed to create container")
	}
	err = c.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return fn.Error("failed to start container")
	}
	return nil
}

func GenerateConnectionScript(clusterConfig *fileclient.AccountClusterConfig) (string, error) {
	t := template.New("connectionScript")
	p, err := t.Parse(`
echo "checking whether k3s server is accepting connections"
while true; do
  lines=$(kubectl get nodes | wc -l)
  if [ "$lines" -lt 2 ]; then
	echo "k3s server is not accepting connections yet, retrying in 1s ..."
	sleep 1
	continue
  fi
  echo "successful, k3s server is now accepting connections"
  break
done
kubectl apply -f {{.InstallCommand.CRDsURL}} --server-side
kubectl create ns kloudlite
cat <<EOF | kubectl apply -f -
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: kloudlite
  namespace: kube-system
spec:
  repo: {{.InstallCommand.ChartRepo}}
  chart: kloudlite-agent
  version: {{.InstallCommand.ChartVersion}}
  targetNamespace: kloudlite
  valuesContent: |-
    accountName: {{.InstallCommand.HelmValues.AccountName}}
    clusterName: {{.InstallCommand.HelmValues.ClusterName}}
    clusterToken: {{.InstallCommand.HelmValues.ClusterToken}}
    kloudliteDNSSuffix: {{.InstallCommand.HelmValues.KloudliteDNSSuffix}}
    messageOfficeGRPCAddr: {{.InstallCommand.HelmValues.MessageOfficeGRPCAddr}}
EOF
`)
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

func (c *client) ConnectClusterToAccount(cConfig *fileclient.AccountClusterConfig) error {
	existingContainer, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter("kl-k3s", "true"),
			dockerLabelFilter("kl-account", c.klfile.AccountName),
		),
	})
	if err != nil {
		return fn.Error("k3s container should exist")
	}
	if len(existingContainer) == 0 {
		return fn.Error("no k3s container found")
	}
	script, err := GenerateConnectionScript(cConfig)
	if err != nil {
		return fn.Error("failed to generate connection script")
	}
	execConfig := container.ExecOptions{
		Cmd: []string{"sh", "-c", script},
	}
	resp, err := c.cli.ContainerExecCreate(context.Background(), existingContainer[0].ID, execConfig)
	if err != nil {
		return fn.Error("failed to create exec")
	}

	err = c.cli.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return fn.Error("failed to start exec")
	}

	for {
		inspectResp, err := c.cli.ContainerExecInspect(context.Background(), resp.ID)
		if err != nil {
			return fn.Error("failed to inspect exec")
		}
		if !inspectResp.Running {
			if inspectResp.ExitCode == 0 {
				return nil
			} else {
				return fn.Error(fmt.Sprintf("command failed with exit code %d", inspectResp.ExitCode))
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *client) EnsureK3SCluster(account string) error {
	err := c.ensureImage(constants.GetK3SImageName())
	if err != nil {
		return err
	}
	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter("kl-k3s", "true"),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	if existingContainers != nil && (len(existingContainers) > 0) {
		if existingContainers[0].Labels["kl-account"] != account {
			err := c.cli.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{})
			if err != nil {
				return fn.Error("failed to stop container")
			}
			err = c.cli.ContainerRemove(context.Background(), existingContainers[0].ID, container.RemoveOptions{
				Force: true,
			})
			if err != nil {
				return fn.Error("failed to remove container")
			}
		} else {
			if existingContainers[0].State != "running" {
				err := c.cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{})
				if err != nil {
					return fn.Error("failed to start container")
				}
			}
			return nil
		}
	}

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"kl-k3s":      "true",
			"kl-account":  c.klfile.AccountName,
		},
		Image: constants.GetK3SImageName(),
		Cmd: []string{
			"server",
			"--tls-san",
			"0.0.0.0",
			"--tls-san",
			fmt.Sprintf("%s.kcluster.local.khost.dev", c.klfile.AccountName),
		},
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			//"/Users/karthik/Downloads/k9s/k9s:/bin/k9s",
			fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", c.klfile.AccountName),
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return fn.Error("failed to create container")
	}
	err = c.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return fn.Error("failed to start container")
	}
	return nil
}

// func (c *client) allWorkspaceContainers() ([]types.Container, error) {
// 	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
// 		All: true,
// 		Filters: filters.NewArgs(
// 			dockerLabelFilter(CONT_MARK_KEY, CONT_MARK_KEY),
// 			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, CONT_WORKSPACE_MARK_KEY),
// 		),
// 	})
// 	if err != nil {
// 		return nil, fn.NewE(err, "failed to list containers")
// 	}

// 	return existingContainers, nil
// }

// func (c *client) ensureContainerRunning(containerId string) error {
// 	cont, err := c.cli.ContainerInspect(context.Background(), containerId)
// 	if err != nil {
// 		return fn.NewE(err, "failed to inspect container")
// 	}

// 	if !cont.State.Running {
// 		return c.cli.ContainerStart(context.Background(), containerId, container.StartOptions{})
// 	}
// 	return nil
// }

// func (c *client) getContainerLogs(ctx context.Context, containerId string) (io.ReadCloser, error) {
// 	return c.cli.ContainerLogs(ctx, containerId, container.LogsOptions{
// 		ShowStdout: true,
// 		ShowStderr: true,
// 		Follow:     false,
// 	})
// }

func (c *client) containerAtPath(path string) (*types.Container, error) {
	defer spinner.Client.UpdateMessage("looking for the container")()

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
			dockerLabelFilter(CONT_PATH_KEY, path),
		),
	})
	if err != nil {
		return nil, fn.Error("failed to list containers")
	}
	if len(existingContainers) == 0 {
		return nil, fn.Error(NO_RUNNING_CONTAINERS)
	}
	return &existingContainers[0], nil
}

func (c *client) waithForSshReady(port int, containerId string) error {
	defer spinner.Client.UpdateMessage("waiting for ssh to be ready")()

	ctx, cf := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cf()

	if c.verbose {
		go func() {
			rc, err := c.cli.ContainerLogs(ctx, containerId, container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Since:      time.Now().Format(time.RFC3339),
				Follow:     true,
			})

			if r := bufio.NewScanner(rc); err == nil {
				for r.Scan() {
					l := r.Text()
					if len(l) > 8 {
						l = l[8:]
					}
					fn.Log(l)
				}
			}
		}()
	}

	t := time.Now()

	for {
		cj, err := c.cli.ContainerInspect(context.TODO(), containerId)
		if err != nil {
			return fn.NewE(err)
		}

		if cj.State != nil && !cj.State.Running {
			rc, err := c.cli.ContainerLogs(context.Background(), containerId, container.LogsOptions{
				ShowStdout: false,
				ShowStderr: true,
				Follow:     false,
				Since:      t.Format(time.RFC3339),
			})

			if r := bufio.NewScanner(rc); err == nil {
				logs := ""
				for r.Scan() {
					l := r.Text()
					if len(l) > 8 {
						l = l[8:]
					}

					logs += l + "\n"
				}

				return fn.NewE(fmt.Errorf("failed to start container"), logs)
			}

			return fn.Errorf("container is not running")
		}

		if err := sshclient.CheckSSHConnection(sshConf("localhost", port)); err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}
