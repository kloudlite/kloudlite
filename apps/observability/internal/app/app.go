package app

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kloudlite/api/apps/observability/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/infra"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
)

type (
	IAMGrpcClient grpc.Client
	InfraClient   grpc.Client
	SessionStore  kv.Repo[*common.AuthSession]
)

var Module = fx.Module(
	"app",

	fx.Provide(
		func(conn IAMGrpcClient) iam.IAMClient {
			return iam.NewIAMClient(conn)
		},
	),

	fx.Provide(func(conn InfraClient) infra.InfraClient {
		return infra.NewInfraClient(conn)
	}),

	fx.Provide(func(cfg *rest.Config) (k8s.Client, error) {
		return k8s.NewClient(cfg, nil)
	}),

	fx.Invoke(func(infraCli infra.InfraClient, kcli k8s.Client, iamCli iam.IAMClient, mux *http.ServeMux, sessStore SessionStore, ev *env.Env, logger logging.Logger) {
		sessionMiddleware := httpServer.NewReadSessionMiddlewareHandler(sessStore, constants.CookieName, constants.CacheSessionPrefix)

		loggingMiddleware := httpServer.NewLoggingMiddleware(logger)

		mux.HandleFunc("/observability/metrics/", loggingMiddleware(sessionMiddleware(func(w http.ResponseWriter, r *http.Request) {
			metricsType := strings.TrimPrefix(r.URL.Path, "/observability/metrics/")

			sess := httpServer.GetHttpSession[*common.AuthSession](r.Context())
			if sess == nil {
				http.Error(w, "not logged in", http.StatusUnauthorized)
				return
			}

			m, ok := r.Context().Value("http-cookies").(map[string]string)
			if !ok {
				m = map[string]string{}
			}

			accountName := m[ev.AccountCookieName]
			if accountName == "" {
				http.Error(w, fmt.Sprintf("no cookie named '%s' present in request", ev.AccountCookieName), http.StatusBadRequest)
				return
			}

			clusterName := r.URL.Query().Get("cluster_name")
			if clusterName == "" {
				http.Error(w, "query param (cluster_name) must be provided", http.StatusBadRequest)
				return
			}

			trackingId := r.URL.Query().Get("tracking_id")
			if trackingId == "" {
				http.Error(w, "query param (tracking_id) must be provided", http.StatusBadRequest)
			}

			can, err := iamCli.Can(r.Context(), &iam.CanIn{
				UserId: string(sess.UserId),
				ResourceRefs: []string{
					iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
				},
				Action: string(iamT.ReadMetrics),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !can.Status {
				http.Error(w, errors.NewEf(err, "unauthorized to view metrics for resources belonging to account (%s)", accountName).Error(), http.StatusUnauthorized)
				return
			}

			st := r.URL.Query().Get("start_time")
			if st == "" {
				st = fmt.Sprintf("%d", time.Now().Add(-3*time.Hour).Unix())
			}

			et := r.URL.Query().Get("end_time")
			if et == "" {
				et = fmt.Sprintf("%d", time.Now().Unix())
			}

			step := r.URL.Query().Get("step")
			if step == "" {
				step = "15s"
			}

			if err := queryProm(ev.PromHttpAddr, PromMetricsType(metricsType), map[string]string{
				"kl_account_name": accountName,
				"kl_cluster_name": clusterName,
				"kl_tracking_id":  trackingId,
			}, st, et, step, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})))

		mux.HandleFunc("/observability/logs", loggingMiddleware(sessionMiddleware(func(w http.ResponseWriter, r *http.Request) {
			sess := httpServer.GetHttpSession[*common.AuthSession](r.Context())
			if sess == nil {
				http.Error(w, "not logged in", http.StatusUnauthorized)
				return
			}

			m, ok := r.Context().Value("http-cookies").(map[string]string)
			if !ok {
				m = map[string]string{}
			}

			accountName := m[ev.AccountCookieName]
			if accountName == "" {
				http.Error(w, fmt.Sprintf("no cookie named '%s' present in request", ev.AccountCookieName), http.StatusBadRequest)
				return
			}

			clusterName := r.URL.Query().Get("cluster_name")
			trackingId := r.URL.Query().Get("tracking_id")

			if !strings.HasPrefix(trackingId, "clus-") {
				out, err := infraCli.GetClusterKubeconfig(r.Context(), &infra.GetClusterIn{
					UserId:      string(sess.UserId),
					UserName:    sess.UserName,
					UserEmail:   sess.UserEmail,
					AccountName: accountName,
					ClusterName: clusterName,
				})
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				cfg, err := k8s.RestConfigFromKubeConfig(out.Kubeconfig)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				kcli, err = k8s.NewClient(cfg, nil)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
			}

			pods, err := ListPods(r.Context(), kcli, map[string]string{constants.ObservabilityTrackingKey: trackingId})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if len(pods) == 0 {
				// it sends http.StatusTooEarly, for the client to retry request after some time
				logger.Infof("no pods found")
				http.Error(w, "no pods found", http.StatusTooEarly)
				return
			}

			closed := false
			go func() {
				for {
					if err := r.Context().Err(); err != nil {
						closed = true
						return
					}
					<-time.After(100 * time.Millisecond)
				}
			}()

			pr, pw := io.Pipe()

			go func() {
				b := bufio.NewReader(pr)
				for !closed {
					msg, err := b.ReadBytes('\n')
					if err != nil {
						if !errors.Is(err, io.EOF) {
							http.Error(w, err.Error(), 500)
						}
						return
					}
					fmt.Fprintf(w, "%s", msg)
					w.(http.Flusher).Flush()
				}
			}()

			if err := StreamLogs(r.Context(), kcli, pods, pw, logger); err != nil {
				http.Error(w, err.Error(), 500)
			}
		})))
	}),
)
