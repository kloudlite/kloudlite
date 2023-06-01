package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/nodectrl/internal/domain/aws"
	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/do"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	"kloudlite.io/apps/nodectrl/internal/env"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

var ProviderClientFx = fx.Module("provider-client-fx",
	fx.Provide(func(env *env.Env, gfs mongogridfs.GridFs) (common.ProviderClient, error) {

		cpd := common.CommonProviderData{}

		if err := utils.Base64YamlDecode(env.ProviderConfig, &cpd); err != nil {
			return nil, err
		}

		switch env.CloudProvider {
		case "aws":

			node := aws.AWSNode{}

			if err := utils.Base64YamlDecode(env.NodeConfig, &node); err != nil {
				return nil, err
			}

			apc := aws.AwsProviderConfig{}

			if err := utils.Base64YamlDecode(env.AWSProviderConfig, &apc); err != nil {
				return nil, err
			}

			return aws.NewAwsProviderClient(node, cpd, apc, gfs), nil
		case "azure":
			panic("not implemented")
		case "do":

			node := do.DoNode{}

			if err := utils.Base64YamlDecode(env.NodeConfig, &node); err != nil {
				return nil, err
			}

			dpc := do.DoProviderConfig{}

			if err := utils.Base64YamlDecode(env.DoProviderConfig, &dpc); err != nil {
				return nil, err
			}

			return do.NewDoProviderClient(node, cpd, dpc, gfs), nil
		case "gcp":
			panic("not implemented")
		}

		return nil, nil
	}),
)
