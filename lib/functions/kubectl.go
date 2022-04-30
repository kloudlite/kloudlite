package functions

import (
	"bytes"
	"os"
	"os/exec"
)

func KubectlApply(stdin ...[]byte) error {
	c := exec.Command("kubectl", "apply", "-f", "-")
	c.Stdin = bytes.NewBuffer(bytes.Join(stdin, []byte("\n---\n")))
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
