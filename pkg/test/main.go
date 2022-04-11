package main

import (
	"fmt"
	"kloudlite.io/pkg/rexec"
)

func main() {
	r := rexec.NewK8sRclient("/Users/abdeshnayak/tmp/kloudlite/data/cluster-test-new/kubeconfig", "wireguard", "deploy/wireguard-deployment")

	r.WriteFile("/tmp/test.txt", []byte("Hello World"))
	r.Run("cat", "/tmp/test.txt")

	out, err := r.Readfile("/tmp/test.txt")

	fmt.Println(string(out), err, "hello")

	// err := r.Writefile("/tmp/test.txt", []byte("hello world2"))

	// out, err := r.Readfile("/tmp/test.txt")

	// out, err := r.Run("ls").Output()
	// fmt.Println(string(out), err)
}
