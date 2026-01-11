package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	zap2 "go.uber.org/zap"
)

// MetricsServer provides HTTP endpoints for GPU and host metrics
type MetricsServer struct {
	CmdExec CommandExecutor
	Logger  *zap2.Logger
	Port    int

	// Cache for GPU metrics with periodic background updates
	cacheMutex    sync.RWMutex
	cachedMetrics *GPUMetricsResponse
	cacheInterval time.Duration
	stopChan      chan struct{}
}

// GPUMetricsResponse is the JSON response for GPU metrics
type GPUMetricsResponse struct {
	Detected          bool    `json:"detected"`
	Model             string  `json:"model,omitempty"`
	DriverVersion     string  `json:"driverVersion,omitempty"`
	Count             int     `json:"count,omitempty"`
	MemoryTotal       int32   `json:"memoryTotal,omitempty"`
	MemoryUsed        int32   `json:"memoryUsed,omitempty"`
	MemoryFree        int32   `json:"memoryFree,omitempty"`
	UtilizationGPU    int32   `json:"utilizationGpu,omitempty"`
	UtilizationMemory int32   `json:"utilizationMemory,omitempty"`
	Temperature       int32   `json:"temperature,omitempty"`
	PowerDraw         float32 `json:"powerDraw,omitempty"`
	PowerLimit        float32 `json:"powerLimit,omitempty"`
}

func (s *MetricsServer) Start() error {
	// Start background goroutine to poll GPU metrics at constant frequency
	s.startBackgroundPolling()

	http.HandleFunc("/metrics/gpu", s.handleGPUMetrics)
	http.HandleFunc("/healthz", s.handleHealthz)

	addr := fmt.Sprintf(":%d", s.Port)
	s.Logger.Info("Starting metrics server", zap2.Int("port", s.Port))
	return http.ListenAndServe(addr, nil)
}

// startBackgroundPolling starts a background goroutine that polls GPU metrics
// at a constant frequency and updates the in-memory cache
func (s *MetricsServer) startBackgroundPolling() {
	// Set default cache interval if not configured
	if s.cacheInterval == 0 {
		s.cacheInterval = 5 * time.Second
	}

	s.stopChan = make(chan struct{})

	go func() {
		s.Logger.Info("Starting background GPU metrics polling", zap2.Duration("interval", s.cacheInterval))

		// Poll immediately on startup
		s.updateCache()

		ticker := time.NewTicker(s.cacheInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateCache()
			case <-s.stopChan:
				s.Logger.Info("Stopping background GPU metrics polling")
				return
			}
		}
	}()
}

// updateCache polls GPU metrics and updates the in-memory cache
func (s *MetricsServer) updateCache() {
	// Check if GPU is detected
	gpuDetected := s.detectGPU()
	if !gpuDetected {
		s.cacheMutex.Lock()
		s.cachedMetrics = &GPUMetricsResponse{
			Detected: false,
		}
		s.cacheMutex.Unlock()
		return
	}

	// Collect GPU metrics
	metrics, err := s.collectGPUMetrics()
	if err != nil {
		s.Logger.Error("Failed to collect GPU metrics in background poll", zap2.Error(err))
		return
	}

	// Update cache with write lock
	s.cacheMutex.Lock()
	s.cachedMetrics = metrics
	s.cacheMutex.Unlock()

	s.Logger.Debug("GPU metrics cache updated",
		zap2.Bool("detected", metrics.Detected),
		zap2.String("model", metrics.Model),
		zap2.Int32("utilizationGpu", metrics.UtilizationGPU))
}

// Stop gracefully stops the background polling goroutine
func (s *MetricsServer) Stop() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}

func (s *MetricsServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *MetricsServer) handleGPUMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Serve from cache with read lock (fast, no nvidia-smi execution)
	s.cacheMutex.RLock()
	cachedMetrics := s.cachedMetrics
	s.cacheMutex.RUnlock()

	// If cache is not yet populated, return a default response
	if cachedMetrics == nil {
		json.NewEncoder(w).Encode(GPUMetricsResponse{
			Detected: false,
		})
		return
	}

	json.NewEncoder(w).Encode(cachedMetrics)
}

func (s *MetricsServer) detectGPU() bool {
	checkScript := `
		if [ -d /sys/bus/pci/devices ]; then
			for device in /sys/bus/pci/devices/*; do
				if [ -f "$device/vendor" ] && [ -f "$device/device" ]; then
					vendor=$(cat "$device/vendor" 2>/dev/null)
					if [ "$vendor" = "0x10de" ]; then
						exit 0
					fi
				fi
			done
		fi
		exit 1
	`
	_, err := s.CmdExec.Execute(checkScript)
	return err == nil
}

func (s *MetricsServer) collectGPUMetrics() (*GPUMetricsResponse, error) {
	// Query nvidia-smi for comprehensive metrics
	metricsScript := "nvidia-smi --query-gpu=name,driver_version,memory.total,memory.used,memory.free,utilization.gpu,utilization.memory,temperature.gpu,power.draw,power.limit --format=csv,noheader,nounits | head -1"

	output, err := s.CmdExec.Execute(metricsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU metrics: %w", err)
	}

	// Parse CSV output
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 10 {
		return nil, fmt.Errorf("unexpected nvidia-smi output format: got %d fields, expected 10", len(parts))
	}

	// Helper to parse int32
	parseInt32 := func(s string) (int32, error) {
		s = strings.TrimSpace(s)
		var val int32
		if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
			return 0, err
		}
		return val, nil
	}

	// Helper to parse float32
	parseFloat32 := func(s string) (float32, error) {
		s = strings.TrimSpace(s)
		var val float32
		if _, err := fmt.Sscanf(s, "%f", &val); err != nil {
			return 0, err
		}
		return val, nil
	}

	memoryTotal, _ := parseInt32(parts[2])
	memoryUsed, _ := parseInt32(parts[3])
	memoryFree, _ := parseInt32(parts[4])
	utilizationGPU, _ := parseInt32(parts[5])
	utilizationMemory, _ := parseInt32(parts[6])
	temperature, _ := parseInt32(parts[7])
	powerDraw, _ := parseFloat32(parts[8])
	powerLimit, _ := parseFloat32(parts[9])

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := s.CmdExec.Execute(countScript)
	count := 1
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count)
	}

	return &GPUMetricsResponse{
		Detected:          true,
		Model:             strings.TrimSpace(parts[0]),
		DriverVersion:     strings.TrimSpace(parts[1]),
		Count:             count,
		MemoryTotal:       memoryTotal,
		MemoryUsed:        memoryUsed,
		MemoryFree:        memoryFree,
		UtilizationGPU:    utilizationGPU,
		UtilizationMemory: utilizationMemory,
		Temperature:       temperature,
		PowerDraw:         powerDraw,
		PowerLimit:        powerLimit,
	}, nil
}
