package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func proxyWithoutPooling(w http.ResponseWriter, req *http.Request) {
	if v := req.Header.Get("kloudlite-proxy"); v != "" {
		remote := *req.URL
		remote.Path, remote.RawPath = "/", "/"
		remote.Scheme = "http"
		remote.Host = fmt.Sprintf("env.%s.%s", v, req.Host)
		fmt.Printf("[REMOTE] [host] %-30s [uri] %-20s \n", remote.Host, req.RequestURI)
		proxy := httputil.NewSingleHostReverseProxy(&remote)
		req.Host = remote.Host
		proxy.ServeHTTP(w, req)
		return
	}

	for _, c := range req.Cookies() {
		if c.Name == "kloudlite-env" {
			remote := *req.URL
			remote.Path, remote.RawPath = "/", "/"
			remote.Scheme = "http"
			remote.Host = fmt.Sprintf("env.%s.%s", c.Value, req.Host)
			fmt.Printf("[REMOTE] [host] %-30s [uri] %-20s \n", remote.Host, req.RequestURI)
			proxy := httputil.NewSingleHostReverseProxy(&remote)
			req.Host = remote.Host
			proxy.ServeHTTP(w, req)
			return
		}
	}

	w.Write([]byte("bad proxy request"))
}

func main() {
	var isDev bool
	var addr string

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&addr, "addr", "0.0.0.0:80", "--addr <host:port>")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("[INCOMING] [host] %-30s [uri] %-20s\n", req.Host, req.RequestURI)
		proxyWithoutPooling(w, req)
	})

	fmt.Printf("starting http server on %s ...\n", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
