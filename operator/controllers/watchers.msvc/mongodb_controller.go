package watchersmsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	labels2 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	watcherMsvc "operators.kloudlite.io/apis/watchers.msvc/v1"
	"operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"os"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

type ConditionType string

const (
	RootPassword string = "ROOT_PASSWORD"
	DbHosts      string = "HOSTS"
	DbUrl        string = "DB_URL"
)

func (c ConditionType) String() string {
	return string(c)
}

// MongoDBReconciler reconciles a HelmMongoDB object
type MongoDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
	mdb    *watcherMsvc.MongoDB
	msvc   *watcherMsvc.MongoDB
	// msvc   *crdsv1.ManagedService
	watchList []types.NamespacedName
	lt        metav1.Time
}

func (r *MongoDBReconciler) IsInQueue(nn types.NamespacedName) bool {
	for _, nName := range r.watchList {
		if nName.String() == nn.String() {
			return true
		}
	}
	return false
}

type HelmMongoDB struct {
	Spec   map[string]interface{} `json:"spec,omitempty"`
	Status struct {
		Conditions []metav1.Condition `json:"conditions,omitempty"`
	} `json:"status"`
}

func (r *MongoDBReconciler) notifyAndDie(ctx context.Context, cond *metav1.Condition) (reconcile.Result, error) {
	meta.SetStatusCondition(&r.msvc.Status.Conditions, *cond)
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *MongoDBReconciler) buildConditions(source string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&r.msvc.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: r.lt,
		Message:            "Not All Checks completed",
	})
	for _, c := range conditions {
		if c.Reason == "" {
			c.Reason = "Unknown"
		}
		if !c.LastTransitionTime.IsZero() {
			if c.LastTransitionTime.Time.Sub(r.lt.Time).Seconds() > 0 {
				r.lt = c.LastTransitionTime
			}
		}
		if c.LastTransitionTime.IsZero() {
			c.LastTransitionTime = r.lt
		}
		if source != "" {
			c.Reason = fmt.Sprintf("%s:%s", source, c.Reason)
			c.Type = fmt.Sprintf("%s%s", source, c.Type)
		}
		meta.SetStatusCondition(&r.msvc.Status.Conditions, c)
	}
}

func (r *MongoDBReconciler) buildOutput(ctx context.Context, depl appsv1.Deployment) error {
	var mongoCfg corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: depl.Namespace, Name: depl.Name}, &mongoCfg); err != nil {
		return err
	}

	body := map[string]string{
		RootPassword: string(mongoCfg.Data["mongodb-root-password"]),
		DbHosts:      fmt.Sprintf("%s.%s.svc.cluster.local", depl.Name, depl.Namespace),
		DbUrl:        fmt.Sprintf("mongodb://%s:%s@%s.%s.svc.cluster.local/admin?authSource=admin", "root", string(mongoCfg.Data["mongodb-root-password"]), depl.Name, depl.Namespace),
	}

	connSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: depl.Namespace,
			Name:      fmt.Sprintf("msvc-%s", r.mdb.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &connSecret, func() error {
		connSecret.StringData = body
		if len(connSecret.OwnerReferences) == 1 && connSecret.OwnerReferences[0].UID != r.mdb.UID {
			r.logger.Debugf("owned by other")
			return nil
		}
		if err := controllerutil.SetControllerReference(r.mdb, &connSecret, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *MongoDBReconciler) ifMongoStandalone(ctx context.Context, _ *HelmMongoDB) error {
	var depl appsv1.Deployment
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: r.mdb.Namespace, Name: fmt.Sprintf("%s-mongodb", r.mdb.Name)}, &depl); err != nil {
		return errors.NewEf(err, "could not get deployment for helm resource")
	}

	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(deplConditions, metav1.Condition{
			Type:               string(cond.Type),
			Status:             metav1.ConditionStatus(cond.Status),
			LastTransitionTime: cond.LastTransitionTime,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	r.buildConditions("Deployment", deplConditions...)
	r.logger.Infof("Hello available: %+v", meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)))
	if !meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)) {
		opts := &client.ListOptions{
			LabelSelector: labels2.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
			Namespace:     depl.Namespace,
		}
		r.logger.Infof("list.options: %+v\n", opts)
		var podsList corev1.PodList
		if err := r.List(ctx, &podsList, opts); err != nil {
			return errors.NewEf(err, "could not list pods for deployment")
		}

		for _, pod := range podsList.Items {
			var podC []metav1.Condition
			for _, condition := range pod.Status.Conditions {
				podC = append(podC, metav1.Condition{
					Type:               string(condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             "NotSpecified",
					Message:            condition.Message,
				})
			}
			r.buildConditions("Pod", podC...)
			var containerC []metav1.Condition
			for _, cs := range pod.Status.ContainerStatuses {
				p := metav1.Condition{
					Type:   fmt.Sprintf("Name%s", cs.Name),
					Status: fn.StatusFromBool(cs.Ready),
				}
				if cs.State.Waiting != nil {
					p.Reason = cs.State.Waiting.Reason
					p.Message = cs.State.Waiting.Message
				}
				if cs.State.Running != nil {
					p.Reason = "Running"
					p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
				}
				containerC = append(containerC, p)
			}
			r.buildConditions("Container", containerC...)
			return nil
		}
	}

	r.buildConditions("", metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: r.lt,
		Reason:             "AllChecksPassed",
		Message:            "Mongo is Ready to use",
	})

	return r.buildOutput(ctx, depl)
}

