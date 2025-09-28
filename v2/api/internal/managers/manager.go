package managers

import (
	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
	"github.com/kloudlite/kloudlite/v2/api/internal/webhooks"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager coordinates repositories and webhooks for handlers
type Manager struct {
	K8sClient             client.Client
	UserRepository        repository.UserRepository
	EnvironmentRepository repository.EnvironmentRepository
	MachineTypeRepository repository.MachineTypeRepository
	WorkMachineRepository repository.WorkMachineRepository
	UserWebhook           *webhooks.UserWebhook
	EnvironmentWebhook    *webhooks.EnvironmentWebhook
	MachineTypeWebhook    *webhooks.MachineTypeWebhook
	WorkMachineWebhook    *webhooks.WorkMachineWebhook
}