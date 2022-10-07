package infraclient

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	CLUSTER_ID = "kl"
)

// rmTFdir implements doProviderClient
// func rmdir(folder string) error {
// 	return execCmd(fmt.Sprintf("rm -rf %q", folder), "")
// }

// makeTFdir implements doProviderClient
func mkdir(folder string) error {
	return execCmd(fmt.Sprintf("mkdir -p %q", folder), "mkdir <terraform_dir>")
}

// destroyNode implements doProviderClient
func destroyNode(folder string, values map[string]string) error {
	vars := []string{"destroy", "-auto-approve", "-no-color"}

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return os.RemoveAll(folder)

	// return err
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

// applyTF implements doProviderClient
func applyTF(folder string, values map[string]string) error {

	vars := []string{"apply", "-auto-approve", "-no-color"}

	fmt.Printf("[#] terraform %s", strings.Join(vars, " "))

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = folder

	return cmd.Run()
}

func execCmd(cmdString string, logStr string) error {
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
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("err occurred: %s\n", err.Error())
		return err
	}
	return nil
}
