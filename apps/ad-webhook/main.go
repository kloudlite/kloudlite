package main

import (
	"encoding/json"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = admissionv1.AddToScheme(scheme)
}

func main() {
	http.HandleFunc("/add-node-selector", func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			klog.Errorf("unexpected content type: %s", contentType)
			http.Error(w, "unexpected content type", http.StatusBadRequest)
			return
		}

		req := admissionv1.AdmissionReview{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			klog.Errorf("failed to decode request body: %v", err)
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		resp := admit(&req)
		if resp != nil {
			klog.Infof("sending response: %s", resp)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				klog.Errorf("failed to encode response: %v", err)
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
			}
			return
		}

		http.Error(w, "no response", http.StatusNotFound)
	})

	server := &http.Server{
		Addr: ":8443",
	}

	klog.Infof("listening on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}
}

func admit(req *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Request.Object.Raw, pod); err != nil {
		klog.Errorf("failed to unmarshal Pod: %v", err)
		return toAdmissionResponse(err)
	}

	pvcFound := false
	for _, volume := range pod.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			pvcFound = true
			break
		}
	}

	if !pvcFound {
		return nil
	}

	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = map[string]string{}
	}

	pod.Spec.NodeSelector["kloudlite.io/stateful"] = "true"
	pod.Spec.PriorityClassName = "stateful"

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		klog.Errorf("failed to marshal Pod: %v", err)
		return toAdmissionResponse(err)
	}

	return &admissionv1.AdmissionResponse{
		UID:     req.Request.UID,
		Allowed: true,
		Patch:   marshaledPod,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
