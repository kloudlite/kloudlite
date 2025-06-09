package types

import "github.com/kloudlite/api/apps/console/internal/entities"

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	PublishConsoleEvent(ctx ConsoleContext, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishEnvironmentResourceEvent(ctx ConsoleContext, envName string, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishResourceEvent(ctx ResourceContext, resourceType entities.ResourceType, name string, update PublishMsg)
	PublishClusterManagedServiceEvent(ctx ConsoleContext, msvcName string, resourceType entities.ResourceType, name string, update PublishMsg)
}
