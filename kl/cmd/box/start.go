package box

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/kloudlite/kl/domain/server"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var (
	foreground bool
	debug      bool
)

type EnvironmentVariable struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

type KLConfigType struct {
	Packages []string              `yaml:"packages" json:"packages"`
	EnvVars  []EnvironmentVariable `yaml:"envVars" json:"envVars"`
	Mounts   map[string]string     `yaml:"mounts" json:"mounts"`
}

type VolMount struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// type FileMounts struct {
// 	MountBasePath string     `yaml:"mountbasepath" json:"mountbasepath"`
// 	Mounts        []VolMount `yaml:"mounts" json:"mounts"`
// }

type FileMountType struct {
	FileMount client.MountType `yaml:"filemount" json:"filemount"`
}

// var fm FM

var imageName string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := startBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
		return
	},
}

func startBox(_ *cobra.Command, args []string) error {
	flag.BoolVar(&foreground, "foreground", false, "--foreground")
	flag.Parse()

	if isPortInUse("1729") {
		fn.Log("Port 1729 is not being used by any other container. Please stop that container first.")
		return nil
	}

	imageName = constants.BoxDockerImage
	if len(args) > 0 {
		imageName = args[0]
	}

	if err := fn.ExecCmd(fmt.Sprintf("docker pull %s", imageName), nil, false); err != nil {
		return err
	}

	fn.Log("starting container...")

	{
		// Global setup
		ensurePublicKey()
		ensureCacheExist()
	}

	{

		envs, mmap, err := server.GetLoadMaps()
		if err != nil {
			return err
		}

		// local setup
		kConf, err := loadConfig()
		if err != nil {
			return err
		}

		fMounts, err := loadFileMount(mmap)
		if err != nil {
			return err
		}

		var ev []EnvironmentVariable
		for k, v := range envs {
			ev = append(ev, EnvironmentVariable{k, v})
		}

		kConf.EnvVars = ev
		kConf.Mounts = fMounts

		ensureBoxExist(*kConf)
		ensureBoxRunning()
	}

	return nil
}

func loadFileMount(mm server.MountMap) (map[string]string, error) {
	kt, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	fm := map[string]string{}

	for _, fe := range kt.FileMount.Mounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	return fm, nil
}

func loadConfig() (*KLConfigType, error) {
	kf, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	// read kl.yml into struct
	klConfig := &KLConfigType{
		Packages: kf.Packages,
	}
	return klConfig, nil
}

func getCwdHash() string {
	cwd, _ := os.Getwd()
	hash := md5.New()
	hash.Write([]byte(cwd))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func ensurePublicKey() {
	currentUser, _ := user.Current()
	sshPath := fmt.Sprintf("%s/.ssh", currentUser.HomeDir)
	if _, err := os.Stat(fmt.Sprintf("%s/id_rsa.pub", sshPath)); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", fmt.Sprintf("%s/id_rsa", sshPath), "-N", "")
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}

func ensureCacheExist() {
	command := exec.Command("docker", "volume", "create", "nix-store")
	err := command.Run()
	if err != nil {
		fn.PrintError(errors.New("Error creating nix-store cache volume"))
	}

	command = exec.Command("docker", "volume", "create", "kl-home-cache")
	err = command.Run()
	if err != nil {
		fn.PrintError(errors.New("Error creating home cache volume"))
	}
}

func ensureBoxExist(klConfig KLConfigType) {
	currentUser, _ := user.Current()
	containerName := "kl-box-" + getCwdHash()
	cwd, _ := os.Getwd()
	o, err := exec.Command("docker", "inspect", containerName).Output()
	startContainer := func() {
		conf, err := json.Marshal(klConfig)
		if err != nil {
			panic(err)
		}

		dockerArgs := []string{"run"}
		if !foreground {
			dockerArgs = append(dockerArgs, "-d")
		}

		dockerArgs = append(dockerArgs, "--name", containerName,
			"-v", fmt.Sprintf("%s/.ssh/id_rsa.pub:/home/kl/.ssh/authorized_keys:z", currentUser.HomeDir),
			"-v", "/var/run/docker.sock:/var/run/docker.sock:ro",
			// "-v", "kl-home-cache:/home:rw",
			// "-v", "nix-store:/nix:rw",
			"-v", "kl-home-cache:/home:rw",
			"-v", "nix-store:/nix:rw",
			"--hostname", "box",
			// "-v", fmt.Sprintf("%s:/home/kl/workspace:rw", cwd),
			"-v", fmt.Sprintf("%s:/home/kl/workspace:ro", cwd),
			"-p", "1729:22",
			imageName, "--", string(conf),
		)

		fmt.Printf("\n\n%+v\n\n", dockerArgs)

		command := exec.Command("docker", dockerArgs...)

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if debug {
			fn.Logf("docker container started with cmd: %s\n", command.String())
		}

		err = command.Run()
		if err != nil {
			fn.PrintError(err)
			fn.PrintError(errors.New("Error running kl-box container"))
		}
	}

	if err != nil {
		startContainer()
	} else {
		// Get all volume mounts
		type Container struct {
			Mounts []struct {
				Type        string `json:"Type"`
				Source      string `json:"Source"`
				Destination string `json:"Destination"`
			}
		}
		var containers []Container
		err := json.Unmarshal(o, &containers)
		if err != nil {
			fn.PrintError(errors.New("Error parsing docker inspect output"))
		}
		for _, container := range containers {
			for _, mount := range container.Mounts {
				if mount.Destination == "/home/kl/workspace" {
					if fmt.Sprintf("/host_mnt%s", cwd) != mount.Source {
						fn.Log("kl-box is running with a different workspace.")
					} else {
						return
					}
				}
			}
		}

		fn.Log("Do you want to reload with current workspace? [y/N] ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" {
			return
		}
		fn.Log("Reloading kl-box container...")
		command := exec.Command(
			"docker",
			"stop", containerName)
		err = command.Run()
		if err != nil {
			fn.PrintError(errors.New("Error stopping kl-box container"))
		}
		command = exec.Command(
			"docker",
			"rm", containerName)
		err = command.Run()
		if err != nil {
			fn.PrintError(errors.New("Error removing kl-box container"))
		}
		startContainer()
	}
}

func ensureBoxRunning() {
	containerName := "kl-box-" + getCwdHash()
	command := exec.Command("docker", "start", containerName)
	err := command.Run()
	if err != nil {
		fn.PrintError(errors.New("Error starting kl-box container"))
	}
}

func isPortInUse(port string) bool {
	command := exec.Command("docker", "ps", "--format", "{{.Ports}}")
	output, err := command.Output()
	if err != nil {
		fn.PrintError(errors.New("Error checking docker containers"))
		return false
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, port) {
			return true
		}
	}
	return false
}

func init() {
	startCmd.Aliases = append(startCmd.Aliases, "s")
}
