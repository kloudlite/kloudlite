package queue

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AnalysisType represents the type of analysis
type AnalysisType string

const (
	AnalysisTypeFull     AnalysisType = "full"     // Both security and quality
	AnalysisTypeSecurity AnalysisType = "security" // Security only
	AnalysisTypeQuality  AnalysisType = "quality"  // Quality only
)

// Job represents an analysis job
type Job struct {
	WorkspaceName string
	Type          AnalysisType
	RequestedAt   time.Time
	Manual        bool // True if manually triggered
}

// JobStatus represents the status of a job
type JobStatus struct {
	WorkspaceName   string
	InProgress      bool
	QueuePosition   int
	LastAnalysis    time.Time
	PendingAnalysis bool
}

// AnalyzeFunc is the function called to analyze a workspace
type AnalyzeFunc func(ctx context.Context, workspaceName string) error

// Queue manages analysis jobs with rate limiting
type Queue struct {
	jobs          chan Job
	maxConcurrent int
	inProgress    sync.Map // map[workspaceName]bool
	queued        sync.Map // map[workspaceName]time.Time
	lastAnalysis  sync.Map // map[workspaceName]time.Time
	logger        *zap.Logger
	analyzeFunc   AnalyzeFunc
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewQueue creates a new analysis queue
func NewQueue(maxConcurrent int, analyzeFunc AnalyzeFunc, logger *zap.Logger) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	return &Queue{
		jobs:          make(chan Job, 100),
		maxConcurrent: maxConcurrent,
		logger:        logger,
		analyzeFunc:   analyzeFunc,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the queue workers
func (q *Queue) Start() {
	for i := 0; i < q.maxConcurrent; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	q.logger.Info("Queue started", zap.Int("workers", q.maxConcurrent))
}

// Stop stops the queue and waits for workers to finish
func (q *Queue) Stop() {
	q.cancel()
	close(q.jobs)
	q.wg.Wait()
	q.logger.Info("Queue stopped")
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(job Job) bool {
	// Check if already in progress
	if _, inProgress := q.inProgress.Load(job.WorkspaceName); inProgress {
		q.logger.Debug("Workspace already being analyzed, skipping",
			zap.String("workspace", job.WorkspaceName))
		return false
	}

	// Check if already queued (unless manual)
	if !job.Manual {
		if _, queued := q.queued.Load(job.WorkspaceName); queued {
			q.logger.Debug("Workspace already queued, skipping",
				zap.String("workspace", job.WorkspaceName))
			return false
		}
	}

	// Mark as queued
	q.queued.Store(job.WorkspaceName, job.RequestedAt)

	// Send to channel (non-blocking)
	select {
	case q.jobs <- job:
		q.logger.Info("Job enqueued",
			zap.String("workspace", job.WorkspaceName),
			zap.String("type", string(job.Type)),
			zap.Bool("manual", job.Manual),
		)
		return true
	default:
		q.queued.Delete(job.WorkspaceName)
		q.logger.Warn("Queue full, dropping job",
			zap.String("workspace", job.WorkspaceName))
		return false
	}
}

// EnqueueAnalysis is a convenience method to enqueue a full analysis
func (q *Queue) EnqueueAnalysis(workspaceName string) bool {
	return q.Enqueue(Job{
		WorkspaceName: workspaceName,
		Type:          AnalysisTypeFull,
		RequestedAt:   time.Now(),
		Manual:        false,
	})
}

// TriggerManualAnalysis triggers an immediate analysis
func (q *Queue) TriggerManualAnalysis(workspaceName string) bool {
	return q.Enqueue(Job{
		WorkspaceName: workspaceName,
		Type:          AnalysisTypeFull,
		RequestedAt:   time.Now(),
		Manual:        true,
	})
}

// GetStatus returns the status for a workspace
func (q *Queue) GetStatus(workspaceName string) JobStatus {
	status := JobStatus{
		WorkspaceName: workspaceName,
	}

	if _, ok := q.inProgress.Load(workspaceName); ok {
		status.InProgress = true
	}

	if _, ok := q.queued.Load(workspaceName); ok {
		status.PendingAnalysis = true
		status.QueuePosition = q.getQueuePosition(workspaceName)
	}

	if lastTime, ok := q.lastAnalysis.Load(workspaceName); ok {
		status.LastAnalysis = lastTime.(time.Time)
	}

	return status
}

// IsInProgress returns true if workspace is currently being analyzed
func (q *Queue) IsInProgress(workspaceName string) bool {
	_, ok := q.inProgress.Load(workspaceName)
	return ok
}

func (q *Queue) worker(id int) {
	defer q.wg.Done()

	q.logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-q.ctx.Done():
			q.logger.Debug("Worker stopping", zap.Int("worker_id", id))
			return

		case job, ok := <-q.jobs:
			if !ok {
				return
			}
			q.processJob(job)
		}
	}
}

func (q *Queue) processJob(job Job) {
	workspaceName := job.WorkspaceName

	// Mark as in progress, remove from queued
	q.queued.Delete(workspaceName)
	q.inProgress.Store(workspaceName, true)

	q.logger.Info("Processing job",
		zap.String("workspace", workspaceName),
		zap.String("type", string(job.Type)),
	)

	startTime := time.Now()

	// Run analysis
	err := q.analyzeFunc(q.ctx, workspaceName)

	duration := time.Since(startTime)

	// Update last analysis time
	q.lastAnalysis.Store(workspaceName, time.Now())

	// Remove from in progress
	q.inProgress.Delete(workspaceName)

	if err != nil {
		q.logger.Error("Job failed",
			zap.String("workspace", workspaceName),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		q.logger.Info("Job completed",
			zap.String("workspace", workspaceName),
			zap.Duration("duration", duration),
		)
	}
}

func (q *Queue) getQueuePosition(workspaceName string) int {
	// This is approximate since we can't easily peek into the channel
	position := 0
	q.queued.Range(func(key, value interface{}) bool {
		if key.(string) != workspaceName {
			if queuedTime, ok := value.(time.Time); ok {
				if requestedTime, exists := q.queued.Load(workspaceName); exists {
					if queuedTime.Before(requestedTime.(time.Time)) {
						position++
					}
				}
			}
		}
		return true
	})
	return position
}

// Stats returns queue statistics
type Stats struct {
	QueuedCount     int
	InProgressCount int
	TotalProcessed  int64
}

// GetStats returns current queue statistics
func (q *Queue) GetStats() Stats {
	stats := Stats{}

	q.queued.Range(func(_, _ interface{}) bool {
		stats.QueuedCount++
		return true
	})

	q.inProgress.Range(func(_, _ interface{}) bool {
		stats.InProgressCount++
		return true
	})

	return stats
}
