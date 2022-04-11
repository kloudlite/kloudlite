package application

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/wgman"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) setupMaster(ip string) error {
	fmt.Println("ssh",
		"-o",
		"StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i",
		fmt.Sprintf("%v/access", i.env.SshKeysPath),
		"root@"+ip,
		"/root/scripts/wait-for-on.sh")

	e := exec.Command(
		"ssh",
		"-o",
		"StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i",
		fmt.Sprintf("%v/access", i.env.SshKeysPath),
		"root@"+ip,
		"/root/scripts/wait-for-on.sh",
	).Run()

	return e
}

func (i *infraClient) checkAndSetupNodeWireguards(ips []string) {

	c := make(chan bool, len(ips))

	fmt.Println(len(ips), "to setup wireguard")

	for ind, ip := range ips {
		go func(ip string) {

			wg := wgman.NewSshWgManager("/etc/wireguard/wg0.conf", ip, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath), false)

			if !wg.IsSetupDone() {
				_ip, e := wg.GetNodeIp()
				if e != nil {
					fmt.Println(e)
					c <- false
					return
				}
				o, e := wg.Init(_ip)
				fmt.Println(o, e)
			}
			fmt.Println(ind, "done")
			c <- true
		}(ip)
	}

	<-c
	fmt.Println("done", "checkAndSetupNodeWireguards")
}

func (i *infraClient) checkAndSetupKubeWireguard(ip, clusterId string) error {

	cmd := exec.Command("kubectl", "wait", "--for=condition=Available=True", "deploy/wireguard-deployment", "-n", "wireguard")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%v/%v/kubeconfig", i.env.DataPath, clusterId))
	err := cmd.Run()

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return err
	}

	wg := wgman.NewKubeWgManager("/config/wg0.conf", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId), "wireguard", "deploy/wireguard-deployment", true)

	o, err := wg.Init(ip)

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return err
	}

	fmt.Println("server publickey: ", o)

	return err

}

func initTerraformInFolder(folder string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		return err
	}
	return nil
}

func applyTerraformInFolder(folder string, values map[string]any) error {
	vars := []string{"apply", "-auto-approve"}

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%v=%v", k, v))
	}

	cmd := exec.Command("terraform", vars...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = folder
	err := cmd.Run()

	if err != nil {
		return err
	}
	return nil
}

func getOutputTerraformInFolder(folder string, key string) (string, error) {
	cmd := exec.Command("terraform", "output", key)
	cmd.Dir = folder
	out, err := cmd.Output()
	return strings.ReplaceAll(strings.ReplaceAll(string(out), "\"", ""), "\n", ""), err
}

func (i *infraClient) waitForSshAvailability(ip string) error {

	count := 0
	for count < 20 {
		e := exec.Command(
			"ssh",
			"-o",
			"StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-i",
			fmt.Sprintf("%v/access", i.env.SshKeysPath),
			"root@"+ip,
			"echo", "hello",
		).Run()

		fmt.Println(e)
		if e == nil {
			return nil
		}

		fmt.Println("waiting for ssh availability")
		time.Sleep(time.Second * 3)
		count++
	}

	return errors.New("not able to access ssh after 10 attempts")
}

func (i *infraClient) installPrimaryMaster(masterIp string, clusterId string) ([]byte, error) {

	i.waitForSshAvailability(masterIp)

	fmt.Println("k3sup", "install", fmt.Sprintf("--ip=%v", masterIp), "--cluster", "--k3s-version=v1.19.1+k3s1", "--user=root")

	cmd := exec.Command(
		"k3sup",
		"install",
		fmt.Sprintf("--ip=%v", masterIp),
		"--cluster",
		"--k3s-version=v1.19.1+k3s1",
		"--user=root",
		"--k3s-extra-args='--disable=traefik'",
	)

	cmd.Dir = fmt.Sprintf("%v/%v", i.env.DataPath, clusterId)

	return cmd.Output()
}

func (i *infraClient) setupAllKubernetes(clusterId string, provider string) ([]byte, error) {
	cmd := exec.Command("bash", "init.sh")
	cmd.Dir = fmt.Sprintf("./infra-scripts/%v/init-scripts", provider)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
	return cmd.Output()
}

func (i *infraClient) installSecondaryMasters(masterIps []string, clusterId string) error {

	masterIp := masterIps[0]

	if len(masterIps) < 2 {
		return nil
	}

	c := make(chan error, len(masterIps)-1)

	for index, ip := range masterIps {

		if index == 0 {
			continue
		}

		go func(ip string) error {

			i.waitForSshAvailability(ip)

			cmd := exec.Command(
				"k3sup",
				"join",
				fmt.Sprintf("--ip=%v", ip),
				fmt.Sprintf("--server-ip=%v", masterIp),
				"--k3s-version=v1.19.1+k3s1",
				"--user=root",
				"--server-user=root",
				"--server",
				"--k3s-extra-args='--disable=traefik'",
			)

			_, err := cmd.Output()
			c <- err

			if err != nil {
				return err
			}

			return nil
		}(ip)

	}

	err := <-c
	fmt.Println("done joining masters")

	return err
}

