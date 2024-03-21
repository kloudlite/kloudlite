package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kloudlite/api/apps/observability/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/kv"
	"go.uber.org/fx"

	"github.com/gofiber/fiber/v2"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
)

type (
	IAMGrpcClient grpc.Client
	SessionStore  kv.Repo[*common.AuthSession]
)

var Module = fx.Module(
	"app",

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Invoke(func(server httpServer.Server, sessStore SessionStore, ev *env.Env) {
		a := server.Raw()
		a.Use(httpServer.NewReadSessionMiddleware(sessStore, constants.CookieName, constants.CacheSessionPrefix))
	}),

	fx.Invoke(
		func(server httpServer.Server, ev *env.Env, sessionRepo kv.Repo[*common.AuthSession], iamCli iam.IAMClient,
		) {
			a := server.Raw()

			a.Get("/observability/metrics/:metric_type", func(c *fiber.Ctx) error {
				sess := httpServer.GetSession[*common.AuthSession](c.Context())
				if sess == nil {
					return fiber.ErrUnauthorized
				}

				m := httpServer.GetHttpCookies(c.Context())
				accountName := m[ev.AccountCookieName]
				if accountName == "" {
					return errors.Newf("no cookie named '%s' present in request", ev.AccountCookieName)
				}

				clusterName := c.Query("cluster_name")
				if clusterName == "" {
					return c.Status(http.StatusBadRequest).JSON(map[string]any{"error": "query param (cluster_name) must be provided"})
				}

				trackingId := c.Query("tracking_id")
				if trackingId == "" {
					return c.Status(http.StatusBadRequest).JSON(map[string]any{"error": "query param (tracking_id) must be provided"})
				}

				can, err := iamCli.Can(c.Context(), &iam.CanIn{
					UserId: string(sess.UserId),
					ResourceRefs: []string{
						iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
					},
					Action: string(iamT.ReadMetrics),
				})
				if err != nil {
					return &fiber.Error{Code: http.StatusUnauthorized, Message: errors.NewEf(err, "unauthorized to view metrics for resources belonging to account (%s)", accountName).Error()}
				}

				if !can.Status {
					return &fiber.Error{Code: http.StatusUnauthorized, Message: fmt.Sprintf("unauthorized to view metrics for resources belonging to account (%s)", accountName)}
				}

				metricType := c.Params("metric_type")

				st := c.Query("start_time", fmt.Sprintf("%d", time.Now().Add(-3*time.Hour).Unix()))
				et := c.Query("end_time", fmt.Sprintf("%d", time.Now().Unix()))
				step := c.Query("step", "5m")

				return queryProm(ev.PromHttpAddr, PromMetricsType(metricType), map[string]string{
					"kl_account_name": accountName,
					"kl_cluster_name": clusterName,
					"kl_tracking_id":  trackingId,
				}, st, et, step, c.Response().BodyWriter())
			})
		},
	),
)
