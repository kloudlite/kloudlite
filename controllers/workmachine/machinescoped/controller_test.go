package machinescoped

import (
	"context"
	"testing"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestShouldReconcileWorkMachineUsesNodeWorkMachineLabel(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-a",
				Labels: map[string]string{"kloudlite.io/workmachine": "wm-a"},
			},
		}).Build(),
		env: Env{NodeName: "node-a"},
	}

	if !r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-a"}}) {
		t.Fatal("expected controller to reconcile workmachine selected by node label")
	}
	if r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-b"}}) {
		t.Fatal("expected controller to skip workmachine not selected by node label")
	}
}

func TestLifecycleStepsDoNotCreateSeparateHostManagerWorkload(t *testing.T) {
	r := &MachineScopedReconciler{}

	for _, step := range r.lifecycleSteps() {
		if step.Name == "ensure-host-manager" {
			t.Fatal("expected integrated workmachine-manager to replace separate host-manager workload")
		}
	}
}

func TestIntegratedHostManagerStepMarksHostManagerReady(t *testing.T) {
	r := &MachineScopedReconciler{}
	var found bool
	for _, step := range r.lifecycleSteps() {
		if step.Condition != workmachineshared.ConditionHostManagerReady {
			continue
		}
		found = true
		result, err := step.OnCreate(context.Background(), workmachineshared.NewStatusSession(&workmachinev1.WorkMachine{}))
		if err != nil {
			t.Fatalf("integrated host manager step returned error: %v", err)
		}
		if result.Requeue || result.RequeueAfter != 0 {
			t.Fatalf("integrated host manager step result = %#v, want zero", result)
		}
	}
	if !found {
		t.Fatal("expected lifecycle to retain HostManagerReady condition for integrated runtime")
	}
}

func TestEnsureIntegratedHostManagerDeletesLegacyHostManagerWorkload(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := testWorkMachine("wm", "wm-test")
	ownerRef := fn.AsOwner(wm, true)
	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "host-manager", Namespace: "wm-test", OwnerReferences: []metav1.OwnerReference{ownerRef}}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "host-manager", Namespace: "wm-test", OwnerReferences: []metav1.OwnerReference{ownerRef}}},
		).Build(),
	}

	result, err := r.ensureIntegratedHostManager(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("ensureIntegratedHostManager returned error: %v", err)
	}
	if result.Requeue || result.RequeueAfter != 0 {
		t.Fatalf("ensureIntegratedHostManager result = %#v, want zero", result)
	}

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Namespace: "wm-test", Name: "host-manager"}, statefulSet); !apiErrors.IsNotFound(err) {
		t.Fatalf("statefulset get error = %v, want not found", err)
	}
	svc := &corev1.Service{}
	if err := r.Get(context.Background(), client.ObjectKey{Namespace: "wm-test", Name: "host-manager"}, svc); !apiErrors.IsNotFound(err) {
		t.Fatalf("service get error = %v, want not found", err)
	}
}

func TestEnsureWorkmachineIngressControllerTargetsIntegratedManager(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := testWorkMachine("wm", "wm-test")
	ownerRef := fn.AsOwner(wm, true)
	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "wm-ingress-controller", Namespace: "wm-test", OwnerReferences: []metav1.OwnerReference{ownerRef}}},
		).Build(),
		Scheme: scheme,
	}

	result, err := r.ensureWorkmachineIngressController(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("ensureWorkmachineIngressController returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("ensureWorkmachineIngressController result = %#v, want no requeue", result)
	}

	service := &corev1.Service{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: "wm-ingress-controller", Namespace: "wm-test"}, service); err != nil {
		t.Fatalf("expected wm-ingress-controller service: %v", err)
	}
	wantedSelector := map[string]string{"app": "workmachine-manager", "kloudlite.io/workmachine": "wm"}
	if len(service.Spec.Selector) != len(wantedSelector) {
		t.Fatalf("service selector = %#v, want %#v", service.Spec.Selector, wantedSelector)
	}
	for key, want := range wantedSelector {
		if got := service.Spec.Selector[key]; got != want {
			t.Fatalf("service selector[%s] = %q, want %q; full selector=%#v", key, got, want, service.Spec.Selector)
		}
	}
	assertServicePort(t, service, "http", 80)
	assertServicePort(t, service, "https", 443)

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: "wm-ingress-controller", Namespace: "wm-test"}, statefulSet); !apiErrors.IsNotFound(err) {
		t.Fatalf("legacy wm-ingress-controller statefulset get error = %v, want not found", err)
	}
}

func TestShouldReconcileWorkMachineKeepsExplicitWorkMachineNameOverride(t *testing.T) {
	r := &MachineScopedReconciler{env: Env{WorkMachineName: "wm-a", NodeName: "node-a"}}

	if !r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-a"}}) {
		t.Fatal("expected explicit workmachine name to match")
	}
	if r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-b"}}) {
		t.Fatal("expected explicit workmachine name to filter other workmachines")
	}
}

func TestCreateHostManagerRBACUsesTargetNamespaceForManagerServiceAccount(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
	}
	wm := testWorkMachine("wm", "wm-test")

	if _, err := r.createHostManagerRBAC(context.Background(), workmachineshared.NewStatusSession(wm)); err != nil {
		t.Fatalf("createHostManagerRBAC returned error: %v", err)
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "hm-wm"}, clusterRoleBinding); err != nil {
		t.Fatalf("get cluster role binding: %v", err)
	}
	assertSingleSubject(t, clusterRoleBinding.Subjects, "wm-test", "workmachine-manager-wm")

	roleBinding := &rbacv1.RoleBinding{}
	if err := r.Get(context.Background(), client.ObjectKey{Namespace: "wm-test", Name: "host-manager"}, roleBinding); err != nil {
		t.Fatalf("get role binding: %v", err)
	}
	assertSingleSubject(t, roleBinding.Subjects, "wm-test", "workmachine-manager-wm")
}

