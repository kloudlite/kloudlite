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
		URL:       fmt.Sprintf(`curl -X POST "%s/image-meta-push" -H "Authorization: %s" -H "Content-Type: application/json" -d '{"image": "imageName:imageTag", "meta": {"repository": "github", "registry": "docker", "author":"kloudlite"}}'`, d.envVars.WebhookURL, encodedToken),
		ScriptURL: fmt.Sprintf(`curl "%s/image-meta-push" | authorization=%s image=imageName:imageTag meta="repository=github,registry=docker,author=kloudlite" sh`, d.envVars.WebhookURL, encodedToken),
	}, nil
}

func (d *domain) CreateRegistryImage(ctx context.Context, accountName string, image string, meta map[string]any) (*entities.RegistryImage, error) {
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
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return createdImage, nil
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
