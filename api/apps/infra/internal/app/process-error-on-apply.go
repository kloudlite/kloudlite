package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	msgOfficeT "github.com/kloudlite/api/apps/message-office/types"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/messaging"
	"github.com/kloudlite/api/pkg/messaging/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, logger *slog.Logger, d domain.Domain) {
	counter := 0
	mu := sync.Mutex{}

	processMsg := func(msg *types.ConsumeMsg) error {
		mu.Lock()
		counter += 1
		mu.Unlock()

		start := time.Now()

		logger := logger.With("subject", msg.Subject, "counter", counter)
		logger.Debug("INCOMING message")

		em, err := msgOfficeT.UnmarshalErrMessage(msg.Payload)
		if err != nil {
			return errors.NewE(err)
		}

		var errObj t.AgentErrMessage
		if err := json.Unmarshal(em.Error, &errObj); err != nil {
			return errors.NewE(err)
		}

		obj := unstructured.Unstructured{Object: errObj.Object}
		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		mlogger := logger.With(
			"GVK", gvkStr,
			"NN", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
			"account", errObj.AccountName,
			"cluster", em.ClusterName,
		)

		if len(strings.TrimSpace(errObj.AccountName)) == 0 {
			mlogger.Warn("message does not contain 'accountName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		if len(strings.TrimSpace(em.ClusterName)) == 0 {
			mlogger.Warn("message does not contain 'clusterName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		mlogger.Debug("validated message")
		defer func() {
			mlogger.Info("PROCESSED message", "took", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		}()

		dctx := domain.InfraContext{
			Context:     context.TODO(),
			UserId:      "sys-user:error-on-apply-worker",
			UserEmail:   "",
			UserName:    "",
			AccountName: em.AccountName,
		}

		opts := domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp}

		gvkstr := obj.GroupVersionKind().String()
		switch gvkstr {
		case globalVpnGVK.String():
			{
				cc, err := fn.JsonConvert[entities.GlobalVPNConnection](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnGlobalVPNConnectionApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnGlobalVPNConnectionDeleteMessage(dctx, em.ClusterName, cc)
			}
		case nodepoolGVK.String():
			{
				nodepool, err := fn.JsonConvert[entities.NodePool](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnNodepoolApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnNodePoolDeleteMessage(dctx, em.ClusterName, nodepool)
			}
		default:
			{
				return errors.Newf("infra error-on-apply reader does not acknowledge resource with gvk (%s)", gvkstr)
			}
		}
	}

	if err := consumer.Consume(processMsg, types.ConsumeOpts{
		OnError: func(err error) error {
			logger.Error("while reading messages, got", "err", err)
			return nil
		},
	}); err != nil {
		logger.Error("when setting up error-on-apply consumer, got", "err", err)
	}
}
