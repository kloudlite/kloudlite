package application

import (
	"encoding/json"
	"fmt"
	"github.com/aymerick/raymond"
	"github.com/pkg/errors"
	"io/ioutil"
	"kloudlite.io/apps/infra/internal/domain"
	libErrors "kloudlite.io/pkg/lib-errors"
	"os"
	"os/exec"
	"strings"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) CreateKubernetes(action domain.SetupClusterAction) (e error) {
	defer libErrors.HandleErr(&e)
	copyTemplateDirCommand := exec.Command(
		"cp",
		"-r",
		fmt.Sprintf("./internal/application/tf/%v", action.Provider),
		fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID),
	)
	copyTemplateDirCommand.Stdout = os.Stdout
	copyTemplateDirCommand.Stderr = os.Stderr
	e = copyTemplateDirCommand.Run()
	libErrors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))
	initCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
		"init",
	)
	initCommand.Stdout = os.Stdout
	initCommand.Stderr = os.Stderr
	e = initCommand.Run()
	libErrors.AssertNoError(e, fmt.Errorf("failed to init terraform check"))
	applyCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
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
	libErrors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	e = i.CreateKubeConfig(action.ClusterID)
	libErrors.AssertNoError(e, fmt.Errorf("unable to create kubeconfig"))
	return e
}

func (i *infraClient) CreateKubeConfig(clusterId string) (e error) {
	fmt.Println("creating kube config")
	defer libErrors.HandleErr(&e)
	fmt.Println("terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, clusterId),
		"output",
		"-json")
	out, e := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, clusterId),
		"output",
		"-json",
	).Output()
	libErrors.AssertNoError(e, fmt.Errorf("should run"))
	var outJson map[string]map[string]interface{}
	e = json.Unmarshal(out, &outJson)
	libErrors.AssertNoError(e, fmt.Errorf("should unmarshal output"))
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
	libErrors.AssertNoError(e, fmt.Errorf("unable to download file"))
	input, e := ioutil.ReadFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId))
	libErrors.AssertNoError(e, fmt.Errorf("should read file"))
	newString := strings.ReplaceAll(string(input), "127.0.0.1", outJson["cluster-ip"]["value"].(string))
	e = ioutil.WriteFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId), []byte(newString), 0644)
	libErrors.AssertNoError(e, fmt.Errorf("should be able to change file"))
	return e
}

func (i *infraClient) UpdateKubernetes(action domain.UpdateClusterAction) (e error) {
	defer libErrors.HandleErr(&e)
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
	libErrors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	return e
}

func (i *infraClient) SetupDOCSI(clusterId string) (e error) {
	defer libErrors.HandleErr(&e)
	if _, err := os.Stat("/tmp/do-api.yaml"); err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		buffer, e := ioutil.ReadFile("./internal/application/csi/do-secret/template.yaml")
		libErrors.AssertNoError(e, fmt.Errorf("template file should exist"))
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
	libErrors.AssertNoError(e, fmt.Errorf("unable to apply Digital Ocean Key"))

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

	libErrors.AssertNoError(e, fmt.Errorf("unable to apply CSI"))
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
	//TODO implement me
	panic("implement me")
}

func (i *infraClient) SetupWireguard(clusterId string) error {
	//TODO implement me
	panic("implement me")
}

func makeInfraClient(env *InfraEnv) domain.InfraClient {
	return &infraClient{
		env,
	}
}
