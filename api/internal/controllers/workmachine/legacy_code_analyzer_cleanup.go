package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const legacyCodeAnalyzerName = "code-analyzer"

func (r *WorkMachineReconciler) cleanupLegacyCodeAnalyzer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      legacyCodeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), service); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete legacy code-analyzer service: %w", err))
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      legacyCodeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), statefulSet); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete legacy code-analyzer statefulset: %w", err))
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      legacyCodeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), serviceAccount); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete legacy code-analyzer service account: %w", err))
	}

	return check.Passed()
}
