package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findSecretVariable(ctx ConsoleContext, name string) (*entities.SecretVariable, error) {
	sec, err := d.secretVariableRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, errors.Newf("secret variable %q not found", name)
	}
	return sec, nil
}

func (d *domain) GetSecretVariableOutputKVs(ctx ConsoleContext, keyrefs []SecretVariableKeyRef) ([]*SecretVariableKeyValueRef, error) {
	filters := repos.Filter{
		fields.AccountName: ctx.AccountName, // Match the account context
	}

	names := make([]any, 0, len(keyrefs))
	for i := range keyrefs {
		names = append(names, keyrefs[i].SvarName)
	}

	filters = d.secretVariableRepo.MergeMatchFilters(filters, map[string]repos.MatchFilter{
		fields.MetadataName: {
			MatchType: repos.MatchTypeArray,
			Array:     names,
		},
	})

	secretVariables, err := d.secretVariableRepo.Find(ctx, repos.Query{Filter: filters})
	if err != nil {
		return nil, errors.NewE(err)
	}

	data := make(map[string]map[string]string)

	for i := range secretVariables {
		m := make(map[string]string, len(secretVariables[i].StringData))
		for k, v := range secretVariables[i].StringData {
			m[k] = v
		}

		data[secretVariables[i].Metadata.Name] = m
	}

	results := make([]*SecretVariableKeyValueRef, 0, len(keyrefs))
	for i := range keyrefs {
		results = append(results, &SecretVariableKeyValueRef{
			SvarName: keyrefs[i].SvarName,
			Key:      keyrefs[i].Key,
			Value:    data[keyrefs[i].SvarName][keyrefs[i].Key],
		})
	}

	return results, nil
}

func (d *domain) GetSecretVariableOutputKeys(ctx ConsoleContext, name string) ([]string, error) {
	secretVariable, err := d.findSecretVariable(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if secretVariable.StringData == nil {
		return nil, errors.Newf("no output keys available for secret variable %q", name)
	}

	results := make([]string, 0, len(secretVariable.StringData))
	for k := range secretVariable.StringData {
		results = append(results, k)
	}

	return results, nil
}

func (d *domain) ListSecretVariables(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.SecretVariable], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}
	return d.secretVariableRepo.FindPaginated(ctx, d.environmentRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) GetSecretVariable(ctx ConsoleContext, name string) (*entities.SecretVariable, error) {
	sec, err := d.findSecretVariable(ctx, name)
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (d *domain) CreateSecretVariable(ctx ConsoleContext, secret entities.SecretVariable) (*entities.SecretVariable, error) {
	secret.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	secret.LastUpdatedBy = secret.CreatedBy
	secret.AccountName = ctx.AccountName

	sec, err := d.secretVariableRepo.Create(ctx, &secret)
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (d *domain) UpdateSecretVariable(ctx ConsoleContext, secret entities.SecretVariable) (*entities.SecretVariable, error) {
	existingSecret, err := d.findSecretVariable(ctx, secret.Metadata.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := repos.Document{
		fc.SecretVariableStringData: secret.StringData,
	}

	updatedSecret, err := d.secretVariableRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: existingSecret.Metadata.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return updatedSecret, nil
}

func (d *domain) DeleteSecretVariable(ctx ConsoleContext, name string) error {
	if _, err := d.findSecretVariable(ctx, name); err != nil {
		return err
	}
	err := d.secretVariableRepo.DeleteOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return err
	}
	return nil
}
