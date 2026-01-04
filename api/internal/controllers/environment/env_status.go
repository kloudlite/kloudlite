package environment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// generateHash generates an 8-character hash from the input string
func generateHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:8]
}

// updateHashAndSubdomain computes and sets the hash and subdomain in environment status
func (r *EnvironmentReconciler) updateHashAndSubdomain(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Compute hash from envName-owner
	hash := generateHash(fmt.Sprintf("%s-%s", environment.Spec.Name, environment.Spec.OwnedBy))

	// Get subdomain from HOSTED_SUBDOMAIN env var (shared across all environments)
	subdomain := os.Getenv("HOSTED_SUBDOMAIN")
	if subdomain == "" {
		logger.Debug("HOSTED_SUBDOMAIN env var not set, subdomain will be empty")
	}

	// Only update if values changed
	if environment.Status.Hash == hash && environment.Status.Subdomain == subdomain {
		return nil
	}

	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.Hash = hash
		environment.Status.Subdomain = subdomain
		return nil
	}, logger)
}

// updateEnvironmentStatus safely updates environment status with retry logic
func (r *EnvironmentReconciler) updateEnvironmentStatus(ctx context.Context, environment *environmentsv1.Environment, state environmentsv1.EnvironmentState, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = state
		environment.Status.Message = message

		now := metav1.Now()
		if state == environmentsv1.EnvironmentStateActive {
			environment.Status.LastActivatedTime = &now
		} else if state == environmentsv1.EnvironmentStateInactive {
			environment.Status.LastDeactivatedTime = &now
		}

		return nil
	}, logger)
}

// addOrUpdateCondition adds or updates a condition in the environment status
func (r *EnvironmentReconciler) addOrUpdateCondition(environment *environmentsv1.Environment, conditionType environmentsv1.EnvironmentConditionType, status metav1.ConditionStatus, reason, message string) {
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	now := metav1.Now()
	newCondition := environmentsv1.EnvironmentCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: &now,
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or add new one
	found := false
	for i, condition := range environment.Status.Conditions {
		if condition.Type == conditionType {
			environment.Status.Conditions[i] = newCondition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, newCondition)
	}
}
