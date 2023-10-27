package domain

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"

	// "github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func (d *domain) GenerateAWSCloudformationTemplateUrl(ctx context.Context, awsAccountId string) (string, error) {
	// TODO: generated installation template url,should be unique for a user, which means externalId should be different for each account

	var result strings.Builder

	result.WriteString("https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?")
	result.WriteString(fmt.Sprintf(`templateURL=%s`, "https://kloudlite-static-assets.s3.ap-south-1.amazonaws.com/public/cloudformation.yml"))
	result.WriteString(fmt.Sprintf(`&stackName=%s`, "kloudlite-access-stack"))
	result.WriteString(fmt.Sprintf(`&param_ExternalId=%s`, "sample"))
	result.WriteString(fmt.Sprintf(`&param_TrustedArn=%s`, "arn:aws:iam::563392089470:root"))

	return result.String(), nil
}

func (d *domain) ValidateAWSAssumeRole(ctx context.Context, awsAccountId string) error {
	sess, err := session.NewSession()
	if err != nil {
		d.logger.Errorf(err, "while creating new session")
		return err
	}

	roleARN := fmt.Sprintf(d.env.KloudliteTenantRoleFormatString, awsAccountId)

	svc := sts.New(sess)

	resp, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn: aws.String(roleARN),
		// WARN: external id should be different for each tenant
		ExternalId:      aws.String(d.env.KloudliteTenantAssumeRoleExternalId),
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

	if isValid, err := pSecret.Validate(); !isValid && err != nil {
		return nil, err
	}

	pSecret.SetGroupVersionKind(schema.FromAPIVersionAndKind("v1", "Secret"))
	pSecret.IncrementRecordVersion()
	pSecret.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	pSecret.LastUpdatedBy = pSecret.CreatedBy

	if err := d.applyK8sResource(ctx, &pSecret.Secret, pSecret.RecordVersion); err != nil {
		return nil, err
	}

	nSecret, err := d.secretRepo.Create(ctx, &pSecret)
	if err != nil {
		return nil, err
	}

	return nSecret, nil
}

func (d *domain) UpdateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, err
	}

	if isValid, err := secret.Validate(); !isValid && err != nil {
		return nil, err
	}

	scrt, err := d.findProviderSecret(ctx, secret.Name)
	if err != nil {
		return nil, err
	}

	scrt.SetGroupVersionKind(schema.FromAPIVersionAndKind("v1", "Secret"))
	scrt.IncrementRecordVersion()
	scrt.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	scrt.Labels = secret.Labels
	scrt.Annotations = secret.Annotations
	scrt.Secret.Data = secret.Secret.Data
	scrt.Secret.StringData = secret.Secret.StringData

	uScrt, err := d.secretRepo.UpdateById(ctx, scrt.Id, scrt)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uScrt.Secret, uScrt.RecordVersion); err != nil {
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

	if err := d.deleteK8sResource(ctx, &cps.Secret); err != nil {
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
