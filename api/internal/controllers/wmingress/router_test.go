package wmingress

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestRouterConcurrentUpdates(t *testing.T) {
	logger := zap.NewNop()
	router := NewRouter(logger, "")

	// Test 1: Concurrent route updates
	t.Run("ConcurrentRouteUpdates", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 100
		updatesPerGoroutine := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < updatesPerGoroutine; j++ {
					routes := []*Route{
						{
							Host:       "example.com",
							Path:       "/",
							PathType:   networkingv1.PathTypePrefix,
							BackendURL: "http://backend:8080",
						},
					}

					if err := router.UpdateRoutes(routes); err != nil {
						t.Errorf("UpdateRoutes failed: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify metrics
		metrics := router.GetMetrics()
		if attempts, ok := metrics["concurrent_update_attempts"].(int64); ok {
			expectedAttempts := int64(numGoroutines * updatesPerGoroutine)
			if attempts != expectedAttempts {
				t.Errorf("Expected %d concurrent update attempts, got %d", expectedAttempts, attempts)
			}
		} else {
			t.Error("concurrent_update_attempts not found in metrics")
		}
	})

	// Test 2: Concurrent updates and reads
	t.Run("ConcurrentUpdatesAndReads", func(t *testing.T) {
		var wg sync.WaitGroup
		numUpdateGoroutines := 10
		numReadGoroutines := 50

		// Start update goroutines
		for i := 0; i < numUpdateGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 100; j++ {
					routes := []*Route{
						{
							Host:       "test.com",
							Path:       "/",
							PathType:   networkingv1.PathTypePrefix,
							BackendURL: "http://backend:8080",
						},
					}

					if err := router.UpdateRoutes(routes); err != nil {
						t.Errorf("UpdateRoutes failed: %v", err)
					}
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Start read goroutines
		for i := 0; i < numReadGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 100; j++ {
					_ = router.GetRoutes()
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		wg.Wait()

		// Verify final state is consistent
		routes := router.GetRoutes()
		if len(routes) != 1 {
			t.Errorf("Expected 1 route, got %d", len(routes))
		}
	})

	// Test 3: Concurrent HTTP requests during route updates
	t.Run("ConcurrentHTTPRequestsDuringUpdates", func(t *testing.T) {
		var wg sync.WaitGroup
		numRequestGoroutines := 20
		numUpdateGoroutines := 5

		// Start request goroutines
		for i := 0; i < numRequestGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 50; j++ {
					req := httptest.NewRequest("GET", "http://test.com/", nil)
					w := httptest.NewRecorder()

					router.ServeHTTP(w, req)
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Start update goroutines
		for i := 0; i < numUpdateGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 20; j++ {
					routes := []*Route{
						{
							Host:       "test.com",
							Path:       "/",
							PathType:   networkingv1.PathTypePrefix,
							BackendURL: "http://backend:8080",
						},
					}

					if err := router.UpdateRoutes(routes); err != nil {
						t.Errorf("UpdateRoutes failed: %v", err)
					}
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		// Verify no data races occurred (would panic or fail on race detector)
		metrics := router.GetMetrics()
		t.Logf("Final metrics: %+v", metrics)
	})

	// Test 4: Conflict detection under high contention
	t.Run("ConflictDetectionUnderContention", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		updatesPerGoroutine := 20

		// Reset metrics
		router.metricsMutex.Lock()
		router.concurrentUpdateAttempts = 0
		router.concurrentUpdateConflicts = 0
		router.metricsMutex.Unlock()

		// Launch many concurrent updates to simulate contention
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < updatesPerGoroutine; j++ {
					routes := []*Route{
						{
							Host:       fmt.Sprintf("host-%d.com", id),
							Path:       "/",
							PathType:   networkingv1.PathTypePrefix,
							BackendURL: fmt.Sprintf("http://backend-%d:8080", id),
						},
					}

					if err := router.UpdateRoutes(routes); err != nil {
						t.Errorf("UpdateRoutes failed: %v", err)
					}
					// Add small delay to increase contention
					time.Sleep(time.Microsecond * 10)
				}
			}(i)
		}

		wg.Wait()

		// Verify metrics were collected
		metrics := router.GetMetrics()
		if attempts, ok := metrics["concurrent_update_attempts"].(int64); ok {
			expectedAttempts := int64(numGoroutines * updatesPerGoroutine)
			if attempts < expectedAttempts {
				t.Errorf("Expected at least %d concurrent update attempts, got %d", expectedAttempts, attempts)
			}
			t.Logf("Concurrent update attempts: %d", attempts)
		}

		// Check for conflict detection (may not always trigger depending on timing)
		if conflicts, ok := metrics["concurrent_update_conflicts"].(int64); ok {
			t.Logf("Detected %d concurrent update conflicts", conflicts)
			// Conflicts are optional - we just verify the tracking works
		}
	})
}

func TestRouterGetRoutesImmutability(t *testing.T) {
	logger := zap.NewNop()
	router := NewRouter(logger, "")

	// Set initial routes
	initialRoutes := []*Route{
		{
			Host:       "example.com",
			Path:       "/",
			PathType:   networkingv1.PathTypePrefix,
			BackendURL: "http://backend:8080",
		},
	}

	if err := router.UpdateRoutes(initialRoutes); err != nil {
		t.Fatalf("UpdateRoutes failed: %v", err)
	}

	// Get routes and modify them
	routes := router.GetRoutes()
	routes[0].BackendURL = "http://modified:9999"

	// Get routes again and verify original is unchanged
	routesAgain := router.GetRoutes()
	if routesAgain[0].BackendURL == "http://modified:9999" {
		t.Error("GetRoutes returned a mutable reference - modification affected internal state")
	}
}
