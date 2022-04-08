package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/aymerick/raymond"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/errors"
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

// func setupSecondaryMasters([]string ips) error{}
// func setupAgents([]string ips) error{}

func initTerraformInFolder(folder string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	// fmt.Println(err)

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

func (i *infraClient) CreateKubernetes(action domain.SetupClusterAction) (e error) {
	defer errors.HandleErr(&e)
	copyTemplateDirCommand := exec.Command(
		"cp",
		"-r",
		fmt.Sprintf("./internal/application/tf/%v", action.Provider),
		fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID),
	)
	copyTemplateDirCommand.Stdout = os.Stdout
	copyTemplateDirCommand.Stderr = os.Stderr
	e = copyTemplateDirCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))

	// fmt.Println(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID))
	e = initTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID))
	errors.AssertNoError(e, fmt.Errorf("unable to init terraform primary"))

	// e = initTerraformInFolder(fmt.Sprintf("%v/%v/secondary-tf", i.env.DataPath, action.ClusterID))
	// errors.AssertNoError(e, fmt.Errorf("unable to init terraform secondary"))

	e = applyTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), map[string]any{
		"cluster-id":         action.ClusterID,
		"do-token":           i.env.DoAPIKey,
		"do-image-id":        i.env.DoImageId,
		"keys-path":          i.env.SshKeysPath,
		"master-nodes-count": action.MastersCount,
		"agent-nodes-count":  action.NodesCount - action.MastersCount,
	})

	// fmt.Println(e.Error())
	errors.AssertNoError(e, fmt.Errorf("unable to apply terraform primary"))

	masterIp, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "master-ip")
	masterIps, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "master-ips")
	agentIps, e := getOutputTerraformInFolder(fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID), "agent-ips")

	errors.AssertNoError(e, fmt.Errorf("unable to get cluster ip"))

	fmt.Println("master out", string(masterIp))
	fmt.Println(masterIps)
	fmt.Println(agentIps)

	e = i.setupMaster(string(masterIp))

	fmt.Println(e)
	errors.AssertNoError(e, fmt.Errorf("unable to setup master"))

	// applyCommand := exec.Command(
	// 	"terraform",
	// 	fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
	// 	"apply",
	// 	"-auto-approve",
	// 	fmt.Sprintf("-var=cluster-id=%v", action.ClusterID),
	// 	fmt.Sprintf("-var=master-nodes-count=%v", action.MastersCount),
	// 	fmt.Sprintf("-var=agent-nodes-count=%v", action.NodesCount-action.MastersCount),
	// 	fmt.Sprintf("-var=keys-path=%v", i.env.SshKeysPath),
	// 	fmt.Sprintf("-var=do-token=%v", i.env.DoAPIKey),
	// 	fmt.Sprintf("-var=do-image-id=%v", i.env.DoImageId),
	// )

	// applyCommand.Stdout = os.Stdout
	// applyCommand.Stderr = os.Stderr
	// e = applyCommand.Run()
	// errors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	// e = i.CreateKubeConfig(action.ClusterID)
	// errors.AssertNoError(e, fmt.Errorf("unable to create kubeconfig"))
	return e
}

func (i *infraClient) CreateKubeConfig(clusterId string) (e error) {

	defer errors.HandleErr(&e)
	out, e := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, clusterId),
		"output",
		"-json",
	).Output()

	errors.AssertNoError(e, fmt.Errorf("should run"))
	var outJson map[string]map[string]interface{}
	e = json.Unmarshal(out, &outJson)
	errors.AssertNoError(e, fmt.Errorf("should unmarshal output"))
	_, e = exec.Command(
		"scp",
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-i",
		fmt.Sprintf("%v/access", i.env.SshKeysPath),
		fmt.Sprintf("root@%v:/etc/rancher/k3s/k3s.yaml", outJson["cluster-ip"]["value"].(string)),
		fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId),
	).Output()
	errors.AssertNoError(e, fmt.Errorf("unable to download file"))
	input, e := ioutil.ReadFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId))
	errors.AssertNoError(e, fmt.Errorf("should read file"))
	newString := strings.ReplaceAll(string(input), "127.0.0.1", outJson["cluster-ip"]["value"].(string))
	e = ioutil.WriteFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId), []byte(newString), 0644)
	errors.AssertNoError(e, fmt.Errorf("should be able to change file"))
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

