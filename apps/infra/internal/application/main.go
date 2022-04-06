package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aymerick/raymond"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/config"
	klerrors "kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/messaging"

	"os"
	"os/exec"
)

type infraClient struct {
	env *InfraEnv
}

func (i *infraClient) CreateKubernetes(action domain.SetupClusterAction) (e error) {
	defer klerrors.HandleErr(&e)
	copyTemplateDirCommand := exec.Command(
		"cp",
		"-r",
		fmt.Sprintf("./internal/application/tf/%v", action.Provider),
		fmt.Sprintf("%v/%v", i.env.DataPath, action.ClusterID),
	)
	copyTemplateDirCommand.Stdout = os.Stdout
	copyTemplateDirCommand.Stderr = os.Stderr
	e = copyTemplateDirCommand.Run()
	klerrors.AssertNoError(e, fmt.Errorf("unable to copy template directory"))
	initCommand := exec.Command(
		"terraform",
		fmt.Sprintf("-chdir=%v/%v", i.env.DataPath, action.ClusterID),
		"init",
	)
	initCommand.Stdout = os.Stdout
	initCommand.Stderr = os.Stderr
	e = initCommand.Run()
	klerrors.AssertNoError(e, fmt.Errorf("failed to init terraform check"))
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
	klerrors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	e = i.CreateKubeConfig(action.ClusterID)
	klerrors.AssertNoError(e, fmt.Errorf("unable to create kubeconfig"))
	return e
}

func (i *infraClient) CreateKubeConfig(clusterId string) (e error) {
	fmt.Println("creating kube config")
	defer klerrors.HandleErr(&e)
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
	klerrors.AssertNoError(e, fmt.Errorf("should run"))
	var outJson map[string]map[string]interface{}
	e = json.Unmarshal(out, &outJson)
	klerrors.AssertNoError(e, fmt.Errorf("should unmarshal output"))
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
	klerrors.AssertNoError(e, fmt.Errorf("unable to download file"))
	input, e := ioutil.ReadFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId))
	klerrors.AssertNoError(e, fmt.Errorf("should read file"))
	newString := strings.ReplaceAll(string(input), "127.0.0.1", outJson["cluster-ip"]["value"].(string))
	e = ioutil.WriteFile(fmt.Sprintf("%v/%v/k3s.yaml", i.env.DataPath, clusterId), []byte(newString), 0644)
	klerrors.AssertNoError(e, fmt.Errorf("should be able to change file"))
	return e
}

func (i *infraClient) UpdateKubernetes(action domain.UpdateClusterAction) (e error) {
	defer klerrors.HandleErr(&e)
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
	klerrors.AssertNoError(e, fmt.Errorf("failed to apply terraform"))
	return e
}

func (i *infraClient) SetupDOCSI(clusterId string) (e error) {
	defer klerrors.HandleErr(&e)
	if _, err := os.Stat("/tmp/do-api.yaml"); err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		buffer, e := ioutil.ReadFile("./internal/application/csi/do-secret/template.yaml")
		klerrors.AssertNoError(e, fmt.Errorf("template file should exist"))
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
	klerrors.AssertNoError(e, fmt.Errorf("unable to apply Digital Ocean Key"))

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

	klerrors.AssertNoError(e, fmt.Errorf("unable to apply CSI"))
	return e
}

func (i *infraClient) SetupCSI(clusterId string, provider string) (e error) {
	switch provider {
	case "do":
		return i.SetupDOCSI(clusterId)
	default:
		return errors.New("provider not supported")
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

func newInfraClient(env *InfraEnv) domain.TF {
	return &infraClient{
		env,
	}
}

type InfraEnv struct {
	DoAPIKey    string `env:"DO_API_KEY", required:"true"`
	DataPath    string `env:"DATA_PATH", required:"true"`
	SshKeysPath string `env:"SSH_KEYS_PATH", required:"true"`
}

var Module = fx.Module("application",
	// Load Env
	fx.Provide(func() (*InfraEnv, error) {
		var envC InfraEnv
		err := config.LoadConfigFromEnv(&envC)
		return &envC, err
	}),

	fx.Provide(newInfraClient),
	domain.Module,

	fx.Provide(func(env *Env, d domain.Domain) (messaging.Consumer[domain.SetupClusterAction], error) {
		return messaging.NewKafkaConsumer[domain.SetupClusterAction](
			[]string{env.KafkaInfraActionTopic},
			env.KafkaBrokers,
			env.KafkaGroupId,
			func(topic string, action domain.SetupClusterAction) error {
				d.CreateCluster(action)
				return nil
			},
		)
	}),

	fx.Invoke(
		func(lf fx.Lifecycle, msgConsumer messaging.Consumer[domain.SetupClusterAction]) error {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return msgConsumer.Subscribe()
				},
				OnStop: func(ctx context.Context) error {
					return msgConsumer.Unsubscribe()
				},
			})
			return nil
		},
	)
)
