package agent

import (
	"bytes"
	"encoding/json"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Run(c *redpanda.Consumer, logger logging.Logger) {
	c.StartConsuming(
		func(m *redpanda.Message) error {
			logger.Infof("action=%s, payload=%s\n", m.Action, m.Payload)
			switch m.Action {
			case "apply", "delete":
				{
					c := exec.Command("kubectl", m.Action, "-f", "-")
					jb, err := json.Marshal(m.Payload)
					if err != nil {
						return errors.NewEf(err, "could not unmarshal into []byte")
					}
					yb, err := yaml.JSONToYAML(jb)
					if err != nil {
						return errors.NewEf(err, "could not convert JSON to YAML")
					}

					c.Stdin = bytes.NewBuffer(yb)
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					return c.Run()
				}
			case "restart":
				{
					b, err := json.Marshal(m.Payload)
					if err != nil {
						return err
					}
					var restartMsg RestartMsg
					if err := json.Unmarshal(b, &restartMsg); err != nil {
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
