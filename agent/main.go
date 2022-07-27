package agent

import (
	"bytes"
	"context"
	"encoding/json"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload"`
}

type ErrMessage struct {
	Error   string         `json:"error"`
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload"`
}

func Run(c *redpanda.Consumer, errProducer *redpanda.Producer, errTopic string, logger logging.Logger) {
	c.StartConsuming(
		func(b []byte, key []byte) error {

			var msg KafkaMessage
			if err := json.Unmarshal(b, &msg); err != nil {
				logger.Errorf(err, "error when unmarshalling []byte to kafkaMessage : %s", b)
				return err
			}
			logger.Infof("action=%s, payload=%s\n", msg.Action, msg.Payload)

			switch msg.Action {
			case "apply", "delete":
				{
					if errX := func() error {
						c := exec.Command("kubectl", msg.Action, "-f", "-")

						pb, err := json.Marshal(msg.Payload)
						if err != nil {
							return errors.NewEf(err, "could not convert msg.Payload into []byte")
						}
						yb, err := yaml.JSONToYAML(pb)
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
							Action:  msg.Action,
							Error:   errX.Error(),
							Payload: msg.Payload,
						}
						b, err := json.Marshal(errMsg)
						if err != nil {
							logger.Errorf(err, "error marshalling ErrMessage to []byte")
							return err
						}
						if err := errProducer.Produce(context.TODO(), errTopic, string(key), b); err != nil {
							return err
						}
						return errX
					}
				}
			default:
				{
					logger.Errorf(nil, "Invalid Action: %s", msg.Action)
				}
			}
			return nil
		},
	)
}
