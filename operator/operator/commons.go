package operator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureAnchor[T rApi.Resource](req *rApi.Request[T]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.GetStatus().Checks
	check := rApi.Check{Generation: req.Object.GetGeneration()}

	anchor := &crdsv1.Anchor{ObjectMeta: metav1.ObjectMeta{Name: req.GetAnchorName(), Namespace: obj.GetNamespace()}}
	if _, err := controllerutil.CreateOrUpdate(ctx, req.GetClient(), anchor, func() error {
		if !fn.IsOwner(anchor, fn.AsOwner(obj)) {
			anchor.SetOwnerReferences(append(anchor.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}

		controllerutil.AddFinalizer(anchor, constants.ForegroundFinalizer)

		anchor.Spec.Type = anchor.Kind
		anchor.Spec.ParentGVK = fn.GVK(anchor)
		return nil
	}); err != nil {
		return req.CheckFailed("AnchorReady", check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks["AnchorReady"] {
		checks["AnchorReady"] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, "anchor", anchor)
	return req.Next()
}
