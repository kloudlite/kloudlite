package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"log"

	flag "github.com/spf13/pflag"

	"github.com/nxtcoder17/kubelet-metrics-reexporter/internal/parser"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var addr string
	var isDev bool

	var nodeName string

	var enrichFromLabels bool
	var enrichFromAnnotations bool
	var enrichTags []string

	var filterPrefixes []string
	var replacePrefixes []string

	flag.StringVar(&addr, "addr", "0.0.0.0:9100", "--addr <host:port>")
	flag.BoolVar(&isDev, "dev", false, "--dev")

	flag.StringVar(&nodeName, "node-name", "", "--node-name <node-name> (if not provided, value is read from NODE_NAME env var)")
	flag.BoolVar(&enrichFromLabels, "enrich-from-labels", false, "--enrich-from-labels")
	flag.BoolVar(&enrichFromAnnotations, "enrich-from-annotations", false, "--enrich-from-annotations")
	flag.StringSliceVar(&enrichTags, "enrich-tag", []string{}, "--enrich-tag <key>=<value> (can be used multiple times)")
	flag.StringSliceVar(&filterPrefixes, "filter-prefix", []string{}, "--filter-prefix <prefix> (can be used multiple times)")
	flag.StringSliceVar(&replacePrefixes, "replace-prefix", []string{}, "--replace-prefix <old>=<new> (can be used multiple times)")
	flag.Parse()

	logger := log.New(os.Stdout, "[kubelet-metrics] ", 0)

	if nodeName == "" {
		logger.Printf("reading nodeName from `NODE_NAME` env-var")
		v, ok := os.LookupEnv("NODE_NAME")
		if !ok {
			log.Fatal("NODE_NAME env var not set")
		}
		nodeName = v
	}
	logger.Printf("NODE NAME is %q", nodeName)

	restCfg, err := func() (*rest.Config, error) {
		if isDev {
			return &rest.Config{Host: "localhost:8080"}, nil
		}
		return rest.InClusterConfig()
	}()
	if err != nil {
		log.Fatalln(err)
	}

	kCli, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		log.Fatalln(err)
	}

	grabPodsByNodeName := func(nodeName string) (map[string]corev1.Pod, error) {
		pl, err := kCli.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		})
		if err != nil {
			return nil, err
		}

		m := make(map[string]corev1.Pod, len(pl.Items))
		for _, pod := range pl.Items {
			m[types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}.String()] = pod
		}
		return m, nil
	}

	parserOpts := parser.ParserOpts{
		EnrichTags: func() map[string]string {
			m := make(map[string]string, len(enrichTags))
			for i := range enrichTags {
				sp := strings.SplitN(enrichTags[i], "=", 2)
				if len(sp) != 2 {
					log.Fatalln("invalid enrich-tag format, should be <key>=<value>")
				}
				m[sp[0]] = sp[1]
			}
			return m
		}(),
		EnrichFromLabels:      enrichFromLabels,
		EnrichFromAnnotations: enrichFromAnnotations,
		FilterPrefixes:        filterPrefixes,
		ReplacePrefixes: func() map[string]string {
			m := make(map[string]string, len(replacePrefixes))
			for i := range replacePrefixes {
				sp := strings.SplitN(replacePrefixes[i], "=", 2)
				if len(sp) != 2 {
					log.Fatalln("invalid replace-prefix format, should be <old>=<new>")
				}
				m[sp[0]] = sp[1]
			}
			return m
		}(),
	}

	http.HandleFunc("/metrics/resource", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		defer func() {
			logger.Printf("GET [%s] at %s, and took %.2fs\n", r.URL, time.Now().Format(time.RFC3339), time.Since(t).Seconds())
		}()

		parserOpts.PodsMap, err = grabPodsByNodeName(nodeName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		mParser := parser.NewParser(kCli, nodeName, parserOpts)

		req := kCli.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/metrics/resource", nodeName))
		b, err := req.DoRaw(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if r.URL.Query().Get("raw") == "true" {
			w.Write(b)
			return
		}

		buff := new(bytes.Buffer)
		if err = mParser.ParseAndEnhanceMetricsInto(b, buff); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		io.Copy(w, buff)
	})

	http.HandleFunc("/metrics/probes", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	})

	logger.Printf("http server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln(err)
	}
}
