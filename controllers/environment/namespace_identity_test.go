package environment

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCleanupWorkspaceConnectionsSameNameDifferentNamespaceOnlyClearsMatchingNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	matchingWorkspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "matching-workspace", Namespace: "wm-alice"},
		Spec: workspacev1.WorkspaceSpec{EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
			EnvironmentRef: corev1.ObjectReference{Name: "shared-env", Namespace: "wm-alice"},
		}},
	}
	otherNamespaceWorkspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "other-namespace-workspace", Namespace: "wm-bob"},
		Spec: workspacev1.WorkspaceSpec{EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
			EnvironmentRef: corev1.ObjectReference{Name: "shared-env", Namespace: "wm-bob"},
		}},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, matchingWorkspace, otherNamespaceWorkspace).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	err := reconciler.cleanupWorkspaceConnections(ctx, env, zap.NewNop())
	assert.NoError(t, err)

	updatedMatching := &workspacev1.Workspace{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "matching-workspace", Namespace: "wm-alice"}, updatedMatching)
	assert.NoError(t, err)
	assert.Nil(t, updatedMatching.Spec.EnvironmentConnection)

	updatedOther := &workspacev1.Workspace{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "other-namespace-workspace", Namespace: "wm-bob"}, updatedOther)
	assert.NoError(t, err)
	assert.NotNil(t, updatedOther.Spec.EnvironmentConnection)
	assert.Equal(t, "wm-bob", updatedOther.Spec.EnvironmentConnection.EnvironmentRef.Namespace)
}

func TestCleanupWorkspaceConnectionsEmptyEnvironmentNamespaceReturnsErrorAndDoesNotClearDefaultNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	defaultNamespaceWorkspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "default-namespace-workspace", Namespace: "wm-default"},
		Spec: workspacev1.WorkspaceSpec{EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
			EnvironmentRef: corev1.ObjectReference{Name: "shared-env", Namespace: "default"},
		}},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, defaultNamespaceWorkspace).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	err := reconciler.cleanupWorkspaceConnections(ctx, env, zap.NewNop())
	assert.ErrorContains(t, err, "environment namespace is required for workspace cleanup")

	updatedWorkspace := &workspacev1.Workspace{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "default-namespace-workspace", Namespace: "wm-default"}, updatedWorkspace)
	assert.NoError(t, err)
	assert.NotNil(t, updatedWorkspace.Spec.EnvironmentConnection)
	assert.Equal(t, "default", updatedWorkspace.Spec.EnvironmentConnection.EnvironmentRef.Namespace)
}

func TestHandleEnvironmentDeactivationSameNameDifferentTargetNamespaceOnlyClearsMatchingTargetNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	matchingWorkspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "matching-workspace", Namespace: "wm-alice"},
		Status: workspacev1.WorkspaceStatus{ConnectedEnvironment: &workspacev1.ConnectedEnvironmentInfo{
			Name:            "shared-env",
			TargetNamespace: "env-alice",
		}},
	}
	otherTargetWorkspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "other-target-workspace", Namespace: "wm-bob"},
		Status: workspacev1.WorkspaceStatus{ConnectedEnvironment: &workspacev1.ConnectedEnvironmentInfo{
			Name:            "shared-env",
			TargetNamespace: "env-bob",
		}},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, matchingWorkspace, otherTargetWorkspace).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	err := reconciler.handleEnvironmentDeactivation(ctx, env, zap.NewNop())
	assert.NoError(t, err)

	updatedMatching := &workspacev1.Workspace{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "matching-workspace", Namespace: "wm-alice"}, updatedMatching)
	assert.NoError(t, err)
	assert.Nil(t, updatedMatching.Status.ConnectedEnvironment)

	updatedOther := &workspacev1.Workspace{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "other-target-workspace", Namespace: "wm-bob"}, updatedOther)
	assert.NoError(t, err)
	assert.NotNil(t, updatedOther.Status.ConnectedEnvironment)
	assert.Equal(t, "env-bob", updatedOther.Status.ConnectedEnvironment.TargetNamespace)
}

