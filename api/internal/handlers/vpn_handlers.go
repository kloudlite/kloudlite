package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"go.uber.org/zap"
)

// VPNHandlers handles HTTP requests for VPN connections
type VPNHandlers struct {
	logger    *zap.Logger
	jwtSecret string
}

// NewVPNHandlers creates a new VPNHandlers
func NewVPNHandlers(logger *zap.Logger, jwtSecret string) *VPNHandlers {
	return &VPNHandlers{
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// VPNConnectResponse represents the VPN connection configuration
type VPNConnectResponse struct {
	CACert   string     `json:"ca_cert"`
	WGConfig string     `json:"wg_config"`
	Hosts    []HostEntry `json:"hosts"`
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// Mock CA Certificate for development
const mockCACert = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKJ7VZ3qN0NYMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMjQwMTAxMDAwMDAwWhcNMjUwMTAxMDAwMDAwWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAwxyz7Z7mN3qQCWQJQJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJ
ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJ
ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJ
ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJ
ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJ
ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJZJwIDAQAB
o1AwTjAdBgNVHQ4EFgQUQJ9Z3qN0NYJ7VZ3qN0NYJ7VZ3qMwHwYDVR0jBBgwFoAU
QJ9Z3qN0NYJ7VZ3qN0NYJ7VZ3qMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsF
AAOCAQEAqN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ
7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3
qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0N
YJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7V
Z3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN
0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qN0NYJ
7VZ3qN0NYJ7VZ3qN0NYJ7VZ3qA==
-----END CERTIFICATE-----`

// Mock WireGuard configuration for development
const mockWGConfig = `[Interface]
PrivateKey = mOcKPr1v4t3K3yT0M0cKPr1v4t3K3yT0M0cKPr1v4t3K3yT0=
Address = 10.42.1.2/24
DNS = 10.43.0.10
ListenPort = 51820

[Peer]
PublicKey = s3Rv3rPu8l1cK3yT0M0cKs3Rv3rPu8l1cK3yT0M0cKs3R=
AllowedIPs = 10.42.0.0/16, 10.43.0.0/16
Endpoint = 127.0.0.1:51821
PersistentKeepalive = 25`

// GetVPNConnect handles GET /api/vpn/connect
// This is a mock implementation that returns test VPN configuration
// In production, this would fetch actual WireGuard config and certificates from Kubernetes
func (h *VPNHandlers) GetVPNConnect(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("VPN connect: Missing authorization header")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header required"})
		return
	}

	// Extract token from Bearer header
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		h.logger.Warn("VPN connect: Invalid authorization header format")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid authorization header format"})
		return
	}

	tokenString := authHeader[len(bearerPrefix):]

	// Parse and validate the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &ConnectionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		h.logger.Warn("VPN connect: Token validation failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return
	}

	claims, ok := token.Claims.(*ConnectionTokenClaims)
	if !ok || !token.Valid {
		h.logger.Warn("VPN connect: Invalid token claims")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token claims"})
		return
	}

	h.logger.Info("VPN connect: Token validated", zap.String("user", claims.Email), zap.String("tokenId", claims.TokenID))

	// TODO: In production, fetch actual configuration from Kubernetes resources:
	// 1. Get WireGuard configuration from WireGuardPeer CRD
	// 2. Get CA certificate from Secret
	// 3. Get host entries from Service/Ingress configurations

	// Return mock VPN configuration
	response := VPNConnectResponse{
		CACert:   mockCACert,
		WGConfig: mockWGConfig,
		Hosts: []HostEntry{
			{
				Hostname: "workspace-dev.kloudlite.local",
				IP:       "10.42.1.10",
			},
			{
				Hostname: "api.kloudlite.local",
				IP:       "10.42.1.11",
			},
			{
				Hostname: "console.kloudlite.local",
				IP:       "10.42.1.12",
			},
			{
				Hostname: "auth.kloudlite.local",
				IP:       "10.42.1.13",
			},
		},
	}

	h.logger.Info("VPN connect: Returning mock configuration", zap.String("user", claims.Email))

	c.JSON(http.StatusOK, response)
}
