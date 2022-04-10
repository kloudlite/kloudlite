package lib

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	typesbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"operators.kloudlite.io/lib/errors"
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
		return nil, errors.Newf("could not parse template %s: %w", filename, e)
	}

	return func(data interface{}) ([]byte, error) {
		w := new(bytes.Buffer)
		e = t.Execute(w, data)
		if e != nil {
			return nil, errors.Newf("could not execute template as %w", e)
		}
		return w.Bytes(), nil
	}, nil
}

func UseJobTemplate(data *JobVars) (*batchv1.Job, error) {
	fn, e := useTemplate("job-template.tmpl.yml")
	if e != nil {
		return nil, errors.NewE(e)
	}
	b, e := fn(data)
	if e != nil {
		return nil, errors.NewE(e)
	}

	var job batchv1.Job
	e = yaml.UnmarshalStrict(b, &job)
	if e != nil {
		return nil, errors.Newf("could not YAML unmarshal template jobVars because %w", e)
	}
	return &job, nil
}

type job struct {
	mgr func(ns string) typesbatchv1.JobInterface
}

func (j *job) Create(ctx context.Context, namespace string, jobVars *JobVars) (*batchv1.Job, error) {
	jobData, err := UseJobTemplate(jobVars)
	if err != nil {
		return nil, errors.NewEf(err, "could not *batchv1.Job from jobVars template")
	}

	kJob, err := j.mgr(namespace).Create(ctx, jobData, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewEf(err, "could not create job")
	}
	return kJob, nil
}

func (j *job) Get(ctx context.Context, namespace string, name string) (*batchv1.Job, error) {
	return j.mgr(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (j *job) Watch(ctx context.Context, namespace string, listOptions metav1.ListOptions) (bool, error) {
	watcher, err := j.mgr(namespace).Watch(context.Background(), listOptions)
	if err != nil {
		return false, errors.NewEf(err, "could not create job watcher")
	}
	defer watcher.Stop()

	for {
		result := <-watcher.ResultChan()
		j := result.Object.(*batchv1.Job)

		switch result.Type {
		case watch.Added:
			logrus.Infof("job (namespace=%s, name=%s) ADDED", j.Name, j.Namespace)

		case watch.Deleted:
			logrus.Infof("job (namespace=%s, name=%s) DELETED", j.Name, j.Namespace)

		case watch.Modified:
			logrus.Infof("job (namespace=%s, name=%s) MODIFIED", j.Name, j.Namespace)
			if j.Status.Succeeded > 0 {
				logrus.Infof("job (namespace=%s, name=%s) MODIFIED (COMPLETED)", j.Name, j.Namespace)
				return true, nil
			}

			if j.Status.Failed > 0 {
				logrus.Infof("job (namespace=%s, name=%s) MODIFIED (FAILED)", j.Name, j.Namespace)
				return false, nil
			}

		default:
			logrus.Infof("Unknown event type %v", result.Type)
			return false, errors.Newf("Unknown event type: %v", result.Type)
		}
	}
}

type Job interface {
	Create(ctx context.Context, namespace string, jobVars *JobVars) (*batchv1.Job, error)
	Watch(ctx context.Context, namespace string, listOptions metav1.ListOptions) (bool, error)
	Get(ctx context.Context, namespace string, name string) (*batchv1.Job, error)
}

func NewJobber(clientset *kubernetes.Clientset) Job {
	return &job{
		mgr: clientset.BatchV1().Jobs,
	}
}
