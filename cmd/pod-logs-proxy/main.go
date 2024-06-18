package main

import (
	"bufio"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var addr string
	var kubeAddr string

	flag.StringVar(&addr, "addr", ":8080", "--addr <host:port>")
	flag.StringVar(&kubeAddr, "kube-addr", "", "--kube-addr <host:port>")
	flag.Parse()

	kcli, err := func() (*kubernetes.Clientset, error) {
		if kubeAddr == "" {
			rcfg, err := rest.InClusterConfig()
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(rcfg)
		}
		return kubernetes.NewForConfig(&rest.Config{Host: kubeAddr})
	}()
	if err != nil {
		panic(err)
	}

	logger := slog.Default()

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthy" {
				next.ServeHTTP(w, r)
				return
			}
			middleware.Logger(next).ServeHTTP(w, r)
		})
	})

	r.HandleFunc("/healthy", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// r.Route("/api/v1/namespaces/{namespace}/pods", func(r chi.Router) {
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		urlp := req.URL.Path

		kreq := kcli.RESTClient().Get().AbsPath(urlp)
		for k, vv := range req.URL.Query() {
			kreq = kreq.Param(k, vv[0])
		}
		logger.Debug("request", "url", kreq.URL(), "query", kreq.URL().Query())
		rc, err := kreq.Stream(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		defer rc.Close()

		reader := bufio.NewReader(rc)
		for {
			b, err := reader.ReadBytes('\n')
			w.Write(b)
			w.(http.Flusher).Flush()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					logger.Error("got non EOF reading from stream", "error", err)
				}
				return
			}
		}
	})
	// })

	logger.Info("starting pod logs proxy http server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
