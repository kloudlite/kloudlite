package domain

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/logs"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/utils"
	"github.com/kloudlite/api/constants"
)

type LogSubscriptionCtx struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	Reader     io.Reader
}

func (d *domain) handleObservabilityLogsMsg(ctx types.Context, subscriptions map[string]LogSubscriptionCtx, msgAny map[string]any) error {
	var msg logs.Message
	b, err := json.Marshal(msgAny)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	if msg.Id == "" {
		msg.Id = "default"
	}

	hash := logs.LogHash(msg.Spec, ctx.Session.UserId, msg.Id)

	switch msg.Event {
	case logs.EventSubscribe:
		{
			if err := d.checkAccountAccess(ctx.Context, msg.Spec.Account, ctx.Session.UserId, iamT.ReadLogs); err != nil {
				return err
			}

			tpk := logs.LogSubsId(msg.Spec, d.env.LogsStreamName)
			d.logger.Debugf("tpk: %s", tpk)

			d.logger.Infof("subscribing to logs for %s", msg.Spec.TrackingId)
			utils.WriteInfo(ctx, "subscribed to logs", msg.Id, types.ForLogs)

			nctx, cf := context.WithCancel(ctx.Context)

			req, err := http.NewRequestWithContext(nctx, http.MethodGet, fmt.Sprintf("http://%s/observability/logs", d.env.ObservabilityApiAddr), nil)
			if err != nil {
				utils.WriteError(ctx, err, msg.Id, types.ForLogs)
			}

			req.AddCookie(&http.Cookie{
				Name:  constants.CookieName,
				Value: string(ctx.Session.Id),
			})

			req.AddCookie(&http.Cookie{Name: "kloudlite-account", Value: msg.Spec.Account})

			qp := req.URL.Query()
			qp.Add("tracking_id", msg.Spec.TrackingId)
			qp.Add("cluster_name", msg.Spec.Cluster)
			req.URL.RawQuery = qp.Encode()

			// resp := &http
			var resp *http.Response = nil

			// max 20 retries for fetching logs
			for i := 1; i < 20; i++ {
				resp, err = http.DefaultClient.Do(req)
				if err != nil {
					utils.WriteError(ctx, err, msg.Id, types.ForLogs)
					defer cf()
					return err
				}

				if resp.StatusCode == http.StatusTooEarly {
					<-time.After(2 * time.Second)
					d.logger.Infof("[%d] retrying fetching logs", i)
					continue
				}
				break
			}

			resp.Close = true
			subscriptions[hash] = LogSubscriptionCtx{
				Context:    nctx,
				CancelFunc: cf,
				Reader:     resp.Body,
			}

			go func() {
				defer resp.Body.Close()
				for {
					if err := nctx.Err(); err != nil {
						fmt.Println("subscription cancelled")
						return
					}
					<-time.After(1 * time.Second)
				}
			}()

			go func() {
				defer resp.Body.Close()
				// defer subscriptions[hash].Close()
				reader := bufio.NewReader(resp.Body)
				for {
					b2, err := reader.ReadBytes('\n')
					if err != nil {
						return
						// return err
					}
					ctx.Write([]byte(types.CreateResponseJson(b2, msg.Id)))
				}
			}()
		}

	case logs.EventUnsubscribe:
		{
			d.logger.Infof("unsubscribing to logs for %s", msg.Spec.TrackingId)
			utils.WriteInfo(ctx, "[logs] subscription cancelled for ", msg.Id, types.ForLogs)
			if v, ok := subscriptions[hash]; ok {
				v.CancelFunc()
			}
		}
	default:
		return fmt.Errorf("invalid event: %s", msg.Event)
	}

	return nil
}
