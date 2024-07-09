package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/env"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/manager"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "--address <host>:<port>")

	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	flag.Parse()

	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}

	ev.IsDev = isDev

	restCfg, err := func() (*rest.Config, error) {
		if isDev {
			return &rest.Config{Host: "localhost:8080", QPS: 200, Burst: 200}, nil
		}
		return rest.InClusterConfig()
	}()
	if err != nil {
		panic(err)
	}

	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	networkingv1.AddToScheme(scheme)

	cli, err := client.New(restCfg, client.Options{
		Scheme: scheme,
		WarningHandler: client.WarningHandlerOptions{
			SuppressWarnings:   true,
			AllowDuplicateLogs: false,
		},
	})
	if err != nil {
		panic(err)
	}

	yamlClient, err := kubectl.NewYAMLClient(restCfg, kubectl.YAMLClientOpts{})
	if err != nil {
		panic(err)
	}

	logger := logging.NewSlogLogger(logging.SlogOptions{
		Writer:        os.Stderr,
		ShowTimestamp: false,
		ShowCaller:    true,
		LogLevel: func() slog.Level {
			if isDev {
				return slog.LevelDebug
			}
			return slog.LevelInfo
		}(),
	})

	mg, err := manager.NewManager(ev, logger, yamlClient.Client(), cli)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	httpLogger := logging.NewHttpLogger(logger, logging.HttpLoggerOptions{
		SilentPaths: []string{"/healthz"},
	})

	r.Use(httpLogger.Use)

	r.Put("/pod/{pod_namespace}/{pod_name}/{pod_ip}", func(w http.ResponseWriter, r *http.Request) {
		podNamespace, podName, podIP := chi.URLParam(r, "pod_namespace"), chi.URLParam(r, "pod_name"), chi.URLParam(r, "pod_ip")
		wgconfig, err := mg.GetWgConfigForReservedPod(r.Context(), manager.WgConfigForReservedPodArgs{
			PodNamespace: podNamespace,
			PodName:      podName,
			PodIP:        podIP,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(wgconfig)
	})

	r.Delete("/pod/{pod_namespace}/{pod_name}", func(w http.ResponseWriter, r *http.Request) {
		podNamespace, podName := chi.URLParam(r, "pod_namespace"), chi.URLParam(r, "pod_name")
		if err := mg.DeregisterPod(r.Context(), podNamespace, podName); err != nil {
			logger.Error("deregistering pod", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Put("/service/{svc_namespace}/{svc_name}", func(w http.ResponseWriter, r *http.Request) {
		if err := mg.ReserveService(r.Context(), chi.URLParam(r, "svc_namespace"), chi.URLParam(r, "svc_name")); err != nil {
			logger.Error("reserving service", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Delete("/service/{svc_namespace}/{svc_name}", func(w http.ResponseWriter, r *http.Request) {
		svcNamespace, svcName := chi.URLParam(r, "svc_namespace"), chi.URLParam(r, "svc_name")
		if err := mg.DeregisterService(r.Context(), svcNamespace, svcName); err != nil {
			logger.Error("deregistering service", "svc", fmt.Sprintf("%s/%s", svcNamespace, svcName), "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	common.PrintReadyBanner()
	logger.Info("starting http server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("failed to start http server", "err", err)
		os.Exit(1)
	}
}
