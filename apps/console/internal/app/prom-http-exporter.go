package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	lokiserver "kloudlite.io/pkg/loki-server"
	"kloudlite.io/pkg/repos"
)

type PrometheusOpts struct {
	Endpoint          string `env:"PROMETHEUS_ENDPOINT" required:"true"`
	BasicAuthUsername string `env:"PROMETHEUS_BASIC_AUTH_USERNAME" required:"false"`
	BasicAuthPassword string `env:"PROMETHEUS_BASIC_AUTH_PASSWORD" required:"false"`
}

type PromMetricsType string

const (
	Cpu    PromMetricsType = "cpu"
	Memory PromMetricsType = "memory"
)

func getPromQuery(resType PromMetricsType, name string) string {
	switch resType {
	case Memory:
		return fmt.Sprintf(`sum(avg_over_time(container_memory_working_set_bytes{pod =~ "%s.*", container != ""} [30s])) /1024/1024`, name)
	case Cpu:
		return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{container!="", pod=~"%s.*"}[2m])) * 1000`, name)
	}
	return ""
}

func metricsQuerySvc(logserver lokiserver.LogServer, promOpts *PrometheusOpts, d domain.Domain, financeClient finance.FinanceClient,
	env *Env,
	cacheClient AuthCacheClient) {
	var app *fiber.App
	app = logserver
	app.Use(
		httpServer.NewSessionMiddleware[*common.AuthSession](
			cacheClient,
			"hotspot-session",
			env.CookieDomain,
			common.CacheSessionPrefix,
		),
	)

	app.Get(
		"/metrics/:metricsType", func(ctx *fiber.Ctx) error {
			metricsType := ctx.Params("metricsType", "")
			if metricsType == "" {
				return ctx.Status(http.StatusInternalServerError).JSON(map[string]string{"error": "metricsType is empty"})
			}

			appId := ctx.Query("appId", "")
			if appId == "" {
				return ctx.Status(http.StatusInternalServerError).JSON(map[string]string{"error": "query params (appId) is empty"})
			}

			app, err := d.GetApp(ctx.Context(), repos.ID(appId))
			if err != nil {
				return err
			}

			if app == nil {
				return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": "app does not exist"})
			}

			project, err := d.GetProjectWithID(ctx.Context(), app.ProjectId)
			if err != nil {
				return err
			}

			cluster, err := financeClient.GetAttachedCluster(
				context.TODO(),
				&finance.GetAttachedClusterIn{AccountId: string(project.AccountId)},
			)

			promQuery := getPromQuery(PromMetricsType(metricsType), app.Name)
			if promQuery == "" {
				return errors.Newf("could not build prom query, invalid (metricsType=%s or name=%s)", metricsType, app.Name)
			}

			// GET http://localhost:9090/api/v1/query_range?query=sum(container_memory_working_set_bytes{pod =~ "kl-project.*",container != ""})/1024/1024&start=1667812368.6&end=1668417168.6&step=2419

			req, err := http.NewRequest(
				http.MethodGet, fmt.Sprintf(
					"%s/api/v1/query_range", strings.Replace(promOpts.Endpoint, "REPLACE_ME", cluster.ClusterId, 1),
				),
				nil,
			)
			if err != nil {
				return err
			}

			qp := req.URL.Query()
			qp.Add("query", promQuery)
			t := time.Now()
			qp.Add("start", fmt.Sprintf("%v", t.Add(-2*24*time.Hour).Unix()))
			qp.Add("end", fmt.Sprintf("%v", t.Unix()))
			qp.Add("step", "900") // 15 minute

			if promOpts.BasicAuthPassword != "" {
				username := func() string {
					if promOpts.BasicAuthUsername != "" {
						return promOpts.BasicAuthUsername
					}
					return cluster.ClusterId
				}()
				req.SetBasicAuth(username, promOpts.BasicAuthPassword)
			}

			req.URL.RawQuery = qp.Encode()

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return ctx.Status(http.StatusInternalServerError).Send(b)
			}

			return ctx.Send(b)
		},
	)
}

func fxMetricsQuerySvc() fx.Option {
	return fx.Invoke(metricsQuerySvc)
}
