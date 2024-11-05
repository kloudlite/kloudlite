package domain

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

type ImageHookPayload struct {
	Image       string         `json:"image"`
	AccountName string         `json:"accountName"`
	Meta        map[string]any `json:"meta"`
}

func encodeAccessToken(accountName string, tokenSecret string) string {
	info := fmt.Sprintf("account=%s", accountName)

	fn.FxErrorHandler()

	h := sha256.New()
	h.Write([]byte(info + tokenSecret))
	sum := fmt.Sprintf("%x", h.Sum(nil))

	info += fmt.Sprintf(";sha256sum=%s", sum)

	return base64.StdEncoding.EncodeToString([]byte(info))
}

func generatePartialWords(word string) []string {
	var partials []string
	for i := 3; i <= len(word); i++ {
		partials = append(partials, word[:i])
	}
	return partials
}

func generateAutocompleteWords(meta map[string]any) string {
	metaString := ""
	for _, value := range meta {
		metaString += fmt.Sprintf("%s ", value)
	}

	words := strings.Fields(metaString)
	var autocompleteWords []string
	for _, word := range words {
		partials := generatePartialWords(word)
		autocompleteWords = append(autocompleteWords, partials...)
	}

	return strings.Join(autocompleteWords, " ")
}

func getImageNameTag(image string) (string, string) {
	parts := strings.Split(image, ":")

	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], "latest"
}

func (d *domain) GetRegistryImageURL(ctx ConsoleContext) (*entities.RegistryImageURL, error) {
	encodedToken := encodeAccessToken(ctx.AccountName, d.envVars.WebhookTokenHashingSecret)

	return &entities.RegistryImageURL{
		URL: []string{
			`export KL_WEBHOOK_TOKEN="paste your token"`,
			strings.TrimSpace(fmt.Sprintf(`
curl -X POST "%s/image/push" \
	-H "Authorization: $KL_WEBHOOK_TOKEN" \
	-H "Content-Type: application/json" \
	-d '{ "image": "<image-name>:<image-tag>", "meta": { "<key>": "<value>" }}'`, d.envVars.WebhookURL)),
		},

		URLExample: []string{
			`export KL_WEBHOOK_TOKEN="super-secret-token"`,
			fmt.Sprintf(`
curl -X POST "%s/image/push" \
	-H "Authorization: $KL_WEBHOOK_TOKEN" \
	-H "Content-Type: application/json" \
	-d '{ "image": "ghcr.io/kloudlite/api/sample:v1.2.3", "meta": { "repo": "kloudlite/sample", "branch": "testing-ci" }}'
`, d.envVars.WebhookURL),
		},

		ScriptURL: []string{
			`export KL_WEBHOOK_TOKEN="paste your token"`,
			fmt.Sprintf(`curl "%s/image-hook.sh" | image=<image-name>:<image-tag> meta="<key-1>=<value-1>,<key-2>=<value-2>" sh`, d.envVars.ImageHookScriptHostedURL),
		},
		ScriptURLExample: []string{
			`export KL_WEBHOOK_TOKEN="super-secret-token"`,
			fmt.Sprintf(`
curl "%s/image-hook.sh" | image=ghcr.io/kloudlite/api/sample:v1.2.3 meta="repo=kloudlite/sample,branch=testing-ci" sh
		`, d.envVars.ImageHookScriptHostedURL),
		},
		KlWebhookAuthToken: encodedToken,
	}, nil
}

func (d *domain) UpsertRegistryImage(ctx context.Context, accountName string, image string, meta map[string]any) (*entities.RegistryImage, error) {
	imageName, imageTag := getImageNameTag(image)

	createdImage, err := d.registryImageRepo.Upsert(ctx, repos.Filter{
		fields.AccountName:        accountName,
		fc.RegistryImageImageName: imageName,
		fc.RegistryImageImageTag:  imageTag,
	}, &entities.RegistryImage{
		AccountName: accountName,
		ImageName:   imageName,
		ImageTag:    imageTag,
		Meta:        meta,
		MetaData:    generateAutocompleteWords(meta),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return createdImage, nil
}

func (d *domain) SearchRegistryImages(ctx ConsoleContext, query string) ([]*entities.RegistryImage, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListRegistryImages); err != nil {
		return nil, errors.NewE(err)
	}

	if query == "" {
		return d.registryImageRepo.Find(ctx, repos.Query{
			Filter: repos.Filter{},
			Sort:   map[string]any{"_id": -1},
			Limit:  fn.New(int64(10)),
		})
	}

	filters := repos.Filter{
		fields.AccountName: ctx.AccountName,
		"$text":            map[string]any{"$search": query},
	}

	searchedImages, err := d.registryImageRepo.Find(ctx, repos.Query{
		Filter: filters,
		Limit:  fn.New(int64(10)),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	return searchedImages, nil
}

func (d *domain) DeleteRegistryImage(ctx ConsoleContext, image string) error {
	imageName, imageTag := getImageNameTag(image)

	matched, err := d.registryImageRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:        ctx.AccountName,
		fc.RegistryImageImageName: imageName,
		fc.RegistryImageImageTag:  imageTag,
	})
	if err != nil {
		return errors.NewE(err)
	}

	if matched == nil {
		return errors.Newf("image not found for account %s", ctx.AccountName)
	}

	err = d.registryImageRepo.DeleteOne(ctx, repos.Filter{
		fields.AccountName:        ctx.AccountName,
		fc.RegistryImageImageName: imageName,
		fc.RegistryImageImageTag:  imageTag,
	})
	if err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) GetRegistryImage(ctx ConsoleContext, image string) (*entities.RegistryImage, error) {
	imageName, imageTag := getImageNameTag(image)
	matched, err := d.registryImageRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:        ctx.AccountName,
		fc.RegistryImageImageName: imageName,
		fc.RegistryImageImageTag:  imageTag,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if matched == nil {
		return nil, errors.Newf("image not found for account %s", ctx.AccountName)
	}

	return matched, nil
}

func (d *domain) ListRegistryImages(ctx ConsoleContext, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.RegistryImage], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListRegistryImages); err != nil {
		return nil, errors.NewE(err)
	}

	filters := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}

	return d.registryImageRepo.FindPaginated(ctx, filters, pq)
}
