package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func httpProxyToHost(w http.ResponseWriter, req *http.Request, remoteHost string) {
	remote := *req.URL
	remote.Path, remote.RawPath = "/", "/"
	remote.Scheme = "http"
	remote.Host = remoteHost
	fmt.Printf("[REMOTE] [host] %-30s [uri] %-20s \n", remote.Host, req.RequestURI)
	proxy := httputil.NewSingleHostReverseProxy(&remote)
	req.Host = remote.Host
	proxy.ServeHTTP(w, req)
}

func proxyWithoutPooling(identifier string, w http.ResponseWriter, req *http.Request) {
	if v := req.Header.Get(identifier); v != "" {
		httpProxyToHost(w, req, fmt.Sprintf("env.%s.%s", v, req.Host))
		return
	}

	for _, c := range req.Cookies() {
		if c.Name == identifier {
			httpProxyToHost(w, req, fmt.Sprintf("env.%s.%s", c.Value, req.Host))
		}
	}

	httpProxyToHost(w, req, fmt.Sprintf("env.%s.%s", "default", req.Host))
	// w.Write([]byte("bad proxy request"))
}

func main() {
	var isDev bool
	var addr string
	var identifier string

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&addr, "addr", "0.0.0.0:80", "--addr <host:port>")
	flag.StringVar(&identifier, "identifier", "kloudlite-env", "--identifier <identifier-name>")
	flag.Parse()

	http.HandleFunc("/.kl/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("[INCOMING] [host] %-30s [uri] %-20s\n", req.Host, req.RequestURI)
		proxyWithoutPooling(identifier, w, req)
	})

	fmt.Printf("starting http server on %s ...\n", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
