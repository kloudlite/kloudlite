package utils

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	Workdir string = "/tmp/tf-workdir"
)

func Base64YamlDecode(in string, out interface{}) error {
	rawDecodedText, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(rawDecodedText, out)
}

func ColorText(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}

func ExecCmd(cmdString string, logStr string) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	if logStr != "" {
		fmt.Printf("[#] %s\n", logStr)
	} else {
		fmt.Printf("[#] %s\n", strings.Join(cmdArr, " "))
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("err occurred: %v\n", err.Error())
		return err
	}
	return nil
}

func ExecCmdWithOutput(cmdString string, logStr string) ([]byte, error) {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, err
	}

	if logStr != "" {
		fmt.Printf("[#] %s\n", logStr)
	} else {
		fmt.Printf("[#] %s\n", strings.Join(cmdArr, " "))
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Stderr = os.Stderr

	return cmd.Output()
}
