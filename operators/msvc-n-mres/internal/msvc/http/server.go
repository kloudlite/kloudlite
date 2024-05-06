package http

import (
	"encoding/json"
	"net/http"
	"strings"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	corev1 "k8s.io/api/core/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}

type ServerArgs struct {
	Addr      string
	K8sClient client.Client
	Route     string
	Logger    logging.Logger
}

func NewServer(args ServerArgs) (*Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	counter := 0
	mux.HandleFunc(args.Route, func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		namespace := strings.TrimSpace(r.URL.Query().Get("namespace"))

		args.Logger.Infof("[%d] request received for %s/%s", counter, namespace, name)
		defer func() {
			counter += 1
		}()

		if name == "" || namespace == "" {
			http.Error(w, "query params 'name' and 'namespace' are required", http.StatusBadRequest)
			return
		}

		var msvc crdsv1.ManagedService

		if err := args.K8sClient.Get(r.Context(), fn.NN(namespace, name), &msvc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if msvc.Output.CredentialsRef.Name == "" {
			http.Error(w, "msvc output output credentials is still empty", http.StatusInternalServerError)
			return
		}

		if msvc.Spec.SharedSecret != nil && *msvc.Spec.SharedSecret != r.Header.Get("kloudlite-shared-secret") {
			http.Error(w, "invalid shared secret", http.StatusUnauthorized)
			return
		}

		var creds corev1.Secret

		if err := args.K8sClient.Get(r.Context(), fn.NN(namespace, msvc.Output.CredentialsRef.Name), &creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(creds.Data)
		if err != nil {
			http.Error(w, "invalid credentials data, failed to marshal .data", http.StatusInternalServerError)
			return
		}

		w.Write(b)
		args.Logger.Infof("request processed for %s/%s", counter, namespace, name)
	})

	return &Server{
		httpServer: &http.Server{Addr: args.Addr, Handler: mux},
	}, nil
}
