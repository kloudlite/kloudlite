package environment

import (
	"context"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addEnvironmentFinalizer(ctx context.Context, c client.Client, env *environmentsv1.Environment) (reconcile.Result, error) {
	if controllerutil.ContainsFinalizer(env, environmentFinalizer) {
		return reconcile.Result{}, nil
	}
	controllerutil.AddFinalizer(env, environmentFinalizer)
	if err := c.Update(ctx, env); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func removeEnvironmentFinalizer(ctx context.Context, c client.Client, key types.NamespacedName) (reconcile.Result, error) {
	env := &environmentsv1.Environment{}
	if err := c.Get(ctx, key, env); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if !controllerutil.ContainsFinalizer(env, environmentFinalizer) {
		return reconcile.Result{}, nil
	}
	base := env.DeepCopy()
	controllerutil.RemoveFinalizer(env, environmentFinalizer)
	if err := c.Patch(ctx, env, client.MergeFrom(base)); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		if apierrors.IsConflict(err) {
			return reconcile.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
