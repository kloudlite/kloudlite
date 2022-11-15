package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type PrometheusOpts struct {
	Endpoint          string `env:"PROMETHEUS_ENDPOINT" required:"true"`
	BasicAuthUsername string `env:"PROMETHEUS_BASIC_AUTH_USERNAME" required:"false"`
	BasicAuthPassword string `env:"PROMETHEUS_BASIC_AUTH_PASSWORD" required:"false"`
}

func metricsQuerySvc(app *fiber.App, promOpts *PrometheusOpts, d domain.Domain, financeClient finance.FinanceClient) {
	app.Get("/metrics/:appId", func(ctx *fiber.Ctx) error {
		appId := ctx.Params("appId", "")
		if appId == "" {
			return ctx.Status(http.StatusBadRequest).JSON(errors.New("appId is empty"))
		}

		app, err := d.GetApp(ctx.Context(), repos.ID(appId))
		if err != nil {
			fmt.Println(err)
		}

		if app == nil {
			return ctx.Status(http.StatusBadRequest).JSON("app does not exist")
		}

		project, err := d.GetProjectWithID(ctx.Context(), app.ProjectId)
		if err != nil {
			return err
		}

		cluster, err := financeClient.GetAttachedCluster(
			context.TODO(),
			&finance.GetAttachedClusterIn{AccountId: string(project.AccountId)},
		)

		// GET
		// http://localhost:9090/api/v1/query_range?query=sum(container_memory_working_set_bytes{pod =~ "kl-project.*", container != ""}) / 1024/1024&start=1667812368.6&end=1668417168.6&step=2419

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/query_range", promOpts.Endpoint), nil)
		if err != nil {
			return err
		}

		qp := req.URL.Query()
		qp.Add("query", fmt.Sprintf(`sum(container_memory_working_set_bytes{pod =~ "%s.*", container != ""}) / 1024/1024`, app.Name))
		t := time.Now()
		qp.Add("start", fmt.Sprintf("%v", t.Add(-2*24*time.Hour).Unix()))
		qp.Add("end", fmt.Sprintf("%v", t.Unix()))
		qp.Add("step", "900") // 1 hour

		if promOpts.BasicAuthPassword != "" {
			req.SetBasicAuth(cluster.ClusterId, promOpts.BasicAuthPassword)
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
		//
		if resp.StatusCode != http.StatusOK {
			return ctx.Status(http.StatusInternalServerError).JSON(string(b))
		}

		return ctx.JSON(string(b))
	})
}

func fxMetricsQuerySvc() fx.Option {
	return fx.Invoke(metricsQuerySvc)
}
