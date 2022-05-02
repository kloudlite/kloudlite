package functions

import (
	"bytes"
	"operators.kloudlite.io/lib/errors"
	"os"
	"os/exec"
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
