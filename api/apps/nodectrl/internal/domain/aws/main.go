package aws

import (
	"fmt"
	"path"
	"time"

	"golang.org/x/net/context"
	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

type AwsProviderConfig struct {
	AccessKey    string `yaml:"accessKey"`
	AccessSecret string `yaml:"accessSecret"`
	AccountId    string `yaml:"accountId"`
}

type AWSNode struct {
	NodeId       string `yaml:"nodeId"`
	Region       string `yaml:"region"`
	InstanceType string `yaml:"instanceType"`
	VPC          string `yaml:"vpc"`
	ImageId      string `yaml:"imageId"`
}

type awsClient struct {
	gfs  mongogridfs.GridFs
	node AWSNode

	accessKey    string
	accessSecret string

	SSHPath     string
	accountId   string
	providerDir string
	tfTemplates string
	labels      map[string]string
	taints      []string
}

func parseValues(a awsClient) map[string]string {
	values := map[string]string{}

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["region"] = a.node.Region
	values["node_id"] = a.node.NodeId
	values["instance_type"] = a.node.InstanceType
	values["keys-path"] = a.SSHPath
	values["ami"] = a.node.ImageId

	return values
}

func (a awsClient) SaveToDbGuranteed(ctx context.Context) {
	for {
		if err := utils.SaveToDb(ctx, a.node.NodeId, a.gfs); err == nil {
			break
		} else {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 20)
	}
}

// NewNode implements ProviderClient
func (a awsClient) NewNode(ctx context.Context) error {

	values := parseValues(a)

	/*
		steps:
			- check if state present in db
			- if present load that to working dir
			- else initialize new tf dir
			- apply terraform
			- upload the final state with defer
	*/

	if err := utils.MakeTfWorkFileReady(ctx, a.node.NodeId, path.Join(a.tfTemplates, "aws"), a.gfs, true); err != nil {
		return err
	}

	defer a.SaveToDbGuranteed(ctx)

	// upload the final state to the db, upsert if db is already present

	// apply the tf file
	if err := func() error {
		if err := utils.InitTFdir(path.Join(utils.Workdir, a.node.NodeId)); err != nil {
			return err
		}

		if err := utils.ApplyTF(path.Join(utils.Workdir, a.node.NodeId), values); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

// DeleteNode implements ProviderClient
func (a awsClient) DeleteNode(ctx context.Context) error {

	values := parseValues(a)

	/*
		steps:
			- check if state present in db
			- if present load that to working dir
			- else initialize new tf dir
			- destroy node with terraform
			- delete final state
	*/

	if err := utils.MakeTfWorkFileReady(ctx, a.node.NodeId, path.Join(a.tfTemplates, "aws"), a.gfs, false); err != nil {
		return err
	}

	// destroy the tf file
	if err := func() error {
		if err := utils.DestroyNode(a.node.NodeId, values); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.zip", a.node.NodeId)

	if err := a.gfs.DeleteAllWithFilename(filename); err != nil {
		return err
	}

	return nil
}

func NewAwsProviderClient(node AWSNode, cpd common.CommonProviderData, apc AwsProviderConfig, gfs mongogridfs.GridFs) common.ProviderClient {
	return awsClient{
		node: node,
		gfs:  gfs,

		accessKey:    apc.AccessKey,
		accessSecret: apc.AccessSecret,
		accountId:    apc.AccountId,

		providerDir: "aws",
		tfTemplates: cpd.TfTemplates,
		labels:      cpd.Labels,
		taints:      cpd.Taints,
		SSHPath:     cpd.SSHPath,
	}
}
