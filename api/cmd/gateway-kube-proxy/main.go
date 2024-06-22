package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var addr string
	var proxyAddr string
	var debug bool
	var authz string

	flag.BoolVar(&debug, "debug", false, "--debug")
	flag.StringVar(&addr, "addr", ":8080", "--addr <host:port>")
	flag.StringVar(&proxyAddr, "proxy-addr", "", "--proxy-addr <host:port>")
	flag.StringVar(&authz, "authz", "", "--authz <authz-token>")
	flag.Parse()

	if authz == "" {
		panic("authz token is required, use --authz <authz-token>")
	}

	logger := slog.Default()

	reverseProxyMap := make(map[string]*httputil.ReverseProxy)

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

	kloudliteAuthzHeader := "X-Kloudlite-Authz"

	r.HandleFunc("/clusters/{cluster_name}/*", func(w http.ResponseWriter, req *http.Request) {
		token := strings.TrimPrefix(req.Header.Get(kloudliteAuthzHeader), "Bearer ")
		logger.Info("request", "method", req.Method, "url", req.URL, "token", token)
		if len(token) != len(authz) || token != authz {
			http.Error(w, "UnAuthorized", http.StatusUnauthorized)
			return
		}

		sp := strings.Split(strings.TrimPrefix(req.URL.Path, "/clusters/"), "/")
		if len(sp) <= 1 {
			http.Error(w, "invalid request", http.StatusForbidden)
			return
		}

		// clusterName := sp[0]
		clusterName := chi.URLParam(req, "cluster_name")

		urlh := strings.ReplaceAll(proxyAddr, "{{.CLUSTER_NAME}}", clusterName)
		urlp := fmt.Sprintf("/%s", strings.Join(sp[1:], "/"))

		reverseProxy, ok := reverseProxyMap[clusterName]
		if !ok {
			reverseProxy = &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = "http"
					req.URL.Host = urlh
					req.URL.Path = urlp
					req.Header.Del(kloudliteAuthzHeader)
					logger.Info("reverse proxy enabled for", "cluster", clusterName, "to-host", urlh, "at path", urlp)
				},
			}
		}

		reverseProxy.ServeHTTP(w, req)
	})

	logger.Info("starting http server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
