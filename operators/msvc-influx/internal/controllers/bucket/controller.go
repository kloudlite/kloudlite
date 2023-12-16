package bucket

import (
	"context"
	"encoding/json"
	"time"

	influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/env"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds"
	IsOwnedByMsvc    string = "is-owned-by-msvc"
	BucketReady      string = "bucket-ready"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyOutput     string = "output"
)

// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &influxdbMsvcv1.Bucket{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(AccessCredsReady, BucketReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconOwnership(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconAccessCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconInfluxBucket(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*influxdbMsvcv1.Bucket]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconOwnership(req *rApi.Request[*influxdbMsvcv1.Bucket]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), fn.NewUnstructured(
			metav1.TypeMeta{
				Kind:       obj.Spec.MsvcRef.Kind,
				APIVersion: obj.Spec.MsvcRef.APIVersion,
			},
		),
	)
	if err != nil {
		return req.CheckFailed(IsOwnedByMsvc, check, err.Error())
	}

	if !fn.IsOwner(obj, fn.AsOwner(msvc)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(msvc)))
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[IsOwnedByMsvc] {
		checks[IsOwnedByMsvc] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func getMsvcOutput(secret *corev1.Secret) (*types.MsvcOutput, error) {
	b, err := json.Marshal(secret)
	if err != nil {
		return nil, err
	}
	var m types.MsvcOutput
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*influxdbMsvcv1.Bucket]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	msvcOutput, err := getMsvcOutput(msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	accessSecretName := "mres-" + obj.Name

	accessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, accessSecretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, accessSecretName).String())
	}

	if accessSecret == nil {
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       accessSecretName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.MresOutput{
					BucketName: obj.Name,
					BucketId:   "",
					OrgId:      "",
					OrgName:    msvcOutput.Org,
					Token:      msvcOutput.Token,
					Uri:        msvcOutput.Uri,
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	b, err := json.Marshal(accessSecret.Data)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	var output types.MresOutput
	if err := json.Unmarshal(b, &output); err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyMsvcOutput, msvcOutput)
	rApi.SetLocal(req, KeyOutput, output)

	return req.Next()
}

func (r *Reconciler) reconInfluxBucket(req *rApi.Request[*influxdbMsvcv1.Bucket]) stepResult.Result {
	// ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	// check := rApi.Check{Generation: obj.Generation}
	//
	// msvcOutput, ok := rApi.GetLocal[*types.MsvcOutput](req, KeyMsvcOutput)
	// if !ok {
	// 	return req.CheckFailed(BucketReady, check, fmt.Sprintf("key %s not found in req.locals", KeyMsvcOutput)).Err(nil)
	// }
	//
	// influxClient := libInflux.NewClient(msvcOutput.Uri, msvcOutput.Token)
	// defer influxClient.Close()
	//
	// exists, err := influxClient.BucketExists(ctx, obj.Spec.Bucket)
	// if err != nil {
	// 	return req.CheckFailed(BucketReady, check, err.Error())
	// }
	//
	// if !exists {
	// 	bucket, err := influxClient.UpsertBucket(ctx, msvcOutput.Org, obj.Name)
	// 	if err != nil {
	// 		return req.FailWithOpError(err)
	// 	}
	//
	// }
	//
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&influxdbMsvcv1.Bucket{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&influxdbMsvcv1.Service{},
	}

	for i := range watchList {
		builder.Watches(
			watchList[i],
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var buckets influxdbMsvcv1.BucketList
					if err := r.List(ctx, &buckets, &client.ListOptions{
						LabelSelector: labels.SelectorFromValidatedSet(
							map[string]string{constants.MsvcNameKey: msvcName},
						),
						Namespace: obj.GetNamespace(),
					},
					); err != nil {
						return nil
					}

					reqs := make([]reconcile.Request, 0, len(buckets.Items))
					for j := range buckets.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(buckets.Items[j].GetNamespace(), buckets.Items[j].GetName())})
					}

					return reqs
				},
			),
		)
	}

	return builder.Complete(r)
}
