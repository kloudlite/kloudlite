package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
	"operators.kloudlite.io/lib/redpanda"
	"os"
	"os/exec"
	"sigs.k8s.io/yaml"
)

type RestartMsg struct {
	v1.TypeMeta `json:",inline"`
	Metadata    struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
}

type KafkaMessage struct {
	Action  string `json:"action"`
	Payload []byte `json:"payload"`
}

type ErrMessage struct {
	ResourceRef string `json:"resourceRef"`
	Action      string `json:"action"`
	Error       string `json:"error"`
}

func Run(c *redpanda.Consumer, errProducer *redpanda.Producer, errTopic string, logger logging.Logger) {
	c.StartConsuming(
		func(b []byte) error {
			var msg KafkaMessage
			if err := json.Unmarshal(b, &msg); err != nil {
				return err
			}
			logger.Infof("action=%s, payload=%s\n", msg.Action, msg.Payload)
			switch msg.Action {
			case "apply", "delete":
				{

					obj := unstructured.Unstructured{}
					if err := json.Unmarshal(msg.Payload, &obj); err != nil {
						return err
					}

					if errX := func() error {
						c := exec.Command("kubectl", msg.Action, "-f", "-")

						yb, err := yaml.JSONToYAML(msg.Payload)
						if err != nil {
							return errors.NewEf(err, "could not convert JSON to YAML")
						}

						c.Stdin = bytes.NewBuffer(yb)
						c.Stdout = os.Stdout
						errStream := bytes.NewBuffer([]byte{})
						c.Stderr = errStream
						if err := c.Run(); err != nil {
							return errors.NewEf(err, errStream.String())
						}
						return nil
					}(); errX != nil {
						errMsg := ErrMessage{
							ResourceRef: fmt.Sprintf(
								"Kind=%s/Namespace=%s/Name=%s",
								obj.GetObjectKind().GroupVersionKind().Kind,
								obj.GetNamespace(),
								obj.GetName(),
							),
							Action: msg.Action,
							Error:  errX.Error(),
						}
						b, errX := json.Marshal(errMsg)
						if err := errProducer.Produce(context.TODO(), errTopic, msg.Action, b); err != nil {
							return err
						}
						return errX
					}
				}
			case "restart":
				{
					var restartMsg RestartMsg
					if err := json.Unmarshal(msg.Payload, &restartMsg); err != nil {
						return err
					}

					switch restartMsg.GroupVersionKind() {
					case crdsv1.GroupVersion.WithKind("App"):
						{
							c := exec.Command(
								"kubectl", "rollout", "restart", "deployments",
								"-n", restartMsg.Metadata.Namespace,
								"-l", "kloudlite.io/app.name", restartMsg.Metadata.Name,
							)
							// kubectl rollout restart deployment -n hotspot -l 'kloudlite.io/app.name=auth-api'
							c.Stdout = nil
							c.Stderr = nil
							if err := c.Run(); err != nil {
								return errors.NewEf(err, "could not restart deployment")
							}
						}
					case crdsv1.GroupVersion.WithKind("Lambda"):
						{
							c := exec.Command(
								"kubectl", "rollout", "restart", "deployments",
								"-n", restartMsg.Metadata.Namespace,
								"-l", "serving.knative.dev/configuration", restartMsg.Metadata.Name,
							)
							// kubectl rollout restart deployment -n hotspot -l 'kloudlite.io/app.name=auth-api'
							c.Stdout = nil
							c.Stderr = nil
							if err := c.Run(); err != nil {
								return errors.NewEf(err, "could not restart deployment")
							}
						}
					case crdsv1.GroupVersion.WithKind("ManagedService"):
						{
							c := exec.Command(
								"kubectl", "rollout", "restart", "deployments",
								"-n", restartMsg.Metadata.Namespace,
								"-l", "kloudlite.io/msvc.name", restartMsg.Metadata.Name,
							)
							c.Stdout = nil
							c.Stderr = nil
							if err := c.Run(); err != nil {
								return errors.NewEf(err, "could not restart deployment")
							}
						}
					}
				}
			default:
				{
				}
			}
			return nil
		},
	)
}
