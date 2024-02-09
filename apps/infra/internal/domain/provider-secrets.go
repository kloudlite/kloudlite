package domain

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	fn "github.com/kloudlite/api/pkg/functions"
	ct "github.com/kloudlite/operator/apis/common-types"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
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

func (d *domain) validateAWSAssumeRole(_ context.Context, paramExternalId string, roleARN string, instanceProfileName string, cfStackName string) error {
	sess, err := session.NewSession()
	sess.Config.Region = aws.String("ap-south-1")
	if err != nil {
		d.logger.Errorf(err, "while creating new session")
		return errors.NewE(err)
	}

	// 1. validating IAM Assume Role
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

	if resp.AssumedRoleUser == nil || resp.AssumedRoleUser.Arn == nil {
		return errors.Newf("AWS assume role (%s) not found", roleARN)
	}

	nsess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-south-1"),
		Credentials: credentials.NewStaticCredentials(*resp.Credentials.AccessKeyId, *resp.Credentials.SecretAccessKey, *resp.Credentials.SessionToken),
	})
	if err != nil {
		return errors.NewE(err)
	}

	cf := cloudformation.New(nsess)
	dso, err := cf.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: &cfStackName,
	})
	if err != nil {
		return errors.NewE(err)
	}

	stackFound := false

	for i := range dso.Stacks {
		if dso.Stacks[i] != nil && *dso.Stacks[i].StackName == cfStackName {
			stackFound = true
			if *dso.Stacks[i].StackStatus != cloudformation.StackStatusCreateComplete {
				return errors.Newf("cloudformation stack (%s) is not completed, yet", cfStackName)
			}
		}
	}

	if !stackFound {
		return errors.Newf("waiting for cloudformation stack to be created")
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

	if err := d.validateAWSAssumeRole(ctx, psecret.AWS.CfParamExternalID, psecret.AWS.GetAssumeRoleRoleARN(), psecret.AWS.CfParamInstanceProfileName, psecret.AWS.CfParamStackName); err != nil {
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
			CreationTimestamp: metav1.Time{
				Time: time.Now(),
			},
			Annotations: map[string]string{
				constants.DescriptionKey: fmt.Sprintf("created by cloudprovider secret %s", ps.Name),
			},
		},
		StringData: stringData,
	}
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

	if err := psecretIn.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	psecretIn.IncrementRecordVersion()
	psecretIn.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	psecretIn.LastUpdatedBy = psecretIn.CreatedBy
	psecretIn.Id = d.secretRepo.NewId()
	switch psecretIn.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			psecretIn.AWS = &entities.AWSSecretCredentials{
				AWSAccountId: psecretIn.AWS.AWSAccountId,
				AccessKey:    psecretIn.AWS.AccessKey,
				SecretKey:    psecretIn.AWS.SecretKey,

				CfParamStackName:           fmt.Sprintf("%s-%s", d.env.AWSCfStackNamePrefix, psecretIn.Id),
				CfParamRoleName:            fmt.Sprintf("%s-%s", d.env.AWSCfRoleNamePrefix, psecretIn.Id),
				CfParamInstanceProfileName: fmt.Sprintf("%s-%s", d.env.AWSCfInstanceProfileNamePrefix, psecretIn.Id),
				CfParamTrustedARN:          d.env.AWSCfParamTrustedARN,
				CfParamExternalID:          fn.CleanerNanoidOrDie(40),
			}

			if err := psecretIn.AWS.Validate(); err != nil {
				return nil, errors.NewE(err)
			}
		}
	default:
		return nil, errors.Newf("unknown cloud provider")
	}

	secret := corev1SecretFromProviderSecret(&psecretIn)
	psecretIn.ObjectMeta = secret.ObjectMeta

	nSecret, err := d.secretRepo.Create(ctx, &psecretIn)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err != nil {
		return nil, errors.NewE(err)
	}
	if err := d.applyK8sResource(ctx, secret, nSecret.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return nSecret, nil
}

// Depricate AWS_SECRET_KEY and AWS_ACCESS_KEY input
func (d *domain) UpdateProviderSecret(ctx InfraContext, providerSecretIn entities.CloudProviderSecret) (*entities.CloudProviderSecret, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	if err := providerSecretIn.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	//switch providerSecretIn.CloudProviderName {
	//case ct.CloudProviderAWS:
	//	{
	//		currScrt.AWS.AccessKey = providerSecretIn.AWS.AccessKey
	//		currScrt.AWS.SecretKey = providerSecretIn.AWS.SecretKey
	//	}
	//}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&providerSecretIn,
	)

	uScrt, err := d.secretRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: providerSecretIn.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, corev1SecretFromProviderSecret(uScrt), uScrt.RecordVersion); err != nil {
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
			fields.AccountName:               ctx.AccountName,
			fc.ClusterSpecCredentialsRefName: secretName,
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
