package handlers

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// DNSServer serves DNS queries using the hosts cache
type DNSServer struct {
	logger      *zap.Logger
	hostsCache  *HostsCache
	upstreamDNS string // CoreDNS address (e.g., "10.43.0.10:53")
	listenAddr  string // e.g., ":53"

	udpServer *dns.Server
	tcpServer *dns.Server
}

// DNSServerConfig holds configuration for the DNS server
type DNSServerConfig struct {
	ListenAddr  string // Listen address (e.g., ":53")
	UpstreamDNS string // Upstream DNS server (e.g., "10.43.0.10:53")
}

// NewDNSServer creates a new DNS server
func NewDNSServer(logger *zap.Logger, hostsCache *HostsCache, cfg DNSServerConfig) *DNSServer {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":53"
	}
	if cfg.UpstreamDNS == "" {
		cfg.UpstreamDNS = "10.43.0.10:53"
	}

	return &DNSServer{
		logger:      logger,
		hostsCache:  hostsCache,
		upstreamDNS: cfg.UpstreamDNS,
		listenAddr:  cfg.ListenAddr,
	}
}

// Start starts the DNS server (both UDP and TCP)
func (d *DNSServer) Start(ctx context.Context) error {
	// Create DNS handler
	handler := dns.HandlerFunc(d.handleQuery)

	// Start UDP server
	d.udpServer = &dns.Server{
		Addr:    d.listenAddr,
		Net:     "udp",
		Handler: handler,
	}

	// Start TCP server
	d.tcpServer = &dns.Server{
		Addr:    d.listenAddr,
		Net:     "tcp",
		Handler: handler,
	}

	// Start servers in goroutines
	errChan := make(chan error, 2)

	go func() {
		d.logger.Info("starting DNS server (UDP)", zap.String("addr", d.listenAddr))
		if err := d.udpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		d.logger.Info("starting DNS server (TCP)", zap.String("addr", d.listenAddr))
		if err := d.tcpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return d.Stop()
	case err := <-errChan:
		return err
	}
}

// Stop stops the DNS server
func (d *DNSServer) Stop() error {
	var lastErr error
	if d.udpServer != nil {
		if err := d.udpServer.Shutdown(); err != nil {
			d.logger.Error("failed to shutdown UDP DNS server", zap.Error(err))
			lastErr = err
		}
	}
	if d.tcpServer != nil {
		if err := d.tcpServer.Shutdown(); err != nil {
			d.logger.Error("failed to shutdown TCP DNS server", zap.Error(err))
			lastErr = err
		}
	}
	return lastErr
}

// handleQuery handles incoming DNS queries
func (d *DNSServer) handleQuery(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = false
	m.RecursionAvailable = true

	for _, q := range r.Question {
		d.logger.Debug("received DNS query",
			zap.String("name", q.Name),
			zap.String("type", dns.TypeToString[q.Qtype]))

		switch q.Qtype {
		case dns.TypeA:
			d.handleARecord(m, q)
		case dns.TypeAAAA:
			// We don't serve AAAA records from cache, forward to upstream
			d.forwardToUpstream(w, r)
			return
		default:
			// Forward other query types to upstream
			d.forwardToUpstream(w, r)
			return
		}
	}

	// If we found an answer, respond; otherwise forward to upstream
	if len(m.Answer) > 0 {
		d.logger.Debug("responding from cache",
			zap.Int("answers", len(m.Answer)))
		w.WriteMsg(m)
	} else {
		d.forwardToUpstream(w, r)
	}
}

// handleARecord handles A record queries
func (d *DNSServer) handleARecord(m *dns.Msg, q dns.Question) {
	// Normalize the query name (remove trailing dot)
	queryName := strings.TrimSuffix(strings.ToLower(q.Name), ".")

	// Look up in hosts cache
	hosts := d.hostsCache.GetHosts()
	for _, host := range hosts {
		if strings.ToLower(host.Hostname) == queryName {
			// Parse the IP
			ip := net.ParseIP(host.IP)
			if ip == nil {
				d.logger.Warn("invalid IP in hosts cache",
					zap.String("hostname", host.Hostname),
					zap.String("ip", host.IP))
				continue
			}

			// Only respond with IPv4 for A records
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}

			rr := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60, // Short TTL since hosts can change
				},
				A: ipv4,
			}
			m.Answer = append(m.Answer, rr)

			d.logger.Debug("found in hosts cache",
				zap.String("hostname", host.Hostname),
				zap.String("ip", host.IP))
			return
		}
	}
}

// forwardToUpstream forwards the query to the upstream DNS server
func (d *DNSServer) forwardToUpstream(w dns.ResponseWriter, r *dns.Msg) {
	client := &dns.Client{
		Net:     "udp",
		Timeout: 5 * time.Second,
	}

	resp, _, err := client.Exchange(r, d.upstreamDNS)
	if err != nil {
		d.logger.Debug("failed to forward to upstream",
			zap.String("upstream", d.upstreamDNS),
			zap.Error(err))

		// Try TCP if UDP fails
		client.Net = "tcp"
		resp, _, err = client.Exchange(r, d.upstreamDNS)
		if err != nil {
			d.logger.Error("failed to forward to upstream (TCP)",
				zap.String("upstream", d.upstreamDNS),
				zap.Error(err))

			// Send SERVFAIL response
			m := new(dns.Msg)
			m.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(m)
			return
		}
	}

	w.WriteMsg(resp)
}
