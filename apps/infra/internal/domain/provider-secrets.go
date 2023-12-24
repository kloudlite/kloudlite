package domain

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"strings"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	ct "github.com/kloudlite/operator/apis/common-types"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateAWSCloudformationTemplateUrl(args entities.AWSSecretCredentials, ev *env.Env) (string, error) {
	qp := []string{
		"templateURL=" + ev.AWSCfStackS3URL,
		"stackName=" + args.CfParamStackName,
		"param_ExternalId=" + args.CfParamExternalID,
		"param_TrustedArn=" + args.CfParamTrustedARN,
		"param_RoleName=" + args.CfParamRoleName,
		"param_InstanceProfileName=" + args.CfParamInstanceProfileName,
	}

	result := bytes.NewBuffer(nil)
	_, err := fmt.Fprintf(result, "https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?")
	if err != nil {
		return "", errors.NewE(err)
	}
	_, err = fmt.Fprint(result, strings.Join(qp, "&"))
	if err != nil {
		return "", errors.NewE(err)
	}
	return result.String(), nil
}

func (d *domain) validateAWSAssumeRole(_ context.Context, paramExternalId string, roleARN string) error {
	sess, err := session.NewSession()
	if err != nil {
		d.logger.Errorf(err, "while creating new session")
		return errors.NewE(err)
	}

	svc := sts.New(sess)

	resp, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		ExternalId:      aws.String(paramExternalId),
		RoleSessionName: aws.String("TestSession"),
	})
	if err != nil {
		d.logger.Errorf(err, "while assuming role, and getting caller identity")
		return errors.NewE(err)
	}

	if resp.AssumedRoleUser.Arn != nil {
		return nil
	}

	return nil
}

type AWSAccessValidationOutput struct {
	Result          bool
	InstallationURL *string
}

func (d *domain) ValidateProviderSecretAWSAccess(ctx InfraContext, name string) (*AWSAccessValidationOutput, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	psecret, err := d.findProviderSecret(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := psecret.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.validateAWSAssumeRole(ctx, psecret.AWS.CfParamExternalID, psecret.AWS.GetAssumeRoleRoleARN()); err != nil {
		installationURL, err := generateAWSCloudformationTemplateUrl(*psecret.AWS, d.env)
		if err != nil {
			return nil, errors.NewE(err)
		}
		return &AWSAccessValidationOutput{
			Result:          false,
			InstallationURL: &installationURL,
		}, nil
	}

	return &AWSAccessValidationOutput{
		Result:          true,
		InstallationURL: nil,
	}, errors.NewE(err)
}

func corev1SecretFromProviderSecret(ps *entities.CloudProviderSecret) *corev1.Secret {
	stringData := map[string]string{}
	if ps.AWS.AccessKey != nil {
		stringData[entities.AccessKey] = *ps.AWS.AccessKey
	}
	if ps.AWS.SecretKey != nil {
		stringData[entities.SecretKey] = *ps.AWS.SecretKey
	}

	if ps.AWS.IsAssumeRoleConfiguration() && ps.AWS.AWSAccountId != nil {
		stringData[entities.AWSAccountId] = *ps.AWS.AWSAccountId
	}
	if ps.AWS.IsAssumeRoleConfiguration() && ps.AWS.CfParamExternalID != "" {
		stringData[entities.AWSAssumeRoleExternalId] = ps.AWS.CfParamExternalID
	}
	if ps.AWS.IsAssumeRoleConfiguration() && ps.AWS.CfParamRoleName != "" {
		stringData[entities.AWAssumeRoleRoleARN] = ps.AWS.GetAssumeRoleRoleARN()
	}
	if ps.AWS.IsAssumeRoleConfiguration() && ps.AWS.CfParamInstanceProfileName != "" {
		stringData[entities.AWSInstanceProfileName] = ps.AWS.CfParamInstanceProfileName
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ps.Name,
			Namespace: ps.Namespace,
			Annotations: map[string]string{
				constants.DescriptionKey: fmt.Sprintf("created by cloudprovider secret %s", ps.Name),
			},
		},
		StringData: stringData,
	}
}

func (d *domain) CreateProviderSecret(ctx InfraContext, psecret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	psecret.AccountName = ctx.AccountName
	psecret.Namespace = accNs

	if err := psecret.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	psecret.IncrementRecordVersion()
	psecret.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	psecret.LastUpdatedBy = psecret.CreatedBy
	psecret.Id = d.secretRepo.NewId()
	switch psecret.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			psecret.AWS = &entities.AWSSecretCredentials{
				AWSAccountId: psecret.AWS.AWSAccountId,
				AccessKey:    psecret.AWS.AccessKey,
				SecretKey:    psecret.AWS.SecretKey,

				CfParamStackName:           fmt.Sprintf("%s-%s", d.env.AWSCfStackNamePrefix, psecret.Id),
				CfParamRoleName:            fmt.Sprintf("%s-%s", d.env.AWSCfRoleNamePrefix, psecret.Id),
				CfParamInstanceProfileName: fmt.Sprintf("%s-%s", d.env.AWSCfInstanceProfileNamePrefix, psecret.Id),
				CfParamTrustedARN:          d.env.AWSCfParamTrustedARN,
				CfParamExternalID:          fn.CleanerNanoidOrDie(40),
			}

			if err := psecret.AWS.Validate(); err != nil {
				return nil, errors.NewE(err)
			}
		}
	default:
		return nil, errors.Newf("unknown cloud provider")
	}
	secret := corev1SecretFromProviderSecret(&psecret)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if err := d.applyK8sResource(ctx, secret, psecret.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	nSecret, err := d.secretRepo.Create(ctx, &psecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nSecret, nil
}

func (d *domain) UpdateProviderSecret(ctx InfraContext, ups entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	if err := ups.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	currScrt, err := d.findProviderSecret(ctx, ups.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	currScrt.IncrementRecordVersion()
	currScrt.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	currScrt.Labels = ups.Labels
	currScrt.Annotations = ups.Annotations

	switch ups.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			currScrt.AWS.AccessKey = ups.AWS.AccessKey
			currScrt.AWS.SecretKey = ups.AWS.SecretKey
		}
	}

	uScrt, err := d.secretRepo.UpdateById(ctx, currScrt.Id, currScrt)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, corev1SecretFromProviderSecret(currScrt), uScrt.RecordVersion); err != nil {
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
			"accountName":              ctx.AccountName,
			"spec.credentialsRef.name": secretName,
		},
	})
	if err != nil {
		return errors.NewE(err)
	}

	if len(clusters) > 0 {
		return errors.Newf("cloud provider secret %q is used by %d cluster(s), deletion is forbidden", secretName, len(clusters))
	}

	if err := d.deleteK8sResource(ctx, corev1SecretFromProviderSecret(cps)); err != nil {
		return errors.NewE(err)
	}
	return d.secretRepo.DeleteById(ctx, cps.Id)
}

func (d *domain) ListProviderSecrets(ctx InfraContext, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListCloudProviderSecrets); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": accNs,
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
	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": accNs,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if scrt == nil {
		return nil, errors.Newf("provider secret with name %q not found", name)
	}

	return scrt, nil
}
