package webhook

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	"github.com/codingconcepts/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/logging"

	fn "github.com/kloudlite/operator/pkg/functions"
	"k8s.io/client-go/dynamic"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type Resource string

const (
	ResourcePod Resource = "pod"
)

const (
	debugWebhookAnnotation string = "kloudlite.io/networking.webhook.debug"
)

type Env struct {
	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY"`
}

type HandlerContext struct {
	context.Context
	Resource
	*slog.Logger

	client          *dynamic.DynamicClient
	CreatedForLabel string
}

type RunArgs struct {
	Addr            string
	LogLevel        string
	KubeRestConfig  *rest.Config
	CreatedForLabel string

	TLSCertFile string
	TLSKeyFile  string
}

func Run(args RunArgs) error {
	start := time.Now()

	if args.CreatedForLabel == "" {
		args.CreatedForLabel = "kloudlite.io/created-for"
	}

	if args.TLSCertFile == "" || args.TLSKeyFile == "" {
		return fmt.Errorf("must provide TLSCertFile and TLSKeyFile")
	}

	logger := logging.NewSlogLogger(logging.SlogOptions{
		Prefix:        "[webhook]",
		ShowCaller:    true,
		ShowDebugLogs: strings.ToLower(args.LogLevel) == "debug",
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)

	httpLogger := logging.NewHttpLogger(logging.HttpLoggerOptions{})
	r.Use(httpLogger.Use)

	dclient, err := dynamic.NewForConfig(args.KubeRestConfig)
	if err != nil {
		return errors.NewEf(err, "creating kubernetes dynamic client")
	}

	r.HandleFunc("/mutate/pod", func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		handleMutate(HandlerContext{
			client:          dclient,
			Context:         r.Context(),
			Resource:        ResourcePod,
			Logger:          logger.With("request-id", requestID),
			CreatedForLabel: args.CreatedForLabel,
		}, w, r)
	})

	server := &http.Server{
		Addr:    args.Addr,
		Handler: r,
	}
	logger.Info("starting http server", "addr", args.Addr)

	common.PrintReadyBanner2(time.Since(start))

	// return server.ListenAndServeTLS("/tmp/tls/tls.crt", "/tmp/tls/tls.key")
	return server.ListenAndServeTLS(args.TLSCertFile, args.TLSKeyFile)
}

func main() {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	var addr string
	flag.StringVar(&addr, "addr", "", "--addr <host:port>")

	var logLevel string
	flag.StringVar(&logLevel, "log-level", "info", "--log-level <debug|warn|info|error>")

	flag.Parse()

	kubeConfig, err := func() (*rest.Config, error) {
		if ev.KubernetesApiProxy == "" {
			return &rest.Config{Host: ev.KubernetesApiProxy}, nil
		}
		return rest.InClusterConfig()
	}()
	if err != nil {
		panic(err)
	}

	if err := Run(RunArgs{
		Addr:           addr,
		LogLevel:       logLevel,
		KubeRestConfig: kubeConfig,
		TLSCertFile:    "/tmp/tls/tls.crt",
		TLSKeyFile:     "/tmp/tls/tls.key",
	}); err != nil {
		panic(err)
	}
}

func handleMutate(ctx HandlerContext, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	review := admissionv1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err = deserializer.Decode(body, nil, &review); err != nil {
		http.Error(w, "could not decode admission review", http.StatusBadRequest)
		return
	}

	response := processPodAdmission(ctx, review)
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func processPodAdmission(ctx HandlerContext, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	ctx.InfoContext(ctx, "pod admission", "ref", review.Request.UID, "op", review.Request.Operation)

	pod := corev1.Pod{}

	switch review.Request.Operation {
	case admissionv1.Create, admissionv1.Delete:
		err := json.Unmarshal(review.Request.Object.Raw, &pod)
		if err != nil {
			return errResponse(ctx, err, review.Request.UID)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}

	if _, ok := pod.Labels[ctx.CreatedForLabel]; ok {
		return mutateAndAllow(review, nil)
	}

	ctx.InfoContext(ctx, "pod-info", "name", pod.Name, "namespace", pod.Namespace)

	gcr := crdsv1.GroupVersion.WithResource("serviceintercepts")
	ul, err := ctx.client.Resource(gcr).Namespace(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errResponse(ctx, err, review.Request.UID)
	}

	var svciList crdsv1.ServiceInterceptList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ul.UnstructuredContent(), &svciList); err != nil {
		return errResponse(ctx, err, review.Request.UID)
	}

	isMatched, err := func() (bool, error) {
		for _, si := range svciList.Items {
			if si.DeletionTimestamp != nil {
				continue
			}

			s, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: pod.Labels})
			if err != nil {
				return false, err
			}

			if !s.Matches(labels.Set(si.Status.Selector)) {
				return true, nil
			}
		}

		return false, nil
	}()
	if err != nil {
		return errResponse(ctx, err, review.Request.UID)
	}

	if !isMatched {
		return mutateAndAllow(review, nil)
	}

	switch review.Request.Operation {
	case admissionv1.Create:
		{
			ctx.Info("[INCOMING] pod", "op", review.Request.Operation, "uid", review.Request.UID, "name", review.Request.Name, "namespace", review.Request.Namespace)

			pod.Spec.NodeSelector = fn.MapMerge(pod.Spec.NodeSelector, map[string]string{"kloudlite.io/no-schedule": "true"})

			lb := pod.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 1)
			}

			lb[ctx.CreatedForLabel] = "intercept"

			patchBytes, err := json.Marshal([]map[string]any{
				{
					"op":    "add",
					"path":  "/metadata/labels",
					"value": lb,
				},
				{
					"op":    "add",
					"path":  "/spec",
					"value": pod.Spec,
				},
			})
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			return mutateAndAllow(review, patchBytes)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}
}

func errResponse(ctx HandlerContext, err error, uid types.UID) admissionv1.AdmissionReview {
	ctx.Error("encountered error", "err", err)
	return admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		},
	}
}

func mutateAndAllow(review admissionv1.AdmissionReview, patch []byte) admissionv1.AdmissionReview {
	patchType := admissionv1.PatchTypeJSONPatch

	resp := admissionv1.AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: true,
	}

	if patch != nil {
		resp.Patch = patch
		resp.PatchType = &patchType
	}

	return admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &resp,
	}
}
