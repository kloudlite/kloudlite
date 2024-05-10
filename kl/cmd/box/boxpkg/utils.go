package boxpkg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

type Container struct {
	Name string
	Path string
}

func (c *client) getContainer() (Container, error) {
	defCr := Container{}

	crlist, err := c.cli.ContainerList(c.cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: "kl-box=true"},
		),
		All: true,
	})
	if err != nil {
		return defCr, err
	}

	if len(crlist) >= 1 {
		if len(crlist[0].Names) >= 1 {

			defCr.Name = crlist[0].Names[0]
			defCr.Path = crlist[0].Labels["kl-box-cwd"]

			if strings.Contains(defCr.Name, "/") {
				s := strings.Split(defCr.Name, "/")
				if len(s) >= 1 {
					defCr.Name = s[1]
				}
			}

			return defCr, nil
		}

		defCr.Name = crlist[0].ID
		defCr.Path = crlist[0].Labels["kl-box-cwd"]

		return defCr, nil
	}

	return defCr, nil
}

func (c *client) ensurePublicKey() error {
	sshPath := path.Join(xdg.Home, ".ssh")
	if _, err := os.Stat(path.Join(sshPath, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path.Join(sshPath, "id_rsa"), "-N", "")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *client) ensureCacheExist() error {
	vlist, err := c.cli.VolumeList(c.cmd.Context(), volume.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "kl-box-nix-store=true",
		}),
	})
	if err != nil {
		return err
	}

	if len(vlist.Volumes) == 0 {
		if _, err := c.cli.VolumeCreate(c.cmd.Context(), volume.CreateOptions{
			Labels: map[string]string{
				"kl-box-nix-store": "true",
			},
			Name: "nix-store",
		}); err != nil {
			return err
		}
	}

	vlist, err = c.cli.VolumeList(c.cmd.Context(), volume.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "kl-box-nix-home-cache=true",
		}),
	})

	if err != nil {
		return err
	}

	if len(vlist.Volumes) == 0 {
		if _, err := c.cli.VolumeCreate(c.cmd.Context(), volume.CreateOptions{
			Labels: map[string]string{
				"kl-box-nix-home-cache": "true",
			},
			Name: "nix-home-cache",
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) startContainer(klConfig KLConfigType, td string) error {
	conf, err := json.Marshal(klConfig)
	if err != nil {
		return err
	}

	dockerArgs := []string{"run"}
	if !c.foreground {
		dockerArgs = append(dockerArgs, "-d")
	}

	sshPath := path.Join(xdg.Home, ".ssh", "id_rsa.pub")

	akByte, err := os.ReadFile(sshPath)
	if err != nil {
		return err
	}

	ak := string(akByte)

	akTmpPath := path.Join(td, "authorized_keys")

	akByte, err = os.ReadFile(path.Join(xdg.Home, ".ssh", "authorized_keys"))
	if err == nil {
		ak += fmt.Sprint("\n", string(akByte))
	}

	if err := os.WriteFile(akTmpPath, []byte(ak), fs.ModePerm); err != nil {
		return err
	}

	switch runtime.GOOS {
	case constants.RuntimeWindows:
		fn.Warn("docker support inside container not implemented yet")
	default:
		dockerArgs = append(dockerArgs, "-v", "/var/run/docker.sock:/var/run/docker.sock:ro")
	}

	cwd, _ := os.Getwd()

	stdErrPath := fmt.Sprintf("%s/stderr.log", td)
	stdOutPath := fmt.Sprintf("%s/stdout.log", td)

	if err := os.WriteFile(stdOutPath, []byte(""), os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(stdErrPath, []byte(""), os.ModePerm); err != nil {
		return err
	}

	dockerArgs = append(dockerArgs,
		"--name", c.containerName,
		"--label", "kl-box=true",
		"--label", fmt.Sprintf("kl-box-cwd=%s", cwd),
		"-v", fmt.Sprintf("%s:/tmp/ssh2/authorized_keys:ro", akTmpPath),
		"-v", "kl-home-cache:/home:rw",
		"-v", "nix-store:/nix:rw",
		"--hostname", "box",
		"-v", fmt.Sprintf("%s:/home/kl/workspace:rw", cwd),
		"-v", fmt.Sprintf("%s:/tmp/stdout.log:rw", stdOutPath),
		"-v", fmt.Sprintf("%s:/tmp/stderr.log:rw", stdErrPath),
		"-p", "1729:22",
		ImageName, "--", string(conf),
	)

	command := exec.Command("docker", dockerArgs...)

	if c.verbose {
		command.Stdout = os.Stdout
	}
	command.Stderr = os.Stderr

	if c.verbose {
		fn.Logf("docker container started with cmd: %s\n", text.Blue(command.String()))
	}

	if _, err := c.cli.ImagePull(c.Context(), ImageName, image.PullOptions{}); err != nil {
		return err
	}

	if err := command.Run(); err != nil {
		return fmt.Errorf("error running kl-box container [%s]", err.Error())
	}

	return nil
}
