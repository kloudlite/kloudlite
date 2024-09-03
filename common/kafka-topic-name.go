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
	RegistryHookTopicName  topicName = "events.webhooks.registry"
)

const (
	sendToAgentSubjectPrefix      = "send-to-agent"
	receiveFromAgentSubjectPrefix = "receive-from-agent"
	//receiveFromWebhookSubjectPrefix = "receive-from-webhook"
)

func SendToAgentSubjectPrefix(accountName string, clusterName string) string {
	return fmt.Sprintf("%s.%s.%s", sendToAgentSubjectPrefix, accountName, clusterName)
}

func ReceiveFromAgentSubjectPrefix(accountName string, clusterName string) string {
	return fmt.Sprintf("%s.%s.%s", receiveFromAgentSubjectPrefix, accountName, clusterName)
}

// func GetKafkaTopicName(accountName string, clusterName string) string {
// 	return fmt.Sprintf("kl-send-to-acc-%s-clus-%s", accountName, clusterName)
// }

// func GetTenantClusterMessagingTopic(accountName string, clusterName string) string {
// 	// return fmt.Sprintf("resource-sync.account-%s.cluster-%s.tenant", accountName, clusterName)
// 	return fmt.Sprintf("%s.%s.%s", SendToAgentSubjectNamePrefix, accountName, clusterName)
// }

func SendToAgentSubjectName(accountName string, clusterName string, gvk string, namespace string, name string) string {
	slug := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s.%s/%s", gvk, namespace, name)))

	return fmt.Sprintf("%s.%s.%s.%s", sendToAgentSubjectPrefix, accountName, clusterName, slug)
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

	GVK       string
	Namespace string
	Name      string
}

func ReceiveFromAgentSubjectName(args ReceiveFromAgentArgs, receiver MessageReceiver, ev platformEvent) string {
	if args.AccountName == "*" && args.ClusterName == "*" {
		slug := "*"
		return fmt.Sprintf("%s.%s.%s.%s.%s.%s", receiveFromAgentSubjectPrefix, args.AccountName, args.ClusterName, slug, receiver, ev)
	}

	slug := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s.%s/%s", args.GVK, args.Namespace, args.Name)))
	return fmt.Sprintf("%s.%s.%s.%s.%s.%s", receiveFromAgentSubjectPrefix, args.AccountName, args.ClusterName, slug, receiver, ev)
}

// func GetPlatformClusterMessagingTopic(accountName string, clusterName string, controller messageReceiver, ev platformEvent) string {
// 	if accountName == "*" && clusterName == "*" {
// 		return fmt.Sprintf("resource-sync.*.*.platform.%s.%s", controller, ev)
// 	}
// 	return fmt.Sprintf("resource-sync.account-%s.cluster-%s.platform.%s.%s", accountName, clusterName, controller, ev)
// }
