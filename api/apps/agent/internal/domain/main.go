package domain

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"

	"kloudlite.io/common"

	"github.com/Masterminds/sprig"
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaReplyTopic string `env:"KAFKA_REPLY_TOPIC" required:"true"`
}

type domain struct {
	applyTemplate func(filename string, data any) ([]byte, error)
	logger        logger.Logger
	producer      messaging.Producer[MessageReply]
	msgTopic      string
}

func kubeApply(b []byte) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBuffer(b)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Message) bucket() string {
	return fmt.Sprintf("%s/%s", m.Namespace, m.ResourceType)
}

func (d *domain) reply(msg *Message, status bool, description string) error {
	return d.producer.SendMessage(
		d.msgTopic,
		msg.bucket(),
		MessageReply{
			Message: description,
			Status:  status,
		},
	)
}

func (d *domain) ProcessMessage(ctx context.Context, msg *Message) error {
	if msg == nil {
		return errors.New("empty message received")
	}
	d.logger.Debugf("MSG: %+v\n", *msg)
	switch msg.ResourceType {
	case common.ResourceProject:
		{
			if spec, ok := msg.Spec.(Project); ok {
				bData, err := d.applyTemplate("project.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", err, string(bData))
				err = kubeApply(bData)
				if err != nil {
					return d.reply(msg, false, "failed to apply resource")
				}
				return d.reply(msg, true, "applied resource, keep listening for resource updates ...")
			}

			return errors.New("malformed spec not of type(Project)")
		}
	case common.ResourceConfig:
		{
			if spec, ok := msg.Spec.(Config); ok {
				bData, e := d.applyTemplate("configs.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Config)")
		}
	case common.ResourceSecret:
		{
			if spec, ok := msg.Spec.(Secret); ok {
				bData, e := d.applyTemplate("secrets.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Secret)")
		}

	case common.ResourceRouter:
		{
			if spec, ok := msg.Spec.(Router); ok {
				bData, e := d.applyTemplate("router.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Router)")
		}

	case common.ResourceGitPipeline:
		{
			if spec, ok := msg.Spec.(Pipeline); ok {
				bData, e := d.applyTemplate("pipeline.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Pipeline)")
		}
	case common.ResourceApp:
		{
			if spec, ok := msg.Spec.(App); ok {
				bData, e := d.applyTemplate("app.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(App)")
		}
	case common.ResourceManagedService:
		{
			if spec, ok := msg.Spec.(ManagedSvc); ok {
				bData, e := d.applyTemplate("msvc.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(ManagedSvc)")
		}
	case common.ResourceManagedResource:
		{
			if spec, ok := msg.Spec.(ManagedRes); ok {
				bData, e := d.applyTemplate("mres.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(ManagedRes)")
		}
	}
	return nil
}

type Domain interface {
	ProcessMessage(ctx context.Context, msg *Message) error
}

var fxDomain = func(logger logger.Logger, env *Env, producer messaging.Producer[MessageReply]) Domain {
	applyTemplate := func(filename string, data any) ([]byte, error) {
		tPath := path.Join(os.Getenv("PWD"), fmt.Sprintf("internal/domain/templates/%s", filename))
		w := new(bytes.Buffer)
		t, e := template.New(filename).Funcs(sprig.TxtFuncMap()).ParseFiles(tPath)
		if e != nil {
			logger.Errorf("could not parse template files as %v", e)
			return nil, e
		}
		e = t.Execute(w, data)
		if e != nil {
			logger.Errorf("could not execute template as %v", e)
			return nil, e
		}
		return w.Bytes(), nil
	}

	return &domain{
		applyTemplate,
		logger,
		producer,
		env.KafkaReplyTopic,
	}
}

var Module = fx.Module("domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
	fx.Invoke(func(lf fx.Lifecycle, producer messaging.Producer[MessageReply]) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return producer.Connect(ctx)
			},
			OnStop: func(ctx context.Context) error {
				producer.Close(ctx)
				return nil
			},
		})
	}),
)
