package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/nxtcoder17/kubelet-metrics-reexporter/internal/kloudlite"
	"github.com/nxtcoder17/kubelet-metrics-reexporter/internal/parser"
	"github.com/nxtcoder17/kubelet-metrics-reexporter/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kubelet_stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	// "github.com/kubernetes/kubelet/blob/master/pkg/apis/stats/v1alpha1/types.go"
	corev1 "k8s.io/api/core/v1"
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

	var shouldValidateMetricLabel bool
	var metricLabelValidationRegex string

	// source: https://prometheus.io/docs/concepts/data_model/
	const MetricLabelValidationRegex = `^[a-zA-Z_][a-zA-Z0-9_]*$`

	flag.StringVar(&addr, "addr", "0.0.0.0:9100", "--addr <host:port>")
	flag.BoolVar(&isDev, "dev", false, "--dev")

	flag.StringVar(&nodeName, "node-name", "", "--node-name <node-name> (if not provided, value is read from NODE_NAME env var)")
	flag.BoolVar(&enrichFromLabels, "enrich-from-labels", false, "--enrich-from-labels")
	flag.BoolVar(&enrichFromAnnotations, "enrich-from-annotations", false, "--enrich-from-annotations")
	flag.StringSliceVar(&enrichTags, "enrich-tag", []string{}, "--enrich-tag <key>=<value> (can be used multiple times)")
	flag.StringSliceVar(&filterPrefixes, "filter-prefix", []string{}, "--filter-prefix <prefix> (can be used multiple times)")
	flag.StringSliceVar(&replacePrefixes, "replace-prefix", []string{}, "--replace-prefix <old>=<new> (can be used multiple times)")

	flag.BoolVar(&shouldValidateMetricLabel, "--validate", true, "--validate")
	flag.StringVar(&metricLabelValidationRegex, "--validation-regex", MetricLabelValidationRegex, "--validation-regex '<regex>'")

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

	getCurrentNode := func() (*corev1.Node, error) {
		node, err := kCli.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	grabPodsByNodeName := func(nodeName string) (types.PodsMap, error) {
		pl, err := kCli.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		})
		if err != nil {
			return nil, err
		}

		return types.ToPodsMap(pl.Items), nil
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

		ShouldValidateMetricLabel: shouldValidateMetricLabel,
		ValidLabelRegexExpr:       metricLabelValidationRegex,
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

		mParser, err := parser.NewParser(kCli, nodeName, parserOpts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

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

	http.HandleFunc("/metrics/kloudlite", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		defer func() {
			logger.Printf("GET [%s] at %s, and took %.2fs\n", r.URL, time.Now().Format(time.RFC3339), time.Since(t).Seconds())
		}()

		req := kCli.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodeName))
		b, err := req.DoRaw(r.Context())
		if err != nil {
			logger.Printf("[ERR] %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if r.URL.Query().Get("raw") == "true" {
			w.Write(b)
			return
		}

		var summary kubelet_stats.Summary
		if err := json.Unmarshal(b, &summary); err != nil {
			logger.Printf("[ERR] %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		podsMap, err := grabPodsByNodeName(nodeName)
		if err != nil {
			logger.Printf("[ERR] %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currNode, err := getCurrentNode()
		if err != nil {
			logger.Printf("[ERR] %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := bytes.NewBuffer(nil)
		kloudlite.Metrics(summary, currNode, podsMap, out)
		io.Copy(w, out)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		defer func() {
			logger.Printf("GET [%s] at %s, and took %.2fs\n", r.URL, time.Now().Format(time.RFC3339), time.Since(t).Seconds())
		}()

		req := kCli.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/metrics", nodeName))
		b, err := req.DoRaw(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if r.URL.Query().Get("raw") == "true" {
			w.Write(b)
			return
		}
	})

	logger.Printf("http server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln(err)
	}
}
