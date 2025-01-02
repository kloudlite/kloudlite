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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/logging"

	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
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
	anchorName     string
	Logger         *slog.Logger
	internalLogger *slog.Logger
	KV             KV

	reconStartTime time.Time
	timerMap       map[string]time.Time

	resourceRefs []ResourceRef
}

func NewRequest[T Resource](ctx context.Context, c client.Client, nn types.NamespacedName, resource T) (*Request[T], error) {
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

	logger := log.FromContext(ctx, "NN", nn.String())

	return &Request[T]{
		ctx:            ctx,
		client:         c,
		Object:         resource,
		Logger:         logging.New(logger),
		internalLogger: logging.New(logger, logging.WithCallDepth(3)),
		anchorName:     anchorName,
		KV:             KV{},
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
	return r.Object.GetAnnotations()[AnnotationShouldReconcileKey] != "false"
}

func (r *Request[T]) EnsureCheckList(expected []CheckMeta) stepResult.Result {
	checkNames := make([]string, len(expected))
	for i := range expected {
		checkNames[i] = expected[i].Name
	}

	if r.Object.GetStatus().Checks == nil {
		r.Object.GetStatus().Checks = make(map[string]Check)
	}

	checksUpdated := ensureChecks(r.Object.GetStatus().Checks, checkNames...)
	if checksUpdated || !slices.Equal(expected, r.Object.GetStatus().CheckList) {
		r.Object.GetStatus().CheckList = expected

		if err := r.client.Status().Update(r.ctx, r.Object); err != nil {
			return stepResult.New().Err(err)
		}
		return stepResult.New().RequeueAfter(1 * time.Second)
	}
	return stepResult.New().Continue(true)
}

func ensureChecks(checks map[string]Check, checkNames ...string) bool {
	updated := false
	for _, name := range checkNames {
		if _, ok := checks[name]; !ok {
			updated = true
			checks[name] = Check{State: WaitingState}
		}
	}
	return updated
}

func (r *Request[T]) EnsureChecks(names ...string) stepResult.Result {
	return stepResult.New().Continue(true)
	// obj, ctx := r.Object, r.Context()
	//
	// checks := fn.MapMerge(obj.GetStatus().Checks)
	// updated := ensureChecks(checks, names...)
	//
	// if updated {
	// 	obj.GetStatus().Checks = checks
	// 	if err := r.client.Status().Update(ctx, obj); err != nil {
	// 		return r.Done().Err(err)
	// 	}
	// }
	// return stepResult.New().Continue(true)
}

func (r *Request[T]) ClearStatusIfAnnotated() stepResult.Result {
	obj := r.Object
	ann := obj.GetAnnotations()

	if v, ok := ann[AnnotationResetCheckKey]; ok {
		if _, ok2 := obj.GetStatus().Checks[v]; ok2 {
			delete(ann, AnnotationResetCheckKey)
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

	if v := ann[AnnotationClearStatusKey]; v == "true" {
		delete(ann, AnnotationClearStatusKey)
		obj.SetAnnotations(ann)
		if err := r.client.Update(r.Context(), obj); err != nil {
			return r.Done().Err(err)
		}

		obj.GetStatus().IsReady = false
		obj.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}
		obj.GetStatus().Checks = nil

		if err := r.client.Status().Update(context.TODO(), obj); err != nil {
			return r.Done().Err(err)
		}
		return r.Done().RequeueAfter(1 * time.Second)
	}
	return r.Next()
}

func (r *Request[T]) RestartIfAnnotated() stepResult.Result {
	ctx, obj := r.Context(), r.Object
	ann := obj.GetAnnotations()
	if v := ann[AnnotationRestartKey]; v == "true" {
		delete(ann, AnnotationRestartKey)
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

func (r *Request[T]) Finalize() stepResult.Result {
	controllerutil.RemoveFinalizer(r.Object, CommonFinalizer)
	controllerutil.RemoveFinalizer(r.Object, ForegroundFinalizer)
	controllerutil.RemoveFinalizer(r.Object, "finalizers.kloudlite.io")
	controllerutil.RemoveFinalizer(r.Object, "finalizers.kloudlite.io/watch") // Keep till all clusters are updated to v1.0.4
	return stepResult.New().Err(r.client.Update(r.ctx, r.Object))
}

func (r *Request[T]) PreReconcile() {
	blue := color.New(color.FgBlue).SprintFunc()
	r.reconStartTime = time.Now()
	r.internalLogger.Info(blue("[reconcile:start] start"))
}

var checkStates = map[State]string{
	WaitingState:   "🔶",
	RunningState:   "⌛",
	ErroredState:   "🔴",
	CompletedState: "🌿",
}

func (r *Request[T]) PostReconcile() {
	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}

	tDiff := time.Since(r.reconStartTime).Seconds()

	isReady := r.Object.GetStatus().IsReady

	r.Object.GetStatus().LastReconcileTime = &metav1.Time{Time: time.Now()}

	if r.Object.GetDeletionTimestamp() == nil {
		r.Object.GetStatus().Resources = r.GetOwnedResources()
	}

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

func (r *Request[T]) LogPreCheck(checkName string) {
	blue := color.New(color.FgBlue).SprintFunc()
	r.timerMap[checkName] = time.Now()
	check, ok := r.Object.GetStatus().Checks[checkName]
	if ok {
		r.internalLogger.Info(fmt.Sprintf(blue("[check:start] %-20s [status] %-5v"), checkName, check.Status))
	}
}

func (r *Request[T]) LogPostCheck(checkName string) {
	tDiff := time.Since(r.timerMap[checkName]).Seconds()
	check, ok := r.Object.GetStatus().Checks[checkName]
	if ok {
		if !check.Status {
			red := color.New(color.FgRed).SprintFunc()
			r.internalLogger.Info(fmt.Sprintf(red("[check:end] (took: %.2fs) %-20s [status] %v [message] %v"), tDiff, checkName, check.Status, check.Message))
			return
		}
		green := color.New(color.FgHiGreen, color.Bold).SprintFunc()
		r.internalLogger.Info(fmt.Sprintf(green("[check:end] (took: %.2fs) %-20s [status] %v"), tDiff, checkName, check.Status))
	}
}

func (r *Request[T]) GetOwnedResources() []ResourceRef {
	return r.resourceRefs
}

func (r *Request[T]) GetOwnedK8sResources() []client.Object {
	kresources := make([]client.Object, len(r.resourceRefs))

	for i := range r.resourceRefs {
		kresources[i] = &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": r.resourceRefs[i].APIVersion,
				"kind":       r.resourceRefs[i].Kind,
				"metadata": map[string]any{
					"name":      r.resourceRefs[i].Name,
					"namespace": r.resourceRefs[i].Namespace,
				},
			},
		}
	}

	return kresources
}

func (r *Request[T]) AddToOwnedResources(refs ...ResourceRef) {
	r.resourceRefs = append(r.resourceRefs, refs...)
}

func (r *Request[T]) CleanupOwnedResources(check *checkWrapper[T]) stepResult.Result {
	resources := r.Object.GetStatus().Resources
	objects := make([]client.Object, 0, len(resources))
	for i := range resources {
		objects = append(objects, &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": resources[i].APIVersion,
			"kind":       resources[i].Kind,
			"metadata": map[string]any{
				"name":      resources[i].Name,
				"namespace": resources[i].Namespace,
			},
		}})
	}

	if err := fn.DeleteAndWait(r.Context(), r.Logger, r.client, objects...); err != nil {
		return check.Failed(err).RequeueAfter(2 * time.Second)
	}

	return check.Completed()
}

/*
INFO: this should only be used for very specific cases, where there is no other way to cleanup owned resources
Like, when deleting ManagedService
  - all managed resources should be deleted, but since owner is already getting deleted, there is no point in their proper cleanup
*/
func (r *Request[T]) ForceCleanupOwnedResources(check *checkWrapper[T]) stepResult.Result {
	ctx := r.Context()
	resources := r.Object.GetStatus().Resources

	objects := make([]client.Object, 0, len(resources))

	for i := range resources {
		res := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": resources[i].APIVersion,
			"kind":       resources[i].Kind,
			"metadata": map[string]any{
				"name":      resources[i].Name,
				"namespace": resources[i].Namespace,
			},
		}}
		objects = append(objects, res)
	}

	if err := fn.ForceDelete(ctx, r.Logger, r.client, objects...); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
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
