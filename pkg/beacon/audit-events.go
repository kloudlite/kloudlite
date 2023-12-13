package beacon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/messaging"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/repos"
	"time"
)

type EventTarget struct {
	ResourceType constants.ResourceType `json:"resourceType" bson:"resourceType"`
	ResourceId   repos.ID               `json:"resourceId" bson:"resourceId"`
}

type status struct {
	Value   string `json:"value" bson:"value"`
	Message string `json:"message" bson:"message"`
}

func StatusOK() status {
	return status{Value: "OK"}
}

func StatusError(err error) status {
	return status{Value: "ERROR", Message: err.Error()}
}

type EventAction struct {
	Action       constants.Action       `json:"action" bson:"action"`
	Status       status                 `json:"status" bson:"status"`
	ResourceType constants.ResourceType `json:"resourceType" bson:"resourceType"`
	ResourceId   repos.ID               `json:"resourceId" bson:"resourceId"`
	Tags         map[string]string      `json:"tags" bson:"tags"`
}

// AuditLogEvent represents who did what and when
type AuditLogEvent struct {
	UserId      repos.ID          `json:"userId" bson:"userId"`
	Email       string            `json:"email" bson:"email"`
	Action      string            `json:"action" bson:"action"`
	Description string            `json:"description" bson:"description"`
	Tags        map[string]string `json:"tags" bson:"tags"`
	Target      EventTarget       `json:"target" bson:"target"`
	When        time.Time         `json:"when" bson:"when"`
}

type beacon struct {
	producer messaging.Producer
	topic    string
}

func getSession(ctx context.Context) (*common.AuthSession, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)

	if session == nil {
		return nil, errors.New("Unauthorized")
	}
	return session, nil
}

func (b *beacon) TriggerEvent(ctx context.Context, accountId repos.ID, event *AuditLogEvent) error {
	eventB, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if event.Tags == nil {
		event.Tags = make(map[string]string, 3)
	}
	event.Tags["accountId"] = string(accountId)

	httpUserAgent := ctx.Value("user-agent")
	if httpUserAgent != nil {
		if ua, ok := httpUserAgent.(string); ok {
			event.Tags["user-agent"] = ua
		}
	}
	if err:=b.producer.Produce(ctx, types.ProduceMsg{
		Subject: b.topic,
		Payload: eventB,
	}); err != nil {
		return err
	}
	return nil
}

func defaultDesc(action, resType, resId string) string {
	return fmt.Sprintf("performed [action=%s] on target [type=%s, id=%s]", action, resType, resId)
}

func (b *beacon) TriggerWithUserCtx(ctx context.Context, accountId repos.ID, act EventAction) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}

	ale := AuditLogEvent{
		UserId:      session.UserId,
		Email:       session.UserEmail,
		Action:      string(act.Action),
		Description: defaultDesc(string(act.Action), string(act.ResourceType), string(act.ResourceId)),
		Tags:        make(map[string]string, 3+len(act.Tags)),
		Target: EventTarget{
			ResourceType: act.ResourceType,
			ResourceId:   act.ResourceId,
		},
		When: time.Now(),
	}

	for k, v := range act.Tags {
		ale.Tags[k] = v
	}

	ale.Tags["accountId"] = string(accountId)
	ale.Tags["sessionId"] = string(session.Id)

	httpUserAgent := ctx.Value("user-agent")
	if httpUserAgent != nil {
		if ua, ok := httpUserAgent.(string); ok {
			ale.Tags["user-agent"] = ua
		}
	}

	eventB, err := json.Marshal(ale)
	if err != nil {
		return err
	}
	if err:=b.producer.Produce(ctx, types.ProduceMsg{
		Subject: b.topic,
		Payload: eventB,
	}); err != nil {
		return err
	}

	return nil
}

func NewBeacon(producer messaging.Producer, topic string) Beacon {
	return &beacon{producer: producer, topic: topic}
}