func (i *infraClient) installAgents(masterIp string, agentIps []string, clusterId string) error {

	// masterIp := masterIps[0]

	if len(agentIps) < 1 {
		return nil
	}

	c := make(chan error, len(agentIps))

	for _, ip := range agentIps {

		go func(ip string) error {
			cmd := exec.Command(
				"k3sup",
				"join",
				fmt.Sprintf("--ip=%v", ip),
				fmt.Sprintf("--server-ip=%v", masterIp),
				"--k3s-version=v1.19.1+k3s1",
				"--user=root",
				"--server-user=root",
			)

			_, err := cmd.Output()
			c <- err

			if err != nil {
				return err
			}

			return nil
		}(ip)

	}

	err := <-c
	fmt.Println("done joining agents")

	return err
}

// TODO: Have to implement this
func (i *infraClient) waitForKubernetesSetup() {}

func (i *infraClient) CreateKubernetes(action domain.SetupClusterAction) (e error) {

	defer errors.HandleErr(&e)

	copyTemplateDirCommand := exec.Command(
		"cp",
		"-r",
		fmt.Sprintf("./infra-scripts/%v/tf/", action.Provider),
		fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID),
	)

	copyTemplateDirCommand.Stdout = os.Stdout
	copyTemplateDirCommand.Stderr = os.Stderr
	e = copyTemplateDirCommand.Run()

	errors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))

	e = initTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID))
	errors.AssertNoError(e, fmt.Errorf("unable to init terraform primary"))

	masterCount := func() int {
		if action.NodesCount > 3 {
			return 3
		}
		return 1
	}()

	e = applyTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), map[string]any{
		"cluster-id":         action.ClusterID,
		"do-token":           i.env.DoAPIKey,
		"keys-path":          i.env.SshKeysPath,
		"master-nodes-count": masterCount,
		"agent-nodes-count":  action.NodesCount - masterCount,
		"do-image-id":        i.env.DoImageId,
	})

	errors.AssertNoError(e, fmt.Errorf("unable to apply terraform primary"))

	// masterFloatingIps, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "master-floating-ips")
	masterIps, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "master-ips")
	agentIps, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "agent-ips")

	errors.AssertNoError(e, fmt.Errorf("unable to get cluster ip"))

	fmt.Println("cluster setup finished")

	masterIp := strings.Split(masterIps, ",")[0]

	_, e = i.installPrimaryMaster(masterIp, action.ClusterID)

	fmt.Println("setup finished", e)
	errors.AssertNoError(e, fmt.Errorf("unable to install primary master"))

	c := make(chan bool, 5)

	go func() {
		i.checkAndSetupNodeWireguards(append(strings.Split(masterIps, ","), strings.Split(agentIps, ",")...))
		fmt.Println("alex", 0)
		c <- true
	}()

	go func() {
		i.checkAndSetupKubeWireguard(masterIp, action.ClusterID)

		fmt.Println("alex", 1)
		c <- true
	}()

	go func() {
		i.setupAllKubernetes(action.ClusterID, action.Provider)

		fmt.Println("alex", 2)
		c <- true
	}()

	go func() {
		i.installSecondaryMasters(strings.Split(masterIps, ","), action.ClusterID)

		fmt.Println("alex", 3)
		c <- true
	}()

	go func() {
		i.installAgents(masterIp, strings.Split(agentIps, ","), action.ClusterID)

		fmt.Println("alex", 4)
		c <- true
	}()

	<-c
	fmt.Println("setup finished")

	fmt.Println(e)
	errors.AssertNoError(e, fmt.Errorf("unable to setup master"))

	// e = i.setupWireguard(masterIp)
	// _, e = i.generateWireguardKey(masterIp)
	errors.AssertNoError(e, fmt.Errorf("unable to setup wireguard"))

	return e
}

func (i *infraClient) UpdateKubernetes(action domain.UpdateClusterAction) (e error) {

	defer errors.HandleErr(&e)
	applyCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v", fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID)),
		"apply",
		"-auto-approve",
		fmt.Sprintf("-var=cluster-id=%v", action.ClusterID),
		fmt.Sprintf("-var=master-nodes-count=%v", action.MastersCount),
		fmt.Sprintf("-var=agent-nodes-count=%v", action.NodesCount-action.MastersCount),
		fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath),
		fmt.Sprintf("-var=do-token=%v", i.env.DoAPIKey),
	)
	applyCommand.Stdout = os.Stdout
	applyCommand.Stderr = os.Stderr
	e = applyCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	return e
}

func fxInfraClient(env *InfraEnv) domain.InfraClient {
	return &infraClient{
		env,
	}
}
