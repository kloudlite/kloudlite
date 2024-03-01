package utils

import (
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
)

func WriteError(ctx types.Context, err error, id string, For types.For) {
	if err := ctx.WriteJSON(types.Response[any]{
		Type:    types.MessageTypeError,
		Message: err.Error(),
		For:     For,
		Id:      id,
	}); err != nil {
		ctx.Logger.Warnf("websocket write: %w", err)
	}
}

func WriteInfo(ctx types.Context, msg string, id string, For types.For) {
	if err := ctx.WriteJSON(types.Response[any]{
		Type:    types.MessageTypeInfo,
		Message: msg,
		Id:      id,
		For:     For,
	}); err != nil {
		ctx.Logger.Warnf("websocket write: %w", err)
	}
}
