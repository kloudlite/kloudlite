package infraclient

import (
	"fmt"
	"time"
)

func InfraClientTest() {
	dop := NewDOProvider(DoProvider{
		ApiToken:  "",
		AccountId: "kl-core",
	}, ProviderEnv{
		ServerUrl:   "",
		SshKeyPath:  "",
		StorePath:   "",
		TfTemplates: "",
		JoinToken:   "",
	})

	var err error

	node := DoNode{
		Region:  "blr1",
		Size:    "s-1vcpu-1gb",
		NodeId:  "node-sample-01",
		ImageId: "ubuntu-18-04-x64",
	}

	err = dop.NewNode(node)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = dop.AttachNode(node)
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 10)

	err = dop.DeleteNode(node)

	if err != nil {
		fmt.Println(err)
		return
	}

}
