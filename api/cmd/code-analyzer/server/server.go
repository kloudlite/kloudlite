package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/queue"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/storage"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/watcher"
	"go.uber.org/zap"
)

// Server is the HTTP API server
type Server struct {
	storage        *storage.Storage
	queue          *queue.Queue
	watcherManager *watcher.Manager
	logger         *zap.Logger
	server         *http.Server
}

// NewServer creates a new HTTP server
func NewServer(
	storage *storage.Storage,
	queue *queue.Queue,
	watcherManager *watcher.Manager,
	logger *zap.Logger,
) *Server {
	return &Server{
		storage:        storage,
		queue:          queue,
		watcherManager: watcherManager,
		logger:         logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Workspaces
	mux.HandleFunc("/workspaces", s.handleWorkspaces)

	// Reports
	mux.HandleFunc("/reports/", s.handleReports)

	// Manual analysis trigger
	mux.HandleFunc("/analyze/", s.handleAnalyze)

	// Status
	mux.HandleFunc("/status/", s.handleStatus)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.logger.Info("Starting HTTP server", zap.String("addr", addr))
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// clearWorkspaceCache clears the findings cache and manifest for a workspace
// This forces a full re-analysis on the next run
func (s *Server) clearWorkspaceCache(workspace string) {
	basePath := s.storage.GetBasePath()
	workspaceDir := filepath.Join(basePath, workspace)

	// Remove findings cache
	findingsCachePath := filepath.Join(workspaceDir, "findings-cache.json")
	if err := os.Remove(findingsCachePath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove findings cache", zap.String("workspace", workspace), zap.Error(err))
	}

	// Remove manifest
	manifestPath := filepath.Join(workspaceDir, "manifest.json")
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove manifest", zap.String("workspace", workspace), zap.Error(err))
	}

	s.logger.Info("Cleared workspace cache for forced re-analysis", zap.String("workspace", workspace))
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status         string `json:"status"`
	ActiveWatchers int    `json:"activeWatchers"`
	QueuedJobs     int    `json:"queuedJobs"`
	InProgressJobs int    `json:"inProgressJobs"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.queue.GetStats()
	watchedWorkspaces := s.watcherManager.GetWatchedWorkspaces()

	resp := HealthResponse{
		Status:         "healthy",
		ActiveWatchers: len(watchedWorkspaces),
		QueuedJobs:     stats.QueuedCount,
		InProgressJobs: stats.InProgressCount,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// WorkspacesResponse represents workspaces list response
type WorkspacesResponse struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

// WorkspaceInfo represents workspace information
type WorkspaceInfo struct {
	Name     string `json:"name"`
	Watching bool   `json:"watching"`
}

func (s *Server) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	watched := s.watcherManager.GetWatchedWorkspaces()
	stored, _ := s.storage.ListWorkspaces()

	// Merge lists
	workspaceMap := make(map[string]bool)
	for _, ws := range watched {
		workspaceMap[ws] = true
	}
	for _, ws := range stored {
		if _, exists := workspaceMap[ws]; !exists {
			workspaceMap[ws] = false
		}
	}

	workspaces := make([]WorkspaceInfo, 0, len(workspaceMap))
	for name, watching := range workspaceMap {
		workspaces = append(workspaces, WorkspaceInfo{
			Name:     name,
			Watching: watching,
		})
	}

	s.writeJSON(w, http.StatusOK, WorkspacesResponse{Workspaces: workspaces})
}

func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /reports/{workspace}/security or /reports/{workspace}/quality or /reports/{workspace}/aggregated
	// Or: /reports/{workspace}/security/history
	path := strings.TrimPrefix(r.URL.Path, "/reports/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "Invalid path. Use /reports/{workspace}/security, /reports/{workspace}/quality, or /reports/{workspace}/aggregated", http.StatusBadRequest)
		return
	}

	workspace := parts[0]
	reportType := parts[1]

	// Handle aggregated reports separately
	if reportType == "aggregated" {
		// Check if requesting history
		if len(parts) >= 3 && parts[2] == "history" {
			history, err := s.storage.GetAggregatedReportHistory(workspace)
			if err != nil {
				s.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.writeJSON(w, http.StatusOK, map[string]interface{}{"history": history})
			return
		}

		// Get latest aggregated report
		report, err := s.storage.GetLatestAggregatedReport(workspace)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if report == nil {
			s.writeError(w, http.StatusNotFound, "No aggregated report found")
			return
		}

		s.writeJSON(w, http.StatusOK, report)
		return
	}

	var rt storage.ReportType
	switch reportType {
	case "security":
		rt = storage.ReportTypeSecurity
	case "quality":
		rt = storage.ReportTypeQuality
	default:
		http.Error(w, "Invalid report type. Use 'security', 'quality', or 'aggregated'", http.StatusBadRequest)
		return
	}

	// Check if requesting history
	if len(parts) >= 3 && parts[2] == "history" {
		history, err := s.storage.GetReportHistory(workspace, rt)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]interface{}{"history": history})
		return
	}

	// Get latest report
	report, err := s.storage.GetLatestReport(workspace, rt)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if report == nil {
		s.writeError(w, http.StatusNotFound, "No report found")
		return
	}

	s.writeJSON(w, http.StatusOK, report)
}

// AnalyzeResponse represents manual analysis trigger response
type AnalyzeResponse struct {
	Queued        bool   `json:"queued"`
	EstimatedTime string `json:"estimatedTime"`
	Message       string `json:"message,omitempty"`
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /analyze/{workspace}
	workspace := strings.TrimPrefix(r.URL.Path, "/analyze/")
	if workspace == "" {
		http.Error(w, "Workspace name required", http.StatusBadRequest)
		return
	}

	// Check if already in progress
	if s.queue.IsInProgress(workspace) {
		s.writeJSON(w, http.StatusOK, AnalyzeResponse{
			Queued:  false,
			Message: "Analysis already in progress",
		})
		return
	}

	// Check for force parameter - clears cache to force full re-analysis
	force := r.URL.Query().Get("force") == "true"
	if force {
		s.clearWorkspaceCache(workspace)
	}

	// Trigger manual analysis
	queued := s.queue.TriggerManualAnalysis(workspace)

	resp := AnalyzeResponse{
		Queued:        queued,
		EstimatedTime: "30-60s",
	}
	if !queued {
		resp.Message = "Failed to queue analysis"
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// StatusResponse represents workspace status response
type StatusResponse struct {
	Workspace       string    `json:"workspace"`
	Watching        bool      `json:"watching"`
	InProgress      bool      `json:"inProgress"`
	PendingAnalysis bool      `json:"pendingAnalysis"`
	QueuePosition   int       `json:"queuePosition"`
	LastAnalysis    time.Time `json:"lastAnalysis,omitempty"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /status/{workspace}
	workspace := strings.TrimPrefix(r.URL.Path, "/status/")
	if workspace == "" {
		http.Error(w, "Workspace name required", http.StatusBadRequest)
		return
	}

	status := s.queue.GetStatus(workspace)

	resp := StatusResponse{
		Workspace:       workspace,
		Watching:        s.watcherManager.IsWatching(workspace),
		InProgress:      status.InProgress,
		PendingAnalysis: status.PendingAnalysis,
		QueuePosition:   status.QueuePosition,
		LastAnalysis:    status.LastAnalysis,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Error encoding JSON response: %v\n", err)
	}
}