func TestHasActiveSnapshotOperationSameNameDifferentNamespaceIgnoresOtherNamespaceActiveOps(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	otherNamespaceRequest := &environmentsv1.EnvironmentCheckpoint{
		ObjectMeta: metav1.ObjectMeta{Name: "checkpoint", Namespace: "env-bob"},
		Spec: environmentsv1.EnvironmentCheckpointSpec{
			EnvironmentName:      "shared-env",
			EnvironmentNamespace: "wm-bob",
			CheckpointName:       "checkpoint-bob",
		},
		Status: environmentsv1.EnvironmentCheckpointStatus{Phase: environmentsv1.EnvironmentCheckpointPhaseCreatingCheckpoint},
	}
	otherNamespaceRestore := &environmentsv1.EnvironmentCheckpointRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "checkpoint-restore", Namespace: "env-bob"},
		Spec: environmentsv1.EnvironmentCheckpointRestoreSpec{
			EnvironmentName:      "shared-env",
			EnvironmentNamespace: "wm-bob",
			CheckpointName:       "checkpoint-bob",
			SourceNamespace:      "env-bob",
		},
		Status: environmentsv1.EnvironmentCheckpointRestoreStatus{Phase: environmentsv1.EnvironmentCheckpointRestorePhaseRestoringData},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, otherNamespaceRequest, otherNamespaceRestore).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	hasActive, err := reconciler.hasActiveSnapshotOperation(ctx, env)
	assert.NoError(t, err)
	assert.False(t, hasActive)
}

func TestHasActiveSnapshotOperationMatchesEnvironmentNameAndNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	matchingRequest := &environmentsv1.EnvironmentCheckpoint{
		ObjectMeta: metav1.ObjectMeta{Name: "checkpoint", Namespace: "env-alice"},
		Spec: environmentsv1.EnvironmentCheckpointSpec{
			EnvironmentName:      "shared-env",
			EnvironmentNamespace: "wm-alice",
			CheckpointName:       "checkpoint-alice",
		},
		Status: environmentsv1.EnvironmentCheckpointStatus{Phase: environmentsv1.EnvironmentCheckpointPhaseCreatingCheckpoint},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, matchingRequest).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	hasActive, err := reconciler.hasActiveSnapshotOperation(ctx, env)
	assert.NoError(t, err)
	assert.True(t, hasActive)
}

func TestHasActiveSnapshotOperationCheckpointEmptyEnvironmentNamespaceDoesNotMatch(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	emptyNamespaceRequest := &environmentsv1.EnvironmentCheckpoint{
		ObjectMeta: metav1.ObjectMeta{Name: "checkpoint", Namespace: "env-alice"},
		Spec: environmentsv1.EnvironmentCheckpointSpec{
			EnvironmentName: "shared-env",
			CheckpointName:  "checkpoint-alice",
		},
		Status: environmentsv1.EnvironmentCheckpointStatus{Phase: environmentsv1.EnvironmentCheckpointPhaseCreatingCheckpoint},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, emptyNamespaceRequest).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	hasActive, err := reconciler.hasActiveSnapshotOperation(ctx, env)
	assert.NoError(t, err)
	assert.False(t, hasActive)
}

func TestHasActiveSnapshotOperationCheckpointRestoreEmptyEnvironmentNamespaceDoesNotMatch(t *testing.T) {
	scheme := testutil.NewTestScheme()
	ctx := context.Background()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-env", Namespace: "wm-alice"},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "env-alice"},
	}
	emptyNamespaceRestore := &environmentsv1.EnvironmentCheckpointRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "checkpoint-restore", Namespace: "env-alice"},
		Spec: environmentsv1.EnvironmentCheckpointRestoreSpec{
			EnvironmentName: "shared-env",
			CheckpointName:  "checkpoint-alice",
			SourceNamespace: "env-alice",
		},
		Status: environmentsv1.EnvironmentCheckpointRestoreStatus{Phase: environmentsv1.EnvironmentCheckpointRestorePhaseRestoringData},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, emptyNamespaceRestore).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	hasActive, err := reconciler.hasActiveSnapshotOperation(ctx, env)
	assert.NoError(t, err)
	assert.False(t, hasActive)
}
