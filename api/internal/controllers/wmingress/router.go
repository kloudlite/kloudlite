package wmingress

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
)

// Route represents a routing rule
type Route struct {
	Host        string
	Path        string
	PathType    networkingv1.PathType
	BackendURL  string
	IngressName string
	Namespace   string
}

// Router handles HTTP request routing
type Router struct {
	logger      *zap.Logger
	routes      []*Route
	routesMutex sync.RWMutex
	httpServer  *http.Server
	httpsServer *http.Server
}

// NewRouter creates a new Router
func NewRouter(logger *zap.Logger) *Router {
	return &Router{
		logger: logger,
		routes: []*Route{},
	}
}

// UpdateRoutes updates the routing table
func (r *Router) UpdateRoutes(routes []*Route) error {
	r.routesMutex.Lock()
	defer r.routesMutex.Unlock()

	r.routes = routes
	r.logger.Info("Routes updated", zap.Int("count", len(routes)))

	return nil
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(req *http.Request) bool {
	return strings.EqualFold(req.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade")
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find matching route (hold lock only during route lookup)
	r.routesMutex.RLock()
	route := r.findMatchingRoute(req)
	r.routesMutex.RUnlock()

	if route == nil {
		http.Error(w, "No route found", http.StatusNotFound)
		return
	}

	isWS := isWebSocketRequest(req)
	if isWS {
		r.logger.Info("Routing WebSocket request",
			zap.String("host", req.Host),
			zap.String("path", req.URL.Path),
			zap.String("backend", route.BackendURL),
		)
	}

	// Create reverse proxy
	backendURL, err := url.Parse(route.BackendURL)
	if err != nil {
		r.logger.Error("Failed to parse backend URL",
			zap.String("url", route.BackendURL),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Handle WebSocket connections specially
	if isWS {
		r.proxyWebSocket(w, req, backendURL)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Customize director to preserve original host and add forwarding headers
	originalDirector := proxy.Director
	proxy.Director = func(proxyReq *http.Request) {
		originalHost := proxyReq.Host
		originalDirector(proxyReq)
		// Set X-Forwarded headers so the backend app knows the original request details
		proxyReq.Header.Set("X-Forwarded-Host", originalHost)
		proxyReq.Header.Set("X-Forwarded-Proto", "https")
		// Keep the original host header so apps can construct correct URLs
		proxyReq.Host = originalHost
	}

	// Enable WebSocket support by flushing responses immediately
	proxy.FlushInterval = -1

	// Custom transport to handle WebSocket upgrades
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false, // Disable HTTP/2 for WebSocket compatibility
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		r.logger.Error("Proxy error",
			zap.String("backend", route.BackendURL),
			zap.Error(err),
		)
		http.Error(w, "Bad gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, req)
}

// proxyWebSocket handles WebSocket upgrade requests by hijacking the connection
func (r *Router) proxyWebSocket(w http.ResponseWriter, req *http.Request, backendURL *url.URL) {
	// Determine backend address
	backendHost := backendURL.Host
	if backendURL.Port() == "" {
		if backendURL.Scheme == "https" || backendURL.Scheme == "wss" {
			backendHost = backendHost + ":443"
		} else {
			backendHost = backendHost + ":80"
		}
	}

	r.logger.Info("Starting WebSocket proxy",
		zap.String("backendHost", backendHost),
		zap.String("backendURL", backendURL.String()),
		zap.String("path", req.URL.Path),
	)

	// Connect to backend
	var backendConn net.Conn
	var err error

	if backendURL.Scheme == "https" || backendURL.Scheme == "wss" {
		backendConn, err = tls.Dial("tcp", backendHost, &tls.Config{
			InsecureSkipVerify: true, // Backend is internal, skip verification
		})
	} else {
		backendConn, err = net.DialTimeout("tcp", backendHost, 30*time.Second)
	}

	if err != nil {
		r.logger.Error("Failed to connect to WebSocket backend",
			zap.String("backend", backendHost),
			zap.Error(err),
		)
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	r.logger.Info("Connected to backend for WebSocket", zap.String("backend", backendHost))

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		r.logger.Error("ResponseWriter does not support hijacking")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		r.logger.Error("Failed to hijack connection", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the original request to backend
	// Preserve original host for origin validation but route to backend
	originalHost := req.Host
	req.URL.Scheme = "http"
	if backendURL.Scheme == "https" || backendURL.Scheme == "wss" {
		req.URL.Scheme = "https"
	}
	req.URL.Host = backendURL.Host
	// Keep the original Host header for origin validation (code-server checks this)
	req.Host = originalHost
	// Add X-Forwarded headers
	req.Header.Set("X-Forwarded-Host", originalHost)
	req.Header.Set("X-Forwarded-Proto", "https")

	// Write the request to backend
	if err := req.Write(backendConn); err != nil {
		r.logger.Error("Failed to write request to backend", zap.Error(err))
		return
	}

	// Read the response from backend
	backendBuf := bufio.NewReader(backendConn)
	resp, err := http.ReadResponse(backendBuf, req)
	if err != nil {
		r.logger.Error("Failed to read response from backend", zap.Error(err))
		return
	}

	r.logger.Info("Backend WebSocket response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
	)

	// Write the response back to client
	if err := resp.Write(clientConn); err != nil {
		r.logger.Error("Failed to write response to client", zap.Error(err))
		return
	}

	// If upgrade was successful, start bidirectional copy
	if resp.StatusCode == http.StatusSwitchingProtocols {
		r.logger.Info("WebSocket connection established",
			zap.String("backend", backendHost),
		)

		// Copy data bidirectionally
		errChan := make(chan error, 2)

		go func() {
			// Copy from client to backend
			// First drain any buffered data from client reader
			if clientBuf.Reader.Buffered() > 0 {
				if _, err := io.CopyN(backendConn, clientBuf.Reader, int64(clientBuf.Reader.Buffered())); err != nil {
					errChan <- err
					return
				}
			}
			_, err := io.Copy(backendConn, clientConn)
			errChan <- err
		}()

		go func() {
			// Copy from backend to client
			// First drain any buffered data from backend
			if backendBuf.Buffered() > 0 {
				if _, err := io.CopyN(clientConn, backendBuf, int64(backendBuf.Buffered())); err != nil {
					errChan <- err
					return
				}
			}
			_, err := io.Copy(clientConn, backendConn)
			errChan <- err
		}()

		// Wait for either direction to finish
		<-errChan
	}
}

// findMatchingRoute finds a route matching the request
func (r *Router) findMatchingRoute(req *http.Request) *Route {
	host := req.Host

	if strings.HasPrefix(host, "vpn.") {
		return &Route{
			Host:       host,
			Path:       "/",
			PathType:   networkingv1.PathTypeExact,
			BackendURL: "localhost:51443",
		}
	}

	// Strip port from host if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	path := req.URL.Path

	var bestMatch *Route
	var bestMatchScore int

	for _, route := range r.routes {
		// Check host match
		if route.Host != "" && !r.matchesHost(route.Host, host) {
			continue
		}

		// Check path match
		if !r.matchesPath(route.Path, route.PathType, path) {
			continue
		}

		// Calculate match score (more specific routes have higher scores)
		score := r.calculateMatchScore(route, host, path)
		if score > bestMatchScore {
			bestMatch = route
			bestMatchScore = score
		}
	}

	return bestMatch
}

// matchesHost checks if the route host matches the request host
func (r *Router) matchesHost(routeHost, requestHost string) bool {
	// Exact match
	if routeHost == requestHost {
		return true
	}

	// Wildcard match (*.example.com)
	if strings.HasPrefix(routeHost, "*.") {
		domain := routeHost[2:]
		return strings.HasSuffix(requestHost, "."+domain)
	}

	return false
}

// matchesPath checks if the route path matches the request path
func (r *Router) matchesPath(routePath string, pathType networkingv1.PathType, requestPath string) bool {
	if routePath == "" {
		routePath = "/"
	}

	switch pathType {
	case networkingv1.PathTypeExact:
		return routePath == requestPath

	case networkingv1.PathTypePrefix:
		// Prefix match
		if routePath == "/" {
			return true
		}
		// Ensure trailing slash handling
		if strings.HasPrefix(requestPath, routePath) {
			// Match if exact or followed by /
			if len(requestPath) == len(routePath) {
				return true
			}
			if requestPath[len(routePath)] == '/' {
				return true
			}
		}
		return false

	case networkingv1.PathTypeImplementationSpecific:
		// Treat as prefix for now
		return strings.HasPrefix(requestPath, routePath)

	default:
		return false
	}
}

// calculateMatchScore calculates a match score for route specificity
func (r *Router) calculateMatchScore(route *Route, host, path string) int {
	score := 0

	// Host specificity
	if route.Host != "" {
		if strings.HasPrefix(route.Host, "*.") {
			score += 10 // Wildcard host
		} else {
			score += 20 // Exact host
		}
	}

	// Path specificity
	if route.PathType == networkingv1.PathTypeExact {
		score += 30 // Exact path
	} else {
		score += len(route.Path) // Longer prefix = more specific
	}

	return score
}

// StartHTTP starts the HTTP server
func (r *Router) StartHTTP(ctx context.Context, port int) error {
	r.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		r.logger.Info("Starting HTTP server", zap.Int("port", port))
		if err := r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Setup graceful shutdown
	go func() {
		<-ctx.Done()
		r.logger.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30)
		defer cancel()
		if err := r.httpServer.Shutdown(shutdownCtx); err != nil {
			r.logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}()

	return nil
}

// StartHTTPS starts the HTTPS server
func (r *Router) StartHTTPS(ctx context.Context, port int, tlsManager *TLSManager) error {
	tlsConfig := tlsManager.GetTLSConfig()
	// Disable HTTP/2 to allow WebSocket connections
	// WebSocket requires HTTP/1.1 for the upgrade handshake
	tlsConfig.NextProtos = []string{"http/1.1"}

	r.httpsServer = &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	go func() {
		r.logger.Info("Starting HTTPS server", zap.Int("port", port))
		// Empty cert/key files since we're using TLSConfig
		if err := r.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			r.logger.Error("HTTPS server error", zap.Error(err))
		}
	}()

	// Setup graceful shutdown
	go func() {
		<-ctx.Done()
		r.logger.Info("Shutting down HTTPS server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30)
		defer cancel()
		if err := r.httpsServer.Shutdown(shutdownCtx); err != nil {
			r.logger.Error("HTTPS server shutdown error", zap.Error(err))
		}
	}()

	return nil
}
