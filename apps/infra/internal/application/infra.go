package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/uuid"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/rexec"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/wgman"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) AddAccount(cxt context.Context, action domain.AddAccountAction) (port string, publicKey string, err error) {
	cmd := exec.Command("bash", "./account-wireguard/init.sh", action.AccountId)
	cmd.Dir = fmt.Sprintf("./infra-scripts/%v/init-scripts", action.Provider)
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId)))
	cmd.Output()

	namespace := fmt.Sprintf("wg-%v", action.AccountId)
	err = i.waitForWireguardAvailability(string(action.ClusterId), namespace)
	if err != nil {
		return
	}
	clusterWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId),
		"wireguard",
		"deploy/wireguard-deployment",
		"/config",
		true,
	)

	clusterWgPublicKey, err := clusterWg.GetPublicKey()

	if err != nil {
		return
	}
	wg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId),
		namespace,
		"deploy/wireguard-deployment",
		"/etc/wireguard",
		true,
	)

	accountWgPublicKey, err := wg.Init(action.AccountIp)
	clusterWgEndpoint := "wireguard-service.wireguard.svc.cluster.local:51820"
	if err != nil {
		return
	}

	accountWgEndpoint := fmt.Sprintf("wireguard-service.%v.svc.cluster.local:51820", namespace)
	if err != nil {
		return
	}

	accountIpSplits := strings.Split(action.AccountIp, ".")
	wg.AddRemotePeer(clusterWgPublicKey, "10.13.13.1/32", &clusterWgEndpoint)
	clusterWg.AddRemotePeer(accountWgPublicKey, fmt.Sprintf("10.12.%v.0/24", accountIpSplits[2]), &accountWgEndpoint)
	if err != nil {
		return
	}
	port, err = i.getWireguardPort(string(action.ClusterId), action.AccountId)
	if err != nil {
		return
	}
	return port, accountWgPublicKey, err
}

func (i *infraClient) getWireguardPort(clusterId string, accountId string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		"services/wireguard-service",
		"-n",
		fmt.Sprintf("wg-%v", accountId),
		"-o",
		"jsonpath=\"{.spec.ports[0].nodePort}\"",
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId)))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(output), "\"", ""), nil
}

func (i *infraClient) GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) (map[string]string, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		fmt.Sprintf("secrets/%v", fmt.Sprintf("mres-%v", resName)),
		"-n",
		namespace,
		"-o",
		"json",
	)
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

	if masterCountstr, err := i.getOutputTerraformInFolder(action.ClusterId, "master-nodes-count"); err == nil {
		masterCount, _ = strconv.Atoi(masterCountstr)
	} else {
		return err
	}

	if agentCountstr, err := i.getOutputTerraformInFolder(action.ClusterId, "agent-nodes-count"); err == nil {
		agentCount, _ = strconv.Atoi(agentCountstr)
	} else {
		return err
	}

	cmd := exec.Command("terraform", "destroy", "-auto-approve", fmt.Sprintf("-var=master-nodes-count=%v", masterCount), fmt.Sprintf("-var=agent-nodes-count=%v", agentCount), fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath))
	cmd.Dir = fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterId)
	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *infraClient) AddPeer(cxt context.Context, action domain.AddPeerAction) (e error) {
	err := i.waitForWireguardAvailability(action.ClusterId, fmt.Sprintf("wg-%v", action.AccountId))
	if err != nil {
		return err
	}
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId),
		fmt.Sprintf("wg-%v", action.AccountId), // namespace
		"deploy/wireguard-deployment",
		"/etc/wireguard",
		true,
	)
	return serverWg.AddRemotePeer(action.PublicKey, fmt.Sprintf("%v/32", action.PeerIp), nil)
}

