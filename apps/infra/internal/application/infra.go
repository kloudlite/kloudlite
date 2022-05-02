package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/rexec"
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

func (i *infraClient) GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) (map[string]string, error) {
	cmd := exec.Command("kubectl", "get", fmt.Sprintf("secrets/%v", fmt.Sprintf("mres-%v", resName)), "-n", namespace, "-o", "json")
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
	output, err := cmd.Output()
	var res struct {
		Data map[string]string `json:"data"`
	}
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, err
	}
	return res.Data, err
}

func (i *infraClient) DeleteCluster(cxt context.Context, action domain.DeleteClusterAction) (e error) {
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

	cmd := exec.Command("terraform", "destroy", "-auto-approve", fmt.Sprintf("-var=master-nodes-count=%v", masterCount), fmt.Sprintf("-var=agent-nodes-count=%v", agentCount), fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath))
	cmd.Dir = fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID)
	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *infraClient) AddPeer(cxt context.Context, action domain.AddPeerAction) (e error) {
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID),
		"wireguard",
		"deploy/wireguard-deployment",
		"/config",
		true,
	)
	return serverWg.AddRemotePeer(action.PublicKey, fmt.Sprintf("%v/32", action.PeerIp), nil)
}

func (i *infraClient) DeletePeer(cxt context.Context, action domain.DeletePeerAction) (e error) {
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID),
		"wireguard",
		"deploy/wireguard-deployment",
		"/config",
		true,
	)

	return serverWg.DeleteRemotePeer(action.PublicKey)

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
		"/config",
		true,
	)
	for _, ip := range nodeIps {
		if ip == "" {
			continue
		}
		func(ip string) {
			wg := wgman.NewSshWgManager(
				"/etc/wireguard/wg0.conf",
				ip,
				"root",
				fmt.Sprintf("%v/access", i.env.SshKeysPath),
				"./",
				false,
			)
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
			e = wg.AddPeer(clusterPublicKey, "10.42.0.0/16,10.43.0.0/16,10.13.13.0/24", &endpoint)
			if e != nil {
				fmt.Println(clusterPublicKey, "10.42.0.0/16,10.43.0.0/16", endpoint)
				panic(e)
			}
			e = serverWg.AddPeer(nodePublicKey, fmt.Sprintf("%v/32", strings.TrimSpace(_ip)), nil)
			if e != nil {
				fmt.Println(fmt.Errorf("unable to add node as peer to wireguard server %v", e))
				return
			}
		}(ip)
	}

	return nil
}

func (i *infraClient) waitForWireguardAvailability(clusterId string) error {

	count := 0
	for count < 200 {
		cmd := exec.Command("kubectl", "wait", "--timeout=3s", "--for=condition=Available=True", "deploy/wireguard-deployment", "-n", "wireguard")
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
		o, e := cmd.Output()
		fmt.Println(e, string(o))
		if e == nil {
			return nil
		}
		fmt.Println("waiting for wireguard pods to be running")
		time.Sleep(time.Second * 6)
		count++
	}

	return errors.New("not able to access wireguard after 12 minute")
}

func (i *infraClient) setupKubeWireguard(ip, clusterId string) (string, error) {

	err := i.waitForWireguardAvailability(clusterId)

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return "", err
	}

	wg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId),
		"wireguard",
		"deploy/wireguard-deployment",
		"/config",
		true,
	)

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
			"mkdir", "-p", "/wg-config",
		).Run()

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
	if file, err := ioutil.ReadFile(fmt.Sprintf("%v/%v/join-token", i.env.DataPath, clusterId)); err == nil {
		return file, nil
	}
	i.waitForSshAvailability(masterIp)
	rClient := rexec.NewSshRclient(masterIp, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath))
	k3sToken, e := fn.CleanerNanoid(32)
	if e != nil {
		return nil, e
	}
	install := rClient.Run("./install.sh", k3sToken)
	_, e = install.Output()
	if e != nil {
		fmt.Println("error on kube server start:", e)
		return nil, e
	}
	e = ioutil.WriteFile(fmt.Sprintf("%v/%v/join-token", i.env.DataPath, clusterId), []byte(k3sToken), 0644)
	if e != nil {
		return nil, e
	}
	configData, e := rClient.Readfile("/etc/rancher/k3s/k3s.yaml")
	if e != nil {
		return nil, e
	}
	all := strings.ReplaceAll(string(configData), "127.0.0.1", masterIp)
	e = ioutil.WriteFile(fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId), []byte(all), 0644)
	return []byte(k3sToken), e
}

