package beacon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type Json map[string]any

type Service interface {
	TriggerSystemEvent(
		ctx context.Context,
		source string,
		eventName string,
		data Json,
	)
	TriggerUserEvent(
		ctx context.Context,
		userId string,
		accountId string,
		source string,
		action string,
		data Json,
	)
}

type SvcImpl struct {
	baseUrl string
}

func (s SvcImpl) TriggerUserEvent(_ context.Context, userId string, accountId string, source string, action string, data Json) {
	postBody, _ := json.Marshal(Json{
		"userId":    userId,
		"accountId": accountId,
		"source":    source,
		"action":    action,
		"data":      data,
	})
	reqBody := bytes.NewBuffer(postBody)
	go http.Post(s.baseUrl, "application/json", reqBody)
	return
}

func (s SvcImpl) TriggerSystemEvent(_ context.Context, source string, eventName string, data Json) {
	postBody, _ := json.Marshal(Json{
		"event":  eventName,
		"source": source,
		"data":   data,
	})
	reqBody := bytes.NewBuffer(postBody)
	go http.Post(s.baseUrl, "application/json", reqBody)
	return
}

func NewService(baseUrl string) Service {
	return &SvcImpl{
		baseUrl: baseUrl,
	}
}
