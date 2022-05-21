package functions

import (
	"bytes"
	"os/exec"

	"operators.kloudlite.io/lib/errors"
)

func KubectlApply(stdin ...[]byte) (stdout *bytes.Buffer, err error) {
	c := exec.Command("kubectl", "apply", "-f", "-")
	wStream, errStream := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
	c.Stdin = bytes.NewBuffer(bytes.Join(stdin, []byte("\n---\n")))
	c.Stdout = wStream
	c.Stderr = errStream
	if err := c.Run(); err != nil {
		return stdout, errors.NewEf(err, errStream.String())
	}
	return stdout, nil
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

func KubectlDelete(namespace, resourceRef string) error {
	c := exec.Command("kubectl", "delete", "-n", namespace, resourceRef)
	errB := bytes.NewBuffer([]byte{})
	outB := bytes.NewBuffer([]byte{})
	c.Stderr = errB
	c.Stdout = outB
	if err := c.Run(); err != nil {
		return errors.NewEf(err, errB.String())
	}
	return nil
}