func (i *infraClient) setupAllKubernetes(clusterId string, provider string) ([]byte, error) {
	i.waitForKubernetesAPIAvailability(clusterId)
	cmd := exec.Command("bash", "init.sh")
	cmd.Dir = fmt.Sprintf("./infra-scripts/%v/init-scripts", provider)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
	return cmd.Output()
}

func (i *infraClient) installSecondaryMasters(masterIps []string, k3sToken string) (err error) {
	masterIp := masterIps[0]
	fmt.Println(masterIps)
	if len(masterIps) < 2 {
		return nil
	}
	//var waitGroup sync.WaitGroup
	//waitGroup.Add(len(masterIps) - 1)
	var outErr error
	for index, ip := range masterIps {
		if index == 0 {
			continue
		}
		func(ip string) error {
			i.waitForSshAvailability(ip)
			nodeName := fmt.Sprintf("master-%v", index)
			rClient := rexec.NewSshRclient(ip, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath))
			run := rClient.Run("./join-master.sh", k3sToken, masterIp, nodeName)
			_, err := run.Output()
			if err != nil {
				fmt.Println("unable to attach", err)
				outErr = err
			}
			//waitGroup.Done()
			return err
		}(ip)
	}
	//waitGroup.Wait()
	return outErr
}

func (i *infraClient) installAgents(masterIp string, agentIps []string, k3sToken string) error {

	if len(agentIps) < 1 {
		return nil
	}
	fmt.Println(agentIps, "HEREERE")
	if agentIps[0] == "" {
		return nil
	}

	//var waitGroup sync.WaitGroup
	//waitGroup.Add(len(agentIps))
	var outErr error
	for index, ip := range agentIps {
		func(ip string) error {
			i.waitForSshAvailability(ip)
			nodeName := fmt.Sprintf("agent-%v", index)
			fmt.Println(ip, nodeName)
			rClient := rexec.NewSshRclient(ip, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath))
			run := rClient.Run("./join-agent.sh", k3sToken, masterIp, nodeName)
			_, err := run.Output()
			if err != nil {
				outErr = err
			}
			//waitGroup.Done()
			return err
		}(ip)
	}
	//waitGroup.Wait()
	return outErr
}

func (i *infraClient) applyCluster(cxt context.Context, clusterId string, provider string, masterNodesCount int, agentNodesCount int) (publicIp string, publicKey string, e error) {
	defer errors.HandleErr(&e)

	if _, err := os.Stat(fmt.Sprintf("%v/%v", i.env.DataPath, clusterId)); os.IsNotExist(err) {

		// TODO: check if cluster already exists
		copyTemplateDirCommand := exec.Command(
			"cp",
			"-r",
			fmt.Sprintf("./infra-scripts/%v/tf/", provider),
			fmt.Sprintf("%v/%v", i.env.DataPath, clusterId),
		)

		copyTemplateDirCommand.Stdout = os.Stdout
		copyTemplateDirCommand.Stderr = os.Stderr
		e = copyTemplateDirCommand.Run()

		errors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))

	}
	e = initTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, clusterId))
	errors.AssertNoError(e, fmt.Errorf("unable to init terraform primary"))

	e = applyTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, clusterId), map[string]any{
		"cluster-id":         clusterId,
		"do-token":           i.env.DoAPIKey,
		"keys-path":          i.env.SshKeysPath,
		"master-nodes-count": masterNodesCount,
		"agent-nodes-count":  agentNodesCount,
		"do-image-id":        i.env.DoImageId,
	})

	errors.AssertNoError(e, fmt.Errorf("unable to apply terraform primary"))

	masterIps, e := i.getOutputTerraformInFolder(clusterId, "master-ips")
	agentIps, e := i.getOutputTerraformInFolder(clusterId, "agent-ips")

	errors.AssertNoError(e, fmt.Errorf("unable to get cluster ip"))

	clusterIp := strings.Split(masterIps, ",")[0]
	var clusterPublicKey string
	var clusterPublicKeyWaitGroup sync.WaitGroup
	clusterPublicKeyWaitGroup.Add(1)

	joinToken, e := i.installPrimaryMaster(clusterIp, clusterId)
	errors.AssertNoError(e, fmt.Errorf("unable to install primary master"))

	var waitGroup sync.WaitGroup
	waitGroup.Add(5)
	go func() {
		clusterPublicKeyWaitGroup.Wait()
		i.setupNodeWireguards(
			append(strings.Split(masterIps, ","), strings.Split(agentIps, ",")...),
			clusterId,
			clusterIp,
			clusterPublicKey,
		)
		fmt.Println("#5 node wireguards done")
		waitGroup.Done()
	}()

	go func() {
		wireguardPub, err := i.setupKubeWireguard(clusterIp, clusterId)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to setup wireguard %v", err))
			waitGroup.Done()
			return
		}
		clusterPublicKey = wireguardPub
		clusterPublicKeyWaitGroup.Done()
		fmt.Println("#4 wireguard setup done")
		waitGroup.Done()
	}()

	go func() {
		kubernetes, _ := i.setupAllKubernetes(clusterId, provider)
		fmt.Println(string(kubernetes))
		fmt.Println("#3 kubernetes setup done")
		waitGroup.Done()
	}()

	go func() {
		i.waitForKubernetesAPIAvailability(clusterId)
		i.installSecondaryMasters(strings.Split(masterIps, ","), string(joinToken))
		fmt.Println("#2 sec masters installed")
		waitGroup.Done()
	}()

	go func() {
		i.installAgents(clusterIp, strings.Split(agentIps, ","), string(joinToken))
		fmt.Println("#1 agents installed")
		waitGroup.Done()
	}()

	waitGroup.Wait()
	fmt.Println("setup finished")

	fmt.Println(e)

	return clusterIp, clusterPublicKey, e
}

