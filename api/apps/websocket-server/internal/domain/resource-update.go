package domain

import (
	"context"
	"encoding/json"
	"fmt"

	iamT "github.com/kloudlite/api/apps/iam/types"
	res_watch "github.com/kloudlite/api/apps/websocket-server/internal/domain/resource_watch"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/types"
	"github.com/kloudlite/api/apps/websocket-server/internal/domain/utils"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/repos"
	mnats "github.com/nats-io/nats.go"
)

func (d *domain) checkAccess(ctx context.Context, rdata *res_watch.ReqData, userId repos.ID) error {
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

func (d *domain) handleResWatchMsg(ctx types.Context, resources *res_watch.ResWatchSubsMap, msgAny map[string]any) error {

	var msg res_watch.Message
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

	rd, err := res_watch.ParseReq(msg.ResPath)

	if err != nil {
		return err
	}

	if err := d.checkAccess(ctx.Context, rd, ctx.Session.UserId); err != nil {
		return err
	}

	switch msg.Event {
	case res_watch.EventSubscribe:
		{
			if _, ok := (*resources)[rd.Topic]; ok {
				return fmt.Errorf("resource already subscribed")
			}

			s, err := d.natsClient.Conn.Subscribe(rd.Topic, func(m *mnats.Msg) {
				if err := ctx.WriteJSON(types.Response[res_watch.Response]{
					Type:    types.MessageTypeResponse,
					For:     types.ForResourceUpdate,
					Data:    res_watch.Response{},
					Message: "update",
					Id:      msg.Id,
				}); err != nil {
					utils.WriteError(ctx, err, msg.Id, types.ForResourceUpdate)
					return
				}
			})

			if err != nil {
				return err
			}

			(*resources)[rd.Topic] = res_watch.ResWatchSubs{
				Sub:      s,
				Resource: *rd,
			}

			utils.WriteInfo(ctx, fmt.Sprintf("subscribed to %s", rd.Topic), msg.Id, types.ForResourceUpdate)
		}
	case res_watch.EventUnsubscribe:
		{
			if s, ok := (*resources)[rd.Topic]; ok {
				if err := s.Sub.Unsubscribe(); err != nil {
					return err
				}

				delete(*resources, rd.Topic)
				utils.WriteInfo(ctx, fmt.Sprintf("unsubscribed from %s", rd.Topic), msg.Id, types.ForResourceUpdate)
				return nil
			}

			utils.WriteError(ctx, fmt.Errorf("resource not found"), msg.Id, types.ForResourceUpdate)
		}
	default:
		return fmt.Errorf("invalid event: %s", msg.Event)
	}

	return nil
}
