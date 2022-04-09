package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/yext/yerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
}

const (
	ConditionReady   = "Ready"
	ConditionAborted = "Aborted"
)

type Bool bool

func (b Bool) Condition() metav1.ConditionStatus {
	if !b {
		return metav1.ConditionFalse
	}
	return metav1.ConditionTrue
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *ProjectReconciler) CheckIfNSExists(ns string) (*corev1.Namespace, bool) {
	result := &corev1.Namespace{}
	e := r.Client.Get(context.Background(), types.NamespacedName{Name: ns}, result)
	if e != nil {
		return nil, false
	}
	return result, true
}

const projectFinalizer = "finalizers.kloudlite.io/project"

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)

	project := &crdsv1.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("resource (name=%s, namespace=%s) could not be found, has been deleted", req.Name, req.Namespace))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	fmt.Println("Received Condition: ", project.Status.Conditions)

	// STEP: is object marked for deletion
	isMarkedForDeletion := project.GetDeletionTimestamp() != nil
	if isMarkedForDeletion {
		fmt.Println("RESOURCE is marked for deletion")
		containsFinalizer := controllerutil.ContainsFinalizer(project, projectFinalizer)
		// STEP: 2: add finalizers if not present
		if !containsFinalizer {
			controllerutil.AddFinalizer(project, projectFinalizer)
			fmt.Println("ADDING resource finalizer")
			err := r.Update(ctx, project)
			if err != nil {
				return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("could not add finalizer to project(%s) because ", project.Name, err))
			}
			return ctrl.Result{}, nil
		}

		// STEP: it will container finalizer for sure
		fmt.Println("EXECUTING resource finalizer")
		err := finalizeProject(project, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(project, projectFinalizer)
		err = r.Update(ctx, project)
		if err != nil {
			return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("could not remove finalizer from project(%s) because %w", project.Name, err))
		}
		return ctrl.Result{}, nil
	}

	if isNotReady(project) || isAborted(project) {
		logger.Info("returning as condition ready=false or aborted = true")
		return ctrl.Result{}, nil
	}

	err := r.setReady(ctx, project, false)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 1}, nil
	}

	// DO the operations below

	_, ok := r.CheckIfNSExists(project.Name)
	fmt.Println("NS Exists: ", ok)
	if !ok {
		// STEP: create resource

		// WARN: extract `hotspot` it into env variable
		jobMgr := r.ClientSet.BatchV1().Jobs("hotspot")

		jobData, err := lib.UseJobTemplate(&lib.JobVars{
			Name:            "job-create",
			Namespace:       "hotspot",
			ServiceAccount:  "hotspot-svc-account",
			Image:           "harbor.dev.madhouselabs.io/kloudlite/jobs/project:latest",
			ImagePullPolicy: "Always",
			Args: []string{
				"create",
				"--name", project.Name,
			},
		})

		if err != nil {
			return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("could not create batchv1.job from template as %w", err))
		}

		job, err := jobMgr.Create(ctx, jobData, metav1.CreateOptions{})
		if err != nil {
			return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("error creating job from jobMgr"))
		}

		status, err := lib.WatchJob(ctx, jobMgr, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", job.ObjectMeta.Name),
		})

		if err != nil {
			return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("watching job failed because %w", err))
		}

		if !status {
			err := r.setAborted(ctx, project, true, "job failed")
			if err != nil {
				return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("job condition update failed as %w", err))
			}
			return ctrl.Result{}, nil
		}

		err = r.setReady(ctx, project, Bool(status))
		if err != nil {
			return ctrl.Result{}, yerrors.Wrap(yerrors.Errorf("job condition update failed as %w", err))
		}
		return ctrl.Result{}, nil
	}

	// STEP: update resource

	fmt.Printf("(project name) %v", project.Name)

	return ctrl.Result{}, nil
}

func finalizeProject(project *crdsv1.Project, logger logr.Logger) error {
	logger.Info("project finalizer triggerred")
	return nil
}

func isReady(project *crdsv1.Project) bool {
	return meta.IsStatusConditionTrue(project.Status.Conditions, ConditionReady)
}

func isNotReady(project *crdsv1.Project) bool {
	return meta.IsStatusConditionFalse(project.Status.Conditions, ConditionReady)
}

func isAborted(project *crdsv1.Project) bool {
	cond := meta.FindStatusCondition(project.Status.Conditions, ConditionAborted)
	// fmt.Printf("cond: +%v\n", cond)
	if cond == nil {
		return false
	}
	return cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == project.GetObjectMeta().GetGeneration()
}

func (r *ProjectReconciler) setReady(ctx context.Context, project *crdsv1.Project, status Bool) error {
	meta.RemoveStatusCondition(&project.Status.Conditions, ConditionAborted)
	meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status.Condition(),
		Reason:  "initialized",
		Message: "",
	})
	return r.updateStatus(ctx, project)
}

func (r *ProjectReconciler) setAborted(ctx context.Context, project *crdsv1.Project, status Bool, msg string) error {
	meta.RemoveStatusCondition(&project.Status.Conditions, ConditionReady)
	meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
		Type:               ConditionAborted,
		Status:             status.Condition(),
		ObservedGeneration: project.ObjectMeta.GetGeneration(),
		Reason:             "final",
		Message:            msg,
	})
	return r.updateStatus(ctx, project)
}

func (r *ProjectReconciler) updateStatus(ctx context.Context, project *crdsv1.Project) error {
	err := r.Status().Update(ctx, project)
	if err != nil {
		return yerrors.Wrap(yerrors.Errorf("could not update conditions on project (%s) as %w", project.Name, err))
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Complete(r)
}
