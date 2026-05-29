package platformscoped

import (
	"context"
	"errors"
	"testing"
	"time"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/reconciler"
	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestSkipCloudPermissionValidation(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")

	if !skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to be skipped")
	}
}

func TestSkipCloudPermissionValidationDefaultsToFalse(t *testing.T) {
	if skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to run by default")
	}
}

func TestSetupCloudProviderRejectsUnsupportedProvider(t *testing.T) {
	_, err := setupCloudProvider(context.Background(), Env{CloudProvider: workmachinev1.CloudProvider("unknown")})
	if err == nil {
		t.Fatal("expected unsupported cloud provider error")
	}
}

func TestValidateProviderPermissionsHonorsSkipToggle(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")
	provider := &recordingProvider{validateErr: errors.New("should be skipped")}

	if err := validateProviderPermissions(context.Background(), provider); err != nil {
		t.Fatalf("expected validation to be skipped, got %v", err)
	}
	if provider.validateCalls != 0 {
		t.Fatalf("expected no validation calls, got %d", provider.validateCalls)
	}
}

func TestPlatformScopedLifecycleStepNamesPreserveOrder(t *testing.T) {
	r := &PlatformScopedReconciler{}
	steps := r.lifecycleSteps()

	got := make([]string, 0, len(steps))
	for _, step := range steps {
		got = append(got, step.Name)
	}

	want := []string{
		"handle-machine-type-change",
		"handle-node-reboot-request",
		"cleanup-workmachine-namespace",
		"setup-cloud-machine",
		"ensure-workmachine-manager",
		"wait-machine-scoped-cleanup",
	}
	wantConditions := []string{
		workmachineshared.ConditionCloudMachineProvisioned,
		workmachineshared.ConditionCloudMachineRunning,
		"",
		workmachineshared.ConditionNodeJoined,
		"",
		workmachineshared.ConditionMachineScopedCleanupComplete,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d steps, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: expected %q, got %q", i, want[i], got[i])
		}
		if steps[i].Condition != wantConditions[i] {
			t.Fatalf("step %d condition: expected %q, got %q", i, wantConditions[i], steps[i].Condition)
		}
	}
}

