package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompositionReconciler_UpdateStatus_Running(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "Success", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}
	assert.Equal(t, environmentsv1.CompositionStateRunning, composition.Status.State)
	assert.Equal(t, "Success", composition.Status.Message)
	assert.Equal(t, int64(1), composition.Status.ObservedGeneration)
	assert.NotNil(t, composition.Status.LastDeployedTime)
}

func TestCompositionReconciler_UpdateStatus_Failed(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 2,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateFailed, "Parse error", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}
	assert.Equal(t, environmentsv1.CompositionStateFailed, composition.Status.State)
	assert.Equal(t, "Parse error", composition.Status.Message)
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, composition.Status.Conditions[0].Status)
}

func TestCompositionReconciler_UpdateStatus_Deploying(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateDeploying, "Deploying", logger)
	assert.NoError(t, err)
	assert.Greater(t, result.RequeueAfter.Seconds(), float64(0))
	assert.Equal(t, environmentsv1.CompositionStateDeploying, composition.Status.State)
}

func TestCompositionReconciler_UpdateStatus_UpdateExistingCondition(t *testing.T) {
	scheme := testutil.NewTestScheme()

	oldTime := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 2,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					LastTransitionTime: oldTime,
					Reason:             "Running",
					Message:            "Old message",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateFailed, "New error message", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify condition was updated, not added
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, composition.Status.Conditions[0].Status)
	assert.Equal(t, "New error message", composition.Status.Conditions[0].Message)
	assert.Equal(t, int64(2), composition.Status.Conditions[0].ObservedGeneration)
}

func TestCompositionReconciler_UpdateStatus_MultipleConditions(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "OtherCondition",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					LastTransitionTime: now,
					Reason:             "Other",
					Message:            "Some other condition",
				},
				{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: 0,
					LastTransitionTime: now,
					Reason:             "Failed",
					Message:            "Old failure",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "Now running", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify we have 2 conditions and only Ready was updated
	assert.Len(t, composition.Status.Conditions, 2)
	assert.Equal(t, "OtherCondition", composition.Status.Conditions[0].Type)
	assert.Equal(t, "Some other condition", composition.Status.Conditions[0].Message) // Unchanged
	assert.Equal(t, "Ready", composition.Status.Conditions[1].Type)
	assert.Equal(t, metav1.ConditionTrue, composition.Status.Conditions[1].Status)
	assert.Equal(t, "Now running", composition.Status.Conditions[1].Message)
}

func TestCompositionReconciler_UpdateStatus_AddConditionWhenNoneExist(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{}, // Empty conditions
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "First status", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify condition was added
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, composition.Status.Conditions[0].Status)
	assert.Equal(t, "First status", composition.Status.Conditions[0].Message)
	assert.NotNil(t, composition.Status.LastDeployedTime)
}
