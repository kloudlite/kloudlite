package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nxtcoder17/go.pkgs/log"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
)

type KV struct {
	data map[string]any
}

func (kv *KV) Set(k string, v any) {
	if kv.data == nil {
		kv.data = make(map[string]any)
	}
	kv.data[k] = v
}

func (kv *KV) Get(k string) (any, error) {
	a, ok := kv.data[k]
	if !ok {
		return nil, fmt.Errorf("key (%s) not found in req.KV", k)
	}
	return a, nil
}

type Request[T Resource] struct {
	ctx            context.Context
	client         client.Client
	Object         T
	Logger         *slog.Logger
	internalLogger *slog.Logger
	KV             KV

	startedAt time.Time
}

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName, resource T) (*Request[T], error) {
	if err := c.Get(ctx, nn, resource); err != nil {
		return nil, err
	}

	// TODO: useful only when reconcilers triggered from envtest as of now
	if resource.GetObjectKind().GroupVersionKind().Kind == "" {
		kinds, _, err := c.Scheme().ObjectKinds(resource)
		if err != nil {
			return nil, err
		}
		if len(kinds) > 0 {
			resource.GetObjectKind().SetGroupVersionKind(kinds[0])
		}
	}

	if resource.GetStatus().Checks == nil {
		resource.GetStatus().Checks = map[string]CheckResult{}
	}

	resource.EnsureGVK()

	return &Request[T]{
		ctx:            ctx,
		client:         c,
		Object:         resource,
		Logger:         log.DefaultLogger().SkipFrames(1).With("NN", nn.String(), "gvk", resource.GetObjectKind().GroupVersionKind().String()).Slog(),
		internalLogger: log.DefaultLogger().SkipFrames(4).With("NN", nn.String(), "gvk", resource.GetObjectKind().GroupVersionKind().String()).Slog(),
		KV:             KV{},
	}, nil
}

func (r *Request[T]) GetClient() client.Client {
	return r.client
}

func (r *Request[T]) EnsureLabelsAndAnnotations() StepResult {
	labels := r.Object.GetEnsuredLabels()
	annotations := r.Object.GetEnsuredAnnotations()

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
			return newStepResult().Err(err)
		}
		return newStepResult().RequeueAfter(500 * time.Millisecond)
	}

	return newStepResult().Continue(true)
}

func (r *Request[T]) ShouldReconcile() bool {
	return r.Object.GetAnnotations()[AnnotationShouldReconcileKey] != "false"
}

func (r *Request[T]) EnsureCheckList(expected []CheckDefinition) StepResult {
	if !slices.Equal(expected, r.Object.GetStatus().CheckList) {
		checks := make(map[string]CheckResult, len(expected))
		for i := range expected {
			checks[expected[i].Name] = CheckResult{State: WaitingState}
		}
		r.Object.GetStatus().Checks = checks
		r.Object.GetStatus().CheckList = expected

		if err := r.client.Status().Update(r.ctx, r.Object); err != nil {
			return newStepResult().Err(err)
		}
		return newStepResult().RequeueAfter(500 * time.Millisecond)
	}

	return newStepResult().Continue(true)
}

//func (r *Request[T]) EnsureChecks(names ...string) step_result.Result {
//	return step_result.New().Continue(true)
//	// obj, ctx := r.Object, r.Context()
//	//
//	// checks := fn.MapMerge(obj.GetStatus().Checks)
//	// updated := ensureChecks(checks, names...)
//	//
//	// if updated {
//	// 	obj.GetStatus().Checks = checks
//	// 	if err := r.client.Status().Update(ctx, obj); err != nil {
//	// 		return r.Done().Err(err)
//	// 	}
//	// }
//	// return step_result.New().Continue(true)
//}

