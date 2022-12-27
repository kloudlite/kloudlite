package main

import (
	"flag"
	"fmt"
	"github.com/codingconcepts/env"
	"log"
	"net/http"
	"net/http/httputil"
)

type envName string

func proxyWithoutPooling(w http.ResponseWriter, req *http.Request) {
	if v := req.Header.Get("kloudlite-proxy"); v != "" {
		remote := req.URL
		remote.Scheme = "http"
		remote.Host = fmt.Sprintf("env.%s.%s", v, req.Host)
		fmt.Printf("proxying to remote: %+v %s\n", remote, remote.String())
		proxy := httputil.NewSingleHostReverseProxy(remote)
		req.Host = remote.Host
		proxy.ServeHTTP(w, req)
		return
	}

	for _, c := range req.Cookies() {
		if c.Name == "kloudlite-env" {
			remote := req.URL
			remote.Scheme = "http"
			remote.Host = fmt.Sprintf("env.%s.%s", c.Value, req.Host)
			fmt.Printf("[remote]: host: %-30s uri: %-20s \n", remote.Host, remote.RequestURI())
			proxy := httputil.NewSingleHostReverseProxy(remote)
			req.Host = remote.Host
			proxy.ServeHTTP(w, req)
			return
		}
	}

	w.Write([]byte("bad proxy request"))
}

type Env struct {
	PrimaryEnvName string `env:"PRIMARY_ENV_NAME"`
}

func main() {
	var isDev bool
	var addr string

	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.StringVar(&addr, "addr", "0.0.0.0:80", "--addr <host:port>")
	flag.Parse()

	ev := &Env{}
	if err := env.Set(ev); err != nil {
		panic(err)
	}

	//scheme := runtime.NewScheme()
	//utilruntime.Must(crdsv1.AddToScheme(scheme))
	//
	//kClient, err := client.New(
	//	func() *rest.Config {
	//		if isDev {
	//			return &rest.Config{Host: devServerAddr}
	//		}
	//		config, err := rest.InClusterConfig()
	//		if err != nil {
	//			panic(err)
	//		}
	//		return config
	//	}(), client.Options{Scheme: scheme},
	//)
	//
	//if err != nil {
	//	panic(err)
	//}

	//var envList crdsv1.SecondaryEnvList
	//if err := kClient.List(context.TODO(), &envList, &client.ListOptions{
	//	LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
	//		"kloudlite.io/primary-env": ev.PrimaryEnvName,
	//	}),
	//}); err != nil {
	//	panic(err)
	//}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("[incoming] host: %-30s uri: %-20s\n", req.Host, req.RequestURI)
		proxyWithoutPooling(w, req)
	})

	fmt.Printf("starting http server on %s ...\n", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