func TestEnsureWorkMachineManagerCreatesPerWorkMachineStatefulSet(t *testing.T) {
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

	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-test", OwnedBy: "karthik"},
	}
	wm.Status.MachineID = "machine-id"
	session := workmachineshared.NewStatusSession(wm)
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(sourceAPIConfig(), sourceAPISecret()).Build(),
		Scheme: scheme,
		env: Env{
			KloudliteInstallationID:  "installation-id",
			InstallationSecret:       "installation-secret",
			JWTSecret:                "jwt-secret",
			HostedSubdomain:          "apps.test",
			TunnelServerImage:        "tunnel-server:test",
			CodeAnalyzerImage:        "code-analyzer:test",
			PodNamespace:             "kloudlite",
			WorkMachineManagerImage:  "workmachine-manager:test",
			SnapshotRegistryEndpoint: "registry.test:5000",
			SnapshotRegistryPrefix:   "wm-snapshots",
			SnapshotRegistryInsecure: "false",
		},
	}

	result, err := r.ensureWorkMachineManager(context.Background(), session)
	if err != nil {
		t.Fatalf("ensureWorkMachineManager returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("ensureWorkMachineManager result = %#v, want no requeue", result)
	}

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: workMachineManagerName("wm"), Namespace: "wm-test"}, statefulSet); err != nil {
		t.Fatalf("expected per-workmachine statefulset: %v", err)
	}
	serviceAccount := &corev1.ServiceAccount{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: workMachineManagerName("wm"), Namespace: "wm-test"}, serviceAccount); err != nil {
		t.Fatalf("expected per-workmachine service account: %v", err)
	}
	configMap := &corev1.ConfigMap{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "api-server-config", Namespace: "wm-test"}, configMap); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected no api-server-config copy in manager namespace, got err=%v", err)
	}
	secret := &corev1.Secret{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "api-server-secret", Namespace: "wm-test"}, secret); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected no api-server-secret copy in manager namespace, got err=%v", err)
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: workMachineManagerName("wm")}, clusterRoleBinding); err != nil {
		t.Fatalf("expected per-workmachine cluster role binding: %v", err)
	}
	if len(clusterRoleBinding.Subjects) != 1 || clusterRoleBinding.Subjects[0].Namespace != "wm-test" {
		t.Fatalf("cluster role binding subjects = %#v, want service account subject in wm-test", clusterRoleBinding.Subjects)
	}
	if statefulSet.Spec.Replicas == nil || *statefulSet.Spec.Replicas != 1 {
		t.Fatalf("replicas = %#v, want 1", statefulSet.Spec.Replicas)
	}
	if statefulSet.Spec.Template.Spec.NodeSelector["kloudlite.io/workmachine"] != "wm" {
		t.Fatalf("nodeSelector = %#v, want kloudlite.io/workmachine=wm", statefulSet.Spec.Template.Spec.NodeSelector)
	}
	container := statefulSet.Spec.Template.Spec.Containers[0]
	if container.Name != "workmachine-manager" {
		t.Fatalf("container name = %q, want workmachine-manager", container.Name)
	}
	if len(container.Command) != 1 || container.Command[0] != "/app/workmachine-manager" {
		t.Fatalf("container command = %#v, want /app/workmachine-manager", container.Command)
	}
	if len(container.Args) != 2 || container.Args[0] != "server" || container.Args[1] != "workmachine-manager" {
		t.Fatalf("container args = %#v, want server workmachine-manager", container.Args)
	}
	if len(container.EnvFrom) != 0 {
		t.Fatalf("expected manager container not to use EnvFrom, got %#v", container.EnvFrom)
	}
	if container.Env[0].Name != "NODE_NAME" {
		t.Fatalf("expected manager to expose NODE_NAME env, got %#v", statefulSet.Spec.Template.Spec.Containers[0].Env)
	}
	if !statefulSet.Spec.Template.Spec.HostPID {
		t.Fatal("expected integrated workmachine-manager pod to run with hostPID enabled")
	}
	if container.SecurityContext == nil || container.SecurityContext.Privileged == nil || !*container.SecurityContext.Privileged {
		t.Fatalf("expected integrated workmachine-manager container to be privileged, got %#v", container.SecurityContext)
	}
	if !hasEnv(container.Env, "WORKMACHINE_NAME", "wm") {
		t.Fatalf("expected WORKMACHINE_NAME env to be set to wm, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "INSTALLATION_KEY", "installation-id") {
		t.Fatalf("expected INSTALLATION_KEY env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "INSTALLATION_SECRET", "installation-secret") {
		t.Fatalf("expected INSTALLATION_SECRET env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "JWT_SECRET", "jwt-secret") {
		t.Fatalf("expected JWT_SECRET env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "HOSTED_SUBDOMAIN", "apps.test") {
		t.Fatalf("expected HOSTED_SUBDOMAIN env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "TUNNEL_SERVER_IMAGE", "tunnel-server:test") {
		t.Fatalf("expected TUNNEL_SERVER_IMAGE env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "CODE_ANALYZER_IMAGE", "code-analyzer:test") {
		t.Fatalf("expected CODE_ANALYZER_IMAGE env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "SNAPSHOT_REGISTRY_ENDPOINT", "registry.test:5000") {
		t.Fatalf("expected SNAPSHOT_REGISTRY_ENDPOINT env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "SNAPSHOT_REGISTRY_PREFIX", "wm-snapshots") {
		t.Fatalf("expected SNAPSHOT_REGISTRY_PREFIX env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "SNAPSHOT_REGISTRY_INSECURE", "false") {
		t.Fatalf("expected SNAPSHOT_REGISTRY_INSECURE env from Env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "REGISTRY_USERNAME", "karthik") {
		t.Fatalf("expected REGISTRY_USERNAME env from WorkMachine owner, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "WM_INGRESS_HTTP_PORT", "80") {
		t.Fatalf("expected WM_INGRESS_HTTP_PORT env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "WM_INGRESS_HTTPS_PORT", "443") {
		t.Fatalf("expected WM_INGRESS_HTTPS_PORT env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "WM_INGRESS_WILDCARD_SECRET_NAME", "kloudlite-wildcard-cert-tls") {
		t.Fatalf("expected WM_INGRESS_WILDCARD_SECRET_NAME env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "WM_INGRESS_WILDCARD_SECRET_NAMESPACE", "wm-test") {
		t.Fatalf("expected WM_INGRESS_WILDCARD_SECRET_NAMESPACE env, got %#v", container.Env)
	}
	if !hasEnv(container.Env, "WM_INGRESS_OWN_NAMESPACE", "wm-test") {
		t.Fatalf("expected WM_INGRESS_OWN_NAMESPACE env, got %#v", container.Env)
	}
	if !hasContainerPort(container.Ports, "http", 80) {
		t.Fatalf("expected integrated ingress http port, got %#v", container.Ports)
	}
	if !hasContainerPort(container.Ports, "https", 443) {
		t.Fatalf("expected integrated ingress https port, got %#v", container.Ports)
	}
	if !hasContainerPort(container.Ports, "ingress-health", 17777) {
		t.Fatalf("expected integrated ingress health port, got %#v", container.Ports)
	}
	if !hasVolumeMount(container.VolumeMounts, "kloudlite-data", "/var/lib/kloudlite") {
		t.Fatalf("expected /var/lib/kloudlite host mount, got %#v", container.VolumeMounts)
	}
	if !hasHostPathVolume(statefulSet.Spec.Template.Spec.Volumes, "host-proc", "/proc") {
		t.Fatalf("expected /proc host path volume, got %#v", statefulSet.Spec.Template.Spec.Volumes)
	}

	deployment := &appsv1.Deployment{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: machineScopedControllerName("wm"), Namespace: "kloudlite"}, deployment); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected no per-workmachine deployment, got err=%v", err)
	}
}

func TestEnsureWorkMachineManagerDeletesLegacyPlatformNamespaceResources(t *testing.T) {
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

	managerName := workMachineManagerName("wm")
	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-test"},
	}
	wm.Status.MachineID = "machine-id"
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "kloudlite"}},
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "kloudlite"}},
		).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite", WorkMachineManagerImage: "workmachine-manager:test"},
	}

	result, err := r.ensureWorkMachineManager(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("ensureWorkMachineManager returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("ensureWorkMachineManager result = %#v, want no requeue", result)
	}

	legacyStatefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "kloudlite"}, legacyStatefulSet); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected legacy platform namespace statefulset to be deleted, got err=%v", err)
	}
	legacyServiceAccount := &corev1.ServiceAccount{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "kloudlite"}, legacyServiceAccount); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected legacy platform namespace service account to be deleted, got err=%v", err)
	}
}

