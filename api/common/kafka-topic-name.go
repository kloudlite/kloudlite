package common

import "fmt"

func GetKafkaTopicName(accountName string, clusterName string) string {
	return fmt.Sprintf("kl-send-to-acc-%s-clus-%s", accountName, clusterName)
}

/*
In case of NATS Jetstream:

	baseName => jetstream name
	  @returns subject name in format {{baseName}}.account-{{accountName}}.cluster-{{clusterName}}

In case of Kafka:

	baseName => topic base name
	@returns topic name in format {{baseName}}-account-{{accountName}}-cluster-{{clusterName}}
*/
func GetTenantClusterMessagingTopic(accountName string, clusterName string) string {
	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.tenant", accountName, clusterName)
}

type platformEvent string

const (
	EventErrorOnApply   platformEvent = "error-on-apply"
	EventResourceUpdate platformEvent = "resource-update"
)

type resourceController string

const (
	KloudliteConsole resourceController = "kloudlite-console"
	KloudliteInfra   resourceController = "kloudlite-infra"
)

func GetPlatformClusterMessagingTopic(accountName string, clusterName string, controller resourceController, ev platformEvent) string {
	if accountName == "*" && clusterName == "*" {
		return fmt.Sprintf("resource-sync.*.*.platform.%s.%s", controller, ev)
	}
	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.platform.%s.%s", accountName, clusterName, controller, ev)
}

// Stream
// {
//   nodepools,
//   jobs,
//   projects,
//   apps,
//   msvc,
//   mres
// }
//
// resource-sync.account-%s.cluster-%s.tenant.{{resource-id}}.{1,2,3}
// resource-sync.account-%s.cluster-%s.platform.{{resource-id}}.{1,2,3}
