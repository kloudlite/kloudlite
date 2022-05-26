/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crds

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/templates"

	"github.com/redhat-cop/operator-utils/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
)

const controllerName = "Account_controller"

// AccountReconciler reconciles a Account object
type AccountReconciler struct {
	util.ReconcilerBase
	Log logr.Logger
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=accounts/finalizers,verbs=update

func (r *AccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("project", req.NamespacedName)
	instance := &crdsv1.Account{}
	err := r.GetClient().Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if ok, err := r.IsValid(instance); !ok {
		return r.ManageError(ctx, instance, err)
	}

	if ok := r.IsInitialized(instance); !ok {
		err := r.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}

	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, controllerName) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		util.RemoveFinalizer(instance, controllerName)
		err = r.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}

	parse, err := templates.Parse(templates.AccountWireguard, instance)
	if err != nil {
		return r.ManageError(ctx, instance, err)
	}
	if _, err := functions.KubectlApplyExec(parse); err != nil {
		return r.ManageError(ctx, instance, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Account{}).
		Complete(r)
}

func (r *AccountReconciler) manageCleanUpLogic(ctx context.Context, instance *crdsv1.Account) error {
	return nil
}
