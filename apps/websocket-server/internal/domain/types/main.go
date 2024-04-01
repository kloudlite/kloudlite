package types

import (
	"context"
	"fmt"
	"sync"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
)

type For string

const (
	ForLogs           For = "logs"
	ForJetstreamLogs  For = "jetstram-logs"
	ForResourceUpdate For = "resource-update"
)

type MessageType string

const (
	MessageTypeError    MessageType = "error"
	MessageTypeResponse MessageType = "response"
	MessageTypeInfo     MessageType = "info"
	MessageTypeWarning  MessageType = "warning"
)

type Response[T any] struct {
	Type MessageType `json:"type"`
	For  For         `json:"for"`

	Data    T      `json:"data"`
	Message string `json:"message"`
	Id      string `json:"id"`
}

func CreateResponseJson(data []byte, id string) string {
	// return map[string]any{
	// 	"type":    "response",
	// 	"for":     "logs",
	// 	"data":    string(data),
	// 	"message": "",
	// 	"id":      id,
	// }
	return fmt.Sprintf(`{"type":"response","for":"logs","data":%s, "id":%q}`, data, id)
}

type Message struct {
	For  For            `json:"for"`
	Data map[string]any `json:"data"`
}

type Context struct {
	Logger  logging.Logger
	Context context.Context
	Session *common.AuthSession
	// Connection *websocket.Conn
	Mutex *sync.Mutex

	Write          func([]byte) error
	WriteJSON      func(interface{}) error
	OnDisconnectFn func() error
}

func (c Context) OnDisconnect() error {
	if c.OnDisconnectFn != nil {
		return c.OnDisconnectFn()
	}
	return nil
}
