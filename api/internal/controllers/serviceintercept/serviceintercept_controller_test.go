package serviceintercept

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
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
			Status: "active",
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

func TestServiceInterceptReconciler_Reconcile_Activation_CreatesEndpoints(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
			PodIP: "10.42.0.100",
		},
	}

	workspaceService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "workspace-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "ssh",
					Port:       22,
					TargetPort: intstr.FromInt(22),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
			},
		},
	}

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
			Status: "active",
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

	k8sClient := testutil.NewFakeClient(scheme, workspace, workspaceService, service, intercept).
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

	// Verify workspace service was updated with intercept port
	updatedWorkspaceService := &corev1.Service{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "workspace-test-workspace",
		Namespace: "workspace-namespace",
	}, updatedWorkspaceService)
	assert.NoError(t, err)

	// Check that intercept port was added
	foundPort := false
	for _, port := range updatedWorkspaceService.Spec.Ports {
		if port.Port == 3000 {
			foundPort = true
			assert.Equal(t, intstr.FromInt(3000), port.TargetPort)
			assert.Equal(t, corev1.ProtocolTCP, port.Protocol)
			break
		}
	}
	assert.True(t, foundPort, "Workspace service should have intercept port 3000")

	// Verify service selector was cleared
	updatedService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.Empty(t, updatedService.Spec.Selector, "Service selector should be cleared")

	// Verify manual Endpoints was created
	endpoints := &corev1.Endpoints{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, endpoints)
	assert.NoError(t, err)
	assert.Len(t, endpoints.Subsets, 1)
	assert.Len(t, endpoints.Subsets[0].Addresses, 1)
	assert.Equal(t, "10.42.0.100", endpoints.Subsets[0].Addresses[0].IP)
	assert.Len(t, endpoints.Subsets[0].Ports, 1)
	assert.Equal(t, int32(3000), endpoints.Subsets[0].Ports[0].Port)

	// Verify status was updated
	updatedIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedIntercept)
	assert.NoError(t, err)
	assert.Equal(t, "Active", updatedIntercept.Status.Phase)
	assert.NotNil(t, updatedIntercept.Status.OriginalServiceSelector)
	assert.Equal(t, "test-app", updatedIntercept.Status.OriginalServiceSelector["app"])
}

func TestServiceInterceptReconciler_Reconcile_Deletion_RestoresService(t *testing.T) {
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

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"intercepts.kloudlite.io/intercepted-by": "test-intercept",
			},
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
			Selector: nil, // Cleared by intercept
		},
	}

	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: "10.42.0.100"},
				},
				Ports: []corev1.EndpointPort{
					{
						Name:     "http",
						Port:     3000,
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
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
			Status: "active",
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
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, service, endpoints, intercept).
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

	// Verify Endpoints was deleted
	deletedEndpoints := &corev1.Endpoints{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, deletedEndpoints)
	assert.True(t, apierrors.IsNotFound(err), "Endpoints should be deleted")

	// Verify service selector was restored
	updatedService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.NotNil(t, updatedService.Spec.Selector)
	assert.Equal(t, "test-app", updatedService.Spec.Selector["app"])

	// Verify annotation was removed
	_, exists := updatedService.Annotations["intercepts.kloudlite.io/intercepted-by"]
	assert.False(t, exists, "Intercept annotation should be removed")

	// Note: Finalizer won't be removed immediately as deletion waits for replacement pods
	// This is the expected behavior - finalizer stays until pods are ready
}

