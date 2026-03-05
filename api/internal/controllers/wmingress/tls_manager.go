package wmingress

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TLSCertificate represents a TLS certificate
type TLSCertificate struct {
	Hosts    []string
	CertPEM  []byte
	KeyPEM   []byte
	SecretID string
}

// TLSManager manages TLS certificates
type TLSManager struct {
	logger       *zap.Logger
	certificates map[string]*TLSCertificate // host -> certificate
	certsMutex   sync.RWMutex
}

// NewTLSManager creates a new TLS manager
func NewTLSManager(logger *zap.Logger) *TLSManager {
	return &TLSManager{
		logger:       logger,
		certificates: make(map[string]*TLSCertificate),
	}
}

// UpdateCertificates updates the certificate store atomically
// Uses the same try-lock pattern as router to detect concurrent updates
func (m *TLSManager) UpdateCertificates(certificates map[string]*TLSCertificate) error {
	// Detect potential concurrent updates by checking if we can acquire the lock immediately
	lockAcquired := make(chan bool, 1)
	go func() {
		m.certsMutex.Lock()
		lockAcquired <- true
	}()

	select {
	case <-lockAcquired:
		// Lock acquired immediately, no conflict
	case <-time.After(10 * time.Millisecond):
		// Lock acquisition took >10ms, indicating potential contention
		// Wait for lock to be acquired
		<-lockAcquired

		m.logger.Warn("Detected concurrent TLS certificate update - potential race condition")
	}

	// Atomic swap of the certificates map
	oldCount := len(m.certificates)
	m.certificates = certificates

	m.logger.Info("TLS certificates updated atomically",
		zap.Int("old_count", oldCount),
		zap.Int("new_count", len(certificates)),
	)

	m.certsMutex.Unlock()

	return nil
}

// GetCertificates returns a copy of the current certificates atomically
func (m *TLSManager) GetCertificates() map[string]*TLSCertificate {
	m.certsMutex.RLock()
	defer m.certsMutex.RUnlock()

	// Return a deep copy to prevent external modifications
	certificatesCopy := make(map[string]*TLSCertificate, len(m.certificates))
	for host, cert := range m.certificates {
		certificatesCopy[host] = &TLSCertificate{
			Hosts:    append([]string{}, cert.Hosts...),
			CertPEM:  append([]byte{}, cert.CertPEM...),
			KeyPEM:   append([]byte{}, cert.KeyPEM...),
			SecretID: cert.SecretID,
		}
	}
	return certificatesCopy
}

// GetMetrics returns TLS manager metrics for monitoring
func (m *TLSManager) GetMetrics() map[string]interface{} {
	m.certsMutex.RLock()
	defer m.certsMutex.RUnlock()

	return map[string]interface{}{
		"certificate_count": len(m.certificates),
	}
}

// GetTLSConfig returns a tls.Config with certificate resolution
func (m *TLSManager) GetTLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: m.getCertificate,
		MinVersion:     tls.VersionTLS12,
	}
}

// getCertificate implements SNI-based certificate selection
func (m *TLSManager) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.certsMutex.RLock()
	defer m.certsMutex.RUnlock()

	serverName := hello.ServerName
	if serverName == "" {
		m.logger.Warn("Client did not provide SNI")
		return nil, fmt.Errorf("no SNI provided")
	}

	// Look for exact match
	if certData, ok := m.certificates[serverName]; ok {
		return m.loadCertificate(certData)
	}

	// Look for wildcard match
	for host, certData := range m.certificates {
		if m.matchesWildcard(host, serverName) {
			return m.loadCertificate(certData)
		}
	}

	m.logger.Warn("No certificate found for host", zap.String("host", serverName))
	return nil, fmt.Errorf("no certificate found for host: %s", serverName)
}

// loadCertificate loads a certificate from PEM data
func (m *TLSManager) loadCertificate(certData *TLSCertificate) (*tls.Certificate, error) {
	cert, err := tls.X509KeyPair(certData.CertPEM, certData.KeyPEM)
	if err != nil {
		m.logger.Error("Failed to load certificate",
			zap.String("secret", certData.SecretID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	return &cert, nil
}

// matchesWildcard checks if a wildcard host matches the server name
func (m *TLSManager) matchesWildcard(host, serverName string) bool {
	// Check for wildcard pattern (*.example.com)
	if len(host) > 2 && host[0] == '*' && host[1] == '.' {
		// Extract the domain part
		domain := host[2:]

		// Check if serverName ends with .domain
		if len(serverName) > len(domain)+1 {
			suffix := serverName[len(serverName)-len(domain)-1:]
			return suffix == "."+domain
		}
	}

	return false
}
