package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operator"

	// account_s3_bucket "github.com/kloudlite/operator/operators/clusters/internal/controllers/account-s3-bucket"
	aws_vpc "github.com/kloudlite/operator/operators/clusters/internal/controllers/aws-vpc"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(clustersv1.AddToScheme)

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
		&aws_vpc.AwsVPCReconciler{
			Env:  ev,
			Name: "aws-vpc:controller",
		},
		&target.ClusterReconciler{
			Env:  ev,
			Name: "clusters:target",
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

				if obj.GetDeletionTimestamp() == nil {
					m[types.ResourceStatusKey] = types.ResourceStatusUpdated
				}

				if obj.GetDeletionTimestamp() != nil {
					m[types.ResourceStatusKey] = func() types.ResourceStatus {
						if types.HasOtherKloudliteFinalizers(obj) {
							return types.ResourceStatusDeleting
						}
						return types.ResourceStatusDeleted
					}()
				}

				msg, err := json.Marshal(types.ResourceUpdate{
					AccountName: accountName,
					ClusterName: clusterName,
					Object:      m,
				})
				if err != nil {
					return err
				}

				_, err = js.Publish(ctx, fmt.Sprintf(ev.NatsClusterUpdateSubjectFormat, accountName, clusterName), msg)
				if err != nil {
					return err
				}
				logger.Infof("published cluster update to nats: %s/%s", obj.Spec.AccountName, obj.Name)
				return nil
			},
		},
	)
}
