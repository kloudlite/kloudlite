package controllers

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"golang.org/x/crypto/ssh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkMachineReconciler reconciles a WorkMachine object
type WorkMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"

// SSH Configuration Constants
const (
	// SSHUserUID is the UID for the SSH server user
	SSHUserUID = "1000"
	// SSHUserGID is the GID for the SSH server user
	SSHUserGID = "1000"
	// SSHUserName is the username for the SSH server
	SSHUserName = "kloudlite"
)

// SSH Jump Host Architecture
//
// This controller implements a secure SSH jump host (bastion) pattern for accessing workspaces:
//
// Authentication Flow:
// 1. User authenticates to jump host using their public key (from WorkMachine.Spec.SSHPublicKeys)
// 2. Jump host forwards TCP connection to workspace
// 3. Workspace authenticates jump host using SSH proxy key pair
//
// Jump Host (workmachine-host-manager):
// - Runs OpenSSH server on port 2222
// - Authorizes users via ssh-authorized-keys ConfigMap (user keys only)
// - Has TCP forwarding enabled (AllowTcpForwarding yes)
// - Does NOT authenticate to workspaces (jump hosts work by TCP forwarding)
// - Password authentication disabled for security
//
// Workspaces:
// - Run OpenSSH servers that authorize the jump host's SSH proxy public key
// - Jump host uses SSH proxy private key to authenticate to workspaces
//
// Security:
// - All SSH keys are validated using ssh.ParseAuthorizedKey() before use
// - Password authentication is disabled (PasswordAuthentication no)
// - Only public key authentication is allowed
// - Jump host runs as non-privileged user (UID/GID 1000)
//
// ConfigMaps:
// - ssh-authorized-keys: Contains user public keys for jump host authentication
// - sshd-config: Contains OpenSSH server configuration with security hardening

