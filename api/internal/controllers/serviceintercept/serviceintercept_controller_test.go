package serviceintercept

import (
	"context"
	"strings"
	"testing"

	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestServiceInterceptReconciler_Reconcile_NotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).
		WithStatusSubresource(&interceptsv1.ServiceIntercept{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &ServiceInterceptReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestServiceInterceptReconciler_Reconcile_AddFinalizer(t *testing.T) {
	scheme := testutil.NewTestScheme()

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, intercept).
		WithStatusSubresource(&interceptsv1.ServiceIntercept{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &ServiceInterceptReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
	}

	// First reconcile should add finalizer
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue || result.RequeueAfter > 0)

	// Verify finalizer was added
	updatedIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedIntercept)
	assert.NoError(t, err)
	assert.Contains(t, updatedIntercept.Finalizers, serviceInterceptFinalizer)
}

func TestServiceInterceptReconciler_Reconcile_Activation_CreatesSOCATPod(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Workspace pod that will receive intercepted traffic
	workspacePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "workspace-namespace",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "workspace-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.42.0.100",
		},
	}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
			PodIP: "10.42.0.100",
		},
	}

	// Headless service created by workspace controller
	headlessService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace-headless",
			Namespace: "workspace-namespace",
			Labels: map[string]string{
				"app":                               "workspace",
				"workspace":                         "test-workspace",
				"workspaces.kloudlite.io/workspace": "test-workspace",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "None",
			Selector: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "port-3000",
					Protocol:   corev1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
				{
					Name:       "port-3001",
					Protocol:   corev1.ProtocolTCP,
					Port:       3001,
					TargetPort: intstr.FromInt(3001),
				},
			},
		},
	}

	// Service to be intercepted
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test-app",
			},
		},
	}

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-intercept",
			Namespace:  "test-namespace",
			Finalizers: []string{serviceInterceptFinalizer},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workspacePod, headlessService, service, intercept).
		WithStatusSubresource(&interceptsv1.ServiceIntercept{}, &workspacesv1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &ServiceInterceptReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
	}

	// Multiple reconciles may be needed
	for i := 0; i < 5; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify workspace headless service exists (created by workspace controller)
	headlessSvc := &corev1.Service{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "workspace-test-workspace-headless",
		Namespace: "workspace-namespace",
	}, headlessSvc)
	assert.NoError(t, err)
	assert.Equal(t, "None", headlessSvc.Spec.ClusterIP, "Should be a headless service")

	// Verify SOCAT pod was created
	socatPod := &corev1.Pod{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service-intercept-test-workspace",
		Namespace: "test-namespace",
	}, socatPod)
	assert.NoError(t, err)

	// Verify SOCAT pod has original service selector labels
	assert.Equal(t, "test-app", socatPod.Labels["app"])

	// Verify SOCAT pod has intercept tracking labels
	assert.Equal(t, "true", socatPod.Labels["intercepts.kloudlite.io/managed"])
	assert.Equal(t, "test-service", socatPod.Labels["intercepts.kloudlite.io/service"])
	assert.Equal(t, "test-workspace", socatPod.Labels["intercepts.kloudlite.io/workspace"])
	assert.Equal(t, "test-intercept", socatPod.Labels["intercepts.kloudlite.io/intercept"])

	// Verify SOCAT pod container configuration
	assert.Len(t, socatPod.Spec.Containers, 1)
	assert.Equal(t, "socat-forwarder", socatPod.Spec.Containers[0].Name)
	assert.Equal(t, "alpine/socat:latest", socatPod.Spec.Containers[0].Image)

	// Verify SOCAT command contains correct port forwarding
	command := strings.Join(socatPod.Spec.Containers[0].Command, " ")
	assert.Contains(t, command, "socat TCP-LISTEN:80")
	assert.Contains(t, command, "workspace-test-workspace-headless.workspace-namespace.svc.cluster.local:3000")

	// Verify original service remains unchanged
	updatedService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedService.Spec.Selector, "Service selector should remain unchanged")
	assert.Equal(t, "test-app", updatedService.Spec.Selector["app"])

	// Verify status was updated
	updatedIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedIntercept)
	assert.NoError(t, err)
	assert.Equal(t, "Active", updatedIntercept.Status.Phase)
	assert.NotEmpty(t, updatedIntercept.Status.SOCATPodName)
	assert.NotNil(t, updatedIntercept.Status.OriginalServiceSelector)
	assert.Equal(t, "test-app", updatedIntercept.Status.OriginalServiceSelector["app"])
}

func TestServiceInterceptReconciler_Reconcile_Deletion_CleansUpSOCATPod(t *testing.T) {
	scheme := testutil.NewTestScheme()

	originalSelector := map[string]string{"app": "test-app"}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-namespace",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
			PodIP: "10.42.0.100",
		},
	}

	// Headless service created by workspace controller (remains unchanged)
	headlessService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace-headless",
			Namespace: "workspace-namespace",
			Labels: map[string]string{
				"app":                               "workspace",
				"workspace":                         "test-workspace",
				"workspaces.kloudlite.io/workspace": "test-workspace",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "None",
			Selector: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "port-3000",
					Protocol:   corev1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
			},
		},
	}

	// Original service (unchanged by intercept)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: originalSelector, // Still has original selector
		},
	}

	// SOCAT pod created by intercept
	socatPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service-intercept-test-workspace",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app":                               "test-app",
				"intercepts.kloudlite.io/managed":   "true",
				"intercepts.kloudlite.io/service":   "test-service",
				"intercepts.kloudlite.io/workspace": "test-workspace",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "socat-forwarder",
					Image:   "alpine/socat:latest",
					Command: []string{"sh", "-c", "socat TCP-LISTEN:80,fork,reuseaddr TCP:workspace-test-workspace-headless.workspace-namespace.svc.cluster.local:3000"},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	now := metav1.Now()
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-intercept",
			Namespace:         "test-namespace",
			Finalizers:        []string{serviceInterceptFinalizer},
			DeletionTimestamp: &now, // Marked for deletion
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:                   "Active",
			OriginalServiceSelector: originalSelector,
			AffectedPodNames:        []string{"test-pod-1"},
			SOCATPodName:            "test-service-intercept-test-workspace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, service, socatPod, headlessService, intercept).
		WithStatusSubresource(&interceptsv1.ServiceIntercept{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &ServiceInterceptReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
	}

	// Reconcile the deletion
	_, _ = reconciler.Reconcile(context.Background(), req)

	// Verify SOCAT pod was deleted
	deletedSOCATPod := &corev1.Pod{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service-intercept-test-workspace",
		Namespace: "test-namespace",
	}, deletedSOCATPod)
	assert.True(t, apierrors.IsNotFound(err), "SOCAT pod should be deleted")

	// Verify workspace headless service still exists (managed by workspace controller)
	existingHeadlessService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "workspace-test-workspace-headless",
		Namespace: "workspace-namespace",
	}, existingHeadlessService)
	assert.NoError(t, err, "Headless service should still exist")

	// Verify original service remains unchanged
	updatedService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.NotNil(t, updatedService.Spec.Selector)
	assert.Equal(t, "test-app", updatedService.Spec.Selector["app"], "Service selector should remain unchanged")
}
