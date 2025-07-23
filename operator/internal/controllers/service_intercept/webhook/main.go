package webhook

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/common"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	"github.com/kloudlite/operator/pkg/errors"
	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"

	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/nxtcoder17/fastlog"
	"k8s.io/client-go/dynamic"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

var logger *fastlog.Logger

type Resource string

const (
	ResourcePod Resource = "pod"
)

const (
	debugWebhookAnnotation string = "kloudlite.io/networking.webhook.debug"
)

type HandlerContext struct {
	context.Context
	*slog.Logger

	client             *dynamic.DynamicClient
	CreatedForLabelKey string
}

type RunArgs struct {
	Addr               string
	Debug              bool
	KubeRestConfig     *rest.Config
	TLSCertFilePath    string
	TLSKeyFilePath     string
	CreatedForLabelKey string
}

func Run(args RunArgs) error {
	start := time.Now()

	logger = fastlog.New(fastlog.Options{
		Writer:        os.Stderr,
		Format:        fastlog.ConsoleFormat,
		ShowDebugLogs: args.Debug,
		ShowCaller:    true,
		EnableColors:  true,
	})

	dclient, err := dynamic.NewForConfig(args.KubeRestConfig)
	if err != nil {
		return errors.NewEf(err, "creating kubernetes dynamic client")
	}

	router := ivy.NewRouter()
	ivy.Logger = logger.Slog()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())

	router.Handle("/mutate/pod", ivy.ToHTTPHandler(func(c *ivy.Context) error {
		body, err := io.ReadAll(c.Body())
		if err != nil {
			return ivy.NewHTTPError(http.StatusBadRequest, "failed to read request body")
		}

		review := admissionv1.AdmissionReview{}
		if _, _, err := codecs.UniversalDeserializer().Decode(body, nil, &review); err != nil {
			return ivy.NewHTTPError(http.StatusBadRequest, "could not decode body into admission review")
		}

		return c.JSON(processPodAdmission(HandlerContext{
			Context:            c,
			Logger:             c.Logger,
			client:             dclient,
			CreatedForLabelKey: args.CreatedForLabelKey,
		}, review))
	}))

	common.PrintReadyBanner2(time.Since(start))

	return http.ListenAndServeTLS(args.Addr, args.TLSCertFilePath, args.TLSKeyFilePath, router)
}

func processPodAdmission(ctx HandlerContext, review admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	ctx.InfoContext(ctx, "pod admission", "ref", review.Request.UID, "op", review.Request.Operation)

	pod := corev1.Pod{}

	switch review.Request.Operation {
	case admissionv1.Create, admissionv1.Delete:
		err := json.Unmarshal(review.Request.Object.Raw, &pod)
		if err != nil {
			return errResponse(err, review.Request.UID)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}

	if _, ok := pod.Labels[ctx.CreatedForLabelKey]; ok {
		return mutateAndAllow(review, nil)
	}

	ctx.InfoContext(ctx, "pod-info", "name", pod.Name, "namespace", pod.Namespace)

	gcr := v1.GroupVersion.WithResource("serviceintercepts")
	ul, err := ctx.client.Resource(gcr).Namespace(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errResponse(err, review.Request.UID)
	}

	var svciList v1.ServiceInterceptList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ul.UnstructuredContent(), &svciList); err != nil {
		return errResponse(err, review.Request.UID)
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
		return errResponse(err, review.Request.UID)
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

			lb[ctx.CreatedForLabelKey] = "intercept"

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
				return errResponse(err, review.Request.UID)
			}

			return mutateAndAllow(review, patchBytes)
		}
	default:
		{
			return mutateAndAllow(review, nil)
		}
	}
}

func errResponse(err error, uid types.UID) admissionv1.AdmissionReview {
	logger.Error("encountered error", "err", err)
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
