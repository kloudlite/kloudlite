package main

import (
	"context"
	"encoding/json"
	"flag"
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
	podBindingIP  string = "kloudlite.io/podbinding.ip"
	podBindingUID string = "kloudlite.io/podbinding.uid"

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
}

func main() {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.HandleFunc("/mutate/pod", func(w http.ResponseWriter, r *http.Request) {
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourcePod, IsDebug: debug}, w, r)
	})

	r.HandleFunc("/mutate/service", func(w http.ResponseWriter, r *http.Request) {
		handleMutate(HandlerContext{Context: r.Context(), Env: ev, Resource: ResourceService, IsDebug: debug}, w, r)
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

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/pod", ctx.Env.GatewayAdminApiAddr), nil)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			var response manager.RegisterPodResult
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return errResponse(err, review.Request.UID)
			}

			wgContainer := corev1.Container{
				Name:  "wg",
				Image: "linuxserver/wireguard",
				Command: []string{
					"sh",
					"-c",
					fmt.Sprintf(`
mkdir -p /config/wg_confs
curl --silent '%s/pod/%s' > /config/wg_confs/wg0.conf
wg-quick down wg0 || echo "starting wireguard"
wg-quick up wg0
%s
`, ctx.GatewayAdminApiAddr, response.PodIP,
						func() string {
							if ctx.IsDebug {
								return "tail -f /dev/null"
							}
							return ""
						}(),
					),
				},
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_ADMIN",
						},
					},
				},
			}

			containersPatch := func() map[string]any {
				if ctx.IsDebug {
					return map[string]any{
						"op":    "add",
						"path":  "/spec/containers",
						"value": append(pod.Spec.Containers, wgContainer),
					}
				}

				return map[string]any{
					"op":    "add",
					"path":  "/spec/initContainers",
					"value": append(pod.Spec.InitContainers, wgContainer),
				}
			}()

			lb := pod.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 2)
			}
			lb[podBindingIP] = response.PodIP
			lb[podBindingUID] = response.PodUID
			pod.SetLabels(lb)

			patchBytes, err := json.Marshal([]map[string]any{
				containersPatch,
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

			if pod.GetDeletionTimestamp() == nil {
				return mutateAndAllow(review, nil)
			}

			pbIP, ok := pod.GetLabels()[podBindingIP]
			if !ok {
				return mutateAndAllow(review, nil)
			}
			pbUID, ok := pod.GetLabels()[podBindingUID]
			if !ok {
				return mutateAndAllow(review, nil)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/pod/%s/%s", ctx.Env.GatewayAdminApiAddr, pbIP, pbUID), nil)
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

func processServiceAdmission(ctx HandlerContext, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	switch review.Request.Operation {
	case admissionv1.Create, admissionv1.Update:
		{
			svc := corev1.Service{}
			err := json.Unmarshal(review.Request.Object.Raw, &svc)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/service/%s/%s", ctx.Env.GatewayAdminApiAddr, svc.Namespace, svc.Name), nil)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			var response manager.RegisterServiceResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return errResponse(err, review.Request.UID)
			}

			lb := svc.GetLabels()
			if lb == nil {
				lb = make(map[string]string, 2)
			}
			lb[svcBindingIP] = response.ServiceBindingIP
			lb[svcBindingUID] = response.ServiceBindingUID
			svc.SetLabels(lb)

			patchBytes, err := json.Marshal([]map[string]any{
				{
					"op":    "add",
					"path":  "/metadata/labels",
					"value": svc.GetLabels(),
				},
			})
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			return mutateAndAllow(review, patchBytes)
		}
	case admissionv1.Delete:
		{
			svc := corev1.Service{}
			err := json.Unmarshal(review.Request.OldObject.Raw, &svc)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			// if svc.GetDeletionTimestamp() == nil {
			// 	return mutateAndAllow(review, nil)
			// }

			svcBindingIP, ok := svc.GetLabels()[svcBindingIP]
			if !ok {
				return mutateAndAllow(review, nil)
			}
			svcBindingUID, ok := svc.GetLabels()[svcBindingUID]
			if !ok {
				return mutateAndAllow(review, nil)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/service/%s/%s", ctx.Env.GatewayAdminApiAddr, svcBindingIP, svcBindingUID), nil)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errResponse(err, review.Request.UID)
			}

			if resp.StatusCode != 200 {
				return errResponse(fmt.Errorf("unexpected status code: %d", resp.StatusCode), review.Request.UID)
			}
			return mutateAndAllow(review, nil)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
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
