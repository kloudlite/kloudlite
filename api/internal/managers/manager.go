package managers

import (
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager coordinates repositories for handlers
type Manager struct {
	K8sClient             client.Client
	UserRepository        repository.UserRepository
	EnvironmentRepository repository.EnvironmentRepository
	MachineTypeRepository repository.MachineTypeRepository
	WorkMachineRepository repository.WorkMachineRepository
	WorkspaceRepository   repository.WorkspaceRepository
}