func (i *infraClient) SetupDOCSI(clusterId string) (e error) {
	defer errors.HandleErr(&e)
	if _, err := os.Stat("/tmp/do-api.yaml"); err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		buffer, e := ioutil.ReadFile("./internal/application/csi/do-secret/template.yaml")
		errors.AssertNoError(e, fmt.Errorf("template file should exist"))
		tpl := raymond.MustParse(string(buffer))
		result := tpl.MustExec(map[string]interface{}{
			"api_key": i.env.DoAPIKey,
		})
		fmt.Sprintf(result)
		e = ioutil.WriteFile("/tmp/do-api.yaml", []byte(result), 0644)
	} else {
		// should not happen
		return err
	}
	fmt.Println(fmt.Sprintf("%v/%v/kube.yaml", i.env.DataPath, clusterId))
	applyDoKeySecretCommand := exec.Command(
		"kubectl",
		"apply",
		"-f", "/tmp/do-api.yaml",
	)
	applyDoKeySecretCommand.Env = os.Environ()
	applyDoKeySecretCommand.Env = append(applyDoKeySecretCommand.Env, fmt.Sprintf("KUBECONFIG=%v/%v/k3s.yaml", i.env.DataPath, clusterId))
	fmt.Println(applyDoKeySecretCommand.Env)
	applyDoKeySecretCommand.Stdout = os.Stdout
	applyDoKeySecretCommand.Stderr = os.Stderr
	e = applyDoKeySecretCommand.Run()
	errors.AssertNoError(e, fmt.Errorf("unable to apply Digital Ocean Key"))

	applyCSICommand1 := exec.Command(
		"kubectl",
		"apply",
		"-f", "./internal/application/csi/crds.yaml",
	)
	applyCSICommand1.Env = os.Environ()
	applyCSICommand1.Env = append(applyDoKeySecretCommand.Env, fmt.Sprintf("KUBECONFIG=%v/%v/k3s.yaml", i.env.DataPath, clusterId))

	applyCSICommand1.Stdout = os.Stdout
	applyCSICommand1.Stderr = os.Stderr
	e = applyCSICommand1.Run()

	applyCSICommand2 := exec.Command(
		"kubectl",
		"apply",
		"-f", "./internal/application/csi",
	)
	applyCSICommand2.Env = os.Environ()
	applyCSICommand2.Env = append(applyDoKeySecretCommand.Env, fmt.Sprintf("KUBECONFIG=%v/%v/k3s.yaml", i.env.DataPath, clusterId))

	applyCSICommand2.Stdout = os.Stdout
	applyCSICommand2.Stderr = os.Stderr
	e = applyCSICommand2.Run()

	errors.AssertNoError(e, fmt.Errorf("unable to apply CSI"))
	return e
}

func (i *infraClient) SetupCSI(clusterId string, provider string) (e error) {
	switch provider {
	case "do":
		return i.SetupDOCSI(clusterId)
	default:
		return fmt.Errorf("provider not supported")
	}
}

func (i *infraClient) SetupOperator(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupMonitoring(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupIngress(clusterId string) error {
	applyHelm := exec.Command("bash", "./internal/application/ingress/init.sh", "install", "-f", "./internal/application/ingress/values.yaml")
	applyHelm.Env = os.Environ()
	applyHelm.Env = append(applyHelm.Env, fmt.Sprintf("KUBECONFIG=%v/%v/k3s.yaml", i.env.DataPath, clusterId))
	applyHelm.Stdout = os.Stdout
	applyHelm.Stderr = os.Stderr
	return applyHelm.Run()
}

func (i *infraClient) SetupWireguard(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func fxInfraClient(env *InfraEnv) domain.InfraClient {
	return &infraClient{
		env,
	}
}
