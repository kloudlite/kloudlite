package daemon

import (
	"encoding/json"
	"fmt"
)

// RPC Protocol - JSON-RPC 2.0 style

// Request represents an RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// Response represents an RPC response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// Error codes
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603
)

// RPC Methods
const (
	MethodPing        = "ping"
	MethodInstallCA   = "install_ca"
	MethodUninstallCA = "uninstall_ca"
	MethodHostsAdd    = "hosts_add"
	MethodHostsRemove = "hosts_remove"
	MethodHostsList   = "hosts_list"
	MethodHostsSync   = "hosts_sync"
	MethodHostsClean  = "hosts_clean"
	MethodHostsFlush  = "hosts_flush"
	MethodVPNConnect  = "vpn_connect"
	MethodVPNQuit     = "vpn_quit"
	MethodStatus      = "status"
)

// Request/Response Parameters

// PingParams - no parameters
type PingParams struct{}

// PingResult - simple pong response
type PingResult struct {
	Message string `json:"message"`
}

// InstallCAParams contains parameters for CA installation
type InstallCAParams struct {
	CertPath string `json:"cert_path"`
}

// InstallCAResult contains result of CA installation
type InstallCAResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UninstallCAParams contains parameters for CA uninstallation
type UninstallCAParams struct {
	CertPath string `json:"cert_path"`
}

// UninstallCAResult contains result of CA uninstallation
type UninstallCAResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HostsAddParams contains parameters for adding a host entry
type HostsAddParams struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Comment  string `json:"comment,omitempty"`
}

// HostsAddResult contains result of adding host entry
type HostsAddResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HostsRemoveParams contains parameters for removing a host entry
type HostsRemoveParams struct {
	Hostname string `json:"hostname"`
}

// HostsRemoveResult contains result of removing host entry
type HostsRemoveResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HostsListParams - no parameters
type HostsListParams struct{}

// HostsEntry represents a host entry
type HostsEntry struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Comment  string `json:"comment,omitempty"`
}

// HostsListResult contains list of host entries
type HostsListResult struct {
	Entries []HostsEntry `json:"entries"`
}

// HostsSyncParams - no parameters
type HostsSyncParams struct{}

// HostsSyncResult contains result of syncing hosts
type HostsSyncResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HostsCleanParams - no parameters
type HostsCleanParams struct{}

// HostsCleanResult contains result of cleaning hosts
type HostsCleanResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HostsFlushParams - no parameters
type HostsFlushParams struct{}

// HostsFlushResult contains result of flushing DNS cache
type HostsFlushResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VPNConnectParams contains parameters for VPN connection
type VPNConnectParams struct {
	Token  string `json:"token,omitempty"`
	Server string `json:"server,omitempty"`
}

// VPNConnectResult contains result of VPN connection
type VPNConnectResult struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	SessionID       string `json:"session_id,omitempty"`
	CACertInstalled bool   `json:"ca_cert_installed"`
	CACertError     string `json:"ca_cert_error,omitempty"`
}

// VPNConnectionSetupResult is used internally to communicate VPN setup result
type VPNConnectionSetupResult struct {
	Error           error
	CACertInstalled bool
	CACertError     string
}

// VPNQuitParams - no parameters needed
type VPNQuitParams struct{}

// VPNQuitResult contains result of VPN disconnection
type VPNQuitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// StatusParams - no parameters
type StatusParams struct{}

// ConnectionStatus represents the status of a VPN connection
type ConnectionStatus struct {
	SessionID string `json:"session_id"`
	Server    string `json:"server"`
	Connected bool   `json:"connected"`
	Uptime    int64  `json:"uptime"`
}

// StatusResult contains daemon status
type StatusResult struct {
	Running     bool               `json:"running"`
	Connections []ConnectionStatus `json:"connections"`
}

// Helper functions for creating requests/responses

// NewRequest creates a new RPC request
func NewRequest(id, method string, params interface{}) (*Request, error) {
	var paramsJSON json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		paramsJSON = data
	}

	return &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(id string, result interface{}) (*Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id string, code int, message string, data string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}
