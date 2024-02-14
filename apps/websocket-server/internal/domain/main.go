package domain

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/logs"
	res_watch "github.com/kloudlite/api/apps/websocket-server/internal/domain/resource_watch"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/utils"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/messaging/nats"
)

func (d *domain) HandleWebSocket(ctx context.Context, c *websocket.Conn) error {
	sess := httpServer.GetSession[*common.AuthSession](ctx)
	if sess == nil {
		return errors.NewE(fmt.Errorf("session not found"))
	}

	mu := sync.Mutex{}

	logsSubs := &logs.LogsSubsMap{}
	rWatchSubs := &res_watch.ResWatchSubsMap{}

	defer func() {
		if err := c.Close(); err != nil {
			d.logger.Warnf("websocket close: %w", err)
		}

		if logsSubs != nil {
			for _, v := range *logsSubs {
				if v.Jc != nil {
					if err := v.Jc.Stop(ctx); err != nil {
						d.logger.Warnf("stop jetstream consumer failed with err: %w", err)
					}
					if err := nats.DeleteConsumer(ctx, d.jetStreamClient, v.Jc); err != nil {
						d.logger.Warnf("deleting jetstream consumer failed with err: %w", err)
					}
				}
			}
		}

		if rWatchSubs != nil {
			for _, v := range *rWatchSubs {
				if v.Sub != nil {
					if err := v.Sub.Unsubscribe(); err != nil {
						d.logger.Warnf("unsubscribe: %w", err)
					}
				}
			}
		}
	}()

	closed := false
	c.SetCloseHandler(func(_ int, _ string) error {
		closed = true
		return nil
	})

	sc := types.Context{
		Context:    ctx,
		Session:    sess,
		Connection: c,
		Mutex:      &mu,
	}

	for {
		if closed {
			break
		}

		var msg types.Message
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				break
			}
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				break
			}

			utils.WriteError(sc, err, "", "")
			continue
		}

		switch msg.For {
		case types.ForLogs:
			if err := d.handleLogsMsg(types.Context{
				Context:    ctx,
				Session:    sess,
				Connection: c,
				Mutex:      &mu,
			}, logsSubs, msg.Data); err != nil {
				utils.WriteError(sc, err, "", types.ForLogs)
			}

		case types.ForResourceUpdate:
			if err := d.handleResWatchMsg(types.Context{
				Context: ctx,
				Session: sess,
				Mutex:   &mu,
			}, rWatchSubs, msg.Data); err != nil {
				utils.WriteError(sc, err, "", types.ForResourceUpdate)
			}

		default:
			utils.WriteError(sc, fmt.Errorf("invalid for: %s", msg.For), "", "")
		}

	}

	return nil
}
