package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
)

// Auto-reconnect constants
const (
	maxConsecutiveFailures = 3
	reconnectPollInterval  = 5 * time.Second // Linear polling interval for reconnection
	hostsPollInterval      = 10 * time.Second
)

// ConnectionHealthMonitor tracks connection health and triggers reconnection
type ConnectionHealthMonitor struct {
	consecutiveFailures int
	mu                  sync.Mutex
	onDisconnected      func()
	wasDisconnected     bool
}

// NewConnectionHealthMonitor creates a new health monitor
func NewConnectionHealthMonitor(onDisconnected func()) *ConnectionHealthMonitor {
	return &ConnectionHealthMonitor{
		onDisconnected: onDisconnected,
	}
}

// RecordFailure records a polling failure and returns true if threshold is exceeded
func (m *ConnectionHealthMonitor) RecordFailure() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consecutiveFailures++
	return m.consecutiveFailures >= maxConsecutiveFailures
}

// RecordSuccess resets the failure count
func (m *ConnectionHealthMonitor) RecordSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.consecutiveFailures > 0 {
		m.consecutiveFailures = 0
	}
	m.wasDisconnected = false
}

// GetFailureCount returns the current failure count
func (m *ConnectionHealthMonitor) GetFailureCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.consecutiveFailures
}

// MarkDisconnected marks the connection as disconnected (only triggers callback once)
func (m *ConnectionHealthMonitor) MarkDisconnected() {
	m.mu.Lock()
	wasAlreadyDisconnected := m.wasDisconnected
	m.wasDisconnected = true
	m.mu.Unlock()

	if !wasAlreadyDisconnected && m.onDisconnected != nil {
		m.onDisconnected()
	}
}

// pollHosts polls the hosts API every 10 seconds and updates /etc/hosts
func (s *Server) pollHosts(ctx context.Context, sessionID string, apiClient *api.Client, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Track current hosts to detect changes
	currentHosts := make(map[string]string) // hostname -> IP

	// Initial fetch
	s.fetchAndUpdateHosts(ctx, sessionID, apiClient, currentHosts)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fetchAndUpdateHosts(ctx, sessionID, apiClient, currentHosts)
		}
	}
}