// Reconcile handles WorkMachine CR reconciliation
func (r *WorkMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "workmachine", "workmachine", req.Name)
	logger.Info("Reconciling WorkMachine")

	// Fetch the WorkMachine instance
	workMachine := &machinesv1.WorkMachine{}
	if err := r.Get(ctx, req.NamespacedName, workMachine); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("WorkMachine not found, probably deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get WorkMachine")
		return ctrl.Result{}, err
	}

	// Check if WorkMachine is being deleted
	if workMachine.DeletionTimestamp != nil {
		logger.Info("WorkMachine is being deleted, starting cleanup")
		return r.handleWorkMachineDeletion(ctx, workMachine, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(workMachine, WorkMachineFinalizerName) {
		logger.Info("Adding finalizer to WorkMachine")
		controllerutil.AddFinalizer(workMachine, WorkMachineFinalizerName)
		if err := r.Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Ensure target namespace exists
	if err := r.ensureNamespace(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure namespace exists")
		return ctrl.Result{}, err
	}

	// Ensure RBAC resources exist for workmachine-node-manager
	if err := r.ensurePackageManagerRBAC(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure workmachine-node-manager RBAC")
		return ctrl.Result{}, err
	}

	// Ensure SSH proxy secret exists
	if err := r.ensureSSHProxySecret(ctx, workMachine, logger); err != nil {
		logger.Error(err, "Failed to ensure SSH proxy secret")
		return ctrl.Result{}, err
	}

	// Ensure SSH host keys secret exists
	if err := r.ensureSSHHostKeysSecret(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure SSH host keys secret")
		return ctrl.Result{}, err
	}

	// Ensure SSH authorized_keys ConfigMap exists
	if err := r.ensureSSHAuthorizedKeysConfigMap(ctx, workMachine, logger); err != nil {
		logger.Error(err, "Failed to ensure SSH authorized_keys ConfigMap")
		return ctrl.Result{}, err
	}

	// Ensure sshd_config ConfigMap exists
	if err := r.ensureSSHDConfigMap(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure sshd_config ConfigMap")
		return ctrl.Result{}, err
	}

	// Ensure workspace-sshd-config ConfigMap exists
	if err := r.ensureWorkspaceSSHDConfigMap(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure workspace-sshd-config ConfigMap")
		return ctrl.Result{}, err
	}

	// Check if namespace is being deleted (but WorkMachine is not)
	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workMachine.Spec.TargetNamespace}, namespace); err == nil {
		if namespace.DeletionTimestamp != nil {
			logger.Info("Namespace is being deleted, but WorkMachine is not - recreating finalizer protection")
			// This shouldn't happen normally because namespace has our finalizer
			// But if it does, we requeue and the deletion will be blocked by the finalizer
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Initialize status if it doesn't exist
	if workMachine.Status.State == "" {
		// First time - set current state to desired state
		workMachine.Status.State = workMachine.Spec.DesiredState
		now := metav1.Now()
		if workMachine.Spec.DesiredState == machinesv1.MachineStateRunning {
			workMachine.Status.StartedAt = &now
		}

		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to initialize WorkMachine status")
			return ctrl.Result{}, err
		}
		logger.Info("Initialized WorkMachine status", "state", workMachine.Status.State)
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	// Check if state transition is needed
	currentState := workMachine.Status.State
	desiredState := workMachine.Spec.DesiredState

	// Ensure package manager deployment exists (regardless of machine state)
	if err := r.ensurePackageManagerDeployment(ctx, workMachine, logger); err != nil {
		logger.Error(err, "Failed to ensure workmachine-node-manager deployment")
		return ctrl.Result{}, err
	}

	if currentState == desiredState {
		// No transition needed, machine is in desired state
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Handle state transitions
	logger.Info("State transition detected", "current", currentState, "desired", desiredState)

	switch currentState {
	case machinesv1.MachineStateStopped:
		if desiredState == machinesv1.MachineStateRunning {
			// Transition to starting
			workMachine.Status.State = machinesv1.MachineStateStarting
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to starting")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to starting")
			// Requeue quickly to move to running
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStarting:
		// Transition to running
		workMachine.Status.State = machinesv1.MachineStateRunning
		now := metav1.Now()
		workMachine.Status.StartedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to running")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now running")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

	case machinesv1.MachineStateRunning:
		if desiredState == machinesv1.MachineStateStopped {
			// Transition to stopping
			workMachine.Status.State = machinesv1.MachineStateStopping
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to stopping")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to stopping")
			// Requeue quickly to move to stopped
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStopping:
		// Transition to stopped
		workMachine.Status.State = machinesv1.MachineStateStopped
		now := metav1.Now()
		workMachine.Status.StoppedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to stopped")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now stopped")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleWorkMachineDeletion handles cleanup when WorkMachine is being deleted
func (r *WorkMachineReconciler) handleWorkMachineDeletion(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) (ctrl.Result, error) {
	namespaceName := workMachine.Spec.TargetNamespace

	// Check for active Workspaces in the target namespace
	workspaceList := &workspacesv1.WorkspaceList{}
	if err := r.List(ctx, workspaceList, client.InNamespace(namespaceName)); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Failed to list Workspaces", "namespace", namespaceName)
			return ctrl.Result{}, err
		}
	}

	// Block deletion if there are active workspaces
	if len(workspaceList.Items) > 0 {
		logger.Info("Deletion blocked: active Workspaces exist in namespace",
			"namespace", namespaceName,
			"workspaceCount", len(workspaceList.Items))

		// Update status with DeletionBlocked condition
		now := metav1.Now()
		workspaceNames := make([]string, len(workspaceList.Items))
		for i, ws := range workspaceList.Items {
			workspaceNames[i] = ws.Name
		}

		message := fmt.Sprintf("Cannot delete WorkMachine: %d active workspace(s) exist: %v", len(workspaceList.Items), workspaceNames)

		// Check if condition already exists
		conditionExists := false
		for i, condition := range workMachine.Status.Conditions {
			if condition.Type == machinesv1.WorkMachineConditionDeletionBlocked {
				workMachine.Status.Conditions[i].Status = metav1.ConditionTrue
				workMachine.Status.Conditions[i].Reason = "ActiveWorkspacesExist"
				workMachine.Status.Conditions[i].Message = message
				workMachine.Status.Conditions[i].LastTransitionTime = &now
				conditionExists = true
				break
			}
		}

		if !conditionExists {
			workMachine.Status.Conditions = append(workMachine.Status.Conditions, machinesv1.WorkMachineCondition{
				Type:               machinesv1.WorkMachineConditionDeletionBlocked,
				Status:             metav1.ConditionTrue,
				LastTransitionTime: &now,
				Reason:             "ActiveWorkspacesExist",
				Message:            message,
			})
		}

		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status with DeletionBlocked condition")
			return ctrl.Result{}, err
		}

		// Requeue to check again later
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// No workspaces, proceed with namespace deletion
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)
	if err == nil {
		// Namespace still exists
		if namespace.DeletionTimestamp != nil {
			// Namespace is being deleted - remove our finalizer to allow it to complete
			if controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
				logger.Info("Removing finalizer from namespace to allow deletion", "namespace", namespaceName)
				controllerutil.RemoveFinalizer(namespace, WorkMachineFinalizerName)
				if err := r.Update(ctx, namespace); err != nil {
					logger.Error(err, "Failed to remove finalizer from namespace")
					return ctrl.Result{}, err
				}
			}
			logger.Info("Namespace is being deleted, waiting for it to be removed", "namespace", namespaceName)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Delete the namespace
		logger.Info("Deleting WorkMachine namespace", "namespace", namespaceName)
		if err := r.Delete(ctx, namespace); err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete namespace", "namespace", namespaceName)
			return ctrl.Result{}, err
		}

		// Requeue to wait for namespace deletion to complete
		logger.Info("Namespace deletion initiated, waiting for completion", "namespace", namespaceName)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get namespace", "namespace", namespaceName)
		return ctrl.Result{}, err
	}

	// Namespace is deleted, remove finalizer
	logger.Info("Namespace is deleted, removing finalizer from WorkMachine")
	controllerutil.RemoveFinalizer(workMachine, WorkMachineFinalizerName)
	if err := r.Update(ctx, workMachine); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("WorkMachine cleanup completed successfully")
	return ctrl.Result{}, nil
}

