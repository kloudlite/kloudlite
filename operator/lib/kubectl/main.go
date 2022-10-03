package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"

	"operators.kloudlite.io/lib/errors"
)

type batchable string

const (
	Deployments  batchable = "deployments"
	Statefulsets batchable = "statefulsets"
)

func Restart(kind batchable, namespace string, labels map[string]string) (int, error) {
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
