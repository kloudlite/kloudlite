package mounter

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
)

func mountFile(_file, data, mountPath string) error {

	file := path.Join(mountPath, _file)

	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path.Dir(file), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(file, []byte(data), os.ModePerm); err != nil {
		fmt.Println("error writing file", err)
	}

	return nil
}

func Mount(mountFiles map[string]string, mountBasePath string) error {

	for k, v := range mountFiles {
		err := mountFile(k, v, mountBasePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func Load(envs map[string]string, args []string) error {

	var cmd *exec.Cmd

	if len(args) > 0 {
		argsWithoutProg := args[1:]
		cmd = exec.Command(args[0], argsWithoutProg...)
	} else {
		cmd = exec.Command("printenv")
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if len(args) > 0 {
		cmd.Env = os.Environ()
	}

	for k, v := range envs {
		if len(args) == 0 {
			fmt.Printf("%s=%q\n", k, v)
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if len(args) == 0 {
		return nil
	}

	return cmd.Run()
}
