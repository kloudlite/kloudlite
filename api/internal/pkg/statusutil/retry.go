package statusutil

import (
	"context"
	"time"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UpdateStatusWithRetry updates the status of a Kubernetes resource with retry logic for conflicts
// It automatically handles resourceVersion conflicts by refetching and retrying
func UpdateStatusWithRetry(
	ctx context.Context,
	c client.Client,
	obj client.Object,
	updateFunc func() error,
	logger *zap.Logger,
) error {
	backoff := wait.Backoff{
		Steps:    5,
		Duration: 10 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}

	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		// Apply the status update
		if err := updateFunc(); err != nil {
			return false, err
		}

		// Attempt to update status
		if err := c.Status().Update(ctx, obj); err != nil {
			lastErr = err
			if apierrors.IsConflict(err) {
				// Resource version conflict - fetch the latest version and retry
				logger.Sugar().Debug("Status update conflict, refetching and retrying",
					"name", obj.GetName(),
					"namespace", obj.GetNamespace(),
					"error", err)

				// Refetch the latest version
				key := types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				}
				if err := c.Get(ctx, key, obj); err != nil {
					logger.Sugar().Error("Failed to refetch resource after conflict",
						"name", obj.GetName(),
						"namespace", obj.GetNamespace(),
						"error", err)
					return false, err
				}

				// Retry with the latest version
				// updateFunc will be called again in the next iteration
				return false, nil
			}
			// For non-conflict errors, don't retry
			logger.Sugar().Error("Failed to update status",
				"name", obj.GetName(),
				"namespace", obj.GetNamespace(),
				"error", err)
			return false, err
		}
		// Success
		return true, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			logger.Sugar().Error("Failed to update status after maximum retries",
				"name", obj.GetName(),
				"namespace", obj.GetNamespace(),
				"error", lastErr)
			return lastErr
		}
		return err
	}

	logger.Sugar().Debug("Successfully updated status",
		"name", obj.GetName(),
		"namespace", obj.GetNamespace())

	return nil
}
