package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
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
		log.Fatal("authz token, must be provided")
	}

	reverseProxyMap := make(map[string]*httputil.ReverseProxy)

	mux := http.NewServeMux()

	kloudliteAuthzHeader := "X-Kloudlite-Authz"

	counter := 1
	mux.HandleFunc("/clusters/", func(w http.ResponseWriter, req *http.Request) {
		token := strings.TrimPrefix(req.Header.Get(kloudliteAuthzHeader), "Bearer ")
		fmt.Println("HERE", token)

		if len(token) != len(authz) || token != authz {
			http.Error(w, "UnAuthorized", http.StatusUnauthorized)
			return
		}

		sp := strings.Split(strings.TrimPrefix(req.URL.Path, "/clusters/"), "/")
		if len(sp) <= 1 {
			http.Error(w, "invalid request", http.StatusForbidden)
			return
		}

		clusterName := sp[0]

		start := time.Now()
		if debug {
			log.Printf("[%d] request received /%s\n", counter, strings.Join(sp[1:], "/"))
		}
		defer func() {
			log.Printf("[%d] (took %.2fs) /%s\n", counter, time.Since(start).Seconds(), strings.Join(sp[1:], "/"))
			counter = counter + 1
		}()

		if clusterName == "" {
			http.Error(w, "kloudlite-cluster is missing", http.StatusForbidden)
			return
		}

		reverseProxy, ok := reverseProxyMap[clusterName]
		if !ok {
			reverseProxy = &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = "http"
					req.URL.Host = strings.ReplaceAll(proxyAddr, "{{.CLUSTER_NAME}}", clusterName)
					req.URL.Path = fmt.Sprintf("/%s", strings.Join(sp[1:], "/"))
					req.Header.Del(kloudliteAuthzHeader)
				},
			}
		}

		reverseProxy.ServeHTTP(w, req)
	})

	log.Fatal(http.ListenAndServe(addr, mux))
}
