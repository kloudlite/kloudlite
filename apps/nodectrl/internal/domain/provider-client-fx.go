package domain

import (
	"os"

	"go.uber.org/fx"

	"kloudlite.io/apps/nodectrl/internal/domain/aws"
	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/do"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	"kloudlite.io/apps/nodectrl/internal/env"
)

var ProviderClientFx = fx.Module("provider-client-fx",
	fx.Provide(func(env *env.Env) (common.ProviderClient, error) {
		const sshDir = "/tmp/ssh"

		if _, err := os.Stat(sshDir); err != nil {
			if e := os.Mkdir(sshDir, os.ModePerm); e != nil {
				return nil, e
			}
		}

		cpd := common.CommonProviderData{}

		if err := utils.Base64YamlDecode(env.ProviderConfig, &cpd); err != nil {
			return nil, err
		}

		switch env.CloudProvider {
		case "aws":

			node := aws.AWSNodeConfig{}

			if err := utils.Base64YamlDecode(env.NodeConfig, &node); err != nil {
				return nil, err
			}

			apc := aws.AwsProviderConfig{}

			if err := utils.Base64YamlDecode(env.AWSProviderConfig, &apc); err != nil {
				return nil, err
			}

			return aws.NewAwsProviderClient(node, cpd, apc)

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

		case "gcp":
			panic("not implemented")
		}

		return nil, nil
	}),
)
