package domain

import (
	"bytes"
	"context"
	"fmt"
"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"

	ct "github.com/kloudlite/operator/apis/common-types"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	fn "kloudlite.io/pkg/functions"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"

	// "github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var assumeRoleRoleNameFormat = "kloudlite-access-role-%s"

func (d *domain) generateAWSCloudformationTemplateUrl(ctx context.Context, stackName string, paramExternalID string, paramRoleName string) (string, error) {
	result := bytes.NewBuffer(nil)

	fmt.Fprintf(result, "https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?")
	fmt.Fprintf(result, "templateURL=%s", d.env.AWSCloudformationStackS3URL)
	fmt.Fprintf(result, "&stackName=%s", stackName)
	fmt.Fprintf(result, "&param_ExternalId=%s", paramExternalID)
	fmt.Fprintf(result, "&param_TrustedArn=%s", d.env.AWSCloudformationParamTrustedARN)
	fmt.Fprintf(result, "&param_RoleName=%s", paramRoleName)

	installationURL := result.String()
	return installationURL, nil
}

func (d *domain) validateAWSAssumeRole(ctx context.Context, awsAccountId string, paramExternalId string) error {
	sess, err := session.NewSession()
	if err != nil {
		d.logger.Errorf(err, "while creating new session")
		return err
	}

	svc := sts.New(sess)

	resp, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn: aws.String(fmt.Sprintf(d.env.AWSAssumeTenantRoleFormatString, awsAccountId)),
		// WARN: external id should be different for each tenant
		ExternalId:      aws.String(paramExternalId),
		RoleSessionName: aws.String("TestSession"),
	})
	if err != nil {
		d.logger.Errorf(err, "while assuming role, and getting caller identity")
		return err
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
		return nil, err
	}

	psecret, err := d.findProviderSecret(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := psecret.Validate(); err != nil {
		return nil, err
	}

	if err := d.validateAWSAssumeRole(ctx, *psecret.AWS.AWSAccountId, psecret.AWS.AWSAssumeRoleExternalId); err != nil {
		installationURL, err := d.generateAWSCloudformationTemplateUrl(ctx, fmt.Sprintf("%s-%s", d.env.AWSCloudformationStackNamePrefix, psecret.Id), psecret.AWS.AWSAssumeRoleExternalId, fmt.Sprintf(assumeRoleRoleNameFormat, psecret.Id))
		if err != nil {
			return nil, err
		}
		return &AWSAccessValidationOutput{
			Result:          false,
			InstallationURL: &installationURL,
		}, nil
	}

	return &AWSAccessValidationOutput{
		Result:          true,
		InstallationURL: nil,
	}, err
}

func corev1SecretFromProviderSecret(ps *entities.CloudProviderSecret) *corev1.Secret {
	stringData := map[string]string{}
	if ps.AWS.AccessKey != nil {
		stringData[entities.AccessKey] = *ps.AWS.AccessKey
	}
	if ps.AWS.SecretKey != nil {
		stringData[entities.SecretKey] = *ps.AWS.SecretKey
	}
	if ps.AWS.AWSAccountId != nil {
		stringData[entities.AWSAccountId] = *ps.AWS.AWSAccountId
	}
	if ps.AWS.AWSAssumeRoleExternalId != "" {
		stringData[entities.AWSAssumeRoleExternalId] = ps.AWS.AWSAssumeRoleExternalId
	}
	if ps.AWS.AWAssumeRoleRoleARN != "" {
		stringData[entities.AWAssumeRoleRoleARN] = ps.AWS.AWAssumeRoleRoleARN
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

func (d *domain) CreateProviderSecret(ctx InfraContext, pSecret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, err
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, err
	}

	pSecret.AccountName = ctx.AccountName
	pSecret.Namespace = accNs

	if err := pSecret.Validate(); err != nil {
		return nil, err
	}

	pSecret.IncrementRecordVersion()
	pSecret.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	pSecret.LastUpdatedBy = pSecret.CreatedBy
	pSecret.Id = d.secretRepo.NewId()
	if pSecret.CloudProviderName == ct.CloudProviderAWS {
		pSecret.AWS.AWAssumeRoleRoleARN = fmt.Sprintf(d.env.AWSAssumeTenantRoleFormatString,
			*pSecret.AWS.AWSAccountId,
			fmt.Sprintf(assumeRoleRoleNameFormat, pSecret.Id),
		)
		pSecret.AWS.AWSAssumeRoleExternalId = fn.CleanerNanoidOrDie(40)
	}

	if err := d.applyK8sResource(ctx, corev1SecretFromProviderSecret(&pSecret), pSecret.RecordVersion); err != nil {
		return nil, err
	}

	nSecret, err := d.secretRepo.Create(ctx, &pSecret)
	if err != nil {
		return nil, err
	}

	return nSecret, nil
}

func (d *domain) UpdateProviderSecret(ctx InfraContext, ups entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, err
	}

	if err := ups.Validate(); err != nil {
		return nil, err
	}

	currScrt, err := d.findProviderSecret(ctx, ups.Name)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if err := d.applyK8sResource(ctx, corev1SecretFromProviderSecret(currScrt), uScrt.RecordVersion); err != nil {
		return nil, err
	}

	return uScrt, nil
}

func (d *domain) DeleteProviderSecret(ctx InfraContext, secretName string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCloudProviderSecret); err != nil {
		return err
	}
	cps, err := d.findProviderSecret(ctx, secretName)
	if err != nil {
		return err
	}

	clusters, err := d.clusterRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountName":              ctx.AccountName,
			"spec.credentialsRef.name": secretName,
		},
	})
	if err != nil {
		return err
	}

	if len(clusters) > 0 {
		return fmt.Errorf("cloud provider secret %q is used by %d cluster(s), deletion is forbidden", secretName, len(clusters))
	}

	if err := d.deleteK8sResource(ctx, corev1SecretFromProviderSecret(cps)); err != nil {
		return err
	}
	return d.secretRepo.DeleteById(ctx, cps.Id)
}

func (d *domain) ListProviderSecrets(ctx InfraContext, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListCloudProviderSecrets); err != nil {
		return nil, err
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, err
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": accNs,
	}
	return d.secretRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCloudProviderSecret); err != nil {
		return nil, err
	}
	return d.findProviderSecret(ctx, name)
}

func (d *domain) findProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error) {
	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, err
	}

	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": accNs,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}

	if scrt == nil {
		return nil, fmt.Errorf("provider secret with name %q not found", name)
	}

	return scrt, nil
}
