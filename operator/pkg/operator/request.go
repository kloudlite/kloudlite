package operator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
)

type Request[T Resource] struct {
	ctx            context.Context
	client         client.Client
	Object         T
	Logger         logging.Logger
	anchorName     string
	internalLogger logging.Logger
	locals         map[string]any

	reconStartTime time.Time
	timerMap       map[string]time.Time

	resourceRefs []ResourceRef
}

type ReconcilerCtx context.Context

func NewReconcilerCtx(parent context.Context, logger logging.Logger) ReconcilerCtx {
	return context.WithValue(parent, "logger", logger)
}

func NewRequest[T Resource](ctx ReconcilerCtx, c client.Client, nn types.NamespacedName, resource T) (*Request[T], error) {
	if err := c.Get(ctx, nn, resource); err != nil {
		return nil, err
	}

	// TODO: useful only when reoncilers triggered from envtest as of now
	if resource.GetObjectKind().GroupVersionKind().Kind == "" {
		kinds, _, err := c.Scheme().ObjectKinds(resource)
		if err != nil {
			return nil, err
		}
		if len(kinds) > 0 {
			resource.GetObjectKind().SetGroupVersionKind(kinds[0])
		}
	}

	logger, ok := ctx.Value("logger").(logging.Logger)
	if !ok {
		panic("no logger passed into NewRequest")
	}

	anchorName := func() string {
		x := strings.ToLower(fmt.Sprintf("%s-%s", resource.GetObjectKind().GroupVersionKind().Kind, resource.GetName()))
		if len(x) >= 63 {
			return x[:62]
		}
		return x
	}()

	if resource.GetStatus().Checks == nil {
		resource.GetStatus().Checks = map[string]Check{}
	}

	return &Request[T]{
		ctx:            ctx,
		client:         c,
		Object:         resource,
		Logger:         logger.WithName(nn.String()).WithKV("NN", nn.String()),
		internalLogger: logger.WithName(nn.String()).WithKV("NN", nn.String()).WithOptions(zap.AddCallerSkip(1)),
		anchorName:     anchorName,
		locals:         map[string]any{},
		timerMap:       map[string]time.Time{},
	}, nil
}

func (r *Request[T]) GetAnchorName() string {
	return r.anchorName
}

func (r *Request[T]) GetClient() client.Client {
	return r.client
}

func (r *Request[T]) EnsureLabelsAndAnnotations() stepResult.Result {
	labels := r.Object.GetEnsuredLabels()
	annotations := r.Object.GetEnsuredAnnotations()

	// if r.Object.GetNamespace() != "" {
	// 	var ns corev1.Namespace
	// 	if err := r.client.Get(r.Context(), fn.NN("", r.Object.GetNamespace()), &ns); err != nil {
	// 		return stepResult.New().Err(errors.NewEf(err, "could not get namespace %q", r.Object.GetNamespace()))
	// 	}
	//
	// 	for k, v := range ns.GetLabels() {
	// 		if strings.HasPrefix(k, "kloudlite.io/") {
	// 			labels[k] = v
	// 		}
	// 	}
	// 	for k, v := range ns.GetAnnotations() {
	// 		if strings.HasPrefix(k, "kloudlite.io/") {
	// 			annotations[k] = v
	// 		}
	// 	}
	// }

	annotations[constants.GVKKey] = r.Object.GetObjectKind().GroupVersionKind().String()

	hasAllLabels := fn.MapContains(r.Object.GetLabels(), labels)
	hasAllAnnotations := fn.MapContains(r.Object.GetAnnotations(), annotations)

	if !hasAllLabels || !hasAllAnnotations {
		x := r.Object.GetLabels()
		if x == nil {
			x = make(map[string]string, len(labels))
		}
		for k, v := range labels {
			x[k] = v
		}
		r.Object.SetLabels(x)

		y := r.Object.GetAnnotations()
		if y == nil {
			y = make(map[string]string, len(annotations))
		}
		for k, v := range annotations {
			y[k] = v
		}
		r.Object.SetAnnotations(y)
		if err := r.client.Update(r.ctx, r.Object); err != nil {
			return stepResult.New().Err(err)
		}
		return stepResult.New().RequeueAfter(1 * time.Second)
	}

	return stepResult.New().Continue(true)
}

func (r *Request[T]) ShouldReconcile() bool {
	return r.Object.GetLabels()[constants.ShouldReconcile] != "false"
}

