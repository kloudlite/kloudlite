package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/websocket/v2"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"

	mnats "github.com/nats-io/nats.go"
)

type SocketService interface {
	HandleWebSocket(ctx context.Context, c *websocket.Conn) error
}

type ReqData struct {
	AccountName string `json:"account"`
	ProjectName string `json:"project"`

	// ResourceName string `json:"resource"`
	// ResourceType string `json:"resource_type"`
	Topic    string `json:"topic"`
	ReqTopic string `json:"req_topic"`
}

func (d *domain) parseReq(rt string) (*ReqData, error) {

	entriesStrs := strings.Split(rt, ".")

	rdata := &ReqData{}

	nTopics := "res-updates"

	for _, entryStr := range entriesStrs {
		entry := strings.Split(entryStr, ":")

		if len(entry) != 2 {
			nTopics += fmt.Sprintf(".%s.*", entry[0])
		} else {
			nTopics += fmt.Sprintf(".%s.%s", entry[0], entry[1])
		}

		if (entry[0] == "account" || entry[0] == "project") && len(entry) == 2 {
			if entry[0] == "account" {
				rdata.AccountName = entry[1]
			}
			if entry[0] == "project" {
				rdata.ProjectName = entry[1]
			}
		}

	}

	rdata.Topic = nTopics
	rdata.ReqTopic = rt
	if rdata.AccountName == "" {
		return nil, fmt.Errorf("invalid topic %s", rt)
	}

	return rdata, nil
}

func (d *domain) checkAccess(ctx context.Context, rdata *ReqData, userId repos.ID) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(userId),
		ResourceRefs: func() []string {
			var refs []string

			if rdata.ProjectName != "" {
				refs = append(refs, iamT.NewResourceRef(rdata.AccountName, iamT.ResourceProject, rdata.ProjectName))
			}

			refs = append(refs, iamT.NewResourceRef(rdata.AccountName, iamT.ResourceAccount, rdata.AccountName))

			return refs

		}(),
		Action: string(func() iamT.Action {
			if rdata.ProjectName != "" {
				return iamT.GetAccount
			} else {
				return iamT.GetProject
			}
		}()),
	})

	if err != nil {
		return err
	}

	if !co.Status {
		return fmt.Errorf("access denied")
	}

	return nil
}

func (d *domain) HandleWebSocket(ctx context.Context, c *websocket.Conn) error {

	sess := httpServer.GetSession[*common.AuthSession](ctx)
	if sess == nil {
		fmt.Println("session not found")
		return fmt.Errorf("session not found")
	}

	defer func() {
		if err:=c.Close(); err!=nil {
			d.logger.Warnf("websocket close: %w", err)
		}
	}()
	log := d.logger

	type Subscription struct {
		resource ReqData
		sub      *mnats.Subscription
		open     bool
	}

	resources := make(map[string]*Subscription)

	type Message struct {
		Event string `json:"event"`
		Data  string `json:"data"`
	}

	// "account:accid.cluster:clusterid.nodepool:nodepoolid"

	type MessageType string

	const (
		MessageTypeError  MessageType = "error"
		MessageTypeUpdate MessageType = "update"
		MessageTypeInfo   MessageType = "info"
	)

	type MessageResponse struct {
		Topic   string      `json:"topic"`
		Message string      `json:"message"`
		Type    MessageType `json:"type"`
	}

	closed := false

	c.SetCloseHandler(func(code int, text string) error {
		closed = true
		return nil
	})

	defer func() {
		for _, r := range resources {
			if err := r.sub.Unsubscribe(); err != nil {
				log.Warnf("websocket unsubscribe: %w", err)
			}
		}
	}()

	// Keep the connection open
	for {

		if closed {
			break
		}

		var message Message
		if err := c.ReadJSON(&message); err != nil {

			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				break
			}
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				break
			}

			if c != nil {
				if err := c.WriteJSON(MessageResponse{
					Type:    MessageTypeError,
					Message: err.Error(),
				}); err != nil {
					log.Warnf("websocket write: %w", err)
				}
			}

			continue
		}

		rd, err := d.parseReq(message.Data)
		if err != nil {

			if c != nil {
				if err := c.WriteJSON(MessageResponse{
					Type:    MessageTypeError,
					Message: err.Error(),
				}); err != nil {
					log.Warnf("websocket write: %w", err)
				}
			}

			continue
		}

		if err := d.checkAccess(ctx, rd, sess.UserId); err != nil {
			if c != nil {
				if err := c.WriteJSON(MessageResponse{
					Type:    MessageTypeError,
					Message: err.Error(),
				}); err != nil {
					log.Warnf("websocket write: %w", err)
				}
			}

			continue
		}

		switch message.Event {
		case "subscribe":
			if _, ok := resources[message.Data]; ok {
				if c != nil {
					if err := c.WriteJSON(MessageResponse{
						Type:    MessageTypeError,
						Message: "resource already subscribed",
					}); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}
			}
			sub, err := d.natsClient.Conn.Subscribe(rd.Topic, func(m *mnats.Msg) {

				rmessage := MessageResponse{
					Topic:   rd.ReqTopic,
					Message: resources[rd.Topic].resource.ReqTopic,
					Type:    MessageTypeUpdate,
				}

				if c != nil && resources[rd.Topic] != nil && resources[rd.Topic].open {
					if err := c.WriteJSON(rmessage); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}

			})
			if err != nil {

				if c != nil {
					if err := c.WriteJSON(MessageResponse{
						Type:    MessageTypeError,
						Message: err.Error(),
					}); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}

				continue
			}
			if c != nil {
				if err := c.WriteJSON(MessageResponse{
					Type:    MessageTypeInfo,
					Message: fmt.Sprintf("subscribed to %s", rd.Topic),
				}); err != nil {
					log.Warnf("websocket write: %w", err)
				}
			}

			resources[rd.Topic] = &Subscription{
				resource: *rd,
				sub:      sub,
				open:     true,
			}

		case "unsubscribe":
			if _, ok := resources[message.Data]; !ok {

				if c != nil {
					if err := c.WriteJSON(MessageResponse{
						Type:    MessageTypeError,
						Message: "resource not subscribed",
					}); err != nil {
						log.Warnf("websocket write: %w", err)
					}
				}

			}

			if resources[rd.Topic].sub != nil {
				if err := resources[rd.Topic].sub.Unsubscribe(); err != nil {
					if c != nil {
						if err := c.WriteJSON(MessageResponse{
							Type:    MessageTypeError,
							Message: err.Error(),
						}); err != nil {
							log.Warnf("websocket write: %w", err)
						}
					}

					break
				}

				delete(resources, message.Data)
			}

		default:
			log.Errorf(fmt.Errorf("websocket read: invalid event %s", message.Event))
		}

	}

	return nil
}

type Domain interface {
	SocketService
}

type domain struct {
	iamClient  iam.IAMClient
	natsClient *nats.Client

	logger logging.Logger
}

func NewDomain(
	iamCli iam.IAMClient,

	logger logging.Logger,
	natsClient *nats.Client,
) Domain {
	return &domain{
		iamClient:  iamCli,
		natsClient: natsClient,

		logger: logger,
	}
}

var Module = fx.Module("domain", fx.Provide(NewDomain))
