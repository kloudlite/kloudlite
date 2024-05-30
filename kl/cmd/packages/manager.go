package packages

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func execPackageCommand(cmd string) error {

	if err := fn.CopyFile("/home/kl/workspace/kl.lock", "/home/kl/.kl/devbox/devbox.lock"); err != nil {
		fn.Warn(err)
	}

	kt, err := client.GetKlFile("")
	if err != nil {
		return err
	}

	b2, err := kt.ToJson()
	if err != nil {
		return err
	}

	if err := os.WriteFile("/home/kl/.kl/devbox/devbox.json", b2, os.ModePerm); err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(cmd))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	command := exec.Command(cmdArr[0], cmdArr[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = "/home/kl/.kl/devbox"

	if err = command.Run(); err != nil {
		return err
	}

	b, err := os.ReadFile("/home/kl/.kl/devbox/devbox.json")
	if err != nil {
		return err
	}

	type KLConfigType struct {
		Packages []string `yaml:"packages" json:"packages"`
	}

	devbox := &KLConfigType{}
	if err := json.Unmarshal(b, devbox); err != nil {
		return err
	}

	kt.Packages = devbox.Packages

	if err := client.WriteKLFile(*kt); err != nil {
		return err
	}

	if err := fn.CopyFile("/home/kl/.kl/devbox/devbox.lock", "/home/kl/workspace/kl.lock"); err != nil {
		fn.Warn(err)
	}

	return client.UpdateDevboxEnvs()
}
