package app

import (
	"context"
	"fmt"
	// "io/ioutil"
	"net/http"
	"strings"
	// "time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	l "github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/fx"
	"kloudlite.io/apps/consolev2.old/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	// "kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	// "kloudlite.io/pkg/repos"
)

type PrometheusOpts struct {
	HttpPort          uint16 `env:"METRICS_HTTP_PORT" required:"true"`
	HttpCors          string `env:"METRICS_HTTP_CORS" required:"true"`
	Endpoint          string `env:"PROMETHEUS_ENDPOINT" required:"true"`
	BasicAuthUsername string `env:"PROMETHEUS_BASIC_AUTH_USERNAME" required:"false"`
	BasicAuthPassword string `env:"PROMETHEUS_BASIC_AUTH_PASSWORD" required:"false"`
}

func (p PrometheusOpts) GetHttpPort() uint16 {
	return p.HttpPort
}

func (p PrometheusOpts) GetHttpCors() string {
	return p.HttpCors
}

type PromMetricsHttpServer struct {
	*fiber.App
}

type PromMetricsType string

const (
	Cpu                PromMetricsType = "cpu"
	Memory             PromMetricsType = "memory"
	NetworkReceived    PromMetricsType = "network-received"
	NetworkTransmitted PromMetricsType = "network-transmitted"
)

func getPromQuery(resType PromMetricsType, namespace string, name string) string {
	switch resType {
	case Memory:
		return fmt.Sprintf(`sum(avg_over_time(container_memory_working_set_bytes{namespace="%s",pod=~"%s.*",container!="POD",image!=""}[30s]))/1024/1024`, namespace, name)
	case Cpu:
		// return fmt.Sprintf(
		//   `
		// 		sum(rate(container_cpu_usage_seconds_total{pod=~"^%s.*", container!=""}[2m])) by (pod, container) /
		// sum(container_spec_cpu_quota{pod=~"^%s.*", container!=""}/container_spec_cpu_period{pod=~"^%s.*", container=!""}) by (pod, container) * 1000
		// 		`, name, name, name,
		// )
		return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s.*", image!="", container!="POD"}[1m])) * 1000`, namespace, name)
	// return fmt.Sprintf(`sum without(cpu) (rate(container_cpu_usage_seconds_total{container!="",pod=~"%s.*",namespace="%s"}[1m])) * 1000`, name, namespace)
	case NetworkTransmitted:
		return fmt.Sprintf("")
	}
	return ""
}

func metricsQuerySvc(app *PromMetricsHttpServer, promOpts *PrometheusOpts, d domain.Domain, financeClient finance.FinanceClient,
	env *Env,
	cacheClient AuthCacheClient) {
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

			// app, err := d.GetApp(ctx.Context(), repos.ID(appId))
			// if err != nil {
			// 	return err
			// }

			if app == nil {
				return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": "app does not exist"})
			}

			// project, err := d.GetProjectWithID(ctx.Context(), app.ProjectId)
			// if err != nil {
			// 	return err
			// }

			// cluster, err := financeClient.GetAttachedCluster(
			// 	context.TODO(),
			// 	&finance.GetAttachedClusterIn{AccountId: string(project.AccountId)},
			// )

			// promQuery := getPromQuery(PromMetricsType(metricsType), strings.ToLower(app.Namespace), strings.ToLower(app.ReadableId))
			// if promQuery == "" {
			// 	return errors.Newf("could not build prom query, invalid (metricsType=%s or name=%s)", metricsType, app.ReadableId)
			// }

			// GET http://localhost:9090/api/v1/query_range?query=sum(container_memory_working_set_bytes{pod =~ "kl-project.*",container != ""})/1024/1024&start=1667812368.6&end=1668417168.6&step=2419

			// req, err := http.NewRequest(
			// 	http.MethodGet, fmt.Sprintf(
			// 		"%s/api/v1/query_range", strings.Replace(promOpts.Endpoint, "REPLACE_ME", cluster.ClusterId, 1),
			// 	),
			// 	nil,
			// )
			// if err != nil {
			// 	return err
			// }

			// qp := req.URL.Query()
			// qp.Add("query", promQuery)
			// t := time.Now()
			// qp.Add("start", fmt.Sprintf("%v", t.Add(-2*24*time.Hour).Unix()))
			// qp.Add("end", fmt.Sprintf("%v", t.Unix()))
			// qp.Add("step", "700") // 15 minute

			// if promOpts.BasicAuthPassword != "" {
			// 	username := func() string {
			// 		if promOpts.BasicAuthUsername != "" {
			// 			return promOpts.BasicAuthUsername
			// 		}
			// 		return cluster.ClusterId
			// 	}()
			// 	req.SetBasicAuth(username, promOpts.BasicAuthPassword)
			// }

			// req.URL.RawQuery = qp.Encode()

			// fmt.Println(req.URL.String())

			// resp, err := http.DefaultClient.Do(req)
			// if err != nil {
			// 	return err
			// }

			// b, err := ioutil.ReadAll(resp.Body)
			// if err != nil {
			// 	return err
			// }

			// if resp.StatusCode != http.StatusOK {
			// 	return ctx.Status(http.StatusInternalServerError).Send(b)
			// }

			// return ctx.Send(b)
			return ctx.Send([]byte(""))
		},
	)
}

func fxMetricsQuerySvc() fx.Option {
	return fx.Module(
		"http-server",
		fx.Provide(
			func() *PromMetricsHttpServer {
				return &PromMetricsHttpServer{App: fiber.New()}
			},
		),

		fx.Invoke(
			func(lf fx.Lifecycle, p *PrometheusOpts, logger logging.Logger, app *PromMetricsHttpServer) {
				app.Use(
					l.New(
						l.Config{
							Format:     "${time} ${status} - ${method} ${latency} \t ${path} \n",
							TimeFormat: "02-Jan-2006 15:04:05",
							TimeZone:   "Asia/Kolkata",
						},
					),
				)
				if p.GetHttpCors() != "" {
					app.Use(
						cors.New(
							cors.Config{
								AllowOrigins:     p.GetHttpCors(),
								AllowCredentials: true,
								AllowMethods: strings.Join(
									[]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
									",",
								),
							},
						),
					)
				}

				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return httpServer.Start(ctx, p.GetHttpPort(), app.App, logger)
						},
						OnStop: func(ctx context.Context) error {
							return app.Shutdown()
						},
					},
				)
			},
		),

		fx.Invoke(metricsQuerySvc),
	)
}
