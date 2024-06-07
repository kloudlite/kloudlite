package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/charmbracelet/log"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/codingconcepts/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/manager"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type Resource string

const (
	ResourcePod     Resource = "pod"
	ResourceService Resource = "service"
)

const (
	podBindingIP        string = "kloudlite.io/podbinding.ip"
	podBindingUID       string = "kloudlite.io/podbinding.uid"
	podReservationToken string = "kloudlite.io/podbinding.reservation-token"

	svcBindingIP  string = "kloudlite.io/servicebinding.ip"
	svcBindingUID string = "kloudlite.io/servicebinding.uid"
)

type Env struct {
	GatewayAdminApiAddr string `env:"GATEWAY_ADMIN_API_ADDR" required:"true"`
}

type HandlerContext struct {
	context.Context
	Env
	Resource
	IsDebug bool
	*slog.Logger
}

func main() {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	var addr string
	flag.StringVar(&addr, "addr", "", "--addr <host:port>")
	flag.Parse()

	log := log.NewWithOptions(os.Stderr, log.Options{ReportCaller: true})
	logger := slog.New(log)

	r := chi.NewRouter()
	// r.Use(httplog.RequestLogger(&httplog.Logger{Logger: logger, Options: httplog.Options{RequestHeaders: false}}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.HandleFunc("/mutate/pod", func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourcePod, IsDebug: debug, Logger: logger.With("request-id", requestID)}, w, r)
	})

	r.HandleFunc("/mutate/service", func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourceService, IsDebug: debug, Logger: logger.With("request-id", requestID)}, w, r)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	logger.Info("starting http server", "addr", addr)
	// err := server.ListenAndServeTLS("/tls/tls.crt", "/tls/tls.key")
	err := server.ListenAndServeTLS("/tmp/tls/tls.crt", "/tmp/tls/tls.key")
	if err != nil {
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

	switch ctx.Resource {
	case ResourcePod:
		{
			response = processPodAdmission(ctx, review)
		}
	case ResourceService:
		{
			response = processServiceAdmission(ctx, review)
		}
	}

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

	switch review.Request.Operation {
	case admissionv1.Create:
		{
			pod := corev1.Pod{}
			err := json.Unmarshal(review.Request.Object.Raw, &pod)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/pod", ctx.Env.GatewayAdminApiAddr), nil)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			var response manager.RegisterPodResult
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			wgContainer := corev1.Container{
				Name:  "wg",
				Image: "linuxserver/wireguard",
				Command: []string{
					"sh",
					"-c",
					fmt.Sprintf(`
mkdir -p /config/wg_confs
curl -X PUT --silent %q > /config/wg_confs/wg0.conf
wg-quick down wg0 || echo "starting wireguard"
wg-quick up wg0
%s
`, fmt.Sprintf("%s/pod/$POD_NAMESPACE/$POD_NAME/$PODBINDING_RESERVATION_TOKEN", ctx.GatewayAdminApiAddr),
						func() string {
							if ctx.IsDebug {
								return "tail -f /dev/null"
							}
							return ""
						}(),
					),
				},
				Env: []corev1.EnvVar{
					{
						Name: "POD_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					},
					{
						Name: "POD_NAMESPACE",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.namespace",
							},
						},
					},
					{
						Name:  "PODBINDING_RESERVATION_TOKEN",
						Value: response.ReservationToken,
					},
				},
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_ADMIN",
						},
					},
				},
			}

			if ctx.IsDebug {
				pod.Spec.Containers = append(pod.Spec.Containers, wgContainer)
			} else {
				pod.Spec.InitContainers = append(pod.Spec.InitContainers, wgContainer)
			}

			lb := pod.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 2)
			}
			lb[podBindingIP] = response.PodBindingIP
			lb[podReservationToken] = response.ReservationToken
			pod.SetLabels(lb)

			// pod.Spec.DNSPolicy = "None"
			// pod.Spec.DNSConfig = &corev1.PodDNSConfig{
			// 	Nameservers: []string{response.DNSNameserver},
			// 	Searches:    []string{},
			// 	Options:     []corev1.PodDNSConfigOption{},
			// }

			patchBytes, err := json.Marshal([]map[string]any{
				{
					"op":    "add",
					"path":  "/metadata/labels",
					"value": pod.GetLabels(),
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
	case admissionv1.Delete:
		{
			pod := corev1.Pod{}
			err := json.Unmarshal(review.Request.OldObject.Raw, &pod)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			if pod.GetDeletionTimestamp() == nil {
				return mutateAndAllow(review, nil)
			}

			pbIP, ok := pod.GetLabels()[podBindingIP]
			if !ok {
				return mutateAndAllow(review, nil)
			}
			pbToken, ok := pod.GetLabels()[podReservationToken]
			if !ok {
				return mutateAndAllow(review, nil)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/pod/%s/%s", ctx.Env.GatewayAdminApiAddr, pbIP, pbToken), nil)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			if resp.StatusCode != 200 {
				return errResponse(ctx, fmt.Errorf("unexpected status code: %d", resp.StatusCode), review.Request.UID)
			}

			return mutateAndAllow(review, nil)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}
}

func processServiceAdmission(ctx HandlerContext, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	switch review.Request.Operation {
	case admissionv1.Create, admissionv1.Update:
		{
			svc := corev1.Service{}
			err := json.Unmarshal(review.Request.Object.Raw, &svc)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/service/%s/%s", ctx.Env.GatewayAdminApiAddr, svc.Namespace, svc.Name), nil)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			var response manager.ReserveServiceResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			lb := svc.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 1)
			}
			lb[svcBindingIP] = response.ServiceBindingIP
			svc.SetLabels(lb)

			patchBytes, err := json.Marshal([]map[string]any{
				{
					"op":    "add",
					"path":  "/metadata/labels",
					"value": svc.GetLabels(),
				},
			})
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			return mutateAndAllow(review, patchBytes)
		}
	case admissionv1.Delete:
		{
			svc := corev1.Service{}
			err := json.Unmarshal(review.Request.OldObject.Raw, &svc)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			svcBindingIP, ok := svc.GetLabels()[svcBindingIP]
			if !ok {
				return mutateAndAllow(review, nil)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/service/%s", ctx.Env.GatewayAdminApiAddr, svcBindingIP), nil)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(ctx, err, review.Request.UID)
			}

			if resp.StatusCode != 200 {
				return errResponse(ctx, fmt.Errorf("unexpected status code: %d", resp.StatusCode), review.Request.UID)
			}
			return mutateAndAllow(review, nil)
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