func (i *infraClient) DeletePeer(cxt context.Context, action domain.DeletePeerAction) (e error) {
	serverWg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId),
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
			_ip, e := wg.GetEmbededWireguardIp()
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
			e = wg.AddPeer(clusterPublicKey, "10.13.13.1/32,10.12.0.0/16", &endpoint)
			if e != nil {
				fmt.Println(clusterPublicKey, "10.42.0.0/16,10.43.0.0/16", endpoint)
				return
			}
			e = serverWg.AddPeer(nodePublicKey, fmt.Sprintf("%v/32", strings.TrimSpace(_ip)), nil)
			fmt.Println(fmt.Sprintf("wireguard for node %v setup %v", ip, nodePublicKey))
			if e != nil {
				fmt.Println(fmt.Errorf("unable to add node as peer to wireguard server %v", e))
				return
			}
		}(ip)
	}

	return nil
}

func (i *infraClient) waitForWireguardAvailability(clusterId, namespace string) error {
	fmt.Println(namespace)
	count := 0
	for count < 200 {
		cmd := exec.Command(
			"kubectl",
			"wait",
			"--timeout=3s",
			"--for=condition=Available=True",
			"deploy/wireguard-deployment",
			"-n",
			namespace,
		)
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

func (i *infraClient) setupKubeWireguard(ip, namespace, clusterId string) (string, error) {

	err := i.waitForWireguardAvailability(clusterId, namespace)

	if err != nil {
		fmt.Println("error on kube server start:", err)
		return "", err
	}

	wg := wgman.NewKubeWgManager(
		"/etc/wireguard/wg0.conf",
		fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, clusterId),
		namespace,
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

type AgentNode struct {
	Name     string
	Ip       string
	PublicIp string
	Dropping bool
}

type ClusterNodesData struct {
	Nodes []*AgentNode
}

type TerraformVariables struct {
	ClusterID  string
	AgentNodes []*AgentNode
	DOToken    string
	KeysPath   string
	DOImageID  string
	Region     string
}

func (i *infraClient) getClusterData(clusterId string) (ClusterNodesData, error) {
	file, err := ioutil.ReadFile(fmt.Sprintf("%v/%v/cluster-nodes.data", i.env.DataPath, clusterId))
	var data ClusterNodesData
	if err == nil {
		err = json.Unmarshal(file, &data)
	}
	err = nil
	return data, err
}

func (i *infraClient) saveClusterData(clusterId string, data ClusterNodesData) error {
	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%v/%v/cluster-nodes.data", i.env.DataPath, clusterId), marshal, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (i *infraClient) generateNewNodeIp(clusterId string, nodeName string) (string, error) {
	data, err := i.getClusterData(clusterId)
	if err != nil {
		return "", err
	}
	saveIp := func(ip string) (string, error) {
		data.Nodes = append(data.Nodes, &AgentNode{Name: nodeName, Ip: ip})
		marshal, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%v/%v/cluster-nodes.data", i.env.DataPath, clusterId), marshal, 0644)
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	for ipind := 4; ipind <= 254; ipind++ {
		ip := fmt.Sprintf("10.13.13.%v", ipind)
		if len(data.Nodes) == 0 {
			return saveIp(ip)
		}
		for ind, node := range data.Nodes {
			if node.Ip == ip {
				break
			}
			if ind == len(data.Nodes)-1 {
				return saveIp(ip)
			}
		}
	}
	return "", errors.New("no more ip available")
}

func applyTerraformInFolder(folder string, values TerraformVariables) error {
	vars := []string{"apply", "-auto-approve"}

	file, err := ioutil.ReadFile(fmt.Sprintf("%v/variables.tmpl", folder))
	if err != nil {
		return err
	}

	parse, err := template.New("").Parse(string(file))
	if err != nil {
		return err
	}
	var variablesData bytes.Buffer
	parse.Execute(&variablesData, values)

	ioutil.WriteFile(fmt.Sprintf("%v/variables.tf", folder), variablesData.Bytes(), 0644)

	cmd := exec.Command("terraform", vars...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = folder
	err = cmd.Run()

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
	install := rClient.Run("./install.sh", k3sToken, i.env.MySQLConnectionStr)
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
			run := rClient.Run("./join-master.sh", k3sToken, masterIp, nodeName, i.env.MySQLConnectionStr)
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

func (i *infraClient) installAgents(masterIp string, agentNodes []*AgentNode, k3sToken string) error {

	if len(agentNodes) < 1 {
		return nil
	}

	var outErr error
	for _, agentNode := range agentNodes {
		if agentNode.Dropping {
			continue
		}
		func(ip string) error {
			fmt.Println("waiting for ssh availability", ip)
			i.waitForSshAvailability(ip)
			rClient := rexec.NewSshRclient(ip, "root", fmt.Sprintf("%v/access", i.env.SshKeysPath))
			run := rClient.Run("./join-agent.sh", k3sToken, masterIp, agentNode.Name)
			_, err := run.Output()
			if err != nil {
				outErr = err
			}
			return err
		}(agentNode.PublicIp)
	}
	return outErr
}

func (i *infraClient) applyCluster(
	cxt context.Context,
	clusterId string,
	provider string,
	region string,
	agentNodes []*AgentNode,
) (publicIp string, publicKey string, e error) {
	defer errors.HandleErr(&e)
	if _, err := os.Stat(fmt.Sprintf("%v/%v", i.env.DataPath, clusterId)); os.IsNotExist(err) {
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

	e = applyTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, clusterId),
		TerraformVariables{
			ClusterID:  clusterId,
			AgentNodes: agentNodes,
			DOToken:    i.env.DoAPIKey,
			KeysPath:   i.env.SshKeysPath,
			DOImageID:  i.env.DoImageId,
			Region:     region,
		})

	errors.AssertNoError(e, fmt.Errorf("unable to apply terraform primary"))

	masterIps, e := i.getOutputTerraformInFolder(clusterId, "master-ips")
	agentIpsMap, e := i.getOutputTerraformInFolder(clusterId, "agent-ips")
	errors.AssertNoError(e, fmt.Errorf("unable to get cluster ips"))
	data, err := i.getClusterData(clusterId)
	errors.AssertNoError(err, fmt.Errorf("unable to get cluster data"))
	for _, agentIpEntry := range strings.Split(agentIpsMap, ",") {
		if agentIpEntry == "" {
			continue
		}
		for _, agentNode := range data.Nodes {
			nodeName := strings.Split(agentIpEntry, ":")[0]
			publicIp := strings.Split(agentIpEntry, ":")[1]
			fmt.Println("node name", nodeName, "public ip", publicIp)
			if agentNode.Name == nodeName {
				agentNode.PublicIp = publicIp
				break
			}
		}
	}
	e = i.saveClusterData(clusterId, data)
	errors.AssertNoError(e, fmt.Errorf("unable to save cluster data"))

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
		var agentIps []string
		for _, agentNode := range data.Nodes {
			if !agentNode.Dropping {
				agentIps = append(agentIps, agentNode.PublicIp)
			}
		}
		err := i.setupNodeWireguards(
			append(strings.Split(masterIps, ","), agentIps...),
			clusterId,
			clusterIp,
			clusterPublicKey,
		)
		if err != nil {
			fmt.Println("[Error] unable to setup wireguard", err)
		}
		fmt.Println("#5 node wireguards done")
		waitGroup.Done()
	}()

	go func() {
		wireguardPub, err := i.setupKubeWireguard("10.13.13.1", "wireguard", clusterId)
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
		_, err := i.setupAllKubernetes(clusterId, provider)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to setup kubernetes %v", err))
		}
		fmt.Println("#3 kubernetes setup done")
		waitGroup.Done()
	}()

	go func() {
		err := i.installSecondaryMasters(strings.Split(masterIps, ","), string(joinToken))
		if err != nil {
			fmt.Println(fmt.Errorf("unable to get cluster data %v", err))
		}
		fmt.Println("#1 secondary masters installed")
		waitGroup.Done()
	}()

	go func() {
		err := i.installAgents(clusterIp, data.Nodes, string(joinToken))
		if err != nil {
			fmt.Println(fmt.Errorf("unable to get cluster data %v", err))
		}
		fmt.Println("#1 agents installed")
		waitGroup.Done()
	}()

	waitGroup.Wait()
	fmt.Println("setup finished")

	fmt.Println(e)

	return clusterIp, clusterPublicKey, e
}

func (i *infraClient) CreateCluster(cxt context.Context, action domain.SetupClusterAction) (publicIp string, publicKey string, e error) {
	agentNodes := make([]*AgentNode, 0)
	for in := 0; in < action.NodesCount; in++ {
		newUUID := uuid.NewUUID()
		nodeName := fmt.Sprintf("agent-%v-%v", in, newUUID)
		ip, e := i.generateNewNodeIp(action.ClusterId, nodeName)
		if e != nil {
			return "", "", e
		}
		agentNodes = append(agentNodes, &AgentNode{
			Name: nodeName,
			Ip:   ip,
		})
	}
	cluster, key, e := i.applyCluster(cxt, action.ClusterId, action.Provider, action.Region, agentNodes)
	i.getClusterData(action.ClusterId)
	return cluster, key, e
}

var dropNodeTimer *time.Timer

func (i *infraClient) UpdateCluster(cxt context.Context, action domain.UpdateClusterAction) (e error) {
	if dropNodeTimer != nil {
		dropNodeTimer.Stop()
	}

	data, e := i.getClusterData(action.ClusterId)
	if e != nil {
		return e
	}

	if action.AgentNodesCount() < len(data.Nodes) {
		for in, node := range data.Nodes {
			if in >= action.AgentNodesCount() {
				node.Dropping = true
			} else {
				node.Dropping = false
			}
		}
	} else {
		for in := len(data.Nodes); in < action.AgentNodesCount(); in++ {
			newUUID := uuid.NewUUID()
			nodeName := fmt.Sprintf("agent-%v-%v", in, newUUID)
			ip, err := i.generateNewNodeIp(action.ClusterId, nodeName)
			if err != nil {
				return err
			}
			data.Nodes = append(data.Nodes, &AgentNode{
				Name:     nodeName,
				Ip:       ip,
				Dropping: false,
			})
		}
	}
	i.saveClusterData(action.ClusterId, data)
	_, _, e = i.applyCluster(cxt, action.ClusterId, action.Provider, action.Region, data.Nodes)

	if e != nil {
		return e
	}

	for _, node := range data.Nodes {
		var cmd *exec.Cmd
		if !node.Dropping {
			cmd = exec.Command(
				"kubectl",
				"uncordon",
				node.Name,
			)
		} else {
			cmd = exec.Command(
				"kubectl",
				"drain",
				node.Name,
				"--ignore-daemonsets",
				"--delete-local-data",
			)
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId)))
		_, e := cmd.Output()
		if e != nil {
			fmt.Println(fmt.Errorf("unable to drain node %v %v", node, e))
			continue
		}
	}

	dropNodeTimer = time.AfterFunc(time.Duration(i.env.NodeDrainTime)*time.Second, func() {
		data, e := i.getClusterData(action.ClusterId)
		if e != nil {
			return
		}
		for _, node := range data.Nodes {
			if node.Dropping {
				cmd := exec.Command(
					"kubectl",
					"delete", "node",
					node.Name,
				)
				cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", fmt.Sprintf("%v/%v/kubeconfig", i.env.DataPath, action.ClusterId)))
				_, e := cmd.Output()
				if e != nil {
					fmt.Println(fmt.Errorf("unable to delete node %v %v", node, e))
					continue
				}
			}
		}
		newNodes := make([]*AgentNode, 0)
		for _, node := range data.Nodes {
			if node.Dropping {
				continue
			}
			newNodes = append(newNodes, node)
		}
		data.Nodes = newNodes
		cluster, key, e := i.applyCluster(cxt, action.ClusterId, action.Provider, action.Region, newNodes)
		fmt.Println(cluster, key, e)
		i.saveClusterData(action.ClusterId, data)
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
