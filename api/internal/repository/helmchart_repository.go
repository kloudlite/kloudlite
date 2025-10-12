package repository

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmChartRepository provides operations for HelmChart resources (namespace-scoped)
type HelmChartRepository interface {
	NamespacedRepository[*environmentsv1.HelmChart, *environmentsv1.HelmChartList]

	// Domain-specific methods
	UpdateStatus(ctx context.Context, name string, namespace string) error
}

// helmChartRepository implements HelmChartRepository
type helmChartRepository struct {
	NamespacedRepository[*environmentsv1.HelmChart, *environmentsv1.HelmChartList]
	client client.Client
}

// NewHelmChartRepository creates a new HelmChartRepository
func NewHelmChartRepository(k8sClient client.Client) HelmChartRepository {
	baseRepo := NewK8sNamespacedRepository(
		k8sClient,
		func() *environmentsv1.HelmChart { return &environmentsv1.HelmChart{} },
		func() *environmentsv1.HelmChartList { return &environmentsv1.HelmChartList{} },
	)

	return &helmChartRepository{
		NamespacedRepository: baseRepo,
		client:               k8sClient,
	}
}

// UpdateStatus refreshes the status of a helm chart
func (r *helmChartRepository) UpdateStatus(ctx context.Context, name string, namespace string) error {
	// Get the helm chart
	chart, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Use status subresource to update only status
	return r.client.Status().Update(ctx, chart)
}
