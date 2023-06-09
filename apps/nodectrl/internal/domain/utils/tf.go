package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GetOutput(folder, key string) ([]byte, error) {
	vars := []string{"output", "-json"}
	fmt.Printf("[#] terraform %s\n", strings.Join(vars, " "))
	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(out))

	var resp map[string]struct {
		Value string `json:"value"`
	}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return []byte(resp[key].Value), nil
}

// destroyNode implements doProviderClient
func DestroyNode(nodeId string, values map[string]string) error {
	dest := path.Join(Workdir, nodeId)
	vars := []string{"destroy", "-auto-approve"}

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = dest

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// applyTF implements doProviderClient
func ApplyTF(folder string, values map[string]string) error {
	vars := []string{"apply", "-auto-approve"}

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	fmt.Printf("[#] terraform %s", strings.Join(vars, " "))

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = folder

	return cmd.Run()
}

func getOutput(folder, key string) ([]byte, error) {
	vars := []string{"output", "-json"}
	fmt.Printf("[#] terraform %s\n", strings.Join(vars, " "))
	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(out))

	var resp map[string]struct {
		Value string `json:"value"`
	}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return []byte(resp[key].Value), nil
}

func InitTFdir(dir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
