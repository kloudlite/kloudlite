package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
)

func (d *domain) ListManagedSvcTemplates() ([]*entities.MsvcTemplate, error) {
	return d.msvcTemplates, nil
}

func (d *domain) GetManagedSvcTemplate(category string, name string) (*entities.MsvcTemplateEntry, error) {
	return d.msvcTemplatesMap[category][name], nil
}
