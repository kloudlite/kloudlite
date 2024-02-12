package utils

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
)

func WriteError(ctx types.Context, err error, id string, For types.For) {
	if ctx.Context != nil {
		ctx.Mutex.Lock()
		if err := ctx.Connection.WriteJSON(types.Response[any]{
			Type:    types.MessageTypeError,
			Message: err.Error(),
			For:     For,
			Id:      id,
		}); err != nil {
			log.Warnf("websocket write: %w", err)
		}
		ctx.Mutex.Unlock()
	}
}

func WriteInfo(ctx types.Context, msg string, id string, For types.For) {
	if ctx.Context != nil {
		ctx.Mutex.Lock()
		if err := ctx.Connection.WriteJSON(types.Response[any]{
			Type:    types.MessageTypeInfo,
			Message: msg,
			Id:      id,
			For:     For,
		}); err != nil {
			log.Warnf("websocket write: %w", err)
		}
		ctx.Mutex.Unlock()
	} else {
		log.Warnf("websocket connection is nil")
	}
}
