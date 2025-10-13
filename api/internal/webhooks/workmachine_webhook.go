package webhooks

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// WorkMachineWebhook handles validation and mutation for WorkMachine resources
type WorkMachineWebhook struct {
	k8sClient client.Client
}

// NewWorkMachineWebhook creates a new WorkMachine webhook
func NewWorkMachineWebhook(k8sClient client.Client) *WorkMachineWebhook {
	return &WorkMachineWebhook{
		k8sClient: k8sClient,
	}
}

// Default implements admission.Defaulter for mutation
func (w *WorkMachineWebhook) Default(ctx context.Context, obj runtime.Object) error {
	machine := obj.(*machinesv1.WorkMachine)

	// Generate name if not provided (one machine per user)
	if machine.Name == "" {
		owner := machine.Spec.OwnedBy
		// Create a deterministic name based on owner
		sanitizedOwner := strings.ReplaceAll(owner, "@", "-")
		sanitizedOwner = strings.ReplaceAll(sanitizedOwner, ".", "-")
		machine.Name = fmt.Sprintf("wm-%s", sanitizedOwner)
	}

	// Set default labels
	if machine.Labels == nil {
		machine.Labels = make(map[string]string)
	}

	// Find user by username or email to get the actual user ID
	var userID, userEmail string
	ownedBy := machine.Spec.OwnedBy

	if strings.Contains(ownedBy, "@") {
		// OwnedBy is an email, lookup user
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err != nil {
			return fmt.Errorf("failed to list users: %v", err)
		}

		for _, user := range userList.Items {
			if user.Spec.Email == ownedBy {
				userID = user.Name
				userEmail = user.Spec.Email
				break
			}
		}

		if userID == "" {
			return fmt.Errorf("user with email %s not found", ownedBy)
		}
	} else {
		// OwnedBy is a username, lookup user
		user := &platformv1alpha1.User{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: ownedBy}, user); err != nil {
			// Try to find by username in labels
			userList := &platformv1alpha1.UserList{}
			if err := w.k8sClient.List(ctx, userList); err != nil {
				return fmt.Errorf("failed to list users: %v", err)
			}

			for _, u := range userList.Items {
				if u.Name == ownedBy {
					userID = u.Name
					userEmail = u.Spec.Email
					break
				}
			}

			if userID == "" {
				return fmt.Errorf("user %s not found", ownedBy)
			}
		} else {
			userID = user.Name
			userEmail = user.Spec.Email
		}
	}

	// Add ownership labels
	machine.Labels["kloudlite.io/owned-by"] = userID
	// Use URL-safe base64 encoding without padding for labels
	encodedEmail := base64.RawURLEncoding.EncodeToString([]byte(userEmail))
	machine.Labels["kloudlite.io/owner-email"] = encodedEmail
	machine.Labels["kloudlite.io/machine-type"] = machine.Spec.MachineType

	return nil
}

// ValidateCreate implements admission.Validator for create operations
func (w *WorkMachineWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	machine := obj.(*machinesv1.WorkMachine)

	// Validate owner exists and has 'user' role
	ownedBy := machine.Spec.OwnedBy
	var foundUser *platformv1alpha1.User

	if strings.Contains(ownedBy, "@") {
		// Check by email
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err != nil {
			return fmt.Errorf("failed to list users: %v", err)
		}

		for _, user := range userList.Items {
			if user.Spec.Email == ownedBy {
				foundUser = &user
				break
			}
		}
	} else {
		// Check by username or ID
		user := &platformv1alpha1.User{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: ownedBy}, user); err == nil {
			foundUser = user
		} else {
			// Try to find by username
			userList := &platformv1alpha1.UserList{}
			if err := w.k8sClient.List(ctx, userList); err != nil {
				return fmt.Errorf("failed to list users: %v", err)
			}

			for _, u := range userList.Items {
				if u.Name == ownedBy {
					foundUser = &u
					break
				}
			}
		}
	}

	if foundUser == nil {
		return fmt.Errorf("owner %s not found", ownedBy)
	}

	// Check if user already has a machine
	machineList := &machinesv1.WorkMachineList{}
	if err := w.k8sClient.List(ctx, machineList); err != nil {
		return fmt.Errorf("failed to list machines: %v", err)
	}

	for _, existingMachine := range machineList.Items {
		if existingMachine.Spec.OwnedBy == ownedBy && existingMachine.Name != machine.Name {
			return fmt.Errorf("user %s already has a work machine: %s", ownedBy, existingMachine.Name)
		}
	}

	// Validate machine type exists and is active
	machineType := &machinesv1.MachineType{}
	if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: machine.Spec.MachineType}, machineType); err != nil {
		return fmt.Errorf("machine type %s not found", machine.Spec.MachineType)
	}

	if !machineType.Spec.Active {
		return fmt.Errorf("machine type %s is not active", machine.Spec.MachineType)
	}

	return nil
}

// ValidateUpdate implements admission.Validator for update operations
func (w *WorkMachineWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	oldMachine := oldObj.(*machinesv1.WorkMachine)
	newMachine := newObj.(*machinesv1.WorkMachine)

	// Prevent changing owner
	if oldMachine.Spec.OwnedBy != newMachine.Spec.OwnedBy {
		return fmt.Errorf("cannot change machine owner")
	}

	// If machine type changed, validate new type
	if oldMachine.Spec.MachineType != newMachine.Spec.MachineType {
		machineType := &machinesv1.MachineType{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: newMachine.Spec.MachineType}, machineType); err != nil {
			return fmt.Errorf("machine type %s not found", newMachine.Spec.MachineType)
		}

		if !machineType.Spec.Active {
			return fmt.Errorf("machine type %s is not active", newMachine.Spec.MachineType)
		}
	}

	return nil
}

// ValidateDelete implements admission.Validator for delete operations
func (w *WorkMachineWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	machine := obj.(*machinesv1.WorkMachine)

	// Check if machine is running
	if machine.Status.State == machinesv1.MachineStateRunning {
		return fmt.Errorf("cannot delete a running machine, please stop it first")
	}

	return nil
}

// InjectDecoder injects the decoder
func (w *WorkMachineWebhook) InjectDecoder(d *admission.Decoder) error {
	return nil
}
