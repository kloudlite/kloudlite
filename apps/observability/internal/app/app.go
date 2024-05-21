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

	fx.Invoke(func(infraCli infra.InfraClient, kcfg *rest.Config, iamCli iam.IAMClient, mux *http.ServeMux, sessStore SessionStore, ev *env.Env, logger logging.Logger) {
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

			k8sCli, err := func() (k8s.Client, error) {
				if strings.HasPrefix(trackingId, "clus-") {
					return k8s.NewClient(kcfg, nil)
				}

				return k8s.NewClient(&rest.Config{
					Host: fmt.Sprintf("http://kloudlite-device-proxy-%s.kl-account-%s.svc.cluster.local:8080/clusters/%s", "default", accountName, clusterName),
					WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
						return httpServer.NewRoundTripperWithHeaders(rt, map[string][]string{
							"X-Kloudlite-Authz": {fmt.Sprintf("Bearer %s", ev.GlobalVPNAuthzSecret)},
						})
					},
				}, nil)
			}()
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create k8s client: %v", err), http.StatusInternalServerError)
				return
			}

			pods, err := ListPods(r.Context(), k8sCli, map[string]string{constants.ObservabilityTrackingKey: trackingId})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			podNames := make([]string, 0, len(pods))
			for _, pod := range pods {
				podNames = append(podNames, pod.Name)
			}

			if err := queryProm(ev.PromHttpAddr, PromMetricsType(metricsType), map[string]PromValue{
				"kl_account_name": {Operator: PromOperatorEqual, Value: accountName},
				"kl_cluster_name": {Operator: PromOperatorEqual, Value: clusterName},
				"kl_tracking_id":  {Operator: PromOperatorEqual, Value: trackingId},
				"pod_name":        {Operator: PromOperatorMatchRegex, Value: fmt.Sprintf("^(%s)$", strings.Join(podNames, ","))},
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

			k8sCli, err := func() (k8s.Client, error) {
				if strings.HasPrefix(trackingId, "clus-") {
					return k8s.NewClient(kcfg, nil)
				}

				return k8s.NewClient(&rest.Config{
					Host: fmt.Sprintf("http://kloudlite-device-proxy-%s.kl-account-%s.svc.cluster.local:8080/clusters/%s", "default", accountName, clusterName),
					WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
						return httpServer.NewRoundTripperWithHeaders(rt, map[string][]string{
							"X-Kloudlite-Authz": {fmt.Sprintf("Bearer %s", ev.GlobalVPNAuthzSecret)},
						})
					},
				}, nil)
			}()
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create k8s client: %v", err), http.StatusInternalServerError)
				return
			}

			pods, err := ListPods(r.Context(), k8sCli, map[string]string{constants.ObservabilityTrackingKey: trackingId})
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
							if !closed {
								http.Error(w, err.Error(), 500)
							}
						}
						return
					}
					fmt.Fprintf(w, "%s", msg)
					w.(http.Flusher).Flush()
				}
			}()

			if err := StreamLogs(r.Context(), k8sCli, pods, pw, logger); err != nil {
				http.Error(w, err.Error(), 500)
			}
		})))
	}),
)
