package infraclient

import (
	"fmt"
)

func testAwsClient() {
	env := GetEnvOrDie()

	awsp := NewAWSProvider(AWSProvider{
		AccessKey:    "***REMOVED***",
		AccessSecret: "***REMOVED***",
		AccountId:    "kl-core",
	}, AWSProviderEnv{
		StorePath:   "/home/vision/tf",
		TfTemplates: "/home/vision/kloudlite/api-go/pkg/infraClient/terraform",
		Secrets:     env.Secret,
	})

	var err error

	node := AWSNode{
		NodeId:       "aws-worker-01",
		Region:       "ap-south-1",
		InstanceType: "m5.large",
		VPC:          "",
	}

	if true {

		if err = awsp.NewNode(node); err != nil {
			fmt.Println(err)
			return
		}

		if err = awsp.AttachNode(node); err != nil {
			fmt.Println(err)
			return
		}

	} else {

		if err = awsp.DeleteNode(node); err != nil {
			fmt.Println(err)
			return
		}

	}

	// time.Sleep(time.Second * 10)

}
