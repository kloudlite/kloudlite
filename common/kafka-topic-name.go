package common

import (
	"encoding/base64"
	"fmt"
)

type topicName string

const (
	GitWebhookTopicName    topicName = "events.webhooks.git"
	AuditEventLogTopicName topicName = "events.audit.event-log"
	NotificationTopicName  topicName = "events.notification"
)

const (
	SendToAgentSubjectNamePrefix      = "send-to-agent"
	ReceiveFromAgentSubjectNamePrefix = "receive-from-agent"
)

func SendToAgentSubjectPrefix(accountName string, clusterName string) string {
	return fmt.Sprintf("%s.%s.%s", SendToAgentSubjectNamePrefix, accountName, clusterName)
}

func ReceiveFromAgentSubjectPrefix(accountName string, clusterName string) string {
	return fmt.Sprintf("%s.%s.%s", ReceiveFromAgentSubjectNamePrefix, accountName, clusterName)
}

// func GetKafkaTopicName(accountName string, clusterName string) string {
// 	return fmt.Sprintf("kl-send-to-acc-%s-clus-%s", accountName, clusterName)
// }

// func GetTenantClusterMessagingTopic(accountName string, clusterName string) string {
// 	// return fmt.Sprintf("resource-sync.account-%s.cluster-%s.tenant", accountName, clusterName)
// 	return fmt.Sprintf("%s.%s.%s", SendToAgentSubjectNamePrefix, accountName, clusterName)
// }

func SendToAgentSubjectName(accountName string, clusterName string, gvk string, namespace string, name string) string {
	slug := base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s.%s/%s", gvk, namespace, name)))

	return fmt.Sprintf("%s.%s.%s.%s", SendToAgentSubjectNamePrefix, accountName, clusterName, slug)
}

type platformEvent string

const (
	EventErrorOnApply   platformEvent = "error-on-apply"
	EventResourceUpdate platformEvent = "resource-update"
)

type MessageReceiver string

const (
	ConsoleReceiver           MessageReceiver = "kloudlite-console"
	InfraReceiver             MessageReceiver = "kloudlite-infra"
	ContainerRegistryReceiver MessageReceiver = "kloudlite-cr"
)

type ReceiveFromAgentArgs struct {
	AccountName string
	ClusterName string
}

func ReceiveFromAgentSubjectName(args ReceiveFromAgentArgs, receiver MessageReceiver, ev platformEvent) string {
	slug := "*"
	return fmt.Sprintf("%s.%s.%s.%s.%s.%s", ReceiveFromAgentSubjectNamePrefix, args.AccountName, args.ClusterName, slug, receiver, ev)
}

// func GetPlatformClusterMessagingTopic(accountName string, clusterName string, controller messageReceiver, ev platformEvent) string {
// 	if accountName == "*" && clusterName == "*" {
// 		return fmt.Sprintf("resource-sync.*.*.platform.%s.%s", controller, ev)
// 	}
// 	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.platform.%s.%s", accountName, clusterName, controller, ev)
// }
