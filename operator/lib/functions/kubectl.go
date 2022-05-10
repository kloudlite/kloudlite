package functions

import (
	"bytes"
	"os"
	"os/exec"

	"operators.kloudlite.io/lib/errors"
)

func KubectlApply(stdin ...[]byte) error {
	c := exec.Command("kubectl", "apply", "-f", "-")
	errB := bytes.NewBuffer([]byte{})
	c.Stdin = bytes.NewBuffer(bytes.Join(stdin, []byte("\n---\n")))
	c.Stdout = os.Stdout
	c.Stderr = errB
	if err := c.Run(); err != nil {
		return errors.NewEf(err, errB.String())
	}
	return nil
}

func KubectlGet(namespace string, resourceRef string) ([]byte, error) {
	c := exec.Command("kubectl", "get", "-o", "json", "-n", namespace, resourceRef)
	errB := bytes.NewBuffer([]byte{})
	outB := bytes.NewBuffer([]byte{})
	c.Stderr = errB
	c.Stdout = outB
	if err := c.Run(); err != nil {
		return nil, errors.NewEf(err, errB.String())
	}
	return outB.Bytes(), nil
}
