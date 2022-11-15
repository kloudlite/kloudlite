package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"operators.kloudlite.io/lib/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type batchable string

const (
	Deployments  batchable = "deployments"
	Statefulsets batchable = "statefulsets"
)

func RolloutRestart(c client.Client, kind Restartable, namespace string, labels map[string]string) error {
	switch kind {
	case Deployment:
		{
			ctx, cancelFn := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancelFn()
			var dl appsv1.DeploymentList
			if err := c.List(
				ctx, &dl, &client.ListOptions{
					Namespace:     namespace,
					LabelSelector: apiLabels.SelectorFromValidatedSet(labels),
				},
			); err != nil {
				return err
			}

			for _, d := range dl.Items {
				if d.Spec.Template.ObjectMeta.Annotations == nil {
					d.Spec.Template.ObjectMeta.Annotations = map[string]string{}
				}
				// [source] (https://stackoverflow.com/a/59051313)
				d.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if err := c.Update(ctx, &d); err != nil {
					return err
				}
			}
		}
	case StatefulSet:
		{
			ctx, cancelFn := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancelFn()

			var sl appsv1.StatefulSetList

			if err := c.List(
				ctx, &sl, &client.ListOptions{
					Namespace:     namespace,
					LabelSelector: apiLabels.SelectorFromValidatedSet(labels),
				},
			); err != nil {
				return err
			}

			for _, d := range sl.Items {
				if d.Spec.Template.ObjectMeta.Annotations == nil {
					d.Spec.Template.ObjectMeta.Annotations = map[string]string{}
				}
				d.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				if err := c.Update(ctx, &d); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func RestartWithKubectl(kind batchable, namespace string, labels map[string]string) (int, error) {
	cmdArgs := []string{
		"rollout", "restart", string(kind),
		"-n", namespace,
	}
	for k, v := range labels {
		cmdArgs = append(cmdArgs, "-l", fmt.Sprintf("%s=%s", k, v))
	}

	// sample cmd: kubectl rollout restart deployment -n hotspot -l 'kloudlite.io/app-n-lambda.name=auth-api'
	c := exec.Command("kubectl", cmdArgs...)
	errStream := bytes.NewBuffer([]byte{})
	c.Stdout = nil
	c.Stderr = errStream
	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), errors.NewEf(err, "could not restart %s, because %s", kind, errStream.String())
		}
	}
	return 0, nil
}

func Scale(kind batchable, namespace string, labels map[string]string, count int) (int, error) {
	cmdArgs := []string{
		"scale", "--replicas", fmt.Sprintf("%d", count),
		"-n", namespace,
		string(kind),
	}
	for k, v := range labels {
		cmdArgs = append(cmdArgs, "-l", fmt.Sprintf("%s=%s", k, v))
	}

	// sample cmd: kubectl rollout restart deployment -n hotspot -l 'kloudlite.io/app-n-lambda.name=auth-api'
	c := exec.Command("kubectl", cmdArgs...)
	errStream := bytes.NewBuffer([]byte{})
	c.Stdout = nil
	c.Stderr = errStream
	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), errors.NewEf(err, "could not restart %s, because %s", kind, errStream.String())
		}
	}
	return 0, nil
}
