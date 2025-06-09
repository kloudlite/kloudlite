package main

import (
	"bufio"
	"errors"
	"flag"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kloudlite/operator/pkg/logging"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var addr string
	var kubeAddr string

	flag.StringVar(&addr, "addr", ":8080", "--addr <host:port>")
	flag.StringVar(&kubeAddr, "kube-addr", "", "--kube-addr <host:port>")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

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

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug})
	httpLogger := logging.NewHttpLogger(logging.HttpLoggerOptions{SilentPaths: []string{"/_healthy"}})

	r := chi.NewRouter()
	r.Use(httpLogger.Use)
	// r.Use(middleware.Logger)

	r.HandleFunc("/_healthy", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
			logger.Debug("read", "line", string(b))
			w.Write(b)
			w.(http.Flusher).Flush()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					logger.Error("got non EOF reading from stream", "err", err)
				}
				return
			}
		}
	})

	logger.Info("starting pod logs proxy http server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
