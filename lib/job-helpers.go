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
	Command         []string
	Args            []string
	Env             map[string]string
}

func useTemplate(filename string) (func(data interface{}) ([]byte, error), error) {
	tPath := path.Join(os.Getenv("PWD"), fmt.Sprintf("lib/templates/%s", filename))
	t, err := template.New(filename).Funcs(sprig.TxtFuncMap()).ParseFiles(tPath)
	if err != nil {
		return nil, errors.NewEf(err, "could not parse template %s", filename)
	}

	return func(data interface{}) ([]byte, error) {
		w := new(bytes.Buffer)
		err = t.Execute(w, data)
		if err != nil {
			return nil, errors.NewEf(err, "could not execute template")
		}
		return w.Bytes(), nil
	}, nil
}

func UseJobTemplate(data *JobVars) (*batchv1.Job, error) {
	fn, err := useTemplate("job-template.tmpl.yml")
	if err != nil {
		return nil, errors.NewE(err)
	}
	b, err := fn(data)
	if err != nil {
		return nil, errors.NewE(err)
	}

	var job batchv1.Job
	err = yaml.UnmarshalStrict(b, &job)
	if err != nil {
		return nil, errors.NewEf(err, "could not YAML unmarshal template jobVars")
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

func (j *job) Delete(ctx context.Context, namespace string, name string) error {
	err := j.mgr(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewEf(err, "could not delete job")
	}
	return nil
}

func (j *job) Get(ctx context.Context, namespace string, name string) (*batchv1.Job, error) {
	gJob, err := j.mgr(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewEf(err, "could not get job")
	}
	return gJob, nil
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

func (j *job) HasFailed(ctx context.Context, namespace string, name string) (*batchv1.JobCondition, error) {
	rJob, err := j.mgr(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewEf(err, "could not get job")
	}
	for _, condition := range rJob.Status.Conditions {
		if condition.Type == batchv1.JobFailed {
			return &condition, nil
		}
	}
	return nil, nil
}

func newBool(b bool) *bool {
	return &b
}

func (j *job) HasSucceeded(ctx context.Context, namespace string, name string) (*bool, error) {
	rJob, err := j.mgr(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewEf(err, "could not get job")
	}
	for _, condition := range rJob.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			return newBool(condition.Status == "True"), nil
		}
	}
	return nil, nil
}

func (j *job) HasCompleted(ctx context.Context, namespace string, name string) (*bool, error) {
	rJob, err := j.mgr(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewEf(err, "could not get job")
	}
	for _, condition := range rJob.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			return newBool(condition.Status == "True"), nil
		}
		if condition.Type == batchv1.JobFailed {
			return newBool(!(condition.Status == "True")), nil
		}
	}
	return nil, nil
}

type Job interface {
	Create(ctx context.Context, namespace string, jobVars *JobVars) (*batchv1.Job, error)
	Delete(ctx context.Context, namespace string, name string) error
	Watch(ctx context.Context, namespace string, listOptions metav1.ListOptions) (bool, error)
	Get(ctx context.Context, namespace string, name string) (*batchv1.Job, error)
	HasSucceeded(ctx context.Context, namespace string, name string) (*bool, error)
	HasFailed(ctx context.Context, namespace string, name string) (*batchv1.JobCondition, error)
	HasCompleted(ctx context.Context, namespace string, name string) (*bool, error)
}

func NewJobber(clientset *kubernetes.Clientset) Job {
	return &job{
		mgr: clientset.BatchV1().Jobs,
	}
}
