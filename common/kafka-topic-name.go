package common

import "fmt"

func GetKafkaTopicName(accountName string, clusterName string) string {
	return fmt.Sprintf("kl-send-to-acc-%s-clus-%s", accountName, clusterName)
}
