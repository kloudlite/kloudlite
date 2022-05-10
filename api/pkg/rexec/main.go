package rexec

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Rclient interface {
	Run(cmd string, args ...string) *exec.Cmd
	Readfile(file string) ([]byte, error)
	WriteFile(path string, content []byte) error
}

type rK8sClient struct {
	kubeConfigPath string
	namespace      string
	name           string
}

func (r *rK8sClient) Run(cmd string, args ...string) *exec.Cmd {
	kubeEnv := fmt.Sprintf("KUBECONFIG=%v", r.kubeConfigPath)
	_cmd := exec.Command("kubectl", append(append([]string{}, "exec", "-n", r.namespace, r.name, "--", cmd), args...)...)
	_cmd.Env = append(_cmd.Env, kubeEnv)
	return _cmd
}

func (r *rK8sClient) Readfile(file string) ([]byte, error) {
	return r.Run("cat", file).Output()
}

func (r *rK8sClient) WriteFile(file string, content []byte) error {
	_content := b64.StdEncoding.EncodeToString(content)
	_cmd := exec.Command("kubectl", "exec", "-n", r.namespace, r.name, "--", "bash", "-c", fmt.Sprintf("echo %v | base64 -d > %v", _content, file))
	kubeEnv := fmt.Sprintf("KUBECONFIG=%v", r.kubeConfigPath)
	_cmd.Env = append(_cmd.Env, kubeEnv)
	_cmd.Stderr = os.Stderr
	_cmd.Stdout = os.Stdout
	return _cmd.Run()
}

func NewK8sRclient(kubeConfigPath, namespace, name string) Rclient {
	return &rK8sClient{kubeConfigPath, namespace, name}
}

type rSshClient struct {
	host        string
	access_path string
	user        string
}

func (r *rSshClient) Run(cmd string, args ...string) *exec.Cmd {
	_args := strings.Split(fmt.Sprintf("%v@%v -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -i %v %v", r.user, r.host, r.access_path, cmd), " ")
	return exec.Command("ssh", append(_args, args...)...)
}

func (r *rSshClient) Readfile(file string) ([]byte, error) {
	return r.Run("cat", file).Output()
}

func (r *rSshClient) WriteFile(file string, content []byte) error {
	_content := b64.StdEncoding.EncodeToString(content)

	_, err := exec.Command("ssh", fmt.Sprintf("%v@%v", r.user, r.host), "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "LogLevel=quiet", "-i", r.access_path, "--", fmt.Sprintf("bash -c \"echo %v | base64 -d > %v\"", _content, file)).Output()
	return err
}

func NewSshRclient(host, user, access_path string) Rclient {
	return &rSshClient{host: host, access_path: access_path, user: user}
}

// to from local to kube
// KUBECONFIG=./kubeconfig kubectl exec -n wireguard deploy/wireguard-deployment -- bash -c "echo $data > abc.txt"
