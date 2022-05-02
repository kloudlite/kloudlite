package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/fx"
	"operators.kloudlite.io/lib/errors"
	"sigs.k8s.io/yaml"
)

type M map[string]interface{}

const MaxRetries = 2

func processMesssage(action string, payload map[string]interface{}) error {
	fmt.Println("#########################")
	fmt.Printf("action: %s, payload %+v\n", action, payload)
	fmt.Println("#########################")
	c := exec.Command("kubectl", action, "-f", "-")
	jb, err := json.Marshal(payload)
	if err != nil {
		return errors.NewEf(err, "could not unmarshal into []byte")
	}
	yb, err := yaml.JSONToYAML([]byte(jb))
	if err != nil {
		return errors.NewEf(err, "could not convert JSON to YAML")
	}

	c.Stdin = bytes.NewBuffer(yb)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

type Message struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

func readMessages(kConsumer *kafka.Consumer) {
	for {
		msg, err := kConsumer.ReadMessage(-1)
		if msg == nil {
			continue
		}
		fmt.Println("###############  Msg received", string(msg.Value))
		if err != nil {
			fmt.Printf("could not read kafka message because %v\n", err)
		}

		hasProcessed := false
		for rCount := 0; rCount < MaxRetries; rCount++ {
			var j Message
			if err2 := json.Unmarshal(msg.Value, &j); err2 != nil {
				fmt.Println(err2)
				break
			}
			fmt.Printf("ACTION %s received\n", j.Action)
			err = processMesssage(j.Action, j.Payload)
			if err != nil {
				fmt.Println("Retrying ....")
				continue
			}
			hasProcessed = true
		}
		if !hasProcessed {
			fmt.Printf("could not process message even after %d retries\n", MaxRetries)
		}
		kConsumer.CommitMessage(msg)
		fmt.Println("committed msg...")
	}
}

var Module = fx.Module("app",
	fx.Invoke(func(lf fx.Lifecycle, consumer *kafka.Consumer) {
		lf.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					readMessages(consumer)
				}()
				return nil
			},
			OnStop: func(context.Context) error {
				return consumer.Close()
			},
		})
	}),
)
