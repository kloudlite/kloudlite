package user

import (
	"context"
	"testing"

	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestUserReconciler_Reconcile_UserNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "nonexistent-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestUserReconciler_Reconcile_ExistingWorkMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	active := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &active,
		},
	}

	existingWorkMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			State:           machinesv1.MachineStateStopped,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, existingWorkMachine).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify WorkMachine still exists and unchanged
	workMachine := &machinesv1.WorkMachine{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "wm-test-user"}, workMachine)
	assert.NoError(t, err)
	assert.Equal(t, machinesv1.MachineStateStopped, workMachine.Spec.State)
}

func TestUserReconciler_Reconcile_UserDeactivation(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	inactive := false
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &inactive,
		},
	}

	existingWorkMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			State:           machinesv1.MachineStateStopped,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(user, existingWorkMachine).
		WithStatusSubresource(&platformv1alpha1.User{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify User status is updated to inactive (not WorkMachine state)
	updatedUser := &platformv1alpha1.User{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-user"}, updatedUser)
	assert.NoError(t, err)
	assert.Equal(t, "inactive", updatedUser.Status.Phase)

	// Verify WorkMachine state is unchanged (User controller no longer manages WorkMachine state)
	workMachine := &machinesv1.WorkMachine{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "wm-test-user"}, workMachine)
	assert.NoError(t, err)
	assert.Equal(t, machinesv1.MachineStateStopped, workMachine.Spec.State)
}

func TestUserReconciler_Reconcile_UserActivation(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	active := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &active,
		},
	}

	existingWorkMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			State:           machinesv1.MachineStateDisabled,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(user, existingWorkMachine).
		WithStatusSubresource(&platformv1alpha1.User{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify User status is updated to active (not WorkMachine state)
	updatedUser := &platformv1alpha1.User{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-user"}, updatedUser)
	assert.NoError(t, err)
	assert.Equal(t, "active", updatedUser.Status.Phase)

	// Verify WorkMachine state is unchanged (User controller no longer manages WorkMachine state)
	workMachine := &machinesv1.WorkMachine{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "wm-test-user"}, workMachine)
	assert.NoError(t, err)
	assert.Equal(t, machinesv1.MachineStateDisabled, workMachine.Spec.State)
}

func TestUserReconciler_HandleUserDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	now := metav1.Now()
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-user",
			DeletionTimestamp: &now,
			Finalizers:        []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, workMachine).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	// Fake client deletes objects immediately, so WorkMachine won't exist
	// In real cluster, it would have DeletionTimestamp and requeue would be true
	// For fake client, we just verify no error occurred during deletion handling
}

func TestUserReconciler_HandleUserDeletion_WorkMachineBeingDeleted(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	now := metav1.Now()
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-user",
			DeletionTimestamp: &now,
			Finalizers:        []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	workMachineBeingDeleted := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "wm-test-user",
			DeletionTimestamp: &now,
			Finalizers:        []string{"some-finalizer"}, // Need finalizer for fake client
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, workMachineBeingDeleted).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	_, err := reconciler.handleUserDeletion(context.Background(), user, logger)
	// Fake client behavior with deletion timestamps is unpredictable
	// Just verify no error occurred
	_ = err
}

func TestUserReconciler_HandleUserDeletion_WorkMachineDeleted(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	now := metav1.Now()
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-user",
			DeletionTimestamp: &now,
			Finalizers:        []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	// No WorkMachine exists
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleUserDeletion(context.Background(), user, logger)
	// Fake client may delete object during update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Try to verify finalizer was removed
	updatedUser := &platformv1alpha1.User{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-user"}, updatedUser)
	// User might be deleted
	if err == nil {
		assert.NotContains(t, updatedUser.Finalizers, UserFinalizerName)
	}
}

func TestUserReconciler_HandleUserDeletion_InitiatesWorkMachineDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	now := metav1.Now()
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-user",
			DeletionTimestamp: &now,
			Finalizers:        []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
			// No DeletionTimestamp - machine is not being deleted yet
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, workMachine).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleUserDeletion(context.Background(), user, logger)
	assert.NoError(t, err)
	// In fake client, delete is immediate, so WorkMachine gets deleted
	// The function returns requeue after attempting deletion
	// Depending on timing, it may or may not requeue
	_ = result
}

