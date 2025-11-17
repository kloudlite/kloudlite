package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
)

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