// ensureNamespace creates the namespace if it doesn't exist and adds finalizer
func (r *WorkMachineReconciler) ensureNamespace(ctx context.Context, namespaceName string, logger logr.Logger) error {
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)

	if err == nil {
		// Namespace already exists, ensure it has the finalizer
		if !controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
			logger.Info("Adding finalizer to existing namespace", "namespace", namespaceName)
			controllerutil.AddFinalizer(namespace, WorkMachineFinalizerName)
			if err := r.Update(ctx, namespace); err != nil {
				logger.Error(err, "Failed to add finalizer to namespace")
				return err
			}
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return err
	}

	// Create the namespace with finalizer
	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"kloudlite.io/managed":     "true",
				"kloudlite.io/workmachine": "true",
			},
			Finalizers: []string{WorkMachineFinalizerName},
		},
	}

	if err := r.Create(ctx, namespace); err != nil {
		return err
	}

	logger.Info("Created namespace for WorkMachine with finalizer", "namespace", namespaceName)
	return nil
}

// ensurePackageManagerRBAC ensures RBAC resources for workmachine-node-manager exist in the namespace
func (r *WorkMachineReconciler) ensurePackageManagerRBAC(ctx context.Context, namespace string, logger logr.Logger) error {
	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
	}

	if err := r.Get(ctx, client.ObjectKey{Name: sa.Name, Namespace: namespace}, sa); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check service account: %w", err)
		}

		// Create service account
		if err := r.Create(ctx, sa); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service account: %w", err)
		}
		logger.Info("Created ServiceAccount for workmachine-node-manager", "namespace", namespace)
	}

	// Create Role with PackageRequest permissions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
		},
	}

	existingRole := &rbacv1.Role{}
	if err := r.Get(ctx, client.ObjectKey{Name: role.Name, Namespace: namespace}, existingRole); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check role: %w", err)
		}

		// Create role
		if err := r.Create(ctx, role); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create role: %w", err)
		}
		logger.Info("Created Role for workmachine-node-manager", "namespace", namespace)
	}

	// Create RoleBinding
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workmachine-node-manager",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "workmachine-node-manager",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "workmachine-node-manager",
		},
	}

	existingRB := &rbacv1.RoleBinding{}
	if err := r.Get(ctx, client.ObjectKey{Name: rb.Name, Namespace: namespace}, existingRB); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check role binding: %w", err)
		}

		// Create role binding
		if err := r.Create(ctx, rb); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create role binding: %w", err)
		}
		logger.Info("Created RoleBinding for workmachine-node-manager", "namespace", namespace)
	}

	return nil
}

