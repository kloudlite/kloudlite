package daemon

import (
	"encoding/json"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
)

// handleInstallCA handles the CA certificate installation request
func (s *Server) handleInstallCA(req *Request) *Response {
	var params InstallCAParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Install to all stores
	stores := []string{"system", "nss", "java"}
	if err := truststore.InstallAll(params.CertPath, stores); err != nil {
		result := InstallCAResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := InstallCAResult{Success: true, Message: "CA certificate installed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleUninstallCA handles the CA certificate uninstallation request
func (s *Server) handleUninstallCA(req *Request) *Response {
	var params UninstallCAParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Uninstall from all stores
	stores := []string{"system", "nss", "java"}
	if err := truststore.UninstallAll(params.CertPath, stores); err != nil {
		result := UninstallCAResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := UninstallCAResult{Success: true, Message: "CA certificate uninstalled successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}
