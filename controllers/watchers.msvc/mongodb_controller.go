package watchersmsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"operators.kloudlite.io/controllers/crds"
	"os"
	"os/exec"
	"strings"
	"time"

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	watcherMsvc "operators.kloudlite.io/apis/watchers.msvc/v1"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
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
	Scheme    *runtime.Scheme
	logger    *zap.SugaredLogger
	mdb       *watcherMsvc.MongoDB
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
	meta.SetStatusCondition(&r.mdb.Status.Conditions, *cond)
	if err := r.Status().Update(ctx, r.mdb); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *MongoDBReconciler) buildConditions(source string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&r.mdb.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: r.lt,
		Message:            "Not All Checks completed",
	})
	for _, c := range conditions {
		if c.Reason == "" {
			c.Reason = "NotSpecified"
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
		meta.SetStatusCondition(&r.mdb.Status.Conditions, c)
	}
}

func (r *MongoDBReconciler) buildStandaloneOutput(ctx context.Context, depl appsv1.Deployment) error {
	var mongoCfg corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: depl.Namespace, Name: depl.Name}, &mongoCfg); err != nil {
		return err
	}

	body := map[string]string{
		RootPassword: string(mongoCfg.Data["mongodb-root-password"]),
		DbHosts:      fmt.Sprintf("%s.%s.svc.cluster.local:27017", depl.Name, depl.Namespace),
		DbUrl:        fmt.Sprintf("mongodb://%s:%s@%s.%s.svc.cluster.local:27017/admin?authSource=admin", "root", string(mongoCfg.Data["mongodb-root-password"]), depl.Name, depl.Namespace),
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

func (r *MongoDBReconciler) buildClusterOutput(ctx context.Context, sts appsv1.StatefulSet, arbiter appsv1.StatefulSet) error {
	var mongoCfg corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: sts.Namespace, Name: sts.Name}, &mongoCfg); err != nil {
		return err
	}

	var hosts []string

	for i := 0; int32(i) < sts.Status.ReadyReplicas; i++ {
		hosts = append(hosts, fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:27017", sts.Name, i, sts.Name, sts.Namespace))
	}
	for i := 0; int32(i) < arbiter.Status.ReadyReplicas; i++ {
		hosts = append(hosts, fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:27017", arbiter.Name, i, arbiter.Name, arbiter.Namespace))
	}

	body := make(map[string]string, 3)
	body[RootPassword] = string(mongoCfg.Data["mongodb-root-password"])
	body[DbHosts] = strings.Join(hosts, ",")
	body[DbUrl] = fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", string(mongoCfg.Data["mongodb-root-password"]), body[DbHosts])

	connSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sts.Namespace,
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
			return errors.NewEf(err, "could not set controller reference over conn secret")
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
			Type:    string(cond.Type),
			Status:  metav1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
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
		return nil
	}

	r.buildConditions("", metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: r.lt,
		Reason:             "AllChecksPassed",
		Message:            "Mongo is Ready to use",
	})

	return r.buildStandaloneOutput(ctx, depl)
}

func (r *MongoDBReconciler) ifMongoReplicaset(ctx context.Context, _ *HelmMongoDB) error {
	var arbiter appsv1.StatefulSet
	if err := r.Get(ctx, types.NamespacedName{Namespace: r.mdb.Namespace, Name: fmt.Sprintf("%s-mongodb-arbiter", r.mdb.Name)}, &arbiter); err != nil {
		return err
	}

	if arbiter.Status.ReadyReplicas != arbiter.Status.Replicas {
		return errors.Newf("sts %s is not ready", arbiter.Name)
	}

	var sts appsv1.StatefulSet
	if err := r.Get(ctx, types.NamespacedName{Namespace: r.mdb.Namespace, Name: fmt.Sprintf("%s-mongodb", r.mdb.Name)}, &sts); err != nil {
		return errors.NewEf(err, "could not find statefulset related to mongodb (cluster) installation")
	}

	if sts.Status.ReadyReplicas != sts.Status.Replicas {
		var pl corev1.PodList
		if err := r.List(ctx, &pl, &client.ListOptions{
			LabelSelector: labels2.SelectorFromValidatedSet(sts.Spec.Selector.MatchLabels),
			Namespace:     r.mdb.Namespace,
		}); err != nil {
			return errors.NewEf(err, "could not list pods with selectors from statefulset")
		}

		for _, pod := range pl.Items {
			var podC []metav1.Condition
			for _, condition := range pod.Status.Conditions {
				podC = append(podC, metav1.Condition{
					Type:               string(condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Message:            condition.Message,
				})
			}
			r.buildConditions("Pod", podC...)

			var initContainerSt []metav1.Condition
			for idx, st := range pod.Status.InitContainerStatuses {
				c := metav1.Condition{
					Type:   fmt.Sprintf("Idx%d", idx),
					Status: fn.StatusFromBool(st.Ready),
				}
				if st.State.Terminated != nil {
					c.Reason = st.State.Terminated.Reason
					c.Message = fmt.Sprintf("terminated with exit-code %d\n", st.State.Terminated.ExitCode)
					c.LastTransitionTime = st.State.Terminated.FinishedAt
				}

				initContainerSt = append(initContainerSt, c)
			}

			r.buildConditions("InitContainer", initContainerSt...)

			var containerC []metav1.Condition
			for idx, cs := range pod.Status.ContainerStatuses {
				p := metav1.Condition{
					Type:   fmt.Sprintf("Idx%d", idx),
					Status: fn.StatusFromBool(cs.Ready),
				}
				if cs.State.Waiting != nil {
					p.Reason = cs.State.Waiting.Reason
					p.Message = fmt.Sprintf("container(%s): %s", cs.Name, cs.State.Waiting.Message)
				}
				if cs.State.Running != nil {
					p.Reason = "Running"
					p.Message = fmt.Sprintf("Container (%s) running since %s",
						cs.Name,
						cs.State.Running.StartedAt.String(),
					)
				}
				containerC = append(containerC, p)
			}
			r.buildConditions("Container", containerC...)
			return nil
		}
		return nil
	}

	// ASSERT: all replicas are ready
	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "ReplicasAreReady",
		Message: "All replicasets are ready",
	})
	return r.buildClusterOutput(ctx, sts, arbiter)
}

func (r *MongoDBReconciler) getMongoDBStatus(ctx context.Context, name types.NamespacedName) error {
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
	r.logger = crds.GetLogger(req.NamespacedName)
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

	if err := r.Status().Update(ctx, r.mdb); err != nil {
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