func sourceAPIConfig() *corev1.ConfigMap {
	immutable := true
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "api-server-config", Namespace: "kloudlite"},
		Data:       map[string]string{"API_HOST": "https://api.test"},
		BinaryData: map[string][]byte{"ca.crt": []byte("ca")},
		Immutable:  &immutable,
	}
}

func sourceAPISecret() *corev1.Secret {
	immutable := true
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-server-secret", Namespace: "kloudlite"},
		Type:       corev1.SecretTypeOpaque,
		Data:       map[string][]byte{"token": []byte("secret-token")},
		Immutable:  &immutable,
	}
}

func hasEnv(envs []corev1.EnvVar, name, value string) bool {
	for _, env := range envs {
		if env.Name == name && env.Value == value {
			return true
		}
	}
	return false
}

func hasContainerPort(ports []corev1.ContainerPort, name string, port int32) bool {
	for _, containerPort := range ports {
		if containerPort.Name == name && containerPort.ContainerPort == port {
			return true
		}
	}
	return false
}

func hasVolumeMount(mounts []corev1.VolumeMount, name, mountPath string) bool {
	for _, mount := range mounts {
		if mount.Name == name && mount.MountPath == mountPath {
			return true
		}
	}
	return false
}

func hasHostPathVolume(volumes []corev1.Volume, name, path string) bool {
	for _, volume := range volumes {
		if volume.Name == name && volume.HostPath != nil && volume.HostPath.Path == path {
			return true
		}
	}
	return false
}

func TestCleanupWorkMachineNamespaceDeletesNamespaceAndRequeues(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-wm"},
	}
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-wm"}}).Build(),
		Scheme: scheme,
	}

	result, err := r.cleanupWorkMachineNamespace(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("cleanupWorkMachineNamespace returned error: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatalf("cleanupWorkMachineNamespace result = %#v, want requeue while namespace deletes", result)
	}

	ns := &corev1.Namespace{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm-wm"}, ns); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected namespace deletion to be requested, got err=%v", err)
	}
}

