package infraclient

import (
	"fmt"
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	Secret string `env:"SECRETS" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}

func testDoClient() {
	env := GetEnvOrDie()

	dop := NewDOProvider(DoProvider{
		ApiToken:  "***REMOVED***",
		AccountId: "kl-core",
	}, DoProviderEnv{
		StorePath:   "/home/vision/tf",
		TfTemplates: "/home/vision/kloudlite/api-go/pkg/infraClient/terraform",
		Secrets:     env.Secret,
		Labels:      map[string]string{
			// "kloudlite.io/region": "blr1",

		},

		SSHPath: "/home/vision/.ssh",
	})

	var err error

	node := DoNode{
		Region: "blr1",
		// Size:   "s-4vcpu-8gb-amd",
		// Size: "s-2vcpu-4gb-amd",
		// Size: "s-1vcpu-1gb-amd",
		Size:    "c-2",
		NodeId:  "try-agent-01",
		ImageId: "ubuntu-22-10-x64",
	}

	// fmt.Println(node, err, dop)

	if false {

		if err = dop.NewNode(node); err != nil {
			fmt.Println(err)
			return
		}

		for {

			if err = dop.AttachNode(node); err != nil {
				fmt.Println(err)

				time.Sleep(time.Second * 5)
				continue
			}

			return
		}

	} else {

		if err = dop.DeleteNode(node); err != nil {
			fmt.Println(err)
			return
		}

	}

}