// ensureSSHProxySecret ensures the SSH key secret exists for workspace access
func (r *WorkMachineReconciler) ensureSSHProxySecret(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) error {
	secretName := "ssh-proxy-key"
	namespace := workMachine.Spec.TargetNamespace
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, secret)

	needsStatusUpdate := false

	if err == nil {
		// Secret already exists, update status if needed
		if publicKeyBytes, ok := secret.Data["public-key"]; ok {
			publicKeyStr := string(publicKeyBytes)
			if workMachine.Status.SSHPublicKey != publicKeyStr {
				workMachine.Status.SSHPublicKey = publicKeyStr
				needsStatusUpdate = true
			}
		}

		if needsStatusUpdate {
			if err := r.Status().Update(ctx, workMachine); err != nil {
				return fmt.Errorf("failed to update WorkMachine status with SSH public key: %w", err)
			}
		}

		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check ssh-proxy-key secret: %w", err)
	}

	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	// Convert to SSH format for public key
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	// Convert to OpenSSH format for private key
	pemBlock := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: marshalED25519PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(pemBlock)

	// Create secret
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-proxy": "true",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"public-key":  publicKeyBytes,
			"private-key": privateKeyBytes,
		},
	}

	if err := r.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create ssh-proxy-key secret: %w", err)
	}

	logger.Info("Created SSH proxy key secret", "namespace", namespace)

	// Update WorkMachine status with the public key
	workMachine.Status.SSHPublicKey = string(publicKeyBytes)
	if err := r.Status().Update(ctx, workMachine); err != nil {
		return fmt.Errorf("failed to update WorkMachine status with SSH public key: %w", err)
	}

	return nil
}

// marshalED25519PrivateKey marshals Ed25519 private key to OpenSSH format
func marshalED25519PrivateKey(key ed25519.PrivateKey) []byte {
	// OpenSSH key format
	magic := []byte("openssh-key-v1\x00")

	var w sshBuffer
	w.writeString("none") // cipher name
	w.writeString("none") // kdf name
	w.writeString("")     // kdf options
	w.writeUint32(1)      // number of keys

	// Public key
	pubKey := key.Public().(ed25519.PublicKey)
	var pubW sshBuffer
	pubW.writeString("ssh-ed25519")
	pubW.writeBytes(pubKey)
	w.writeBytes(pubW.bytes())

	// Private key
	var privW sshBuffer
	privW.writeUint32(0) // check int 1
	privW.writeUint32(0) // check int 2
	privW.writeString("ssh-ed25519")
	privW.writeBytes(pubKey)
	privW.writeBytes(append(key.Seed(), pubKey...))
	privW.writeString("") // comment

	// Add padding
	for i := 0; i < (8 - (len(privW.bytes()) % 8)); i++ {
		privW.writeByte(byte(i + 1))
	}

	w.writeBytes(privW.bytes())

	return append(magic, w.bytes()...)
}

// sshBuffer helper for SSH key encoding
type sshBuffer struct {
	buf []byte
}

