package kubectl

import (
	"bytes"
	"fmt"
	"operators.kloudlite.io/lib/errors"
	"os/exec"
)

type restartable string

const (
	Deployments  restartable = "deployments"
	Statefulsets restartable = "statefulsets"
)

func Restart(kind restartable, namespace string, labels map[string]string) error {
	cmdArgs := []string{
		"rollout", "restart", string(kind),
		"-n", namespace,
	}
	for k, v := range labels {
		cmdArgs = append(cmdArgs, "-l", fmt.Sprintf("%s=%s", k, v))
	}

	// sample cmd: kubectl rollout restart deployment -n hotspot -l 'kloudlite.io/app.name=auth-api'
	c := exec.Command("kubectl", cmdArgs...)
	errStream := bytes.NewBuffer([]byte{})
	c.Stdout = nil
	c.Stderr = errStream
	if err := c.Run(); err != nil {
		return errors.NewEf(err, "could not restart deployment, because %s", errStream.String())
	}
	return nil
}
