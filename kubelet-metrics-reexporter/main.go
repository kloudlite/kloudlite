package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/kloudlite/kubelet-metrics-reexporter/internal/kloudlite"
	"github.com/kloudlite/kubelet-metrics-reexporter/internal/parser"
	"github.com/kloudlite/kubelet-metrics-reexporter/pkg/k8s"
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

	extraTags := make(map[string]string, len(enrichTags))
	for i := range enrichTags {
		s := strings.SplitN(enrichTags[i], "=", 2)
		if len(s) != 2 {
			continue
		}
		extraTags[s[0]] = s[1]
	}

	fmt.Printf("extra tags: %+v\n", extraTags)

	kcli, err := k8s.NewClient(restCfg)
	if err != nil {
		logger.Fatalln(err)
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

		mParser, err := parser.NewParser(r.Context(), kcli, nodeName, parserOpts)
		if err != nil {
			logger.Printf("[ERROR]: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := kcli.MetricsResource(r.Context(), nodeName)
		if err != nil {
			logger.Printf("[ERROR]: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

		if r.URL.Query().Get("raw") == "true" {
			b, err := kcli.StatsSummaryRaw(r.Context(), nodeName)
			if err != nil {
				logger.Printf("[ERROR]: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(b)
			return
		}

		out := bytes.NewBuffer(nil)
		ma, err := kloudlite.NewMetricsAggregator(r.Context(), kcli, nodeName, extraTags)
		if err != nil {
			logger.Printf("[ERROR]: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ma.WriteNodeMetrics(out)
		ma.WritePodMetrics(out)
		io.Copy(w, out)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		defer func() {
			logger.Printf("GET [%s] at %s, and took %.2fs\n", r.URL, time.Now().Format(time.RFC3339), time.Since(t).Seconds())
		}()

		b, err := kcli.Metrics(r.Context(), nodeName)
		if err != nil {
			logger.Printf("[ERROR]: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	logger.Printf("http server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln(err)
	}
}