func TestCleanupWorkMachineNamespaceReturnsWhenNamespaceGone(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-wm"},
	}
	r := &PlatformScopedReconciler{Client: fake.NewClientBuilder().WithScheme(scheme).Build(), Scheme: scheme}

	result, err := r.cleanupWorkMachineNamespace(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("cleanupWorkMachineNamespace returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("cleanupWorkMachineNamespace result = %#v, want no requeue", result)
	}
}

func TestReconcilePlatformDeletionWaitsBeforeDeletingManagerWhenMachineCleanupMissing(t *testing.T) {
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

	wm := deletingWorkMachine("wm", "wm-test")
	managerName := workMachineManagerName("wm")
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-test"}},
			&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "wm-test"}},
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "wm-test"}},
			&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: managerName}},
		).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite"},
	}

	result, err := r.reconcilePlatform(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("reconcilePlatform returned error: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatalf("reconcilePlatform result = %#v, want requeue while waiting for machine cleanup", result)
	}

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "wm-test"}, statefulSet); err != nil {
		t.Fatalf("expected manager statefulset to remain while cleanup is missing: %v", err)
	}
	serviceAccount := &corev1.ServiceAccount{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "wm-test"}, serviceAccount); err != nil {
		t.Fatalf("expected manager service account to remain while cleanup is missing: %v", err)
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName}, clusterRoleBinding); err != nil {
		t.Fatalf("expected manager cluster role binding to remain while cleanup is missing: %v", err)
	}
}

func TestReconcilePlatformDeletionContinuesWhenCleanupConditionMissingButManagerNamespaceGone(t *testing.T) {
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

	wm := deletingWorkMachine("wm", "wm-test")
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite"},
	}

	result, err := r.reconcilePlatform(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("reconcilePlatform returned error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("reconcilePlatform result = %#v, want no wait when manager namespace is already gone", result)
	}
}

func TestReconcilePlatformDeletionContinuesWhenCleanupConditionMissingButManagerGone(t *testing.T) {
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

	wm := deletingWorkMachine("wm", "wm-test")
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-test"}}).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite"},
	}

	result, err := r.reconcilePlatform(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("reconcilePlatform returned error: %v", err)
	}
	if result.RequeueAfter != 5*time.Second {
		t.Fatalf("reconcilePlatform result = %#v, want namespace deletion requeue when manager is gone", result)
	}
	namespace := &corev1.Namespace{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm-test"}, namespace); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected platform to delete namespace after manager is gone, got err=%v", err)
	}
}

func TestReconcilePlatformDeletionDeletesManagerAfterMachineCleanupComplete(t *testing.T) {
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

	wm := deletingWorkMachine("wm", "wm-test")
	wm.Status.Conditions = []metav1.Condition{{
		Type:               workmachineshared.ConditionMachineScopedCleanupComplete,
		Status:             metav1.ConditionTrue,
		Reason:             workmachineshared.ReasonReconciled,
		ObservedGeneration: wm.Generation,
	}}
	managerName := workMachineManagerName("wm")
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "wm-test"}},
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: managerName, Namespace: "wm-test"}},
			&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: managerName}},
		).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite"},
	}

	result, err := r.reconcilePlatform(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("reconcilePlatform returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("reconcilePlatform result = %#v, want no requeue after machine cleanup", result)
	}

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "wm-test"}, statefulSet); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected manager statefulset deletion after cleanup, got err=%v", err)
	}
	serviceAccount := &corev1.ServiceAccount{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName, Namespace: "wm-test"}, serviceAccount); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected manager service account deletion after cleanup, got err=%v", err)
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: managerName}, clusterRoleBinding); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected manager cluster role binding deletion after cleanup, got err=%v", err)
	}
}

func TestReconcileReturnPreservesReconcileErrorOverPatchConflictRequeue(t *testing.T) {
	reconcileErr := errors.New("real reconcile failure")
	conflictRequeue := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileReturn(ctrl.Result{}, reconcileErr, conflictRequeue, nil)

	if !errors.Is(err, reconcileErr) {
		t.Fatalf("error = %v, want reconcile error %v", err, reconcileErr)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero result when preserving reconcile error", result)
	}
}

func TestReconcileReturnUsesPatchConflictRequeueWithoutReconcileError(t *testing.T) {
	conflictRequeue := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileReturn(ctrl.Result{}, nil, conflictRequeue, nil)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != conflictRequeue {
		t.Fatalf("result = %#v, want patch conflict requeue %#v", result, conflictRequeue)
	}
}