func (r *Request[T]) EnsureChecks(names ...string) stepResult.Result {
	obj, ctx, checks := r.Object, r.Context(), r.Object.GetStatus().Checks
	nChecks := len(checks)

	if checks == nil {
		checks = map[string]Check{}
	}

	for i := range names {
		if _, ok := checks[names[i]]; !ok {
			checks[names[i]] = Check{}
		}
	}

	if nChecks != len(checks) {
		obj.GetStatus().Checks = checks
		if err := r.client.Status().Update(ctx, obj); err != nil {
			return r.Done().Err(err)
			// return r.FailWithOpError(err)
		}
	}

	return stepResult.New().Continue(true)
}

func (r *Request[T]) ClearStatusIfAnnotated() stepResult.Result {
	obj := r.Object
	ann := obj.GetAnnotations()

	if v, ok := ann[constants.ResetCheckKey]; ok {
		if _, ok2 := obj.GetStatus().Checks[v]; ok2 {
			delete(ann, constants.ResetCheckKey)
			obj.SetAnnotations(ann)
			if err := r.client.Update(context.TODO(), obj); err != nil {
				return stepResult.New().Err(err)
			}

			delete(obj.GetStatus().Checks, v)
			if err := r.client.Status().Update(context.TODO(), obj); err != nil {
				return stepResult.New().Err(err)
			}
			return r.Done().RequeueAfter(2 * time.Second)
		}
	}

	if v := ann[constants.ClearStatusKey]; v == "true" {
		delete(ann, constants.ClearStatusKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(r.Context(), obj); err != nil {
			return r.Done().Err(err)
		}

		obj.GetStatus().IsReady = false
		obj.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
		obj.GetStatus().Checks = nil
		obj.GetStatus().Message = nil
		// obj.GetStatus().Messages = nil
		// obj.GetStatus().Conditions = nil
		// obj.GetStatus().OpsConditions = nil
		// obj.GetStatus().ChildConditions = nil
		// obj.GetStatus().DisplayVars = nil
		// obj.GetStatus().GeneratedVars = rawJson.RawJson{}

		if err := r.client.Status().Update(context.TODO(), obj); err != nil {
			return r.Done().Err(err)
		}
		return r.Done().RequeueAfter(0 * time.Second)
	}
	return r.Next()
}

func (r *Request[T]) RestartIfAnnotated() stepResult.Result {
	ctx, obj := r.Context(), r.Object
	ann := obj.GetAnnotations()
	if v := ann[constants.RestartKey]; v == "true" {
		delete(ann, constants.RestartKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(ctx, obj); err != nil {
			return r.Done().Err(err)
		}

		if err := fn.RolloutRestart(r.client, fn.Deployment, obj.GetNamespace(), obj.GetEnsuredLabels()); err != nil {
			return stepResult.New().Err(err)
		}
		if err := fn.RolloutRestart(r.client, fn.StatefulSet, obj.GetNamespace(), obj.GetEnsuredLabels()); err != nil {
			return stepResult.New().Err(err)
		}
		return r.Done().RequeueAfter(500 * time.Millisecond)
	}

	return r.Next()
}

func (r *Request[T]) EnsureFinalizers(finalizers ...string) stepResult.Result {
	obj := r.Object

	if !fn.ContainsFinalizers(obj, finalizers...) {
		for i := range finalizers {
			controllerutil.AddFinalizer(obj, finalizers[i])
		}
		if err := r.client.Update(r.Context(), obj); err != nil {
			return r.Done().Err(err)
		}
		return stepResult.New()
	}
	return stepResult.New().Continue(true)
}

func (r *Request[T]) CheckFailed(name string, check Check, msg string) stepResult.Result {
	check.Status = false
	check.Message = msg
	if r.Object.GetStatus().Checks == nil {
		r.Object.GetStatus().Checks = make(map[string]Check, 1)
	}
	r.Object.GetStatus().Checks[name] = check
	r.Object.GetStatus().Message.Set(name, check.Message)
	r.Object.GetStatus().IsReady = false
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
	if err := r.client.Status().Update(r.ctx, r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	return stepResult.New()
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) Done(result ...ctrl.Result) stepResult.Result {
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
	if err := r.client.Status().Update(context.TODO(), r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	if len(result) > 0 {
		return stepResult.New().RequeueAfter(result[0].RequeueAfter)
	}
	return stepResult.New()
}

func (r *Request[T]) Next() stepResult.Result {
	return stepResult.New().Continue(true)
}

func (r *Request[T]) UpdateStatus() stepResult.Result {
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
	checks := r.Object.GetStatus().Checks

	for name := range checks {
		if checks[name].Status {
			if err := r.Object.GetStatus().Message.Delete(name); err != nil {
				return stepResult.New().Err(err)
			}

			if r.Object.GetStatus().Message.Len() == 0 {
				r.Object.GetStatus().Message = nil
			}
		}
	}

	if err := r.client.Status().Update(r.Context(), r.Object); err != nil {
		return stepResult.New().Err(err)
	}
	return stepResult.New().Continue(true)
}

func (r *Request[T]) Finalize() stepResult.Result {
	controllerutil.RemoveFinalizer(r.Object, constants.CommonFinalizer)
	controllerutil.RemoveFinalizer(r.Object, constants.ForegroundFinalizer)
	return stepResult.New().Err(r.client.Update(r.ctx, r.Object))
}

func (r *Request[T]) LogPreReconcile() {
	blue := color.New(color.FgBlue).SprintFunc()
	r.reconStartTime = time.Now()
	r.internalLogger.Infof(blue("[new] reconcilation start"))
}

func (r *Request[T]) LogPostReconcile() {
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}

	tDiff := time.Since(r.reconStartTime).Seconds()

	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
	r.Object.GetStatus().Resources = r.GetOwnedResources()

	defer func() {
		if err := r.client.Status().Update(r.Context(), r.Object); err != nil {
			if !apiErrors.IsNotFound(err) && !apiErrors.IsConflict(err) {
				red := color.New(color.FgHiRed, color.Bold).SprintFunc()
				r.internalLogger.Infof(red("[end] (took: %.2fs) reconcilation in progress, as status update failed"), tDiff)
			}
		}
	}()

	if !r.Object.GetStatus().IsReady {
		yellow := color.New(color.FgHiYellow, color.Bold).SprintFunc()
		r.internalLogger.Infof(yellow("[end] (took: %.2fs) reconcilation complete"), tDiff)
		return
	}

	green := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	r.internalLogger.Infof(green("[end] (took: %.2fs) reconcilation complete"), tDiff)
}

func (r *Request[T]) LogPreCheck(checkName string) {
	blue := color.New(color.FgBlue).SprintFunc()
	r.timerMap[checkName] = time.Now()
	check, ok := r.Object.GetStatus().Checks[checkName]
	if ok {
		r.internalLogger.Infof(blue("[check:start] %-20s [status] %-5v"), checkName, check.Status)
	}
}

func (r *Request[T]) LogPostCheck(checkName string) {
	tDiff := time.Since(r.timerMap[checkName]).Seconds()
	check, ok := r.Object.GetStatus().Checks[checkName]
	if ok {
		if !check.Status {
			red := color.New(color.FgRed).SprintFunc()
			r.internalLogger.Infof(red("[check:end] (took: %.2fs) %-20s [status] %v [message] %v"), tDiff, checkName, check.Status, check.Message)
		}
		green := color.New(color.FgHiGreen, color.Bold).SprintFunc()
		r.internalLogger.Infof(green("[check:end] (took: %.2fs) %-20s [status] %v"), tDiff, checkName, check.Status)
	}
}

func (r *Request[T]) GetOwnedResources() []ResourceRef {
	return r.resourceRefs
}

func (r *Request[T]) AddToOwnedResources(refs ...ResourceRef) {
	r.resourceRefs = append(r.resourceRefs, refs...)
}

func (r *Request[T]) CleanupOwnedResources() stepResult.Result {
	ctx, obj := r.Context(), r.Object
	check := Check{Generation: r.Object.GetGeneration()}

	checkName := "cleanupLogic"

	resources := r.GetOwnedResources()

	for i := range resources {
		res := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": resources[i].APIVersion,
			"kind":       resources[i].Kind,
			"metadata": map[string]any{
				"name":      resources[i].Name,
				"namespace": resources[i].Namespace,
			},
		}}

		if err := r.client.Get(ctx, client.ObjectKeyFromObject(res), res); err != nil {
			if !apiErrors.IsNotFound(err) {
				return r.CheckFailed("CleanupResource", check, err.Error()).Err(nil)
			}
			return r.CheckFailed("CleanupResource", check,
				fmt.Sprintf("waiting for deletion of owned resource gvk=%s, nn=%s", res.GetObjectKind().GroupVersionKind().String(), fn.NN(res.GetNamespace(), res.GetName())),
			).Err(nil)
		}

		if res.GetDeletionTimestamp() == nil {
			if err := r.client.Delete(ctx, res); err != nil {
				return r.CheckFailed("CleanupResource", check, err.Error()).Err(nil)
			}
		}
	}

	check.Status = true
	if check != obj.GetStatus().Checks[checkName] {
		obj.GetStatus().Checks[checkName] = check
		r.UpdateStatus()
	}
	return r.Next()
}

func ParseResourceRef(obj client.Object) ResourceRef {
	return ResourceRef{
		TypeMeta: metav1.TypeMeta{
			Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
			APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		},
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}
