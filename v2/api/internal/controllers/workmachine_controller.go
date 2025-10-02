package controllers

import (
	"context"
	"time"

	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkMachineReconciler reconciles a WorkMachine object
type WorkMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles WorkMachine CR reconciliation
func (r *WorkMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "workmachine", "workmachine", req.Name)
	logger.Info("Reconciling WorkMachine")

	// Fetch the WorkMachine instance
	workMachine := &machinesv1.WorkMachine{}
	if err := r.Get(ctx, req.NamespacedName, workMachine); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("WorkMachine not found, probably deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get WorkMachine")
		return ctrl.Result{}, err
	}

	// Initialize status if it doesn't exist
	if workMachine.Status.State == "" {
		// First time - set current state to desired state
		workMachine.Status.State = workMachine.Spec.DesiredState
		now := metav1.Now()
		if workMachine.Spec.DesiredState == machinesv1.MachineStateRunning {
			workMachine.Status.StartedAt = &now
		}

		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to initialize WorkMachine status")
			return ctrl.Result{}, err
		}
		logger.Info("Initialized WorkMachine status", "state", workMachine.Status.State)
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	// Check if state transition is needed
	currentState := workMachine.Status.State
	desiredState := workMachine.Spec.DesiredState

	if currentState == desiredState {
		// No transition needed, machine is in desired state
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Handle state transitions
	logger.Info("State transition detected", "current", currentState, "desired", desiredState)

	switch currentState {
	case machinesv1.MachineStateStopped:
		if desiredState == machinesv1.MachineStateRunning {
			// Transition to starting
			workMachine.Status.State = machinesv1.MachineStateStarting
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to starting")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to starting")
			// Requeue quickly to move to running
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStarting:
		// Transition to running
		workMachine.Status.State = machinesv1.MachineStateRunning
		now := metav1.Now()
		workMachine.Status.StartedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to running")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now running")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

	case machinesv1.MachineStateRunning:
		if desiredState == machinesv1.MachineStateStopped {
			// Transition to stopping
			workMachine.Status.State = machinesv1.MachineStateStopping
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to stopping")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to stopping")
			// Requeue quickly to move to stopped
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStopping:
		// Transition to stopped
		workMachine.Status.State = machinesv1.MachineStateStopped
		now := metav1.Now()
		workMachine.Status.StoppedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to stopped")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now stopped")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&machinesv1.WorkMachine{}).
		Complete(r)
}
