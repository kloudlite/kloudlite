package logs

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/repos"
)

type Event string

const (
	EventSubscribe   Event = "subscribe"
	EventUnsubscribe Event = "unsubscribe"
)

type MsgData struct {
	Account       string  `json:"account"`
	Cluster       string  `json:"cluster"`
	TrackingId    string  `json:"trackingId"`
	Since         *string `json:"since,omitempty"`
	RecordVersion *int    `json:"recordVersion"`
}

type Message struct {
	Event Event   `json:"event"`
	Spec  MsgData `json:"spec"`
	Id    string  `json:"id"`
}

type Response struct {
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
	PodName       string    `json:"podName"`
	ContainerName string    `json:"containerName"`
}

type LogsSubsMap map[string]LogsSubs
type LogsSubs struct {
	Jc       *msg_nats.JetstreamConsumer
	Id       string
	Resource MsgData
}

func LogHash(md MsgData, userId repos.ID, sid string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s", md.Account, md.Cluster, md.TrackingId, userId))))
}

func parseTime(since string) (time.Time, error) {
	now := time.Now()

	// Split the string into the numeric and duration type parts
	length := len(since)
	if length < 2 {
		return now, fmt.Errorf("invalid expiration format")
	}

	durationValStr := since[:length-1]
	durationVal, err := strconv.Atoi(durationValStr)
	if err != nil {
		return now, fmt.Errorf("invalid duration value: %v", err)
	}

	durationType := since[length-1]

	switch durationType {
	case 'm':
		return now.Add(-time.Duration(durationVal) * time.Minute), nil
	case 'h':
		return now.Add(-time.Duration(durationVal) * time.Hour), nil
	case 'd':
		return now.AddDate(0, 0, -durationVal), nil
	case 'w':
		return now.AddDate(0, 0, -durationVal*7), nil
	case 'M':
		return now.AddDate(0, -durationVal, 0), nil
	default:
		return now, fmt.Errorf("invalid duration type: %v, available types: m, h, d, w, M", durationType)
	}
}

func ParseSince(since *string) (*time.Time, error) {
	if since == nil {
		return nil, nil
	}

	if *since == "" {
		return nil, nil
	}

	t, err := parseTime(*since)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &t, nil
}

func LogSubsId(md MsgData, logStreamName string) string {

	if md.RecordVersion == nil {
		return fmt.Sprintf("%s.%s.%s.%s.>", logStreamName, md.Account, md.Cluster, md.TrackingId)
	}

	return fmt.Sprintf("%s.%s.%s.%s.%d.>", logStreamName, md.Account, md.Cluster, md.TrackingId, md.RecordVersion)
}
