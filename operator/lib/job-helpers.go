package lib

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	errors "github.com/yext/yerrors"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typesbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"sigs.k8s.io/yaml"
)

type JobVars struct {
	Name            string
	Namespace       string
	ServiceAccount  string
	Image           string
	ImagePullPolicy string
	Args            []string
	Env             map[string]string
}

func useTemplate(filename string) (func(data interface{}) ([]byte, error), error) {
	tPath := path.Join(os.Getenv("PWD"), fmt.Sprintf("lib/templates/%s", filename))
	t, e := template.New(filename).Funcs(sprig.TxtFuncMap()).ParseFiles(tPath)
	if e != nil {
		return nil, errors.Wrap(errors.Errorf("could not parse template %s: %w", filename, e))
	}

	return func(data interface{}) ([]byte, error) {
		w := new(bytes.Buffer)
		e = t.Execute(w, data)
		if e != nil {
			return nil, errors.Wrap(errors.Errorf("could not execute template as %v", e))
		}
		return w.Bytes(), nil

	}, nil
}

func UseJobTemplate(data *JobVars) (*batchv1.Job, error) {
	fn, e := useTemplate("job-template.tmpl.yml")
	if e != nil {
		return nil, errors.Wrap(e)
	}
	b, e := fn(data)
	if e != nil {
		return nil, errors.Wrap(e)
	}

	var job batchv1.Job
	e = yaml.UnmarshalStrict(b, &job)
	if e != nil {
		return nil, errors.Wrap(e)
	}
	return &job, nil
}

func WatchJob(ctx context.Context, jobMgr typesbatchv1.JobInterface, listoptions metav1.ListOptions) (bool, error) {
	watcher, err := jobMgr.Watch(context.Background(), listoptions)
	if err != nil {
		return false, errors.Wrap(errors.Errorf("could not watch job because %W", err))
	}
	for {
		result := <-watcher.ResultChan()

		switch result.Type {
		case watch.Added:
			fmt.Println("(job) ADDED")

		case watch.Deleted:
			fmt.Println("(job) DELETED")

		case watch.Modified:
			fmt.Println("(job) MODIFIED")
			j := result.Object.(*batchv1.Job)

			if j.Status.Succeeded > 0 {
				fmt.Println("(job) COMPLETED")
				return true, nil
			}

			if j.Status.Failed > 0 {
				fmt.Println("(job) FAILED")
				return false, nil
			}

		default:
			return false, fmt.Errorf("Unknown event type: %v", result.Type)
		}
	}
}