func (w *sshBuffer) writeUint32(v uint32) {
	w.buf = append(w.buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (w *sshBuffer) writeString(s string) {
	w.writeBytes([]byte(s))
}

func (w *sshBuffer) writeBytes(b []byte) {
	w.writeUint32(uint32(len(b)))
	w.buf = append(w.buf, b...)
}

func (w *sshBuffer) writeByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *sshBuffer) bytes() []byte {
	return w.buf
}

// ensureSSHHostKeysSecret ensures the SSH host keys secret exists for SSH servers
// This secret contains RSA, ECDSA, and Ed25519 host keys that are shared across
// all workspace pods and the host-manager pod to maintain consistent SSH server identity
func (r *WorkMachineReconciler) ensureSSHHostKeysSecret(ctx context.Context, namespace string, logger logr.Logger) error {
	secretName := "ssh-host-keys"
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, secret)

	if err == nil {
		// Secret already exists
		logger.Info("SSH host keys secret already exists", "namespace", namespace)
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check ssh-host-keys secret: %w", err)
	}

	logger.Info("Generating SSH host keys", "namespace", namespace)

	// Generate RSA 2048-bit key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate ECDSA 256-bit key
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// Generate Ed25519 key
	ed25519Public, ed25519Private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	// Marshal RSA key
	rsaPrivateBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	rsaSSHPublicKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to create RSA SSH public key: %w", err)
	}
	rsaPublicBytes := ssh.MarshalAuthorizedKey(rsaSSHPublicKey)

	// Marshal ECDSA key
	ecdsaPrivateBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}
	ecdsaPrivatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecdsaPrivateBytes,
	})

	ecdsaSSHPublicKey, err := ssh.NewPublicKey(&ecdsaKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to create ECDSA SSH public key: %w", err)
	}
	ecdsaPublicBytes := ssh.MarshalAuthorizedKey(ecdsaSSHPublicKey)

	// Marshal Ed25519 key (use OpenSSH format)
	ed25519PrivatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: marshalED25519PrivateKey(ed25519Private),
	})

	ed25519SSHPublicKey, err := ssh.NewPublicKey(ed25519Public)
	if err != nil {
		return fmt.Errorf("failed to create Ed25519 SSH public key: %w", err)
	}
	ed25519PublicBytes := ssh.MarshalAuthorizedKey(ed25519SSHPublicKey)

	// Create secret with all host keys
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-host-keys": "true",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ssh_host_rsa_key":         rsaPrivateBytes,
			"ssh_host_rsa_key.pub":     rsaPublicBytes,
			"ssh_host_ecdsa_key":       ecdsaPrivatePEM,
			"ssh_host_ecdsa_key.pub":   ecdsaPublicBytes,
			"ssh_host_ed25519_key":     ed25519PrivatePEM,
			"ssh_host_ed25519_key.pub": ed25519PublicBytes,
		},
	}

	if err := r.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create ssh-host-keys secret: %w", err)
	}

	logger.Info("Created SSH host keys secret", "namespace", namespace)
	return nil
}

// ensureSSHAuthorizedKeysConfigMap ensures the SSH authorized_keys ConfigMap exists
func (r *WorkMachineReconciler) ensureSSHAuthorizedKeysConfigMap(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) error {
	namespace := workMachine.Spec.TargetNamespace
	configMapName := "ssh-authorized-keys"

	// Build authorized_keys content with user keys from WorkMachine spec
	// Validate each SSH key before adding to authorized_keys
	var authorizedKeys strings.Builder
	first := true
	for i, key := range workMachine.Spec.SSHPublicKeys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}

		// Validate SSH key format using golang.org/x/crypto/ssh
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmedKey)); err != nil {
			logger.Error(err, "Invalid SSH public key format", "index", i, "key", trimmedKey[:min(50, len(trimmedKey))]+"...")
			continue // Skip invalid keys but don't fail the entire reconciliation
		}

		if !first {
			authorizedKeys.WriteString("\n")
		}
		authorizedKeys.WriteString(trimmedKey)
		first = false
	}

	// Check if ConfigMap exists
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, configMap)

	if err == nil {
		// ConfigMap exists, update if needed
		if configMap.Data["authorized_keys"] != authorizedKeys.String() {
			configMap.Data["authorized_keys"] = authorizedKeys.String()
			if err := r.Update(ctx, configMap); err != nil {
				return fmt.Errorf("failed to update ssh-authorized-keys configmap: %w", err)
			}
			logger.Info("Updated SSH authorized_keys ConfigMap", "namespace", namespace)
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check ssh-authorized-keys configmap: %w", err)
	}

	// Create ConfigMap with authorized_keys
	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-config": "true",
			},
		},
		Data: map[string]string{
			"authorized_keys": authorizedKeys.String(),
		},
	}

	if err := r.Create(ctx, configMap); err != nil {
		return fmt.Errorf("failed to create ssh-authorized-keys configmap: %w", err)
	}

	logger.Info("Created SSH authorized_keys ConfigMap", "namespace", namespace)
	return nil
}

