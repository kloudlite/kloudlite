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
  "encoding/json"

  "k8s.io/apimachinery/pkg/runtime"
  ""
  "operators.kloudlite.io/lib/harbor"
  "operators.kloudlite.io/lib/logging"
  rApi "operators.kloudlite.io/lib/operator"
  stepResult "operators.kloudlite.io/lib/operator/step-result"
  ctrl "sigs.k8s.io/controller-runtime"
  "sigs.k8s.io/controller-runtime/pkg/client"
  "operators.kloudlite.io/lib/kubectl"
)

type {{$reconType}} struct {
  client.Client
  Scheme    *runtime.Scheme
  env       *env.Env
  harborCli *harbor.Client
  logger    logging.Logger
  Name      string
}

func (r *{{$reconType}}) GetName() string {
  return r.Name
}

const (
  HelmReady        string = "helm-ready"
  StsReady         string = "sts-ready"
  AccessCredsReady string = "access-creds-ready"
)

const (
  KeyRootPassword string = "root-password"
)


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
  if step := req.EnsureChecks(HelmReady, StsReady, AccessCredsReady); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := r.reconAccessCreds(req); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := r.reconHelm(req); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  if step := r.reconSts(req); !step.ShouldProceed() {
    return step.ReconcilerResponse()
  }

  req.Object.Status.IsReady = true
  return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *{{$reconType}}) finalize(req *rApi.Request[*{{$kindObjName}}]) stepResult.Result {
  return req.Finalize()
}

func (r *{{$reconType}}) reconAccessCreds(req *rApi.Request[*{{$kindObjName}}]) stepResult.Result {
  ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

  check := rApi.Check{Generation: obj.Generation}
  secretName := "msvc-" + obj.Name
  scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
  if err != nil {
    req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
  }

  if scrt == nil {
    rootPassword := fn.CleanerNanoid(40)
    b, err := templates.Parse(
      templates.Secret, map[string]any{
        "name":       secretName,
        "namespace":  obj.Namespace,
        "labels":     obj.GetLabels(),
        "owner-refs": []metav1.OwnerReference{fn.AsOwner(obj)},
        "string-data": map[string]string{
          "ROOT_PASSWORD": rootPassword,
          // TODO: user
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

  if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
    obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
    if err := r.Update(ctx, obj); err != nil {
      return req.FailWithOpError(err)
    }
    return req.Done().RequeueAfter(2 * time.Second)
  }

  check.Status = true
  if check != checks[AccessCredsReady] {
    checks[AccessCredsReady] = check
    return req.UpdateStatus()
  }

  rApi.SetLocal(req, KeyRootPassword, string(scrt.Data["ROOT_PASSWORD"]))
  return req.Next()
}

func (r *{{$reconType}}) reconHelm(req *rApi.Request[*{{$kindObjName}}]) stepResult.Result {
  ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
  check := rApi.Check{Generation: obj.Generation}

  helmRes, err := rApi.Get(
    ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(/* TODO: (user) */),
  )
  if err != nil {
    req.Logger.Infof("helm reosurce (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
  }

  rootPassword, ok := rApi.GetLocal[string](req, KeyRootPassword)
  if !ok {
    return req.CheckFailed(HelmReady, check, err.Error())
  }


  if helmRes == nil || check.Generation > checks[HelmReady].Generation {
    storageClass, err := obj.Spec.CloudProvider.GetStorageClass(ct.Xfs)
    if err != nil {
      return req.CheckFailed(HelmReady, check, err.Error())
    }

    b, err := templates.Parse(
      // TODO: (user)
    )

    if err != nil {
      return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
    }

    if err := fn.KubectlApplyExec(ctx, b); err != nil {
      return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
    }

    checks[HelmReady] = check
    return req.UpdateStatus()
  }

  cds, err := conditions.FromObject(helmRes)
  if err != nil {
    return req.CheckFailed(HelmReady, check, err.Error())
  }

  deployedC := meta.FindStatusCondition(cds, "Deployed")
  if deployedC == nil {
    return req.Done()
  }

  if deployedC.Status == metav1.ConditionFalse {
    return req.CheckFailed(HelmReady, check, deployedC.Message)
  }

  if deployedC.Status == metav1.ConditionTrue {
    check.Status = true
  }

  if check != checks[HelmReady] {
    checks[HelmReady] = check
    return req.UpdateStatus()
  }

  return req.Next()
}

func (r *{{$reconType}}) reconSts(req *rApi.Request[*{{$kindObjName}}]) stepResult.Result {
  ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
  check := rApi.Check{Generation: obj.Generation}
  var stsList appsv1.StatefulSetList

  if err := r.List(
    ctx, &stsList, &client.ListOptions{
      LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.MsvcNameKey: obj.Name}),
      Namespace: obj.Namespace,
    },
  ); err != nil {
    return req.CheckFailed(StsReady, check, err.Error())
  }

  for i := range stsList.Items {
    item := stsList.Items[i]
    if item.Status.AvailableReplicas != item.Status.Replicas {
      check.Status = false

      var podsList corev1.PodList
      if err := r.List(
        ctx, &podsList, &client.ListOptions{
          LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
            constants.MsvcNameKey: obj.Name,
          }),
        },
      ); err != nil {
        return req.FailWithOpError(err)
      }

      messages := rApi.GetMessagesFromPods(podsList.Items...)
      if len(messages) > 0 {
        b, err := json.Marshal(messages)
        if err != nil {
          return req.CheckFailed(StsReady, check, err.Error())
        }
        return req.CheckFailed(StsReady, check, string(b))
      }
    }
  }

  check.Status = true
  if check != checks[StsReady] {
    checks[StsReady] = check
    return req.UpdateStatus()
  }

  return req.Next()
}

func (r *{{$reconType}}) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
  r.Client = mgr.GetClient()
  r.Scheme = mgr.GetScheme()
  r.logger = logger.WithName(r.Name)
  r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

  builder := ctrl.NewControllerManagedBy(mgr).For(&{{$kindObjName}}{})
  builder.Owns(&corev1.Secret{})
  builder.Owns(fn.NewUnstructured(...))

  builder.Watches(
    &source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(
      func(obj client.Object) []reconcile.Request {
        v, ok := obj.GetLabels()[constants.MsvcNameKey]
        if !ok {
          return nil
        }
        return []reconcile.Request{fn.NN(obj.GetNamespace(), v)}
      },
    ),
  )


  return builder.Complete(r)
}