func TestServiceInterceptReconciler_Reconcile_InactiveStatus_ClearsEndpoints(t *testing.T) {
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
			Selector: nil, // Cleared by intercept
		},
	}

	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: "10.42.0.100"},
				},
				Ports: []corev1.EndpointPort{
					{
						Name:     "http",
						Port:     3000,
						Protocol: corev1.ProtocolTCP,
					},
				},
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
			Status: "inactive", // Changed to inactive
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
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, service, endpoints, intercept).
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

	// Reconcile with inactive status
	for i := 0; i < 3; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify Endpoints was deleted
	deletedEndpoints := &corev1.Endpoints{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, deletedEndpoints)
	assert.True(t, apierrors.IsNotFound(err), "Endpoints should be deleted when inactive")

	// Verify service selector was restored
	updatedService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.NotNil(t, updatedService.Spec.Selector)
	assert.Equal(t, "test-app", updatedService.Spec.Selector["app"])

	// Verify status was updated to Inactive
	updatedIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedIntercept)
	assert.NoError(t, err)
	assert.Equal(t, "Inactive", updatedIntercept.Status.Phase)
}

func TestServiceInterceptReconciler_Reconcile_WorkspaceNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "test-app"},
		},
	}

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-intercept",
			Namespace:  "test-namespace",
			Finalizers: []string{serviceInterceptFinalizer},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "nonexistent-workspace",
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

	k8sClient := testutil.NewFakeClient(scheme, service, intercept).
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

	// Reconcile should return error for missing workspace
	result, err := reconciler.Reconcile(context.Background(), req)
	// Error is expected when workspace is not found
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	// Controller will requeue due to error
	assert.True(t, result.Requeue || result.RequeueAfter > 0 || err != nil)
}

func TestServiceInterceptReconciler_Reconcile_ServiceNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

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

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-intercept",
			Namespace:  "test-namespace",
			Finalizers: []string{serviceInterceptFinalizer},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "nonexistent-service",
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

	k8sClient := testutil.NewFakeClient(scheme, workspace, intercept).
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

	// Reconcile should return error for missing service
	result, err := reconciler.Reconcile(context.Background(), req)
	// Error is expected when service is not found
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	// Controller will requeue due to error
	assert.True(t, result.Requeue || result.RequeueAfter > 0 || err != nil)
}

func TestServiceInterceptReconciler_Reconcile_MultiplePortMappings(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
			PodIP: "10.42.0.100",
		},
	}

	workspaceService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "workspace-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "ssh",
					Port:       22,
					TargetPort: intstr.FromInt(22),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
			},
		},
	}

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
				{
					Name:       "grpc",
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
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
			Status: "active",
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
				{
					ServicePort:   9090,
					WorkspacePort: 3001,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workspaceService, service, intercept).
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

	// Verify Endpoints has all ports
	endpoints := &corev1.Endpoints{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, endpoints)
	assert.NoError(t, err)
	assert.Len(t, endpoints.Subsets, 1)
	assert.Len(t, endpoints.Subsets[0].Addresses, 1)
	assert.Equal(t, "10.42.0.100", endpoints.Subsets[0].Addresses[0].IP)
	assert.Len(t, endpoints.Subsets[0].Ports, 2)

	// Verify both ports are present
	portMap := make(map[int32]bool)
	for _, port := range endpoints.Subsets[0].Ports {
		portMap[port.Port] = true
	}
	assert.True(t, portMap[3000], "Port 3000 should be in endpoints")
	assert.True(t, portMap[3001], "Port 3001 should be in endpoints")
}

func TestServiceInterceptReconciler_SkipReInterceptionDuringDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Service that was previously intercepted
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"intercepts.kloudlite.io/intercepted-by": "test-intercept",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "test-app"},
		},
	}

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

	now := metav1.Now()
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-intercept",
			Namespace:         "test-namespace",
			Finalizers:        []string{serviceInterceptFinalizer},
			DeletionTimestamp: &now, // Being deleted
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
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
			Phase: "Active",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, service, workspace, intercept).
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

	// Reconcile deletion
	_, _ = reconciler.Reconcile(context.Background(), req)

	// Verify service selector was restored
	updatedService := &corev1.Service{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-service",
		Namespace: "test-namespace",
	}, updatedService)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedService.Spec.Selector)

	// Verify annotation was removed
	_, exists := updatedService.Annotations["intercepts.kloudlite.io/intercepted-by"]
	assert.False(t, exists)
}
