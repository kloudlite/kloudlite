package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"net/http"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/service-intercept/internal/controllers/svci"
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
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/logging"

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
	podBindingIP        string = "kloudlite.io/podbinding.ip"
	podReservationToken string = "kloudlite.io/podbinding.reservation-token"

	svcBindingIPLabel            string = "kloudlite.io/servicebinding.ip"
	svcReservationTokenLabel     string = "kloudlite.io/servicebinding.reservation-token"
	kloudliteWebhookTriggerLabel string = "kloudlite.io/webhook.trigger"
)

const (
	debugWebhookAnnotation string = "kloudlite.io/networking.webhook.debug"
)

type Env struct {
	GatewayAdminApiAddr string `env:"GATEWAY_ADMIN_API_ADDR" required:"true"`

	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY" required:"true"`
}

type Flags struct {
	WgImage           string
	WgImagePullPolicy string
}

type HandlerContext struct {
	context.Context
	Env
	Flags
	Resource
	*slog.Logger

	client *dynamic.DynamicClient
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

	var flags Flags

	flag.StringVar(&flags.WgImage, "wg-image", "ghcr.io/kloudlite/hub/wireguard:latest", "--wg-image <image>")

	flag.StringVar(&flags.WgImagePullPolicy, "wg-image-pull-policy", "IfNotPresent", "--wg-image-pull-policy <image-pull-policy>")

	flag.Parse()

	logger := logging.NewSlogLogger(logging.SlogOptions{
		Prefix:        "[webhook]",
		ShowCaller:    true,
		ShowDebugLogs: logLevel == "debug",
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)

	httpLogger := logging.NewHttpLogger(logging.HttpLoggerOptions{})
	r.Use(httpLogger.Use)

	dclient, err := dynamic.NewForConfig(&rest.Config{
		Host: ev.KubernetesApiProxy,
	})
	if err != nil {
		panic(err)
	}

	r.HandleFunc("/mutate/pod", func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		handleMutate(HandlerContext{
			client:  dclient,
			Context: r.Context(), Env: ev, Flags: flags, Resource: ResourcePod, Logger: logger.With("request-id", requestID)}, w, r)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	logger.Info("starting http server", "addr", addr)

	common.PrintReadyBanner()

	if err := server.ListenAndServeTLS("/tmp/tls/tls.crt", "/tmp/tls/tls.key"); err != nil {
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

	var response admissionv1.AdmissionReview

	response = processPodAdmission(ctx, review)
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

	gcr := crdsv1.GroupVersion.WithResource("serviceintercepts")
	ul, err := ctx.client.Resource(gcr).Namespace(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errResponse(ctx, err, review.Request.UID)
	}

	var svciList crdsv1.ServiceInterceptList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ul.UnstructuredContent(), svciList); err != nil {
		return errResponse(ctx, err, review.Request.UID)
	}

	isMatched, err := func() (bool, error) {
		for _, si := range svciList.Items {
			s, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: si.Status.Selector})
			if err != nil {
				return false, err
			}

			if !s.Matches(labels.Set(pod.Labels)) {
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

			if pod.GetLabels()[constants.KloudliteGatewayEnabledLabel] == "false" {
				return mutateAndAllow(review, nil)
			}

			pod.Spec.NodeName = "non-existent"

			lb := pod.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 1)
			}

			lb[svci.CreatedForLabel] = "intercept"

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
		// Request:  review.Request,
		Response: &resp,
	}
}
