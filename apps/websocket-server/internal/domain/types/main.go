package types

import (
	"context"
	"sync"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/logging"
)

type For string

const (
	ForLogs           For = "logs"
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

type Message struct {
	For  For            `json:"for"`
	Data map[string]any `json:"data"`
}

type Context struct {
	Logger  logging.Logger
	Context context.Context
	Session *common.AuthSession
	// Connection *websocket.Conn
	Mutex     *sync.Mutex
	WriteJSON func(interface{}) error
}
