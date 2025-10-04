package webhooks

import (
	"context"
	"testing"

	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/pkg/apis/platform/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWorkmachineWebhook_ValidateCreate_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	// Create test user and machine type
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, machineType).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
			DesiredState:    machinesv1.MachineStateStopped,
		},
	}

	err := webhook.ValidateCreate(context.Background(), machine)
	assert.NoError(t, err, "Should validate successfully")
}

func TestWorkmachineWebhook_ValidateCreate_OwnerNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "nonexistent-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateCreate(context.Background(), machine)
	assert.Error(t, err, "Should fail when owner not found")
	assert.Contains(t, err.Error(), "owner nonexistent-user not found")
}

func TestWorkmachineWebhook_ValidateCreate_OwnerByEmail(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			Active: true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, machineType).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateCreate(context.Background(), machine)
	assert.NoError(t, err, "Should validate successfully with email as owner")
}

func TestWorkmachineWebhook_ValidateCreate_MachineTypeNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "nonexistent-type",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateCreate(context.Background(), machine)
	assert.Error(t, err, "Should fail when machine type not found")
	assert.Contains(t, err.Error(), "machine type nonexistent-type not found")
}

func TestWorkmachineWebhook_ValidateCreate_MachineTypeInactive(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deprecated-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Deprecated Type",
			Active:      false,
			Resources: machinesv1.MachineResources{
				CPU:    "2",
				Memory: "4Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, machineType).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "deprecated-type",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateCreate(context.Background(), machine)
	assert.Error(t, err, "Should fail when machine type is inactive")
	assert.Contains(t, err.Error(), "machine type deprecated-type is not active")
}

func TestWorkmachineWebhook_ValidateCreate_UserAlreadyHasMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			Active: true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	existingMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user, machineType, existingMachine).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	newMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateCreate(context.Background(), newMachine)
	assert.Error(t, err, "Should fail when user already has a machine")
	assert.Contains(t, err.Error(), "already has a work machine")
}

func TestWorkmachineWebhook_ValidateUpdate_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-8",
		},
		Spec: machinesv1.MachineTypeSpec{
			Active: true,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "16Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(machineType).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	oldMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	newMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-8",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateUpdate(context.Background(), oldMachine, newMachine)
	assert.NoError(t, err, "Should allow machine type change")
}

func TestWorkmachineWebhook_ValidateUpdate_OwnerChangeNotAllowed(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	oldMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "user1",
			MachineType:     "standard-4",
			TargetNamespace: "wm-user1",
		},
	}

	newMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "user2",
			MachineType:     "standard-4",
			TargetNamespace: "wm-user2",
		},
	}

	err := webhook.ValidateUpdate(context.Background(), oldMachine, newMachine)
	assert.Error(t, err, "Should not allow owner change")
	assert.Contains(t, err.Error(), "cannot change machine owner")
}

func TestWorkmachineWebhook_ValidateUpdate_InactiveMachineType(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	inactiveMachineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "inactive-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			Active: false,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "16Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(inactiveMachineType).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	oldMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	newMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "inactive-type",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.ValidateUpdate(context.Background(), oldMachine, newMachine)
	assert.Error(t, err, "Should not allow switching to inactive machine type")
	assert.Contains(t, err.Error(), "is not active")
}

func TestWorkmachineWebhook_ValidateDelete_RunningMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	runningMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
		Status: machinesv1.WorkMachineStatus{
			State: machinesv1.MachineStateRunning,
		},
	}

	err := webhook.ValidateDelete(context.Background(), runningMachine)
	assert.Error(t, err, "Should not allow deleting running machine")
	assert.Contains(t, err.Error(), "cannot delete a running machine")
}

func TestWorkmachineWebhook_ValidateDelete_StoppedMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	stoppedMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
		Status: machinesv1.WorkMachineStatus{
			State: machinesv1.MachineStateStopped,
		},
	}

	err := webhook.ValidateDelete(context.Background(), stoppedMachine)
	assert.NoError(t, err, "Should allow deleting stopped machine")
}

func TestWorkmachineWebhook_Default_GenerateName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.Default(context.Background(), machine)
	assert.NoError(t, err)
	assert.NotEmpty(t, machine.Name, "Should generate name")
	assert.Contains(t, machine.Name, "wm-", "Generated name should have wm- prefix")
}

func TestWorkmachineWebhook_Default_AddLabels(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.Default(context.Background(), machine)
	assert.NoError(t, err)
	assert.NotNil(t, machine.Labels)
	assert.Equal(t, "test-user", machine.Labels["kloudlite.io/owned-by"])
	assert.NotEmpty(t, machine.Labels["kloudlite.io/owner-email"])
	assert.Equal(t, "standard-4", machine.Labels["kloudlite.io/machine-type"])
}

func TestWorkmachineWebhook_Default_EmailAsOwner(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.Default(context.Background(), machine)
	assert.NoError(t, err)
	assert.NotNil(t, machine.Labels)
	assert.Equal(t, "test-user", machine.Labels["kloudlite.io/owned-by"])
}

func TestWorkmachineWebhook_Default_UserNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewWorkMachineWebhook(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "nonexistent@example.com",
			MachineType:     "standard-4",
			TargetNamespace: "wm-test-user",
		},
	}

	err := webhook.Default(context.Background(), machine)
	assert.Error(t, err, "Should fail when user not found during mutation")
	assert.Contains(t, err.Error(), "not found")
}