func TestCreateHostManagerRBACPreservesUnownedLegacyHostManagerServiceAccount(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "host-manager", Namespace: "wm-test"}}).Build(),
	}
	wm := testWorkMachine("wm", "wm-test")

	if _, err := r.createHostManagerRBAC(context.Background(), workmachineshared.NewStatusSession(wm)); err != nil {
		t.Fatalf("createHostManagerRBAC returned error: %v", err)
	}

	serviceAccount := &corev1.ServiceAccount{}
	if err := r.Get(context.Background(), client.ObjectKey{Namespace: "wm-test", Name: "host-manager"}, serviceAccount); err != nil {
		t.Fatalf("get service account: %v", err)
	}
}

func TestDeleteNamespaceRemovesChildFinalizersBeforeDelete(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := environmentv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := packagesv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workspacev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := testWorkMachine("wm", "wm-test")
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "workspace-a",
			Finalizers: []string{"workspaces.kloudlite.io/finalizer"},
		},
		Spec: workspacev1.WorkspaceSpec{WorkmachineName: "wm"},
	}
	environment := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "environment-a",
			Finalizers: []string{"environments.kloudlite.io/finalizer"},
		},
		Spec: environmentv1.EnvironmentSpec{WorkMachineName: "wm", TargetNamespace: "env-a"},
	}
	packageRequest := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "package-a",
			Namespace:  "wm-test",
			Finalizers: []string{"workspaces.kloudlite.io/package-cleanup"},
		},
	}

	clientBuilder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-test"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-a"}},
		workspace,
		environment,
		packageRequest,
	)
	clientBuilder.WithInterceptorFuncs(interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			switch o := obj.(type) {
			case *workspacev1.Workspace:
				if controllerHasFinalizer(o, "workspaces.kloudlite.io/finalizer") {
					t.Fatalf("workspace delete called before finalizer removal: %#v", o.Finalizers)
				}
			case *environmentv1.Environment:
				if controllerHasFinalizer(o, "environments.kloudlite.io/finalizer") {
					t.Fatalf("environment delete called before finalizer removal: %#v", o.Finalizers)
				}
			case *packagesv1.PackageRequest:
				if controllerHasFinalizer(o, "workspaces.kloudlite.io/package-cleanup") {
					t.Fatalf("package request delete called before finalizer removal: %#v", o.Finalizers)
				}
			}
			return c.Delete(ctx, obj, opts...)
		},
	})

	r := &MachineScopedReconciler{Client: clientBuilder.Build(), Scheme: scheme}
	if err := r.initSharedRuntimeState(); err != nil {
		t.Fatalf("init shared runtime state: %v", err)
	}

	if _, err := r.deleteNamespace(context.Background(), workmachineshared.NewStatusSession(wm)); err != nil {
		t.Fatalf("deleteNamespace returned error: %v", err)
	}
}

func TestDeleteNamespaceLeavesWorkMachineNamespaceForPlatformCleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := environmentv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := packagesv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workspacev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := testWorkMachine("wm", "wm-test")
	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-test"}}).Build(),
		Scheme: scheme,
	}

	result, err := r.deleteNamespace(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("deleteNamespace returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("deleteNamespace result = %#v, want no requeue after child cleanup", result)
	}

	namespace := &corev1.Namespace{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm-test"}, namespace); err != nil {
		t.Fatalf("expected workmachine namespace to remain for platform cleanup: %v", err)
	}
	if namespace.DeletionTimestamp != nil {
		t.Fatalf("expected machine scoped cleanup not to delete workmachine namespace, deletionTimestamp=%v", namespace.DeletionTimestamp)
	}
}

func testWorkMachine(name, targetNamespace string) *workmachinev1.WorkMachine {
	return &workmachinev1.WorkMachine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: workmachinev1.GroupVersion.String(),
			Kind:       "WorkMachine",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: "wm-uid"},
		Spec: workmachinev1.WorkMachineSpec{
			TargetNamespace: targetNamespace,
		},
	}
}

func assertSingleSubject(t *testing.T, subjects []rbacv1.Subject, namespace, name string) {
	t.Helper()
	if len(subjects) != 1 {
		t.Fatalf("subjects = %#v, want exactly one subject", subjects)
	}
	if subjects[0].Kind != "ServiceAccount" || subjects[0].Namespace != namespace || subjects[0].Name != name {
		t.Fatalf("subject = %#v, want ServiceAccount/%s/%s", subjects[0], namespace, name)
	}
}

func assertServicePort(t *testing.T, service *corev1.Service, name string, port int32) {
	t.Helper()
	for _, servicePort := range service.Spec.Ports {
		if servicePort.Name == name && servicePort.Port == port {
			return
		}
	}
	t.Fatalf("expected service port %s=%d, got %#v", name, port, service.Spec.Ports)
}

func controllerHasFinalizer(obj client.Object, finalizer string) bool {
	for _, existing := range obj.GetFinalizers() {
		if existing == finalizer {
			return true
		}
	}
	return false
}
