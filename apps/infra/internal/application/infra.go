package application

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/wgman"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) DeleteCluster(action domain.DeleteClusterAction) (e error) {
	//TODO implement me
	var masterCount, agentCount int

	if masterCountstr, err := i.getOutputTerraformInFolder(action.ClusterID, "master-nodes-count"); err == nil {
		masterCount, _ = strconv.Atoi(masterCountstr)
	} else {
		return err
	}

	if agentCountstr, err := i.getOutputTerraformInFolder(action.ClusterID, "agent-nodes-count"); err == nil {
		agentCount, _ = strconv.Atoi(agentCountstr)
	} else {
		return err
	}

	cmd := exec.Command("terraform", "destroy", "-auto-approve", fmt.Sprintf("-var=master-nodes-count=%v", masterCount), fmt.Sprintf("-var=agent-nodes-count=%v", agentCount))
	cmd.Dir = fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID)
	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *infraClient) AddPeer(action domain.AddPeerAction) (e error) {
	//TODO implement me
	// panic("implement me")
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID),
		"wireguard",
		"deploy/wireguard-deployment",
		true,
	)

	return serverWg.AddRemotePeer(action.PublicKey, fmt.Sprintf("%v/32", action.PeerIp), nil)

}

func (i *infraClient) DeletePeer(action domain.DeletePeerAction) (e error) {
	//TODO implement me
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID),
		"wireguard",
		"deploy/wireguard-deployment",
		true,
	)

	return serverWg.DeleteRemotePeer(action.PublicKey)

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

func (i *infraClient) setupNodeWireguards(
	nodeIps []string,
	clusterId string,
	clusterIp string,
	clusterPublicKey string,
) (err error) {

	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId),
		"wireguard",
		"deploy/wireguard-deployment",
		true,
	)
	for _, ip := range nodeIps {
		func(ip string) {
			wg := wgman.NewSshWgManager("/etc/wireguard/wg0.conf", ip, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath), false)

			// if !wg.IsSetupDone() {
			if true {
				_ip, e := wg.GetNodeIp()
				if e != nil {

					fmt.Println(fmt.Errorf("unable to get nodeip %v", e))
					return
				}
				nodePublicKey, e := wg.Init(_ip)
				if e != nil {

					fmt.Println(fmt.Errorf("failed to setup wireguard for node %v", e))
					return
				}
				endpoint := fmt.Sprintf("%v:31820", clusterIp)
				e = wg.AddPeer(clusterPublicKey, "10.42.0.0/16,10.43.0.0/16", &endpoint)
				if e != nil {
					fmt.Println(clusterPublicKey, "10.42.0.0/16,10.43.0.0/16", endpoint)
					panic(e)

					fmt.Println(fmt.Errorf("failed to add peer to wireguard for node %v", e))
					return
				}
				e = serverWg.AddPeer(nodePublicKey, fmt.Sprintf("%v/32", strings.TrimSpace(_ip)), nil)
				if e != nil {

					fmt.Println(fmt.Errorf("unable to add node as peer to wireguard server %v", e))
					return
				}
			}

		}(ip)
	}

	return nil
}

func (i *infraClient) setupKubeWireguard(ip, clusterId string) (string, error) {

	cmd := exec.Command("kubectl", "wait", "--for=condition=Available=True", "deploy/wireguard-deployment", "-n", "wireguard")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%v/%v/kubeconfig", i.env.DataPath, clusterId))
	err := cmd.Run()

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return "", err
	}

	wg := wgman.NewKubeWgManager("/etc/wireguard/wg0.conf", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId), "wireguard", "deploy/wireguard-deployment", true)

	o, err := wg.Init(ip)

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return o, err
	}
	return o, err
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

func (i *infraClient) getOutputTerraformInFolder(clusterId string, key string) (string, error) {
	cmd := exec.Command("terraform", "output", key)
	cmd.Dir = fmt.Sprintf("%v/%v", i.env.DataPath, clusterId)
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

	cmd := exec.Command(
		"k3sup",
		"install",
		fmt.Sprintf("--ip=%v", masterIp),
		"--cluster",
		"--k3s-version=v1.19.1+k3s1",
		"--user=root",
		"--k3s-extra-args='--disable=traefik'",
		"--k3s-extra-args='--node-name=master'",
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

func (i *infraClient) installSecondaryMasters(masterIps []string, clusterId string) (err error) {
	masterIp := masterIps[0]
	if len(masterIps) < 2 {
		return nil
	}
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(masterIps) - 1)
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
			cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
			_, err := cmd.Output()
			waitGroup.Done()
			if err != nil {
				return err
			}
			return nil
		}(ip)
	}
	waitGroup.Wait()
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

func (i *infraClient) CreateCluster(action domain.SetupClusterAction) (publicIp string, publicKey string, e error) {

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

	masterIps, e := i.getOutputTerraformInFolder(action.ClusterID, "master-ips")
	agentIps, e := i.getOutputTerraformInFolder(action.ClusterID, "agent-ips")

	errors.AssertNoError(e, fmt.Errorf("unable to get cluster ip"))

	clusterIp := strings.Split(masterIps, ",")[0]
	var clusterPublicKey string
	var clusterPublicKeyWaitGroup sync.WaitGroup
	clusterPublicKeyWaitGroup.Add(1)

	out, e := i.installPrimaryMaster(clusterIp, action.ClusterID)
	fmt.Println(string(out), e)

	errors.AssertNoError(e, fmt.Errorf("unable to install primary master"))

	var waitGroup sync.WaitGroup
	waitGroup.Add(5)
	go func() {
		clusterPublicKeyWaitGroup.Wait()
		i.setupNodeWireguards(
			append(strings.Split(masterIps, ","), strings.Split(agentIps, ",")...),
			action.ClusterID,
			clusterIp,
			clusterPublicKey,
		)
		waitGroup.Done()
	}()

	go func() {
		wireguardPub, err := i.setupKubeWireguard(clusterIp, action.ClusterID)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to setup wireguard %v", err))
			waitGroup.Done()
			return
		}
		clusterPublicKey = wireguardPub
		clusterPublicKeyWaitGroup.Done()
		waitGroup.Done()
	}()

	go func() {
		kubernetes, _ := i.setupAllKubernetes(action.ClusterID, action.Provider)
		fmt.Println(string(kubernetes))
		waitGroup.Done()
	}()

	go func() {
		i.installSecondaryMasters(strings.Split(masterIps, ","), action.ClusterID)
		waitGroup.Done()
	}()

	go func() {
		i.installAgents(clusterIp, strings.Split(agentIps, ","), action.ClusterID)
		waitGroup.Done()
	}()

	waitGroup.Wait()
	fmt.Println("setup finished")

	fmt.Println(e)

	return clusterIp, clusterPublicKey, e
}

func (i *infraClient) UpdateCluster(action domain.UpdateClusterAction) (e error) {
	panic("implement me")
	return nil
}

func fxInfraClient(env *InfraEnv) domain.InfraClient {
	return &infraClient{
		env,
	}
}
