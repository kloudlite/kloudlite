package res_watch

import (
	"fmt"
	mnats "github.com/nats-io/nats.go"
	"strings"
)

type ResWatchSubsMap map[string]ResWatchSubs
type ResWatchSubs struct {
	Sub      *mnats.Subscription
	Resource ReqData
}

type Event string

const (
	EventSubscribe   Event = "subscribe"
	EventUnsubscribe Event = "unsubscribe"
)

type Message struct {
	ResPath string
	Id      string
	Event   Event
}

type ReqData struct {
	AccountName string `json:"account"`
	ProjectName string `json:"project"`

	// ResourceName string `json:"resource"`
	// ResourceType string `json:"resource_type"`
	Topic    string `json:"topic"`
	ReqTopic string `json:"req_topic"`
}

type Response struct {
}

func ParseReq(rt string) (*ReqData, error) {

	entriesStrs := strings.Split(rt, ".")

	rdata := &ReqData{}

	nTopics := "res-updates"

	for _, entryStr := range entriesStrs {
		entry := strings.Split(entryStr, ":")

		if len(entry) != 2 {
			nTopics += fmt.Sprintf(".%s.*", entry[0])
		} else {
			nTopics += fmt.Sprintf(".%s.%s", entry[0], entry[1])
		}

		if (entry[0] == "account" || entry[0] == "project") && len(entry) == 2 {
			if entry[0] == "account" {
				rdata.AccountName = entry[1]
			}
			if entry[0] == "project" {
				rdata.ProjectName = entry[1]
			}
		}

	}

	rdata.Topic = nTopics
	rdata.ReqTopic = rt
	if rdata.AccountName == "" {
		return nil, fmt.Errorf("invalid topic %s", rt)
	}

	return rdata, nil
}
