package domain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/Masterminds/sprig"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/shared"
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
	case shared.RESOURCE_PROJECT:
		{
			if spec, ok := msg.Spec.(Project); ok {
				bData, e := d.applyTemplate("project.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Project)")
		}
	case shared.RESOURCE_CONFIG:
		{
			if spec, ok := msg.Spec.(Config); ok {
				bData, e := d.applyTemplate("configs.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Config)")
		}
	case shared.RESOURCE_SECRET:
		{
			if spec, ok := msg.Spec.(Secret); ok {
				bData, e := d.applyTemplate("secrets.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(Secret)")
		}
	case shared.RESOURCE_GIT_PIPELINE:
		{
		}
	case shared.RESOURCE_APP:
		{
			if spec, ok := msg.Spec.(App); ok {
				bData, e := d.applyTemplate("app.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(App)")
		}
	case shared.RESOURCE_MANAGED_SERVICE:
		{
			if spec, ok := msg.Spec.(ManagedSvc); ok {
				bData, e := d.applyTemplate("msvc.tmpl.yml", spec)
				fmt.Printf("error (%v) Data (%v)\n", e, string(bData))
				return kubeApply(bData)
			}
			return errors.New("malformed spec not of type(ManagedSvc)")
		}
	case shared.RESOURCE_MANAGED_RESOURCE:
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