func (r *Request[T]) ClearStatusIfAnnotated() StepResult {
	obj := r.Object
	ann := obj.GetAnnotations()

	if v, ok := ann[AnnotationResetCheckKey]; ok {
		if _, ok2 := obj.GetStatus().Checks[v]; ok2 {
			delete(ann, AnnotationResetCheckKey)
			obj.SetAnnotations(ann)
			if err := r.client.Update(context.TODO(), obj); err != nil {
				return newStepResult().Err(err)
			}

			delete(obj.GetStatus().Checks, v)
			if err := r.client.Status().Update(context.TODO(), obj); err != nil {
				return newStepResult().Err(err)
			}
			return newStepResult().RequeueAfter(2 * time.Second)
		}
	}

	if v := ann[AnnotationClearStatusKey]; v == "true" {
		delete(ann, AnnotationClearStatusKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(r.Context(), obj); err != nil {
			return newStepResult().Err(err)
		}

		obj.GetStatus().IsReady = false
		obj.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
		obj.GetStatus().Checks = nil

		if err := r.client.Status().Update(context.TODO(), obj); err != nil {
			return newStepResult().Err(err)
		}
		return newStepResult().RequeueAfter(1 * time.Second)
	}
	return newStepResult().Continue(true)
}

func (r *Request[T]) RestartIfAnnotated() StepResult {
	ctx, obj := r.Context(), r.Object
	ann := obj.GetAnnotations()
	if v := ann[AnnotationRestartKey]; v == "true" {
		delete(ann, AnnotationRestartKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(ctx, obj); err != nil {
			return newStepResult().Err(err)
		}

		if err := fn.RolloutRestart(r.client, fn.Deployment, obj.GetNamespace(), obj.GetEnsuredLabels()); err != nil {
			return newStepResult().Err(err)
		}
		if err := fn.RolloutRestart(r.client, fn.StatefulSet, obj.GetNamespace(), obj.GetEnsuredLabels()); err != nil {
			return newStepResult().Err(err)
		}
		return newStepResult().RequeueAfter(500 * time.Millisecond)
	}

	return newStepResult().Continue(true)
}

func (r *Request[T]) EnsureFinalizers(finalizers ...string) StepResult {
	obj := r.Object

	if !fn.ContainsFinalizers(obj, finalizers...) {
		for i := range finalizers {
			controllerutil.AddFinalizer(obj, finalizers[i])
		}
		if err := r.client.Update(r.Context(), obj); err != nil {
			return newStepResult().Err(err)
		}
		return newStepResult().RequeueAfter(500 * time.Millisecond)
	}
	return newStepResult().Continue(true)
}

func (r *Request[T]) Context() context.Context {
	return r.ctx
}

func (r *Request[T]) statusUpdate() error {
	return r.client.Status().Update(r.ctx, r.Object)
}

//func (r *Request[T]) Done(result ...ctrl.Result) step_result.Result {
//	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
//	if err := r.client.Status().Update(context.TODO(), r.Object); err != nil {
//		return step_result.New().Err(err)
//	}
//	if len(result) > 0 {
//		return step_result.New().RequeueAfter(result[0].RequeueAfter)
//	}
//	return step_result.New()
//}
//
//func (r *Request[T]) Next() step_result.Result {
//	return step_result.New().Continue(true)
//}
//
//func (r *Request[T]) Finalize() step_result.Result {
//	controllerutil.RemoveFinalizer(r.Object, Finalizer)
//	controllerutil.RemoveFinalizer(r.Object, ForegroundFinalizer)
//	controllerutil.RemoveFinalizer(r.Object, "finalizers.kloudlite.io")
//	controllerutil.RemoveFinalizer(r.Object, "finalizers.kloudlite.io/watch") // Keep till all clusters are updated to v1.0.4
//	return step_result.New().Err(r.client.Update(r.ctx, r.Object))
//}

func (r *Request[T]) PreReconcile() {
	blue := color.New(color.FgBlue).SprintFunc()
	r.startedAt = time.Now()
	r.internalLogger.Info(blue("[reconcile:start] start"))
}

var checkStates = map[CheckState]string{
	WaitingState: "🔶",
	RunningState: "⌛",
	ErroredState: "🔴",
	PassedState:  "🌿",
}

func (r *Request[T]) PostReconcile() {
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}

	tDiff := time.Since(r.startedAt).Seconds()

	isReady := r.Object.GetStatus().IsReady

	if isReady {
		r.Object.GetStatus().LastReadyGeneration = r.Object.GetGeneration()
	}

	if err := r.client.Status().Update(r.Context(), r.Object); err != nil {
		if !apiErrors.IsNotFound(err) && !apiErrors.IsConflict(err) {
			red := color.New(color.FgHiRed, color.Bold).SprintFunc()
			r.internalLogger.Info(fmt.Sprintf(red("[reconcile:end] (took: %.2fs) reconcilation in progress, as status update failed"), tDiff))
		}
	}

	m := make(map[string]string, len(r.Object.GetAnnotations()))
	maps.Copy(m, r.Object.GetAnnotations())

	m[AnnotationResourceReady] = func() string {
		readyMsg := strconv.FormatBool(isReady)

		generationMsg := fmt.Sprintf("%d", r.Object.GetStatus().LastReadyGeneration)
		if !isReady && r.Object.GetGeneration() != r.Object.GetStatus().LastReadyGeneration {
			generationMsg = fmt.Sprintf("%d -> %d", r.Object.GetStatus().LastReadyGeneration, r.Object.GetGeneration())
		}

		deletionMsg := ""
		if r.Object.GetDeletionTimestamp() != nil {
			deletionMsg = ", being deleted"
		}

		return fmt.Sprintf("%s (%s%s)", readyMsg, generationMsg, deletionMsg)
	}()

	m[AnnotationResourceChecks] = func() string {
		checks := make([]string, 0, len(r.Object.GetStatus().Checks))
		currentCheck := ""
		keys := fn.MapKeys(r.Object.GetStatus().Checks)
		slices.Sort(keys)
		for _, k := range keys {
			if r.Object.GetStatus().Checks[k].State == RunningState || r.Object.GetStatus().Checks[k].State == ErroredState {
				currentCheck = k
			}
			checks = append(checks, checkStates[r.Object.GetStatus().Checks[k].State])
		}

		if currentCheck != "" {
			return fmt.Sprintf("%s (%s)", strings.Join(checks, ""), currentCheck)
		}
		return strings.Join(checks, "")
	}()

	if !fn.MapEqual(r.Object.GetAnnotations(), m) {
		r.Object.SetAnnotations(m)
		if err := r.client.Update(r.Context(), r.Object); err != nil {
			r.internalLogger.Info("[reconcile:end] failed to update resource annotations")
		}
	}

	if !isReady {
		yellow := color.New(color.FgHiYellow, color.Bold).SprintFunc()
		r.internalLogger.Info(fmt.Sprintf(yellow("[reconcile:end] (took: %.2fs) complete"), tDiff))
		return
	}

	green := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	r.internalLogger.Info(fmt.Sprintf(green("[reconcile:end] (took: %.2fs) complete"), tDiff))
}

//func (r *Request[T]) LogPreCheck(checkName string) {
//	blue := color.New(color.FgBlue).SprintFunc()
//	r.timerMap[checkName] = time.Now()
//	if check, ok := r.Object.GetStatus().Checks[checkName]; ok {
//		r.internalLogger.Info(fmt.Sprintf(blue("[check:start]"), "check.name", checkName, "check.state", check.State))
//	}
//}
//
//func (r *Request[T]) LogPostCheck(checkName string) {
//	check, ok := r.Object.GetStatus().Checks[checkName]
//	if ok {
//		fg := color.New(color.FgHiGreen, color.Bold).SprintFunc()
//		args := []string{
//			"check.name", checkName,
//			"check.state", check.State.String(),
//			"check.time_taken", fmt.Sprintf("%.2fs", time.Since(r.timerMap[checkName]).Seconds()),
//		}
//		if check.State == FailedState || check.State == ErroredState {
//			fg = color.New(color.FgRed).SprintFunc()
//			args = append(args, "check.message", check.Message)
//		}
//		// FIXME: must be args...
//		r.internalLogger.Info(fg("check end"), args)
//	}
//}

//func (r *Request[T]) GetOwnedResources() []ResourceRef {
//	return r.resourceRefs
//}

//func (r *Request[T]) GetOwnedK8sResources() []client.Object {
//	kresources := make([]client.Object, len(r.resourceRefs))
//
//	for i := range r.resourceRefs {
//		kresources[i] = &unstructured.Unstructured{
//			Object: map[string]any{
//				"apiVersion": r.resourceRefs[i].APIVersion,
//				"kind":       r.resourceRefs[i].Kind,
//				"metadata": map[string]any{
//					"name":      r.resourceRefs[i].Name,
//					"namespace": r.resourceRefs[i].Namespace,
//				},
//			},
//		}
//	}
//
//	return kresources
//}

//func (r *Request[T]) AddToOwnedResources(refs ...ResourceRef) {
//	r.resourceRefs = append(r.resourceRefs, refs...)
//}

//func (r *Request[T]) CleanupOwnedResources(check *Check[T]) step_result.Result {
//	resources := r.Object.GetStatus().Resources
//	objects := make([]client.Object, 0, len(resources))
//	for i := range resources {
//		objects = append(objects, &unstructured.Unstructured{Object: map[string]any{
//			"apiVersion": resources[i].APIVersion,
//			"kind":       resources[i].Kind,
//			"metadata": map[string]any{
//				"name":      resources[i].Name,
//				"namespace": resources[i].Namespace,
//			},
//		}})
//	}
//
//	if err := fn.DeleteAndWait(r.Context(), r.Logger, r.client, objects...); err != nil {
//		return check.Errored(err).RequeueAfter(2 * time.Second)
//	}
//
//	return check.Passed()
//}

/*
INFO: this should only be used for very specific cases, where there is no other way to cleanup owned resources
Like, when deleting ManagedService
  - all managed resources should be deleted, but since owner is already getting deleted, there is no point in their proper cleanup
*/
//func (r *Request[T]) ForceCleanupOwnedResources(check *Check[T]) step_result.Result {
//	ctx := r.Context()
//	resources := r.Object.GetStatus().Resources
//
//	objects := make([]client.Object, 0, len(resources))
//
//	for i := range resources {
//		res := &unstructured.Unstructured{Object: map[string]any{
//			"apiVersion": resources[i].APIVersion,
//			"kind":       resources[i].Kind,
//			"metadata": map[string]any{
//				"name":      resources[i].Name,
//				"namespace": resources[i].Namespace,
//			},
//		}}
//		objects = append(objects, res)
//	}
//
//	if err := fn.ForceDelete(ctx, r.Logger, r.client, objects...); err != nil {
//		return check.Errored(err)
//	}
//
//	return check.Passed()
//}
//
//func ParseResourceRef(obj client.Object) ResourceRef {
//	return ResourceRef{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
//			APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
//		},
//		Namespace: obj.GetNamespace(),
//		Name:      obj.GetName(),
//	}
//}
