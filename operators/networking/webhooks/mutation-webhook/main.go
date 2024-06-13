package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/codingconcepts/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	podIPLabel string = "kloudlite.io/pod.ip"
)

type Env struct {
	GatewayAdminApiAddr string `env:"GATEWAY_ADMIN_API_ADDR" required:"true"`
}

type HandlerContext struct {
	context.Context
	Env
	Resource
}

func main() {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.HandleFunc("/mutate/pod", func(w http.ResponseWriter, r *http.Request) {
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourcePod}, w, r)
	})

	r.HandleFunc("/mutate/service", func(w http.ResponseWriter, r *http.Request) {
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourceService}, w, r)
	})
	server := &http.Server{
		Addr:    ":8443",
		Handler: r,
	}
	fmt.Println("Starting server on port 8443")
	// err := server.ListenAndServeTLS("/tls/tls.crt", "/tls/tls.key")
	err := server.ListenAndServeTLS("/tmp/tls/tls.crt", "/tmp/tls/tls.key")
	if err != nil {
		panic(err)
	}
}

func handleMutate(ctx HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to mutate: %s", ctx.Resource)
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
	log.Printf("admission request: %s", review.Request.Operation)

	switch review.Request.Operation {
	case admissionv1.Create:
		{
			pod := corev1.Pod{}
			err := json.Unmarshal(review.Request.Object.Raw, &pod)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/peer/pod/%s/%s", ctx.Env.GatewayAdminApiAddr, pod.Namespace, pod.GenerateName), nil)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			// b is pod IP
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			initContainer := corev1.Container{
				Name:  "wg",
				Image: "linuxserver/wireguard",
				Command: []string{
					"sh",
					"-c",
					fmt.Sprintf(`
mkdir -p /config/wg_confs
curl --silent '%s/pod/wg-config/%s' > /config/wg_confs/wg0.conf
wg-quick down wg0 || echo "starting wireguard"
wg-quick up wg0`, ctx.GatewayAdminApiAddr, b),
				},
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_ADMIN",
						},
					},
				},
			}

			pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)

			lb := pod.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 1)
			}
			lb[podIPLabel] = string(b)
			pod.SetLabels(lb)

			patchBytes, err := json.Marshal([]map[string]any{
				{
					"op":    "add",
					"path":  "/spec/initContainers",
					"value": pod.Spec.InitContainers,
				},
				{
					"op":    "add",
					"path":  "/metadata/labels",
					"value": pod.GetLabels(),
				},
			})
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			return mutateAndAllow(review, patchBytes)
		}
	case admissionv1.Delete:
		{
			pod := corev1.Pod{}
			err := json.Unmarshal(review.Request.OldObject.Raw, &pod)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			podIP, ok := pod.GetLabels()[podIPLabel]
			if !ok {
				return mutateAndAllow(review, nil)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/peer/pod/%s", ctx.Env.GatewayAdminApiAddr, podIP), nil)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			if resp.StatusCode != 200 {
				return mutateAndAllow(review, nil)
				// return errResponse(fmt.Errorf("unexpected status code: %d", resp.StatusCode), review.Request.UID)
			}

			return mutateAndAllow(review, nil)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}
}

func processServiceAdmission(_ context.Context, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	pod := corev1.Service{}
	err := json.Unmarshal(review.Request.Object.Raw, &pod)
	if err != nil {
		return errResponse(err, review.Request.UID)
	}

	patchType := admissionv1.PatchTypeJSONPatch

	return admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		// Request:  review.Request,
		Response: &admissionv1.AdmissionResponse{
			UID:       review.Request.UID,
			Allowed:   true,
			Patch:     nil,
			PatchType: &patchType,
		},
	}
}

func errResponse(err error, uid types.UID) admissionv1.AdmissionReview {
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
