package domain

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/nats-io/nats.go/jetstream"
)

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

func parseSince(since *string) (*time.Time, error) {
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

type LogsReqData struct {
	AccountName string  `json:"account"`
	ClusterName string  `json:"cluster"`
	TrackingId  string  `json:"trackingId"`
	Since       *string `json:"since,omitempty"`
}

func (d *domain) newJetstreamConsumerForLog(ctx context.Context, subject string, consumerId string, since *string) (*msg_nats.JetstreamConsumer, error) {
	t, err := parseSince(since)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if t != nil {
		return msg_nats.NewJetstreamConsumer(ctx, d.jetStreamClient, msg_nats.JetstreamConsumerArgs{
			Stream: d.env.LogsStreamName,
			ConsumerConfig: msg_nats.ConsumerConfig{
				DeliverPolicy: jetstream.DeliverByStartTimePolicy,
				OptStartTime:  t,
				Name:          consumerId,
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

func getLogHash(ld LogsReqData, userId repos.ID) string {
	uuid := uuid.New().String()
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s-%s", ld.AccountName, ld.ClusterName, ld.TrackingId, userId, uuid))))
}

func (d *domain) getLogSubsId(ld LogsReqData) string {
	return fmt.Sprintf("%s.%s.%s.%s.>", d.env.LogsStreamName, ld.AccountName, ld.ClusterName, ld.TrackingId)
}

func (d *domain) HandleWebSocketForLogs(ctx context.Context, c *websocket.Conn) error {
	sess := httpServer.GetSession[*common.AuthSession](ctx)
	if sess == nil {
		return errors.NewE(fmt.Errorf("session not found"))
	}

	defer func() {
		if err := c.Close(); err != nil {
			d.logger.Warnf("websocket close: %w", err)
		}
	}()

	log := d.logger

	type Subscription struct {
		resource LogsReqData
		jc       *msg_nats.JetstreamConsumer
		open     bool
	}

	resources := make(map[string]*Subscription)

	type Message struct {
		Event string      `json:"event"`
		Data  LogsReqData `json:"data"`
	}

	type MessageType string

	const (
		MessageTypeError  MessageType = "error"
		MessageTypeUpdate MessageType = "update"
		MessageTypeInfo   MessageType = "info"
		MessageTypeLog    MessageType = "log"
	)
	type MsgSpec struct {
		PodName       string `json:"podName"`
		ContainerName string `json:"containerName"`
	}

	type MessageResponse struct {
		Timestamp time.Time   `json:"timestamp"`
		Message   string      `json:"message"`
		Spec      *MsgSpec    `json:"spec,omitempty"`
		Type      MessageType `json:"type"`
	}

	closed := false

	writeError := func(c *websocket.Conn, err error) error {
		if c != nil {
			return c.WriteJSON(MessageResponse{
				Type:    MessageTypeError,
				Message: err.Error(),
			})
		}
		return nil
	}

	writeInfo := func(c *websocket.Conn, msg string) error {
		if c != nil {
			return c.WriteJSON(MessageResponse{
				Type:    MessageTypeInfo,
				Message: msg,
			})
		}
		return nil
	}

	for {
		if closed {
			break
		}

		var msg Message
		if err := c.ReadJSON(&msg); err != nil {

			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				break
			}
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				break
			}

			if err := writeError(c, err); err != nil {
				log.Warnf("websocket write: %w", err)
			}

			continue
		}

		if err := d.checkAccountAccess(ctx, msg.Data.AccountName, sess.UserId, iamT.GetAccount); err != nil {
			if err := writeError(c, err); err != nil {
				log.Warnf("websocket write: %w", err)
			}
			continue
		}

		hash := getLogHash(msg.Data, sess.UserId)

		switch msg.Event {
		case "subscribe":

			if _, ok := resources[hash]; ok {
				if err := writeError(
					c, errors.Newf("already subscribed to logs for account: %s, cluster: %s, trackingId: %s",
						msg.Data.AccountName, msg.Data.ClusterName, msg.Data.TrackingId,
					),
				); err != nil {
					log.Warnf("websocket write: %w", err)
				}
				continue
			}

			jc, err := d.newJetstreamConsumerForLog(ctx, d.getLogSubsId(msg.Data), hash, msg.Data.Since)
			if err != nil {
				if err := writeError(c, err); err != nil {
					log.Warnf("websocket write: %w", err)
				}
				continue
			}

			resources[hash] = &Subscription{
				resource: msg.Data,
				jc:       jc,
				open:     true,
			}

			go func() {

				if err := writeInfo(c, "subscribed to logs"); err != nil {
					log.Warnf("websocket write: %w", err)
				}

				if err := jc.Consume(
					func(msg *types.ConsumeMsg) error {
						if c != nil {
							var resp MessageResponse
							if err := json.Unmarshal(msg.Payload, &resp); err != nil {
								if err := writeError(c, err); err != nil {
									log.Warnf("websocket write: %w", err)
								}
							}
							resp.Type = MessageTypeLog
							sp := strings.Split(msg.Subject, ".")
							resp.Spec = &MsgSpec{
								PodName:       sp[len(sp)-2],
								ContainerName: sp[len(sp)-1],
							}
							if err := c.WriteJSON(resp); err != nil {
								log.Warnf("websocket write: %w", err)
							}
						}

						return nil
					},
					types.ConsumeOpts{
						OnError: func(err error) error {
							if err := writeError(c, err); err != nil {
								log.Warnf("websocket write: %w", err)
							}

							return err
						},
					},
				); err != nil {
					if err := writeError(c, err); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}

			}()

		case "unsubscribe":
			if _, ok := resources[hash]; !ok {
				if err := writeError(
					c, errors.Newf("not subscribed to logs for account: %s, cluster: %s, trackingId: %s",
						msg.Data.AccountName, msg.Data.ClusterName, msg.Data.TrackingId,
					),
				); err != nil {
					log.Warnf("websocket write: %w", err)
				}

				continue
			}

			if resources[hash].jc != nil {
				if err := resources[hash].jc.Stop(ctx); err != nil {
					if err := writeError(c, err); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}
			}

		default:
			if err := writeError(
				c, errors.Newf("invalid event: %s, available events: subscribe, unsubscribe", msg.Event),
			); err != nil {
				log.Warnf("websocket write: %w", err)
			}
		}

	}

	return nil
}
