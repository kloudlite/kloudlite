package shared

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

// PodDeletionTracker tracks ongoing pod deletions to prevent race conditions
// when multiple controllers try to delete the same pod concurrently
type PodDeletionTracker struct {
	mu       sync.RWMutex
	deletions map[types.UID]DeletionInfo
	logger   *zap.Logger

	metrics *DeletionMetrics
}

// DeletionInfo stores information about an ongoing deletion
type DeletionInfo struct {
	PodName      string
	Namespace    string
	StartedAt    time.Time
	Controller   string
	Completed    bool
	CompletionError error
}

// DeletionMetrics tracks statistics for pod deletion operations
type DeletionMetrics struct {
	mu                sync.RWMutex
	TotalDeletions    int64
	SuccessfulDeletes int64
	FailedDeletes     int64
	RaceConditions    int64 // When another controller already deleting the same pod
}

// NewPodDeletionTracker creates a new pod deletion tracker
func NewPodDeletionTracker(logger *zap.Logger) *PodDeletionTracker {
	return &PodDeletionTracker{
		deletions: make(map[types.UID]DeletionInfo),
		logger:    logger,
		metrics:   &DeletionMetrics{},
	}
}

// TryStartDeletion attempts to start deleting a pod. Returns true if deletion should proceed,
// false if the pod is already being deleted by another controller.
func (t *PodDeletionTracker) TryStartDeletion(podUID types.UID, podName, namespace, controller string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if deletion is already in progress
	if info, exists := t.deletions[podUID]; exists {
		if !info.Completed {
			// Another controller is already deleting this pod
			t.metrics.RaceConditions++
			t.logger.Warn("Pod deletion race condition detected",
				zap.String("pod", podName),
				zap.String("namespace", namespace),
				zap.String("attempting_controller", controller),
				zap.String("existing_controller", info.Controller),
				zap.Time("deletion_started", info.StartedAt),
				zap.Duration("deletion_duration", time.Since(info.StartedAt)))
			return false
		}
	}

	// Start tracking this deletion
	t.deletions[podUID] = DeletionInfo{
		PodName:    podName,
		Namespace:  namespace,
		StartedAt:  time.Now(),
		Controller: controller,
		Completed:  false,
	}
	t.metrics.TotalDeletions++

	t.logger.Debug("Started tracking pod deletion",
		zap.String("pod", podName),
		zap.String("namespace", namespace),
		zap.String("controller", controller))

	return true
}

// CompleteDeletion marks a deletion as complete (success or failure)
func (t *PodDeletionTracker) CompleteDeletion(podUID types.UID, podName, namespace string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if info, exists := t.deletions[podUID]; exists {
		info.Completed = true
		info.CompletionError = err
		t.deletions[podUID] = info

		if err == nil {
			t.metrics.SuccessfulDeletes++
		} else {
			t.metrics.FailedDeletes++
		}

		t.logger.Debug("Completed pod deletion tracking",
			zap.String("pod", podName),
			zap.String("namespace", namespace),
			zap.String("controller", info.Controller),
			zap.Duration("duration", time.Since(info.StartedAt)),
			zap.Error(err))

		// Clean up old completed entries after 5 minutes
		go t.cleanupOldEntries()
	}
}

// cleanupOldEntries removes completed deletion entries older than 5 minutes
func (t *PodDeletionTracker) cleanupOldEntries() {
	time.Sleep(5 * time.Minute)
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for uid, info := range t.deletions {
		if info.Completed && info.StartedAt.Before(cutoff) {
			delete(t.deletions, uid)
		}
	}
}

// GetMetrics returns a copy of the current metrics
func (t *PodDeletionTracker) GetMetrics() DeletionMetrics {
	t.metrics.mu.RLock()
	defer t.metrics.mu.RUnlock()

	return DeletionMetrics{
		TotalDeletions:    t.metrics.TotalDeletions,
		SuccessfulDeletes: t.metrics.SuccessfulDeletes,
		FailedDeletes:     t.metrics.FailedDeletes,
		RaceConditions:    t.metrics.RaceConditions,
	}
}

// IsDeleting returns true if the pod is currently being deleted
func (t *PodDeletionTracker) IsDeleting(podUID types.UID) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if info, exists := t.deletions[podUID]; exists {
		return !info.Completed
	}
	return false
}
