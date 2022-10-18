package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goombaio/namegenerator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	elasticsearchMsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/env"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/types"
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
	HelmReady        string = "helm-ready"
	StsReady         string = "sts-ready"
	AccessCredsReady string = "access-creds-ready"
)

const (
	KeyOutput   string = "output"
	KibanaReady string = "kibana-ready"
)

// +kubebuilder:rbac:groups=elasticsearc.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticsearc.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elasticsearc.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &elasticsearchMsvcv1.Service{})
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

	if step := r.reconKibana(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*elasticsearchMsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*elasticsearchMsvcv1.Service]) stepResult.Result {
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
				"string-data": types.MsvcOutput{
					Username: "elastic",
					Password: rootPassword,
					Port:     9200,
					Hosts:    fmt.Sprintf("%s-headless.%s.svc.cluster.local", obj.Name, obj.Namespace),
					Uri:      fmt.Sprintf("http://%s-headless.%s.svc.cluster.local:9200", obj.Name, obj.Namespace),
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

	secret, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyOutput, secret)
	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*elasticsearchMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmElasticType),
	)
	if err != nil {
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	output, ok := rApi.GetLocal[*types.MsvcOutput](req, KeyOutput)
	if !ok {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.ElasticSearch, map[string]any{
				"obj":              obj,
				"storage-class":    fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Ext4),
				"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"elastic-password": output.Password,
				"labels": map[string]string{
					constants.MsvcNameKey: obj.Name,
					constants.AccountRef:  obj.GetAnnotations()[constants.AccountRef],
				},
			},
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

func (r *Reconciler) reconSts(req *rApi.Request[*elasticsearchMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	var stsList appsv1.StatefulSetList

	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.MsvcNameKey: obj.Name}),
			Namespace:     obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(StsReady, check, err.Error())
	}

	if len(stsList.Items) == 0 {
		return req.CheckFailed(StsReady, check, fmt.Sprintf("no statefulset found, yet ...")).Err(nil)
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(
						map[string]string{
							constants.MsvcNameKey: obj.Name,
						},
					),
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
			return req.CheckFailed(StsReady, check, fmt.Sprintf("waiting for pods to start ..."))
		}
	}

	check.Status = true
	if check != checks[StsReady] {
		checks[StsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconKibana(req *rApi.Request[*elasticsearchMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	kibanaRes, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name+"-kibana"), &elasticsearchMsvcv1.Kibana{})
	if err != nil {
		kibanaRes = nil
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(KibanaReady, check, err.Error())
		}
	}

	if obj.Spec.KibanaEnabled {
		if kibanaRes == nil {
			req.Logger.Infof("would be creating kibana resource")
		}

		if kibanaRes == nil {
			msvcOutput, ok := rApi.GetLocal[*types.MsvcOutput](req, KeyOutput)
			if !ok {
				return req.CheckFailed(KibanaReady, check, errors.NotInLocals(KeyOutput).Error()).Err(nil)
			}

			seed := time.Now().UTC().UnixNano()
			nameGenerator := namegenerator.NewNameGenerator(seed)
			if err := r.Create(
				ctx, &elasticsearchMsvcv1.Kibana{
					ObjectMeta: metav1.ObjectMeta{
						Name:            obj.Name + "-kibana",
						Namespace:       obj.Namespace,
						OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
						Labels: map[string]string{
							constants.AccountRef:  obj.GetAnnotations()[constants.AccountRef],
							constants.MsvcNameKey: obj.Name,
						},
					},
					Spec: elasticsearchMsvcv1.KibanaSpec{
						ReplicaCount: 1,
						Region:       obj.Spec.Region,
						Resources: ct.Resources{
							Cpu: ct.CpuT{
								Min: "200m",
								Max: "400m",
							},
							Memory: "800Mi",
						},
						ElasticUrl: msvcOutput.Uri,
						Expose: elasticsearchMsvcv1.Expose{
							Enabled:         true,
							Domain:          nameGenerator.Generate() + "-" + nameGenerator.Generate() + ".crewscale.kl-client.kloudlite.io",
							BasicAuthSecret: "basic-auth-creds",
						},
					},
				},
			); err != nil {
				return req.CheckFailed(KibanaReady, check, err.Error()).Err(nil)
			}
			checks[KibanaReady] = check
			return req.Done()
		}

		if !kibanaRes.Status.IsReady {
			check.Status = false
			if kibanaRes.Status.Message.RawMessage != nil {
				b, err := json.Marshal(kibanaRes.Status.Message)
				if err != nil {
					return req.CheckFailed(KibanaReady, check, errors.NewEf(err, "could not marshal into json").Error()).Err(nil)
				}
				check.Message = string(b)
			}
			checks[KibanaReady] = check
			return req.UpdateStatus()
		}

		check.Status = true
		if check != checks[KibanaReady] {
			checks[KibanaReady] = check
			return req.UpdateStatus()
		}
		return req.Next()
	}

	if kibanaRes != nil {
		req.Logger.Infof("kibana is not enabled, but kibana exists, so would be deleting it")
		if err := r.Delete(
			ctx, &elasticsearchMsvcv1.Kibana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      obj.Name,
					Namespace: obj.Namespace,
				},
			},
		); err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.CheckFailed(KibanaReady, check, err.Error()).Err(nil)
			}
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&elasticsearchMsvcv1.Service{})
	builder.Owns(fn.NewUnstructured(constants.HelmElasticType))
	builder.Owns(&elasticsearchMsvcv1.Kibana{})
	builder.Owns(&corev1.Secret{})

	builder.Watches(
		&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				value, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{
					{NamespacedName: fn.NN(obj.GetNamespace(), value)},
				}
			},
		),
	)

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
