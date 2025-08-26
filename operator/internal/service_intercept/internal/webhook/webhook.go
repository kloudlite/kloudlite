package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	v1 "github.com/kloudlite/operator/api/v1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"

	fn "github.com/kloudlite/operator/pkg/functions"
)

type Resource string

const (
	ResourcePod Resource = "pod"
)

type MutationWebhook struct {
	Debug bool

	Scheme     *runtime.Scheme
	KubeClient client.Client

	ShouldIgnorePod func(pod *corev1.Pod) bool
}

func (mw *MutationWebhook) Handler() (http.Handler, error) {
	router := ivy.NewRouter()
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())

	decoder := serializer.NewCodecFactory(mw.Scheme).UniversalDeserializer()

	router.Handle("/mutate/pod", ivy.ToHTTPHandler(func(c *ivy.Context) error {
		body, err := io.ReadAll(c.Body())
		if err != nil {
			return ivy.NewHTTPError(http.StatusBadRequest, "failed to read request body")
		}

		review := admissionv1.AdmissionReview{}
		if _, _, err := decoder.Decode(body, nil, &review); err != nil {
			return ivy.NewHTTPError(http.StatusBadRequest, "could not decode body into admission review")
		}

		return c.JSON(mw.ProcessPodAdmission(c, review))
	}))

	return router, nil
}

func (mw *MutationWebhook) isServiceInterceptOverPod(ctx context.Context, pod *corev1.Pod) (bool, error) {
	var svciList v1.ServiceInterceptList
	if err := mw.KubeClient.List(ctx, &svciList, client.InNamespace(pod.Namespace)); err != nil {
		return false, err
	}

	for _, si := range svciList.Items {
		if si.DeletionTimestamp != nil {
			continue
		}

		s, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: pod.Labels})
		if err != nil {
			return false, err
		}

		if !s.Matches(apiLabels.Set(si.Status.Selector)) {
			return true, nil
		}
	}

	return false, nil
}

func (mw *MutationWebhook) ProcessPodAdmission(ctx *ivy.Context, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	pod := &corev1.Pod{}

	switch review.Request.Operation {
	case admissionv1.Create, admissionv1.Delete:
		err := json.Unmarshal(review.Request.Object.Raw, pod)
		if err != nil {
			return mw.Failed(review, fmt.Errorf("failed to unmarshal admission review into a corev1.Pod: %w", err))
		}
	default:
		{
			return mw.Success(review, nil)
		}
	}

	logger := ctx.Logger.With(
		"review.uid", review.Request.UID,
		"review.operation", review.Request.Operation,
		"name/generate-name", func() string {
			if pod.Name == "" {
				return pod.GenerateName
			}
			return pod.Name
		}(),
		"namespace", pod.Namespace,
	)

	if mw.ShouldIgnorePod(pod) {
		logger.Debug("ignoring review request, as pod statisfies ignoring constraints")
		return mw.Success(review, nil)
	}

	isMatched, err := mw.isServiceInterceptOverPod(ctx, pod)
	if err != nil {
		return mw.Failed(review, err)
	}

	if !isMatched {
		return mw.Success(review, nil)
	}

	switch review.Request.Operation {
	case admissionv1.Create:
		{
			logger.Info("[INCOMING] pod")

			pod.Spec.NodeSelector = fn.MapMerge(
				pod.Spec.NodeSelector,
				map[string]string{v1.ProjectDomain + "/service-intercept.no-schedule": "true"})

			// lb := pod.GetLabels()
			// if lb == nil {
			// 	lb = make(map[string]string, 1)
			// }
			//
			// lb[ctx.CreatedForLabelKey] = "intercept"

			patchBytes, err := json.Marshal([]map[string]any{
				// {
				// 	"op":    "add",
				// 	"path":  "/metadata/labels",
				// 	"value": lb,
				// },
				{
					"op":    "add",
					"path":  "/spec",
					"value": pod.Spec,
				},
			})
			if err != nil {
				return mw.Failed(review, err)
			}

			return mw.Success(review, patchBytes)
		}
	default:
		return mw.Success(review, nil)
	}
}

func (mw *MutationWebhook) Failed(review admissionv1.AdmissionReview, err error) admissionv1.AdmissionReview {
	return admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		},
	}
}

func (mw *MutationWebhook) Success(review admissionv1.AdmissionReview, patch []byte) admissionv1.AdmissionReview {
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
