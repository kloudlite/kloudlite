package snapshot

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// How often to re-check registry connectivity
	storeCheckInterval = 5 * time.Minute
)

// SnapshotStoreReconciler reconciles SnapshotStore resources
type SnapshotStoreReconciler struct {
	client.Client
	Logger *zap.Logger
}

func (r *SnapshotStoreReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("snapshotStore", req.Name),
	)

	// Fetch SnapshotStore (cluster-scoped)
	store := &snapshotv1.SnapshotStore{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, store); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotStore", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check registry connectivity
	ready, message := r.checkRegistry(store, logger)

	// Update status
	now := metav1.Now()
	store.Status.Ready = ready
	store.Status.Message = message
	store.Status.LastChecked = &now

	if err := r.Status().Update(ctx, store); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("SnapshotStore status updated",
		zap.Bool("ready", ready),
		zap.String("message", message))

	// Re-check periodically
	return reconcile.Result{RequeueAfter: storeCheckInterval}, nil
}

// checkRegistry verifies the OCI registry is accessible
func (r *SnapshotStoreReconciler) checkRegistry(store *snapshotv1.SnapshotStore, logger *zap.Logger) (bool, string) {
	endpoint := store.Spec.Registry.Endpoint
	if endpoint == "" {
		return false, "Registry endpoint not configured"
	}

	// Build URL for health check
	scheme := "https"
	if store.Spec.Registry.Insecure {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/v2/", scheme, endpoint)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		logger.Warn("Registry check failed", zap.String("url", url), zap.Error(err))
		return false, fmt.Sprintf("Cannot connect to registry: %v", err)
	}
	defer resp.Body.Close()

	// OCI registry should return 200 or 401 (if auth required) for /v2/
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
		return true, "Registry is accessible"
	}

	return false, fmt.Sprintf("Registry returned unexpected status: %d", resp.StatusCode)
}

func (r *SnapshotStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotStore{}).
		Complete(r)
}
