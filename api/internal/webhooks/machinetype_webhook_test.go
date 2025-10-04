package webhooks

import (
	"context"
	"testing"

	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMachineTypeWebhook_ValidateCreate_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

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

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.NoError(t, err, "Should validate successfully")
}

func TestMachineTypeWebhook_ValidateCreate_InvalidName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Standard_4", // Invalid: uppercase and underscore
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

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with invalid name")
	assert.Contains(t, err.Error(), "invalid machine type name")
}

func TestMachineTypeWebhook_ValidateCreate_MissingDisplayName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			Category: "general",
			Active:   true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with missing display name")
	assert.Contains(t, err.Error(), "displayName is required")
}

func TestMachineTypeWebhook_ValidateCreate_InvalidCategory(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "invalid-category",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with invalid category")
	assert.Contains(t, err.Error(), "invalid category")
}

func TestMachineTypeWebhook_ValidateCreate_MissingCPU(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with missing CPU")
	assert.Contains(t, err.Error(), "CPU is required")
}

func TestMachineTypeWebhook_ValidateCreate_MissingMemory(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU: "4",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with missing memory")
	assert.Contains(t, err.Error(), "memory is required")
}

func TestMachineTypeWebhook_ValidateCreate_InvalidCPUQuantity(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "invalid-cpu",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with invalid CPU quantity")
	assert.Contains(t, err.Error(), "invalid CPU quantity")
}

func TestMachineTypeWebhook_ValidateCreate_InvalidMemoryQuantity(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

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
				Memory: "invalid-memory",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with invalid memory quantity")
	assert.Contains(t, err.Error(), "invalid memory quantity")
}

func TestMachineTypeWebhook_ValidateCreate_InvalidGPUQuantity(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-machine",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "GPU Machine",
			Category:    "gpu",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "32Gi",
				GPU:    "invalid-gpu",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.Error(t, err, "Should fail with invalid GPU quantity")
	assert.Contains(t, err.Error(), "invalid GPU quantity")
}

func TestMachineTypeWebhook_ValidateCreate_WithGPU(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-machine",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "GPU Machine",
			Category:    "gpu",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "32Gi",
				GPU:    "1",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), machineType)
	assert.NoError(t, err, "Should validate successfully with GPU")
}

func TestMachineTypeWebhook_ValidateCreate_DuplicateDisplayName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	existingType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard Machine",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingType).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	newType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard Machine",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "16Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), newType)
	assert.Error(t, err, "Should fail with duplicate display name")
	assert.Contains(t, err.Error(), "already exists")
}

func TestMachineTypeWebhook_ValidateCreate_MultipleDefaults(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	existingDefault := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Default Type",
			Category:    "general",
			Active:      true,
			IsDefault:   true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingDefault).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	newDefault := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-default",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "New Default",
			Category:    "general",
			Active:      true,
			IsDefault:   true,
			Resources: machinesv1.MachineResources{
				CPU:    "8",
				Memory: "16Gi",
			},
		},
	}

	err := webhook.ValidateCreate(context.Background(), newDefault)
	assert.Error(t, err, "Should fail when another default exists")
	assert.Contains(t, err.Error(), "already marked as default")
}

func TestMachineTypeWebhook_ValidateUpdate_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	oldType := &machinesv1.MachineType{
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

	newType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU Updated",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateUpdate(context.Background(), oldType, newType)
	assert.NoError(t, err, "Should allow valid update")
}

func TestMachineTypeWebhook_ValidateDelete_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unused-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Unused Type",
			Category:    "general",
			Active:      false,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.ValidateDelete(context.Background(), machineType)
	assert.NoError(t, err, "Should allow deleting unused machine type")
}

func TestMachineTypeWebhook_ValidateDelete_InUse(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "in-use-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "In-Use Type",
			Category:    "general",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         "test-user",
			MachineType:     "in-use-type",
			TargetNamespace: "wm-test-user",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(machineType, workMachine).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	err := webhook.ValidateDelete(context.Background(), machineType)
	assert.Error(t, err, "Should not allow deleting machine type in use")
	assert.Contains(t, err.Error(), "work machines are using it")
}

func TestMachineTypeWebhook_Default_SetCategory(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.Default(context.Background(), machineType)
	assert.NoError(t, err)
	assert.Equal(t, "general", machineType.Spec.Category, "Should set default category")
}

func TestMachineTypeWebhook_Default_AddLabels(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "compute-optimized",
			Active:      true,
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.Default(context.Background(), machineType)
	assert.NoError(t, err)
	assert.NotNil(t, machineType.Labels)
	assert.Equal(t, "compute-optimized", machineType.Labels["kloudlite.io/machine-type-category"])
	assert.Equal(t, "true", machineType.Labels["kloudlite.io/machine-type-active"])
}

func TestMachineTypeWebhook_Default_SetPriority(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	webhook := NewMachineTypeWebhook(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard-4",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Standard 4 CPU",
			Category:    "general",
			Resources: machinesv1.MachineResources{
				CPU:    "4",
				Memory: "8Gi",
			},
		},
	}

	err := webhook.Default(context.Background(), machineType)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), machineType.Spec.Priority, "Should set default priority")
}

func TestIsValidMachineTypeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid lowercase", "standard-4", true},
		{"valid with numbers", "gpu-2x-large", true},
		{"valid single char", "a", true},
		{"invalid uppercase", "Standard-4", false},
		{"invalid underscore", "standard_4", false},
		{"invalid starts with hyphen", "-standard", false},
		{"invalid ends with hyphen", "standard-", false},
		{"valid no hyphens", "standard4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMachineTypeName(tt.input)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateResources(t *testing.T) {
	t.Run("valid resources", func(t *testing.T) {
		resources := &machinesv1.MachineResources{
			CPU:    "4",
			Memory: "8Gi",
		}
		err := validateResources(resources)
		assert.NoError(t, err)
	})

	t.Run("missing CPU", func(t *testing.T) {
		resources := &machinesv1.MachineResources{
			Memory: "8Gi",
		}
		err := validateResources(resources)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU is required")
	})

	t.Run("missing memory", func(t *testing.T) {
		resources := &machinesv1.MachineResources{
			CPU: "4",
		}
		err := validateResources(resources)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory is required")
	})

	t.Run("valid with GPU", func(t *testing.T) {
		resources := &machinesv1.MachineResources{
			CPU:    "8",
			Memory: "32Gi",
			GPU:    "2",
		}
		err := validateResources(resources)
		assert.NoError(t, err)
	})
}
