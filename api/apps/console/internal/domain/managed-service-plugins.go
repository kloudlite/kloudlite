package domain

import "github.com/kloudlite/api/apps/console/internal/entities"

// GetManagedPluginTemplate implements Domain.
func (d *domain) GetManagedServicePlugin(category string, plugin string) (*entities.ManagedServicePlugin, error) {
	return d.managedServicePluginsMap[category][plugin], nil
}

// ListManagedServicePlugins implements Domain.
func (d *domain) ListManagedServicePlugins() ([]*entities.ManagedServicePlugins, error) {
	return d.managedServicePlugins, nil
}
