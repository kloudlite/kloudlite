package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func httpProxyToHost(counter int64, w http.ResponseWriter, req *http.Request, remoteHost string) {
	remote := *req.URL
	remote.Path, remote.RawPath = "/", "/"
	remote.Scheme = "http"
	remote.Host = remoteHost
	fmt.Printf("(%d) [INCOMING] [host] %-30s [uri] %-20s\n", counter, req.Host, req.RequestURI)
	fmt.Printf("(%d) [REMOTE  ] [host] %-30s [uri] %-20s\n", counter, remote.Host, req.RequestURI)
	proxy := httputil.NewSingleHostReverseProxy(&remote)
	req.Host = remote.Host
	proxy.ServeHTTP(w, req)
}

func proxyWithoutPooling(counter int64, identifier string, w http.ResponseWriter, req *http.Request) {
	if v := req.Header.Get(identifier); v != "" {
		httpProxyToHost(counter, w, req, fmt.Sprintf("env.%s.%s", v, req.Host))
		return
	}

	for _, c := range req.Cookies() {
		if c.Name == identifier {
			httpProxyToHost(counter, w, req, fmt.Sprintf("env.%s.%s", c.Value, req.Host))
		}
	}

	httpProxyToHost(counter, w, req, fmt.Sprintf("env.%s.%s", "default", req.Host))
}

func main() {
	var isDev bool
	var addr string
	var identifier string

	var counter int64

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&addr, "addr", "0.0.0.0:80", "--addr <host:port>")
	flag.StringVar(&identifier, "identifier", "kloudlite-workspace", "--identifier <identifier-name>")
	flag.Parse()

	http.HandleFunc("/.kl/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		counter += 1
		proxyWithoutPooling(counter, identifier, w, req)
	})

	fmt.Printf("starting http server on %s ...\n", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
