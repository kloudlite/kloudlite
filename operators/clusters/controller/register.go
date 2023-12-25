package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"
	account_s3_bucket "github.com/kloudlite/operator/operators/clusters/internal/controllers/account-s3-bucket"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(clustersv1.AddToScheme)
	// mgr.RegisterWebhooks(&clustersv1.Cluster{})

	logger := mgr.Operator().Logger

	nc, err := nats.Connect(ev.NatsURL, func(opts *nats.Options) error {
		name := "clusters-operator"
		*opts = nats.Options{
			Name:           name,
			Servers:        []string{ev.NatsURL},
			AllowReconnect: true,
			MaxReconnect:   -1,
			ReconnectWait:  3 * time.Second,
			PingInterval:   3 * time.Second,
			ClosedCB: func(*nats.Conn) {
				logger.Infof("[%s] connection closed with nats server", name)
			},
			DisconnectedCB: func(*nats.Conn) {
				logger.Infof("[%s] disconnected with nats server", opts.Name)
			},
			ConnectedCB: func(*nats.Conn) {
				logger.Infof("[%s] connected to nats server", opts.Name)
			},
			ReconnectedCB: func(*nats.Conn) {
				logger.Infof("[%s] reconnected to nats server", opts.Name)
			},
			RetryOnFailedConnect: true,
			Compression:          true,
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	mgr.RegisterControllers(
		&target.ClusterReconciler{
			Client: nil,
			Scheme: &runtime.Scheme{},
			Env:    ev,
			Name:   "clusters:target",
			NotifyOnClusterUpdate: func(ctx context.Context, obj *clustersv1.Cluster) error {
				var m map[string]any
				b, err := json.Marshal(obj)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(b, &m); err != nil {
					return err
				}

				accountName := obj.Spec.AccountName
				clusterName := obj.Name

				msg, err := json.Marshal(types.ResourceUpdate{AccountName: accountName, ClusterName: clusterName, Object: m})
				if err != nil {
					return err
				}

				_, err = js.Publish(ctx, fmt.Sprintf(ev.NatsClusterUpdateSubjectFormat, accountName, clusterName), msg)
				if err != nil {
					return err
				}
				logger.Infof("[%s] published cluster update to nats: %s", fmt.Sprintf("%s/%s", obj.Spec.AccountName, obj.Name))
				return nil
			},
		},
		&account_s3_bucket.Reconciler{Name: "clusters:account-s3-bucket", Env: ev},
	)
}
