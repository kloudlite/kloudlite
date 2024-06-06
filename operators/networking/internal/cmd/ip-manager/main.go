package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/env"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/manager"
	"github.com/kloudlite/operator/pkg/kubectl"
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
			return &rest.Config{Host: "localhost:8080"}, nil
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

	manager, err := manager.NewManager(ev, yamlClient.Client(), cli)
	if err != nil {
		panic(err)
	}

	log.SetReportTimestamp(false)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}
			middleware.Logger(next).ServeHTTP(w, r)
		})
	})

	r.Post("/pod", func(w http.ResponseWriter, r *http.Request) {
		s, err := manager.RegisterPod(r.Context())
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(s)
		if err != nil {
			log.Error("unmarshalling", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	r.Get("/pod/{pod_ip}", func(w http.ResponseWriter, r *http.Request) {
		podIP := chi.URLParam(r, "pod_ip")
		wgconfig, err := manager.GetWgConfigForPod(r.Context(), podIP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(wgconfig)
		// if err := manager.RestartWireguard(); err != nil {
		// 	log.Error(err)
		// }
	})

	r.Delete("/pod/{pb_ip}/{pb_uid}", func(w http.ResponseWriter, r *http.Request) {
		pbIP, pbUID := chi.URLParam(r, "pb_ip"), chi.URLParam(r, "pb_uid")
		if err := manager.DeregisterPod(r.Context(), pbIP, pbUID); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		// if err := manager.RestartWireguard(); err != nil {
		// 	log.Error(err)
		// }
	})

	r.Post("/service/{svc_namespace}/{svc_name}", func(w http.ResponseWriter, r *http.Request) {
		result, err := manager.RegisterService(r.Context(), chi.URLParam(r, "svc_namespace"), chi.URLParam(r, "svc_name"))
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Error("marshalling result", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Put("/service/{svc_binding_name}", func(w http.ResponseWriter, r *http.Request) {
		if err := manager.RegisterAndSyncNginxStreams(r.Context(), chi.URLParam(r, "svc_binding_name")); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Delete("/service/{svc_binding_ip}/{svc_binding_uid}", func(w http.ResponseWriter, r *http.Request) {
		svcBindingIP, svcBindingUID := chi.URLParam(r, "svc_binding_ip"), chi.URLParam(r, "svc_binding_uid")
		if err := manager.DeregisterService(r.Context(), svcBindingIP, svcBindingUID); err != nil {
			log.Error("while deregistering service", "svc", svcBindingIP, "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Infof("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
