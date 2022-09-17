package infraclient

import (
	"fmt"
)

func testAwsClient() {

	awsp := NewAWSProvider(AWSProvider{
		AccessKey:    "",
		AccessSecret: "",
		AccountId:    "kl-core",
	}, AWSProviderEnv{
		ServerUrl:   "***REMOVED***",
		SshKeyPath:  "/home/vision/tf/ssh",
		StorePath:   "/home/vision/tf",
		TfTemplates: "/home/vision/kloudlite/api-go/pkg/infraClient/terraform",
		JoinToken:   "",
	})

	var err error

	node := AWSNode{
		NodeId:       "node-sample-02",
		Region:       "ap-south-1",
		InstanceType: "t2.micro",
		VPC:          "",
		// AMI:          "ami-0d70546e43a941d70",
		AMI: "ami-068257025f72f470d",
	}

	// if err = awsp.NewNode(node); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if err = awsp.AttachNode(node); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	if err = awsp.UnattachNode(node); err != nil {
		fmt.Println(err)
		return
	}

	// time.Sleep(time.Second * 10)

	if err = awsp.DeleteNode(node); err != nil {
		fmt.Println(err)
		return
	}

}