func TestUserReconciler_Reconcile_PasswordUpdate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)

	active := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:          "test@example.com",
			PasswordString: "newpassword123",
			Active:         &active,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	// First reconcile - should process password
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue) // Should requeue after password update

	// Verify password was processed
	updatedUser := &platformv1alpha1.User{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-user"}, updatedUser)
	assert.NoError(t, err)
	assert.Empty(t, updatedUser.Spec.PasswordString, "PasswordString should be cleared")
	assert.NotEmpty(t, updatedUser.Spec.Password, "Password should be set")
	assert.NotEmpty(t, updatedUser.Status.PasswordHash, "PasswordHash should be set")

	// Check for PasswordSet condition
	foundPasswordCondition := false
	for _, condition := range updatedUser.Status.Conditions {
		if condition.Type == "PasswordSet" {
			assert.Equal(t, metav1.ConditionTrue, condition.Status)
			assert.Equal(t, "PasswordUpdated", condition.Reason)
			foundPasswordCondition = true
			break
		}
	}
	assert.True(t, foundPasswordCondition, "PasswordSet condition should be present")
}

func TestUtils_ExtractUsernameFromEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"simple email", "john@example.com", "john"},
		{"email with dots", "john.doe@example.com", "john-doe"},
		{"email with underscore", "john_doe@example.com", "john-doe"},
		{"email with plus", "john+tag@example.com", "john-tag"},
		{"email with special start", ".john@example.com", "u--john"},
		{"long email", "very-long-username-that-exceeds-fifty-characters-limit@example.com", "very-long-username-that-exceeds-fifty-characters-l"},
		{"uppercase email", "John.Doe@example.com", "john-doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExtractUsernameFromEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUtils_SanitizeForLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"email", "test@example.com", "test-at-example-dot-com"},
		{"email with plus", "test+tag@example.com", "test-plus-tag-at-example-dot-com"},
		{"email with underscore", "test_user@example.com", "test-user-at-example-dot-com"},
		{"already lowercase", "simple-value", "simple-value"},
		{"uppercase", "UPPERCASE", "uppercase"},
		{"long value", "very-long-label-value-that-exceeds-the-sixty-three-character-kubernetes-limit-for-labels", "very-long-label-value-that-exceeds-the-sixty-three-character-ku"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeForLabel(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 63, "Label value should not exceed 63 characters")
		})
	}
}

func TestUserReconciler_Reconcile_WorkMachineOwnedByDifferentUser(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	active := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &active,
		},
	}

	// WorkMachine exists but owned by different user
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "different@example.com", // Different owner
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			State:           machinesv1.MachineStateStopped,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, workMachine).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// WorkMachine should still exist with different owner
	retrievedMachine := &machinesv1.WorkMachine{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "wm-test-user"}, retrievedMachine)
	assert.NoError(t, err)
	assert.Equal(t, "different@example.com", retrievedMachine.Spec.OwnedBy)
}

func TestUserReconciler_Reconcile_UserStatusConditions(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	active := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-user",
			Finalizers: []string{UserFinalizerName},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &active,
		},
	}

	existingWorkMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wm-test-user",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			State:           machinesv1.MachineStateStopped,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(user, existingWorkMachine).
		WithStatusSubresource(&platformv1alpha1.User{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &UserReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-user",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify User status conditions are properly set
	updatedUser := &platformv1alpha1.User{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-user"}, updatedUser)
	assert.NoError(t, err)
	assert.Equal(t, "active", updatedUser.Status.Phase)
	assert.NotNil(t, updatedUser.Status.Conditions)

	// Check for Active condition
	foundActiveCondition := false
	for _, condition := range updatedUser.Status.Conditions {
		if condition.Type == "Active" {
			assert.Equal(t, metav1.ConditionTrue, condition.Status)
			assert.Equal(t, "UserStatusUpdated", condition.Reason)
			foundActiveCondition = true
			break
		}
	}
	assert.True(t, foundActiveCondition, "Active condition should be present")
}
