package do

import (
	"context"
	"fmt"
	"path"
	"time"

	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

type DoProviderConfig struct {
	ApiToken  string `yaml:"apiToken"`
	AccountId string `yaml:"accountId"`
}

type DoNode struct {
	Region  string `yaml:"region"`
	Size    string `yaml:"size"`
	NodeId  string `yaml:"nodeId"`
	ImageId string `yaml:"imageId"`
}

type doClient struct {
	gfs  mongogridfs.GridFs
	node DoNode

	apiToken string

	SSHPath     string
	accountId   string
	providerDir string
	tfTemplates string
	labels      map[string]string
	taints      []string
}

func parseValues(d doClient) map[string]string {
	values := map[string]string{}

	values["do-token"] = d.apiToken
	values["accountId"] = d.accountId

	values["do-image-id"] = "ubuntu-22-10-x64"
	values["nodeId"] = d.node.NodeId
	values["size"] = d.node.Size
	values["keys-path"] = d.SSHPath

	return values
}

// SaveToDbGuranteed implements ProviderClient
func (d doClient) SaveToDbGuranteed(ctx context.Context) {
	for {
		if err := utils.SaveToDb(ctx, d.node.NodeId, d.gfs); err == nil {
			break
		} else {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 20)
	}
}

// NewNode implements ProviderClient
func (d doClient) NewNode(ctx context.Context) error {
	values := parseValues(d)

	/*
		steps:
			- check if state present in db
			- if present load that to working dir
			- else initialize new tf dir
			- apply terraform
			- upload the final state with defer
	*/

	if err := utils.MakeTfWorkFileReady(ctx, d.node.NodeId, path.Join(d.tfTemplates, "do"), d.gfs, true); err != nil {
		return err
	}

	defer d.SaveToDbGuranteed(ctx)

	// apply the tf file
	if err := func() error {
		if err := utils.InitTFdir(path.Join(utils.Workdir, d.node.NodeId)); err != nil {
			return err
		}

		if err := utils.ApplyTF(path.Join(utils.Workdir, d.node.NodeId), values); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

// DeleteNode implements ProviderClient
func (d doClient) DeleteNode(ctx context.Context) error {
	values := parseValues(d)

	/*
		steps:
			- check if state present in db
			- if present load that to working dir
			- else initialize new tf dir
			- destroy node with terraform
			- delete final state
	*/

	if err := utils.MakeTfWorkFileReady(ctx, d.node.NodeId, path.Join(d.tfTemplates, "aws"), d.gfs, false); err != nil {
		return err
	}

	// destroy the tf file
	if err := func() error {
		if err := utils.DestroyNode(d.node.NodeId, values); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.zip", d.node.NodeId)

	if err := d.gfs.DeleteAllWithFilename(filename); err != nil {
		return err
	}

	return nil
}

func NewDoProviderClient(node DoNode, cpd common.CommonProviderData, dpc DoProviderConfig, gfs mongogridfs.GridFs) common.ProviderClient {
	return doClient{
		node: node,
		gfs:  gfs,

		apiToken:  dpc.ApiToken,
		accountId: dpc.AccountId,

		providerDir: "do",

		tfTemplates: cpd.TfTemplates,
		labels:      cpd.Labels,
		taints:      cpd.Taints,
		SSHPath:     cpd.SSHPath,
	}
}
