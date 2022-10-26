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
		Labels: map[string]string{
			// "kloudlite.io/region": "blr1",
		},
	})

	var err error

	node := DoNode{
		Region: "blr1",
		Size:   "s-4vcpu-8gb-amd",
		// Size: "s-2vcpu-4gb-amd",
		// Size: "s-1vcpu-1gb-amd",
		// Size:    "c-2",
		NodeId:  "kl-auto-scaler",
		ImageId: "117388514",
	}

	// fmt.Println(node, err, dop)

	if false {

		if err = dop.NewNode(node); err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Second * 30)

		for {

			if err = dop.AttachNode(node); err != nil {
				fmt.Println(err)

				time.Sleep(time.Second * 30)
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

	// time.Sleep(time.Second * 10)
}