func (r *MongoDBReconciler) ifMongoReplicaset(_ context.Context, _ *HelmMongoDB) error {
	// TODO: Implement me
	panic("implement me")
}

func (r *MongoDBReconciler) getMongoDBStatus(ctx context.Context, name types.NamespacedName) error {
	r.logger.Infof("HEllo here")
	w := bytes.NewBuffer([]byte{})
	command := exec.Command("kubectl", "get", "-o", "json", "-n", name.Namespace, fmt.Sprintf("mongodbs.msvc.kloudlite.io/%s", name.Name))
	command.Stderr = os.Stderr
	command.Stdout = w

	if err := command.Run(); err != nil {
		return errors.NewEf(err, "could not EXEC getMongoDBStatus command")
	}
	var hm HelmMongoDB
	if err := json.Unmarshal(w.Bytes(), &hm); err != nil {
		return errors.NewEf(err, "could not unmarshal into HelmMongoDBStatus")
	}

	r.buildConditions("Helm", hm.Status.Conditions...)

	if !meta.IsStatusConditionTrue(hm.Status.Conditions, "Deployed") {
		return nil
	}

	// Helm controller has installed CRD
	// STEP: go and find deployment/pod read their conditions and throw them into managed service
	v, ok := hm.Spec["architecture"]
	if !ok || v == "standalone" {
		return r.ifMongoStandalone(ctx, &hm)
	}
	return r.ifMongoReplicaset(ctx, &hm)
}

// +kubebuilder:rbac:groups=watchers.msvc.kloudlite.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watchers.msvc.kloudlite.io,resources=mongodbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watchers.msvc.kloudlite.io,resources=mongodbs/finalizers,verbs=update

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = controllers.GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true).With("MongoDBWatcher", "true")

	logger.Infof("Reconcilation Request Received")

	var mdb watcherMsvc.MongoDB
	if err := r.Get(ctx, req.NamespacedName, &mdb); err != nil {
		logger.Infof("HERE error occurred %+v", err)
		if apiErrors.IsNotFound(err) {
			// INFO: might have been deleted
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	mdb.Status.Conditions = []metav1.Condition{}
	r.mdb = &mdb
	r.msvc = &mdb
	logger.Infof("HERE")

	// var msvc crdsv1.ManagedService
	// if err := r.Get(ctx, req.NamespacedName, &msvc); err != nil {
	// 	if apiErrors.IsNotFound(err) {
	// 		// INFO: might have been deleted
	// 		return reconcileResult.OK()
	// 	}
	// 	return reconcileResult.Failed()
	// }
	//
	// r.msvc = &msvc

	if err := r.getMongoDBStatus(ctx, req.NamespacedName); err != nil {
		logger.Info("\nmongodb reconcilation failed", err)
		return r.notifyAndDie(ctx, &metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: r.lt,
			Reason:             "ErrDuringReconcilation",
			Message:            err.Error(),
		})
	}

	fmt.Println("msvc.status.conditions", len(r.msvc.Status.Conditions))
	for _, condition := range r.msvc.Status.Conditions {
		r.logger.Infof("c.Type: %s c.Status %s c.Reason %s, timestamp: %s\n", condition.Type, condition.Status, condition.Reason, condition.LastTransitionTime)
	}

	if err := r.Status().Update(ctx, r.msvc); err != nil {
		logger.Info("failed", err.Error())
		return reconcileResult.FailedE(err)
	}
	r.watchList = []types.NamespacedName{}
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.lt = metav1.Time{
		Time: time.Now(),
	}
	r.watchList = []types.NamespacedName{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&watcherMsvc.MongoDB{}).
		Watches(&source.Kind{
			Type: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "msvc.kloudlite.io/v1",
					"kind":       "MongoDB",
				},
			},
		}, handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			var reqs []reconcile.Request
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()},
			})
			fmt.Println("reconciliations: ", reqs)
			return reqs
		})).
		Watches(&source.Kind{
			Type: &appsv1.Deployment{},
		}, handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			labels := o.GetLabels()
			if s := labels["app.kubernetes.io/component"]; s != "mongodb" {
				return nil
			}
			if s := labels["app.kubernetes.io/name"]; s != "mongodb" {
				return nil
			}
			resourceName := labels["app.kubernetes.io/instance"]
			nn := types.NamespacedName{Namespace: o.GetNamespace(), Name: resourceName}
			if !r.IsInQueue(nn) {
				r.watchList = append(r.watchList, nn)
				return []reconcile.Request{{
					NamespacedName: nn,
				}}
			}
			fmt.Println("\nDB deployment reconciliations: ", r.watchList)
			return nil
		})).
		Complete(r)
}