// ensureSSHDConfigMap ensures the sshd_config ConfigMap exists with secure configuration
// This ConfigMap contains the OpenSSH server configuration for the jump host
func (r *WorkMachineReconciler) ensureSSHDConfigMap(ctx context.Context, namespace string, logger logr.Logger) error {
	configMapName := "sshd-config"

	// Define secure sshd_config content
	sshdConfig := `# OpenSSH Server Configuration for WorkMachine Jump Host
# This configuration enables secure SSH access with jump host (bastion) functionality

# Network Configuration
Port 2222
ListenAddress 0.0.0.0

# Authentication
PermitRootLogin no
PubkeyAuthentication yes
PasswordAuthentication no
PermitEmptyPasswords no
ChallengeResponseAuthentication no

# SSH Jump Host / Bastion Configuration
AllowTcpForwarding yes
GatewayPorts yes
X11Forwarding no

# Security
StrictModes yes
MaxAuthTries 3
MaxSessions 10

# Logging
SyslogFacility AUTH
LogLevel VERBOSE

# Environment
AcceptEnv LANG LC_*

# Subsystems
Subsystem sftp /usr/lib/ssh/sftp-server
`

	// Check if ConfigMap exists
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, configMap)

	if err == nil {
		// ConfigMap exists, update if needed
		if configMap.Data["sshd_config"] != sshdConfig {
			configMap.Data["sshd_config"] = sshdConfig
			if err := r.Update(ctx, configMap); err != nil {
				return fmt.Errorf("failed to update sshd-config configmap: %w", err)
			}
			logger.Info("Updated sshd_config ConfigMap", "namespace", namespace)
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check sshd-config configmap: %w", err)
	}

	// Create ConfigMap with sshd_config
	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-config": "true",
			},
		},
		Data: map[string]string{
			"sshd_config": sshdConfig,
		},
	}

	if err := r.Create(ctx, configMap); err != nil {
		return fmt.Errorf("failed to create sshd-config configmap: %w", err)
	}

	logger.Info("Created sshd_config ConfigMap", "namespace", namespace)
	return nil
}

// ensureWorkspaceSSHDConfigMap creates a ConfigMap with SSHD configuration override for workspaces
// This configures workspaces to use authorized_keys from the mounted ConfigMap location
func (r *WorkMachineReconciler) ensureWorkspaceSSHDConfigMap(ctx context.Context, namespace string, logger logr.Logger) error {
	configMapName := "workspace-sshd-config"

	// SSHD drop-in config to override AuthorizedKeysFile location
	// This will be mounted to /etc/ssh/sshd_config.d/ in workspace containers
	// StrictModes is disabled to allow ConfigMap-mounted authorized_keys (root-owned directory)
	sshdConfigOverride := `# Kloudlite Workspace SSH Configuration
# Override authorized keys location to use mounted ConfigMap
AuthorizedKeysFile /etc/ssh/kl-authorized-keys/authorized_keys
# Disable StrictModes to allow ConfigMap-mounted directories (owned by root)
StrictModes no
`

	// Check if ConfigMap exists
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, configMap)

	if err == nil {
		// ConfigMap exists, update if needed
		if configMap.Data["99-kl-authorized-keys.conf"] != sshdConfigOverride {
			configMap.Data["99-kl-authorized-keys.conf"] = sshdConfigOverride
			if err := r.Update(ctx, configMap); err != nil {
				return fmt.Errorf("failed to update workspace-sshd-config configmap: %w", err)
			}
			logger.Info("Updated workspace-sshd-config ConfigMap", "namespace", namespace)
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check workspace-sshd-config configmap: %w", err)
	}

	// Create ConfigMap with SSHD override
	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/ssh-config":       "true",
				"kloudlite.io/workspace-config": "true",
			},
		},
		Data: map[string]string{
			"99-kl-authorized-keys.conf": sshdConfigOverride,
		},
	}

	if err := r.Create(ctx, configMap); err != nil {
		return fmt.Errorf("failed to create workspace-sshd-config configmap: %w", err)
	}

	logger.Info("Created workspace-sshd-config ConfigMap", "namespace", namespace)
	return nil
}