// fetchAndUpdateHosts fetches hosts from API and updates /etc/hosts
func (s *Server) fetchAndUpdateHosts(ctx context.Context, sessionID string, apiClient *api.Client, currentHosts map[string]string) {
	hosts, err := apiClient.GetHosts()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to fetch hosts: %v\n", sessionID, err)
		return
	}

	// Build map of new hosts
	newHosts := make(map[string]string)
	for _, host := range hosts {
		newHosts[host.Hostname] = host.IP
	}

	// Remove hosts that no longer exist
	for hostname := range currentHosts {
		if _, exists := newHosts[hostname]; !exists {
			fmt.Printf("[Session %s] Removing host: %s\n", sessionID, hostname)
			if err := s.hostsManager.Remove(hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Add or update hosts
	for hostname, ip := range newHosts {
		if currentIP, exists := currentHosts[hostname]; !exists || currentIP != ip {
			fmt.Printf("[Session %s] Adding/updating host: %s -> %s\n", sessionID, hostname, ip)
			if err := s.hostsManager.Add(hostname, ip, fmt.Sprintf("# kltun session %s", sessionID)); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to add host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Update current hosts map
	for k := range currentHosts {
		delete(currentHosts, k)
	}
	for k, v := range newHosts {
		currentHosts[k] = v
	}

	fmt.Printf("[Session %s] ✓ Hosts updated (%d entries)\n", sessionID, len(currentHosts))
}

// pollHostsFromTunnel polls the hosts from tunnel server every 10 seconds and updates /etc/hosts
// It now includes health monitoring to trigger reconnection when the tunnel becomes unreachable
func (s *Server) pollHostsFromTunnel(ctx context.Context, sessionID string, tunnelClient *api.TunnelClient, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(hostsPollInterval)
	defer ticker.Stop()

	// Track current hosts to detect changes
	currentHosts := make(map[string]string) // hostname -> IP

	// Get the connection object for state updates
	s.connMutex.RLock()
	conn := s.connections[sessionID]
	s.connMutex.RUnlock()

	// Create health monitor with disconnection callback
	healthMonitor := NewConnectionHealthMonitor(func() {
		fmt.Printf("[Session %s] Connection lost - WorkMachine may be stopped\n", sessionID)
		fmt.Printf("[Session %s] Entering reconnection mode, will auto-reconnect when available...\n", sessionID)

		if conn != nil {
			conn.SetState(StateReconnecting)
			// Signal the reconnection goroutine
			select {
			case conn.ReconnectChan <- struct{}{}:
			default:
				// Channel already has a signal pending
			}
		}
	})

	// Initial fetch
	s.fetchAndUpdateHostsFromTunnelWithMonitor(ctx, sessionID, tunnelClient, currentHosts, healthMonitor)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip polling if we're in reconnecting state (reconnection loop handles it)
			if conn != nil && conn.GetState() == StateReconnecting {
				continue
			}
			s.fetchAndUpdateHostsFromTunnelWithMonitor(ctx, sessionID, tunnelClient, currentHosts, healthMonitor)
		}
	}
}

// fetchAndUpdateHostsFromTunnel fetches hosts from tunnel server and updates /etc/hosts
func (s *Server) fetchAndUpdateHostsFromTunnel(ctx context.Context, sessionID string, tunnelClient *api.TunnelClient, currentHosts map[string]string) {
	hosts, err := tunnelClient.GetHosts()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to fetch hosts from tunnel server: %v\n", sessionID, err)
		return
	}

	// Build map of new hosts
	newHosts := make(map[string]string)
	for _, host := range hosts {
		newHosts[host.Hostname] = host.IP
	}

	// Remove hosts that no longer exist
	for hostname := range currentHosts {
		if _, exists := newHosts[hostname]; !exists {
			fmt.Printf("[Session %s] Removing host: %s\n", sessionID, hostname)
			if err := s.hostsManager.Remove(hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Add or update hosts
	for hostname, ip := range newHosts {
		if currentIP, exists := currentHosts[hostname]; !exists || currentIP != ip {
			fmt.Printf("[Session %s] Adding/updating host: %s -> %s\n", sessionID, hostname, ip)
			if err := s.hostsManager.Add(hostname, ip, fmt.Sprintf("# kltun session %s", sessionID)); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to add host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Update current hosts map
	for k := range currentHosts {
		delete(currentHosts, k)
	}
	for k, v := range newHosts {
		currentHosts[k] = v
	}

	fmt.Printf("[Session %s] ✓ Hosts updated from tunnel server (%d entries)\n", sessionID, len(currentHosts))
}

// fetchAndUpdateHostsFromTunnelWithMonitor fetches hosts with health monitoring for auto-reconnect
func (s *Server) fetchAndUpdateHostsFromTunnelWithMonitor(ctx context.Context, sessionID string, tunnelClient *api.TunnelClient, currentHosts map[string]string, healthMonitor *ConnectionHealthMonitor) {
	hosts, err := tunnelClient.GetHosts()
	if err != nil {
		failureCount := healthMonitor.GetFailureCount() + 1
		fmt.Printf("[Session %s] Warning: Failed to fetch hosts (%d/%d failures): %v\n",
			sessionID, failureCount, maxConsecutiveFailures, err)

		if healthMonitor.RecordFailure() {
			healthMonitor.MarkDisconnected()
		}
		return
	}

	// Success - reset failure count
	healthMonitor.RecordSuccess()

	// Build map of new hosts
	newHosts := make(map[string]string)
	for _, host := range hosts {
		newHosts[host.Hostname] = host.IP
	}

	// Remove hosts that no longer exist
	for hostname := range currentHosts {
		if _, exists := newHosts[hostname]; !exists {
			fmt.Printf("[Session %s] Removing host: %s\n", sessionID, hostname)
			if err := s.hostsManager.Remove(hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Add or update hosts
	for hostname, ip := range newHosts {
		if currentIP, exists := currentHosts[hostname]; !exists || currentIP != ip {
			fmt.Printf("[Session %s] Adding/updating host: %s -> %s\n", sessionID, hostname, ip)
			if err := s.hostsManager.Add(hostname, ip, fmt.Sprintf("# kltun session %s", sessionID)); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to add host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Update current hosts map
	for k := range currentHosts {
		delete(currentHosts, k)
	}
	for k, v := range newHosts {
		currentHosts[k] = v
	}

	fmt.Printf("[Session %s] ✓ Hosts updated from tunnel server (%d entries)\n", sessionID, len(currentHosts))
}