func (i *infraClient) CreateCluster(cxt context.Context, action domain.SetupClusterAction) (publicIp string, publicKey string, e error) {
	return i.applyCluster(cxt, action.ClusterID, action.Provider, action.MasterNodesCount(), action.AgentNodesCount())
}

func (i *infraClient) UpdateCluster(cxt context.Context, action domain.UpdateClusterAction) (e error) {
	masterNodesCountStr, e := i.getOutputTerraformInFolder(action.ClusterID, "master-nodes-count")
	agentNodesCountStr, e := i.getOutputTerraformInFolder(action.ClusterID, "agent-nodes-count")

	masterNodesCount, e := strconv.Atoi(masterNodesCountStr)
	agentNodesCount, e := strconv.Atoi(agentNodesCountStr)

	dropNodes := make([]string, 0)

	firstPassExpectedMasterCount := masterNodesCount
	firstPassExpectedAgentCount := agentNodesCount

	if action.MasterNodesCount() < masterNodesCount {
		for i := action.MasterNodesCount(); i < masterNodesCount; i++ {
			dropNodes = append(dropNodes, fmt.Sprintf("master-%v", i))
		}
	} else {
		firstPassExpectedMasterCount = action.MasterNodesCount()
	}

	if action.AgentNodesCount() < agentNodesCount {
		for i := action.AgentNodesCount(); i < agentNodesCount; i++ {
			dropNodes = append(dropNodes, fmt.Sprintf("agent-%v", i))
		}
	} else {
		firstPassExpectedAgentCount = action.AgentNodesCount()
	}

	_, _, e = i.applyCluster(cxt, action.ClusterID, action.Provider, firstPassExpectedMasterCount, firstPassExpectedAgentCount)

	if e != nil {
		return e
	}

	for _, node := range dropNodes {
		cmd := exec.Command(
			"kubectl",
			"drain",
			node,
			"--ignore-daemonsets",
			"--delete-local-data",
		)
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID)))
		_, e := cmd.Output()
		if e != nil {
			fmt.Println(fmt.Errorf("unable to drain node %v %v", node, e))
			continue
		}
	}

	time.AfterFunc(time.Duration(i.env.NodeDrainTime)*time.Second, func() {
		for _, node := range dropNodes {
			cmd := exec.Command(
				"kubectl",
				"delete", "node",
				node,
			)
			cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterID)))
			_, e := cmd.Output()
			if e != nil {
				fmt.Println(fmt.Errorf("unable to drain node %v %v", node, e))
				continue
			}
		}
		i.applyCluster(cxt, action.ClusterID, action.Provider, action.MasterNodesCount(), action.AgentNodesCount())
	})
	return e
}

func (i *infraClient) waitForKubernetesAPIAvailability(clusterId string) error {
	count := 0
	for count < 20 {
		cmd := exec.Command(
			"kubectl",
			"get",
			"nodes",
		)
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
		e := cmd.Run()
		if e == nil {
			return nil
		}

		fmt.Println("waiting for kubernetes availability")
		time.Sleep(time.Second * 3)
		count++
	}

	return errors.New("not able to access kubernetes after 10 attempts")
}

func fxInfraClient(env *InfraEnv) domain.InfraClient {
	return &infraClient{
		env,
	}
}
