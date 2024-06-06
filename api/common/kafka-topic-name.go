package common

import "fmt"

type topicName string

const (
	GitWebhookTopicName    topicName = "events.webhooks.git"
	AuditEventLogTopicName topicName = "events.audit.event-log"
	NotificationTopicName  topicName = "events.notification"
)

func GetKafkaTopicName(accountName string, clusterName string) string {
	return fmt.Sprintf("kl-send-to-acc-%s-clus-%s", accountName, clusterName)
}

func GetTenantClusterMessagingTopic(accountName string, clusterName string) string {
	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.tenant", accountName, clusterName)
}

type platformEvent string

const (
	EventErrorOnApply   platformEvent = "error-on-apply"
	EventResourceUpdate platformEvent = "resource-update"
)

type messageReceiver string

const (
	ConsoleReceiver           messageReceiver = "kloudlite-console"
	InfraReceiver             messageReceiver = "kloudlite-infra"
	ContainerRegistryReceiver messageReceiver = "kloudlite-cr"
)

func GetPlatformClusterMessagingTopic(accountName string, clusterName string, controller messageReceiver, ev platformEvent) string {
	if accountName == "*" && clusterName == "*" {
		return fmt.Sprintf("resource-sync.*.*.platform.%s.%s", controller, ev)
	}
	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.platform.%s.%s", accountName, clusterName, controller, ev)
}