// ensurePackageManagerDeployment ensures the workmachine-host-manager deployment exists in the WorkMachine's namespace
func (r *WorkMachineReconciler) ensurePackageManagerDeployment(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) error {
	namespace := workMachine.Spec.TargetNamespace
	deploymentName := "workmachine-host-manager"
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)

	if err == nil {
		// Deployment already exists
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check workmachine-host-manager deployment: %w", err)
	}

	// Create the Deployment
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	privileged := true
	replicas := int32(1)

	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                       "workmachine-host-manager",
				"kloudlite.io/package-mgmt": "true",
				"kloudlite.io/workmachine":  workMachine.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "workmachine-host-manager",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "workmachine-host-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "workmachine-node-manager",
					InitContainers: []corev1.Container{
						{
							Name:            "setup-ssh-key",
							Image:           "busybox:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								fmt.Sprintf("cp /ssh-key-source/private-key /ssh-key-target/id_ed25519 && chown %s:%s /ssh-key-target/id_ed25519 && chmod 600 /ssh-key-target/id_ed25519", SSHUserUID, SSHUserGID),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "ssh-proxy-key",
									MountPath: "/ssh-key-source",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-key-volume",
									MountPath: "/ssh-key-target",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "workmachine-node-manager",
							Image:           "kloudlite/workmachine-node-manager:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: namespace,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/var/lib/kloudlite/nix-store",
								},
								{
									Name:      "workspace-homes",
									MountPath: "/var/lib/kloudlite/workspace-homes",
								},
							},
						},
						{
							Name:            "ssh-server",
							Image:           "linuxserver/openssh-server:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "PUID",
									Value: SSHUserUID,
								},
								{
									Name:  "PGID",
									Value: SSHUserGID,
								},
								{
									Name:  "PASSWORD_ACCESS",
									Value: "false",
								},
								{
									Name:  "USER_PASSWORD",
									Value: "kloudlite123",
								},
								{
									Name:  "USER_NAME",
									Value: SSHUserName,
								},
								{
									Name:  "SUDO_ACCESS",
									Value: "false",
								},
								{
									Name:  "TCP_FORWARDING",
									Value: "true",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "ssh",
									ContainerPort: 2222,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "ssh-key-volume",
									MountPath: "/config/.ssh/id_ed25519",
									SubPath:   "id_ed25519",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-authorized-keys",
									MountPath: "/config/.ssh/authorized_keys",
									SubPath:   "authorized_keys",
									ReadOnly:  true,
								},
								{
									Name:      "sshd-config",
									MountPath: "/etc/ssh/sshd_config",
									SubPath:   "sshd_config",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_rsa_key",
									SubPath:   "ssh_host_rsa_key",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_rsa_key.pub",
									SubPath:   "ssh_host_rsa_key.pub",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_ecdsa_key",
									SubPath:   "ssh_host_ecdsa_key",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_ecdsa_key.pub",
									SubPath:   "ssh_host_ecdsa_key.pub",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_ed25519_key",
									SubPath:   "ssh_host_ed25519_key",
									ReadOnly:  true,
								},
								{
									Name:      "ssh-host-keys",
									MountPath: "/etc/ssh/ssh_host_ed25519_key.pub",
									SubPath:   "ssh_host_ed25519_key.pub",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nix-store",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/nix-store",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "workspace-homes",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/workspace-homes",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "ssh-proxy-key",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "ssh-proxy-key",
								},
							},
						},
						{
							Name: "ssh-key-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "ssh-authorized-keys",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "ssh-authorized-keys",
									},
								},
							},
						},
						{
							Name: "sshd-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "sshd-config",
									},
								},
							},
						},
						{
							Name: "ssh-host-keys",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "ssh-host-keys",
									DefaultMode: func() *int32 { m := int32(0600); return &m }(),
								},
							},
						},
					},
				},
			},
		},
	}

	if err := r.Create(ctx, deployment); err != nil {
		return fmt.Errorf("failed to create workmachine-host-manager deployment: %w", err)
	}

	logger.Info("Created workmachine-host-manager Deployment", "namespace", namespace)
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&machinesv1.WorkMachine{}).
		Complete(r)
}
