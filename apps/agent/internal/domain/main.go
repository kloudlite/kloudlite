package domain

import (
	"bytes"
	"fmt"
	"kloudlite.io/common"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/Masterminds/sprig"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
)

type domain struct {
	applyTemplate func(filename string, data any) ([]byte, error)
}

func kubeApply(b []byte) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBuffer(b)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (d *domain) ProcessMessage(msg *Message) error {
	switch msg.ResourceType {
	case common.ResourceProject:
		{
			if spec, ok := msg.Spec.(Project); ok {
				bData, e := d.applyTemplate("project.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
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
	ProcessMessage(msg *Message) error
}

var fxDomain = func(logger logger.Logger) Domain {
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
	}
}

var Module = fx.Module("domain",
	fx.Provide(fxDomain),
)
