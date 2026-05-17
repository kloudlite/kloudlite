package workmachine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// UsageEvent represents a billable event sent to the console API.
type UsageEvent struct {
	EventType    string                 `json:"event_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// UsageReporter sends usage events to the console API for billing.
type UsageReporter struct {
	consoleBaseURL  string
	installationKey string
	httpClient      *http.Client
	logger          *zap.Logger
}

// NewUsageReporter creates a new UsageReporter.
func NewUsageReporter(consoleBaseURL, installationKey string, logger *zap.Logger) *UsageReporter {
	return &UsageReporter{
		consoleBaseURL:  consoleBaseURL,
		installationKey: installationKey,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		logger:          logger,
	}
}

// ReportEvent sends a usage event to the console API.
// This is fire-and-forget: errors are logged but do not block the caller.
func (r *UsageReporter) ReportEvent(ctx context.Context, event UsageEvent) {
	if r == nil {
		return
	}

	body, err := json.Marshal(event)
	if err != nil {
		r.logger.Error("failed to marshal usage event", zap.Error(err))
		return
	}

	url := fmt.Sprintf("%s/api/installations/usage-event", r.consoleBaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		r.logger.Error("failed to create usage event request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-installation-key", r.installationKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("failed to send usage event",
			zap.String("event_type", event.EventType),
			zap.String("resource_id", event.ResourceID),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		r.logger.Error("usage event rejected",
			zap.String("event_type", event.EventType),
			zap.String("resource_id", event.ResourceID),
			zap.Int("status", resp.StatusCode),
		)
	} else {
		r.logger.Debug("usage event reported",
			zap.String("event_type", event.EventType),
			zap.String("resource_id", event.ResourceID),
		)
	}
}
