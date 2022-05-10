package crds

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	labels2 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	t "operators.kloudlite.io/lib/types"

	msvcv1 "operators.kloudlite.io/apis/msvc/v1"
	watcherMsvc "operators.kloudlite.io/apis/watchers.msvc/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	lib.MessageSender
	logger *zap.SugaredLogger
	msvc   *crdsv1.ManagedService
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

const msvcFinalizer = "finalizers.kloudlite.io/managed-service"

func (r *ManagedServiceReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.msvc.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrUnknown",
		Message: err.Error(),
	})

	if err := r.notify(); err != nil {
		return reconcileResult.FailedE(err)
	}
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.msvc.LogRef()))
	}
	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) notify() error {
	err := r.SendMessage(r.msvc.LogRef(), lib.MessageReply{
		Key:        r.msvc.LogRef(),
		Conditions: r.msvc.Status.Conditions,
		Status:     meta.IsStatusConditionTrue(r.msvc.Status.Conditions, "Ready"),
	})
	if err != nil {
		return errors.NewEf(err, "could not send message into kafka")
	}
	return nil
}

func (r *ManagedServiceReconciler) notifyAndUpdate(ctx context.Context) (ctrl.Result, error) {
	if err := r.notify(); err != nil {
		return reconcileResult.FailedE(err)
	}
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

var availableMsvc = map[ManagedServiceType]bool{
	MongoDBStandalone: true,
	MongoDBCluster:    true,
	ElasticSearch:     false,
	MySqlStandalone:   false,
	MySqlCluster:      false,
}

type Service struct {
	ApiVersion string `json:"api_version,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Metadata   struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	Spec struct {
		Inputs t.KV `json:"inputs,omitempty"`
	} `json:"spec"`
}

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true)

	msvc := &crdsv1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, msvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !msvc.HasLabels() {
		msvc.EnsureLabels()
		if err := r.Update(ctx, msvc); err != nil {
			return ctrl.Result{}, err
		}
		return reconcileResult.OK()
	}

	r.msvc = msvc
	if msvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, msvc)
	}

	// if isAvail, ok := availableMsvc[ManagedServiceType(msvc.Spec.Kind)]; !ok || !isAvail {
	// 	logger.Info("Invalid ManagedSvc Kind: ", msvc.Spec.Kind)
	// 	return reconcileResult.Failed()
	// }
	//
	svc := Service{
		ApiVersion: msvc.Spec.ApiVersion,
		Kind:       msvc.Spec.Kind,
	}
	svc.Metadata.Name = msvc.Name
	svc.Metadata.Namespace = msvc.Namespace
	svc.Spec.Inputs = msvc.Spec.Inputs

	b, err := json.Marshal(svc)
	if err != nil {
		return r.notifyAndDie(ctx, err)
	}

	if err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply service %s:%s", svc.ApiVersion, svc.Kind))
	}

	// switch ManagedServiceType(msvc.Spec.Kind) {
	// case MongoDBStandalone:
	// 	{
	// 		var helmSecret corev1.Secret
	// 		nn := types.NamespacedName{Namespace: msvc.Namespace, Name: fmt.Sprintf("%s-mongodb", msvc.Name)}
	// 		if err := r.Get(ctx, nn, &helmSecret); err != nil {
	// 			logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", nn.String())
	// 		}
	// 		x, ok := helmSecret.Data["mongodb-root-password"]
	// 		msvc.Spec.Inputs.Set("root_password", fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40)))
	//
	// 		b, err := templates.Parse(templates.MongoDBStandalone, msvc)
	// 		if err != nil {
	// 			r.logger.Info(err)
	// 			return r.notifyAndDie(ctx, err)
	// 		}
	// 		watcher, err := templates.Parse(templates.MongoDBWatcher, msvc)
	// 		if err != nil {
	// 			r.logger.Info(err)
	// 			return r.notifyAndDie(ctx, err)
	// 		}
	// 		if err := fn.KubectlApply(b, watcher); err != nil {
	// 			return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply managed service"))
	// 		}
	// 	}
	//
	// case MongoDBCluster:
	// 	{
	// 		var helmSecret corev1.Secret
	// 		nn := types.NamespacedName{Namespace: msvc.Namespace, Name: fmt.Sprintf("%s-mongodb", msvc.Name)}
	// 		if err := r.Get(ctx, nn, &helmSecret); err != nil {
	// 			logger.Info("helm release %s is not available yet, assuming resource not yet installed, so installing", nn.String())
	// 		}
	//
	// 		x, ok := helmSecret.Data["mongodb-root-password"]
	// 		msvc.Spec.Inputs.Set("root_password", fn.IfThenElse(ok, string(x), fn.CleanerNanoid(40)))
	// 		y, ok2 := helmSecret.Data["mongodb-replica-set-key"]
	// 		msvc.Spec.Inputs.Set("replica_set_key", fn.IfThenElse(ok2, string(y), fn.CleanerNanoid(41)))
	//
	// 		b, err := templates.Parse(templates.MongoDBCluster, msvc)
	// 		if err != nil {
	// 			r.logger.Info(err)
	// 			return r.notifyAndDie(ctx, err)
	// 		}
	// 		watcher, err := templates.Parse(templates.MongoDBWatcher, msvc)
	// 		if err != nil {
	// 			r.logger.Info(err)
	// 			return r.notifyAndDie(ctx, err)
	// 		}
	// 		if err := fn.KubectlApply(b, watcher); err != nil {
	// 			return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply managed service"))
	// 		}
	// 	}
	// }

	mWatcher := &watcherMsvc.MongoDB{}
	if err := r.Get(ctx, req.NamespacedName, mWatcher); err != nil {
	}
	if mWatcher.Name != "" {
		logger.Infof("watcher: %+v", mWatcher.Name)
		msvc.Status.Conditions = mWatcher.Status.Conditions
		return r.notifyAndUpdate(ctx)
	}

	logger.Infof("applied managed service")
	return r.notifyAndUpdate(ctx)
}

func (r *ManagedServiceReconciler) finalize(ctx context.Context, msvc *crdsv1.ManagedService) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(msvc, msvcFinalizer) {
		controllerutil.RemoveFinalizer(msvc, msvcFinalizer)
		if err := r.Update(ctx, msvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if controllerutil.ContainsFinalizer(msvc, finalizers.Foreground.String()) {
		var mdb msvcv1.MongoDB
		if err := r.Get(ctx, types.NamespacedName{Namespace: msvc.Namespace, Name: "MongoDB"}, &mdb); err != nil {
			if apiErrors.IsNotFound(err) {
				controllerutil.RemoveFinalizer(msvc, finalizers.Foreground.String())
				if err := r.Update(ctx, msvc); err != nil {
					return reconcileResult.FailedE(err)
				}
				return reconcileResult.OK()
			}
		}
	}

	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.ManagedService{}).
		Owns(&msvcv1.MongoDB{}).
		Watches(&source.Kind{
			Type: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "msvc.kloudlite.io/v1",
					"kind":       "MongoDB",
				},
			},
		}, handler.EnqueueRequestsFromMapFunc(func(c client.Object) []reconcile.Request {
			var msvcList crdsv1.ManagedServiceList

			ref := fmt.Sprintf("msvc.kloudlite.io/ref: %s-%s-%s", c.GetNamespace(), c.GetObjectKind().GroupVersionKind().Kind, c.GetName())
			fmt.Printf("label is| %s\n", ref)
			if err := r.List(context.TODO(), &msvcList, &client.ListOptions{
				LabelSelector: labels2.SelectorFromValidatedSet(map[string]string{
					"msvc.kloudlite.io/ref": ref,
				}),
			}); err != nil {
				return nil
			}
			var reqs []reconcile.Request
			for _, item := range msvcList.Items {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()},
				})
			}
			fmt.Println("reqs:", reqs)
			return reqs
		})).
		Complete(r)
}