func TestReconcileReturnsRequeueWithoutErrorOnFinalizerUpdateConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	wm := &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	conflict := k8sErrors.NewConflict(schema.GroupResource{Group: workmachinev1.GroupVersion.Group, Resource: "workmachines"}, "wm", errors.New("stale object"))
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(wm).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return conflict
			},
		}).Build(),
		Scheme: scheme,
	}

	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKey{Name: "wm"}})
	if err != nil {
		t.Fatalf("Reconcile error = %v, want nil for conflict requeue", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatalf("Reconcile result = %#v, want conflict requeue", result)
	}
}

func TestUpdateNodeIPLabelsRetriesConflictWithoutError(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	updateCalls := 0
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "wm", Labels: map[string]string{}}}).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if _, ok := obj.(*corev1.Node); ok {
					updateCalls++
					if updateCalls == 1 {
						return k8sErrors.NewConflict(schema.GroupResource{Resource: "nodes"}, obj.GetName(), errors.New("node changed"))
					}
				}
				return c.Update(ctx, obj, opts...)
			},
		}).Build(),
		Scheme: scheme,
	}

	r.updateNodeIPLabels(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "wm", Labels: map[string]string{}}}, true, &workmachinev1.MachineInfo{PublicIP: "1.2.3.4", PrivateIP: "10.0.0.5"})

	if updateCalls < 2 {
		t.Fatalf("updateCalls = %d, want retry after conflict", updateCalls)
	}
	node := &corev1.Node{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm"}, node); err != nil {
		t.Fatalf("get node: %v", err)
	}
	if node.Labels[NodeLabelPublicIP] != "1.2.3.4" || node.Labels[NodeLabelPrivateIP] != "10.0.0.5" {
		t.Fatalf("node labels = %#v, want updated IP labels", node.Labels)
	}
}

func TestClearNodeIPLabelsRetriesConflictWithoutError(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	updateCalls := 0
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "wm", Labels: map[string]string{NodeLabelPublicIP: "1.2.3.4", NodeLabelPrivateIP: "10.0.0.5"}}}).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if _, ok := obj.(*corev1.Node); ok {
					updateCalls++
					if updateCalls == 1 {
						return k8sErrors.NewConflict(schema.GroupResource{Resource: "nodes"}, obj.GetName(), errors.New("node changed"))
					}
				}
				return c.Update(ctx, obj, opts...)
			},
		}).Build(),
		Scheme: scheme,
	}

	r.clearNodeIPLabels(context.Background(), &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "wm", Labels: map[string]string{NodeLabelPublicIP: "1.2.3.4", NodeLabelPrivateIP: "10.0.0.5"}}})

	if updateCalls < 2 {
		t.Fatalf("updateCalls = %d, want retry after conflict", updateCalls)
	}
	node := &corev1.Node{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm"}, node); err != nil {
		t.Fatalf("get node: %v", err)
	}
	if node.Labels[NodeLabelPublicIP] != "" || node.Labels[NodeLabelPrivateIP] != "" {
		t.Fatalf("node labels = %#v, want IP labels removed", node.Labels)
	}
}

type recordingProvider struct {
	validateCalls int
	validateErr   error
}

func (p *recordingProvider) ValidatePermissions(context.Context) error {
	p.validateCalls++
	return p.validateErr
}

func (p *recordingProvider) CreateMachine(context.Context, *workmachinev1.WorkMachine) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) GetMachineStatus(context.Context, string) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) StartMachine(context.Context, string) error  { return nil }
func (p *recordingProvider) StopMachine(context.Context, string) error   { return nil }
func (p *recordingProvider) RebootMachine(context.Context, string) error { return nil }
func (p *recordingProvider) IncreaseVolumeSize(context.Context, string, int32) error {
	return nil
}
func (p *recordingProvider) ChangeMachine(context.Context, string, string) error { return nil }
func (p *recordingProvider) DeleteMachine(context.Context, string) error         { return nil }

func deletingWorkMachine(name, targetNamespace string) *workmachinev1.WorkMachine {
	deletionTime := metav1.Now()
	return &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Generation:        1,
			Finalizers:        []string{reconciler.Finalizer},
			DeletionTimestamp: &deletionTime,
		},
		Spec: workmachinev1.WorkMachineSpec{TargetNamespace: targetNamespace},
	}
}
