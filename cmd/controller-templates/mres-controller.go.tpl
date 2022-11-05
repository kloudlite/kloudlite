{{- /*variables*/ -}}
{{- $package := get . "package" -}}
{{- $kind := get . "kind" -}}
{{- $kindPkg := get . "kind-pkg" -}}
{{- $kindPlural := get . "kind-plural" -}}
{{- $apiGroup := get . "api-group" -}}

{{- $reconType := printf "%sReconciler" .kind -}}
{{- $kindObjName := printf "%s.%s" $kindPkg $kind -}}

package {{$package}}

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	libMongo "operators.kloudlite.io/lib/mongo"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"operators.kloudlite.io/lib/kubectl"
)

type {{$reconType}} struct {
	client.Client
	Scheme    *runtime.Scheme
	logger    logging.Logger
	Name      string
}

func (r *{{$reconType}}) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds"
	IsOwnedByMsvc    string = "is-owned-by-msvc"
)

// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}}/status,verbs=get;update;patch
// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}}/finalizers,verbs=update

func (r *{{$reconType}}) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &{{$kindPkg}}.{{$kind}}{})

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
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(AccessCredsReady, DBUserReady); !step.ShouldProceed() {
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

	if step := r.reconDBCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDBUser(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *{{$reconType}}) finalize(req *rApi.Request[*{{$kindPkg}}.{{$kind}}]) stepResult.Result {
	return req.Finalize()
}

func (r *{{$reconType}}) reconOwnership(req *rApi.Request[*{{$kindPkg}}.{{$kind}}]) stepResult.Result {
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

func (r *{{$reconType}}) reconDBCreds(req *rApi.Request[*{{$kindPkg}}.{{$kind}}]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	accessSecretName := "mres-" + obj.Name

	accessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, accessSecretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, accessSecretName).String())
	}

	// msvc output ref
	msvcOutput, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output does not exist").Error())
	}

	if accessSecret == nil {
		dbPasswd := fn.CleanerNanoid(40)
		hosts := string(msvcOutput.Data["DB_HOSTS"])

		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       accessSecretName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": Output{},
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

	var output Output
		if err := json.Unmarshal(b, &output); err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, "output", output)

	return req.Next()
}

func (r *{{$reconType}}) reconDBUser(req *rApi.Request[*{{$kindPkg}}.{{$kind}}]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	accessCreds, ok := rApi.GetLocal[Output](req, "output")
	if !ok {
		return req.CheckFailed(DBUserReady, check, "key 'output' does not exist in req-locals")
	}

	return req.Next()
}

func (r *{{$reconType}}) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&{{$kindPkg}}.{{$kind}}{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList {{$kindPkg}}.{{$kind}}List
					if err := r.List(
						context.TODO(), &dbList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{constants.MsvcNameKey: msvcName},
							),
							Namespace: obj.GetNamespace(),
						},
					); err != nil {
						return nil
					}

					reqs := make([]reconcile.Request, 0, len(dbList.Items))
					for j := range dbList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(dbList.Items[j].GetNamespace(), dbList.Items[j].GetName())})
					}

					return reqs
				},
			),
		)
	}

	return builder.Complete(r)
}
