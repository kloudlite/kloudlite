package domain

import (
	"fmt"
	"strings"
	"time"

	fn "github.com/kloudlite/api/pkg/functions"

	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	ct "github.com/kloudlite/operator/apis/common-types"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func corev1SecretFromProviderSecret(ps *entities.CloudProviderSecret) (*corev1.Secret, error) {
	stringData := map[string]string{}

	switch ps.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			switch ps.AWS.AuthMechanism {
			case clustersv1.AwsAuthMechanismSecretKeys:
				{
					if err := fn.JsonConversion(ps.AWS.AuthSecretKeys, &stringData); err != nil {
						return nil, err
					}
				}
			case clustersv1.AwsAuthMechanismAssumeRole:
				{
					if err := fn.JsonConversion(ps.AWS.AssumeRoleParams.AwsAssumeRoleParams, &stringData); err != nil {
						return nil, err
					}
				}
			default:
				{
					return nil, fmt.Errorf("unknown aws auth mechanism (%s)", ps.AWS.AuthMechanism)
				}
			}
		}
	case ct.CloudProviderGCP:
		{
			if err := ps.GCP.Validate(); err != nil {
				return nil, err
			}

			if err := fn.JsonConversion(ps.GCP.GCPCredentials, &stringData); err != nil {
				return nil, err
			}
		}
	default:
		{
			return nil, fmt.Errorf("unknown cloudprovider (%s)", ps.CloudProviderName)
		}
	}

	stringData["cloudprovider"] = string(ps.CloudProviderName)

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:              ps.Name,
			Namespace:         ps.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Annotations: map[string]string{
				constants.DescriptionKey: fmt.Sprintf("created by cloudprovider secret %s", ps.Name),
			},
		},
		StringData: stringData,
	}, nil
}

func (d *domain) CreateProviderSecret(ctx InfraContext, psecretIn entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return nil, errors.NewE(err)
	}

	psecretIn.AccountName = ctx.AccountName
	psecretIn.Namespace = accNs

	psecretIn.Id = d.secretRepo.NewId()

	switch psecretIn.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			if psecretIn.AWS == nil {
				return nil, fmt.Errorf("aws vars must be set")
			}

			switch psecretIn.AWS.AuthMechanism {
			case clustersv1.AwsAuthMechanismSecretKeys:
				{
					psecretIn.AWS.CfParamStackName = fmt.Sprintf("%s-%s", d.env.AWSCfStackNamePrefix, psecretIn.Id)
					psecretIn.AWS.CfParamRoleName = fmt.Sprintf("%s-%s", d.env.AWSCfRoleNamePrefix, psecretIn.Id)
					psecretIn.AWS.CfParamInstanceProfileName = fmt.Sprintf("%s-%s", d.env.AWSCfInstanceProfileNamePrefix, psecretIn.Id)

					if psecretIn.AWS.AuthSecretKeys == nil {
						psecretIn.AWS.AuthSecretKeys = &entities.AWSAuthSecretKeys{}
					}
					psecretIn.AWS.AuthSecretKeys.CfParamUserName = fmt.Sprintf("kloudlite-user-%s", psecretIn.Id)
				}
			case clustersv1.AwsAuthMechanismAssumeRole:
				{
					if psecretIn.AWS.AssumeRoleParams == nil {
						return nil, fmt.Errorf("aws assume role params, must be set")
					}
					psecretIn.AWS.CfParamStackName = fmt.Sprintf("%s-%s", d.env.AWSCfStackNamePrefix, psecretIn.Id)
					psecretIn.AWS.CfParamRoleName = fmt.Sprintf("%s-%s", d.env.AWSCfRoleNamePrefix, psecretIn.Id)
					psecretIn.AWS.CfParamInstanceProfileName = fmt.Sprintf("%s-%s", d.env.AWSCfInstanceProfileNamePrefix, psecretIn.Id)
					psecretIn.AWS.AssumeRoleParams.CfParamTrustedARN = d.env.AWSCfParamTrustedARN

					psecretIn.AWS.AssumeRoleParams.ExternalID = fn.CleanerNanoidOrDie(40)
					psecretIn.AWS.AssumeRoleParams.RoleARN = psecretIn.AWS.GetAssumeRoleRoleARN()
				}
			default:
				{
					return nil, fmt.Errorf("unknown aws auth mechanism (%s)", psecretIn.AWS.AuthMechanism)
				}
			}

			// if err := psecretIn.AWS.Validate(); err != nil {
			// 	return nil, errors.NewE(err)
			// }
		}
	case ct.CloudProviderGCP:
		{
			if err := psecretIn.GCP.Validate(); err != nil {
				return nil, errors.NewE(err)
			}
		}
	default:
		return nil, errors.Newf("unknown cloud provider")
	}

	psecretIn.IncrementRecordVersion()
	psecretIn.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	psecretIn.LastUpdatedBy = psecretIn.CreatedBy

	secret, _ := corev1SecretFromProviderSecret(&psecretIn)
	psecretIn.ObjectMeta = secret.ObjectMeta

	nSecret, err := d.secretRepo.Create(ctx, &psecretIn)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, secret, nSecret.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return nSecret, nil
}

func (d *domain) UpdateProviderSecret(ctx InfraContext, providerSecretIn entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	fieldsPatch := map[string]any{}
	switch providerSecretIn.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			if providerSecretIn.AWS.AuthMechanism == clustersv1.AwsAuthMechanismSecretKeys {
				if providerSecretIn.AWS.AuthSecretKeys != nil {
					fieldsPatch[fc.CloudProviderSecretAwsAuthSecretKeysAccessKey] = strings.TrimSpace(providerSecretIn.AWS.AuthSecretKeys.AccessKey)
					fieldsPatch[fc.CloudProviderSecretAwsAuthSecretKeysSecretKey] = strings.TrimSpace(providerSecretIn.AWS.AuthSecretKeys.SecretKey)
				}
			}
		}
	}

	patchForUpdate := common.PatchForUpdate(ctx, &providerSecretIn, common.PatchOpts{
		XPatch: fieldsPatch,
	})

	uScrt, err := d.secretRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: providerSecretIn.Name}, patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	realSecret, err := corev1SecretFromProviderSecret(uScrt)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, realSecret, uScrt.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return uScrt, nil
}

func (d *domain) DeleteProviderSecret(ctx InfraContext, secretName string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCloudProviderSecret); err != nil {
		return errors.NewE(err)
	}
	cps, err := d.findProviderSecret(ctx, secretName)
	if err != nil {
		return errors.NewE(err)
	}

	clusters, err := d.clusterRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName: ctx.AccountName,
			// fc.ClusterSpecCredentialsRefName: secretName,
			fc.ClusterSpecAwsCredentialsSecretRefName: secretName,
		},
	})
	if err != nil {
		return errors.NewE(err)
	}

	if len(clusters) > 0 {
		return errors.Newf("cloud provider secret %q is used by %d cluster(s), deletion is forbidden", secretName, len(clusters))
	}

	realSecret, err := corev1SecretFromProviderSecret(cps)
	if err != nil {
		return errors.NewE(err)
	}
	if err := d.deleteK8sResource(ctx, realSecret); err != nil {
		return errors.NewE(err)
	}
	return d.secretRepo.DeleteById(ctx, cps.Id)
}

func (d *domain) ListProviderSecrets(ctx InfraContext, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListCloudProviderSecrets); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}
	return d.secretRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findProviderSecret(ctx, name)
}

func (d *domain) findProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if scrt == nil {
		return nil, errors.Newf("provider secret with name %q not found", name)
	}

	return scrt, nil
}
