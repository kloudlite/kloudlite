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
  "time"

  "k8s.io/apimachinery/pkg/runtime"
  "github.com/kloudlite/operator/pkg/harbor"
  "github.com/kloudlite/operator/pkg/logging"
  rApi "github.com/kloudlite/operator/pkg/operator"
  stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
  ctrl "sigs.k8s.io/controller-runtime"
  "sigs.k8s.io/controller-runtime/pkg/client"
  "github.com/kloudlite/operator/pkg/kubectl"
)

type {{$reconType}} struct {
  client.Client
  Scheme    *runtime.Scheme
  Env       *env.Env
  logger    logging.Logger
  Name      string
  yamlClient *kubectl.YAMLClient
}

func (r *{{$reconType}}) GetName() string {
  return r.Name
}

// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}}/status,verbs=get;update;patch
// +kubebuilder:rbac:groups={{$apiGroup}},resources={{$kindPlural}}/finalizers,verbs=update

func (r *{{$reconType}}) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
  req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &{{$kindObjName}}{})
  if err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
  }

  if req.Object.GetDeletionTimestamp() != nil {
    if x := r.finalize(req); !x.ShouldProceed() {
      return x.ReconcilerResponse()
    }
    return ctrl.Result{}, nil
  }

	req.LogPreReconcile()
	defer req.LogPostReconcile()

  if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  req.Object.Status.IsReady = true
  return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *{{$reconType}}) finalize(req *rApi.Request[*{{$kindObjName}}]) stepResult.Result {
  return req.Finalize()
}

func (r *{{$reconType}}) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
  r.Client = mgr.GetClient()
  r.Scheme = mgr.GetScheme()
  r.logger = logger.WithName(r.Name)
  r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

  builder := ctrl.NewControllerManagedBy(mgr).For(&{{$kindObjName}}{})
  return builder.Complete(r)
}
