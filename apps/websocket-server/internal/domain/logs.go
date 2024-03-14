package domain

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/logs"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/utils"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/messaging/nats"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	msg_types "github.com/kloudlite/api/pkg/messaging/types"

	"github.com/nats-io/nats.go/jetstream"
)

func (d *domain) newJetstreamConsumerForLog(ctx context.Context, subject string, consumerId string, since *string) (*msg_nats.JetstreamConsumer, error) {
	t, err := logs.ParseSince(since)
	if err != nil {
		return nil, errors.NewE(err)
	}

	id := uuid.New().String()
	cid := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s", consumerId, id))))

	if t != nil {
		return msg_nats.NewJetstreamConsumer(ctx, d.jetStreamClient, msg_nats.JetstreamConsumerArgs{
			Stream: d.env.LogsStreamName,
			ConsumerConfig: msg_nats.ConsumerConfig{
				DeliverPolicy: jetstream.DeliverByStartTimePolicy,
				OptStartTime:  t,
				Name:          cid,
				Description:   "this is an ephemeral consumer which dispatches logs to a websocket client",
				FilterSubjects: []string{
					subject,
				},
			},
		})
	}

	return msg_nats.NewJetstreamConsumer(ctx, d.jetStreamClient, msg_nats.JetstreamConsumerArgs{
		Stream: d.env.LogsStreamName,
		ConsumerConfig: msg_nats.ConsumerConfig{
			Name:        consumerId,
			Description: "this is an ephemeral consumer which dispatches logs to a websocket client",
			FilterSubjects: []string{
				subject,
			},
		},
	})
}

func (d *domain) handleLogsMsg(ctx types.Context, logsSubs *logs.LogsSubsMap, msgAny map[string]any) error {
	log := d.logger

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

			if _, ok := (*logsSubs)[hash]; ok {
				if (*logsSubs)[hash].Jc != nil {
					if err := (*logsSubs)[hash].Jc.Stop(ctx.Context); err != nil {
						return err
					}
				}
			}

			jc, err := d.newJetstreamConsumerForLog(ctx.Context, logs.LogSubsId(msg.Spec, d.env.LogsStreamName), hash, msg.Spec.Since)
			if err != nil {
				return err
			}

			if (*logsSubs) == nil {
				*logsSubs = make(logs.LogsSubsMap)
			}

			(*logsSubs)[hash] = logs.LogsSubs{
				Jc:       jc,
				Id:       msg.Id,
				Resource: msg.Spec,
			}

			go func() {
				utils.WriteInfo(ctx, "subscribed to logs", msg.Id, types.ForLogs)

				if err := jc.Consume(
					func(m *msg_types.ConsumeMsg) error {

						var data logs.Response
						var resp types.Response[logs.Response]
						if err := json.Unmarshal(m.Payload, &data); err != nil {
							return err
						}

						resp.Type = types.MessageTypeResponse
						resp.Id = msg.Id
						sp := strings.Split(m.Subject, ".")

						data.PodName = sp[len(sp)-2]
						data.ContainerName = sp[len(sp)-1]

						resp.Data = data
						resp.For = types.ForLogs

						if err := ctx.WriteJSON(resp); err != nil {
							log.Warnf("websocket write: %w", err)
						}

						return nil
					},
					msg_types.ConsumeOpts{
						OnError: func(err error) error {
							utils.WriteError(ctx, err, msg.Id, types.ForLogs)
							return err
						},
					},
				); err != nil {
					utils.WriteError(ctx, err, msg.Id, types.ForLogs)
				}
			}()

		}

	case logs.EventUnsubscribe:
		{

			ctx.Mutex.Lock()
			if res, ok := (*logsSubs)[hash]; ok {
				if res.Jc != nil {
					if err := res.Jc.Stop(ctx.Context); err != nil {
						return err
					}

					if err := nats.DeleteConsumer(ctx.Context, d.jetStreamClient, res.Jc); err != nil {
						return err
					}

					delete(*logsSubs, hash)
				}
				ctx.Mutex.Unlock()
				utils.WriteInfo(ctx, "[logs] subscription cancelled for ", msg.Id, types.ForLogs)
			} else {
				ctx.Mutex.Unlock()
				utils.WriteWarn(ctx, fmt.Errorf("[logs] no subscription found for account: %s, cluster: %s, trackingId: %s",
					msg.Spec.Account, msg.Spec.Cluster, msg.Spec.TrackingId), msg.Id, types.ForLogs)
			}

		}
	default:
		return fmt.Errorf("invalid event: %s", msg.Event)
	}

	return nil
}
