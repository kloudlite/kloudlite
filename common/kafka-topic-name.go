package common

import "fmt"

func GetKafkaTopicName(accountName string, clusterName string) string {
	return fmt.Sprintf("clus-%s-%s-incoming", accountName, clusterName)
}
