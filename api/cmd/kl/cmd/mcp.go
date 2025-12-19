package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/devbox"
	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/workspace"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MCP Protocol Types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCP Capability Types
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool Types
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCP Server
type MCPServer struct {
	client *workspace.Client
}

func NewMCPServer() (*MCPServer, error) {
	if err := InitClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize workspace client: %w", err)
	}
	return &MCPServer{client: WsClient}, nil
}

func (s *MCPServer) GetTools() []Tool {
	return []Tool{
		// Package Management
		{
			Name:        "kl_pkg_search",
			Description: "Search for Nix packages available to install",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {Type: "string", Description: "Search query for package name"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "kl_pkg_add",
			Description: "Add one or more packages to the workspace. Uses the latest version from nixos-unstable channel.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"packages": {Type: "string", Description: "Comma-separated list of package names to add (e.g., 'nodejs,python3,go')"},
				},
				Required: []string{"packages"},
			},
		},
		{
			Name:        "kl_pkg_install",
			Description: "Install a specific package with version control options",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"package": {Type: "string", Description: "Package name to install"},
					"version": {Type: "string", Description: "Specific version (optional)"},
					"channel": {Type: "string", Description: "Nixpkgs channel (optional, e.g., 'nixos-unstable')"},
				},
				Required: []string{"package"},
			},
		},
		{
			Name:        "kl_pkg_uninstall",
			Description: "Remove a package from the workspace",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"package": {Type: "string", Description: "Package name to remove"},
				},
				Required: []string{"package"},
			},
		},
		{
			Name:        "kl_pkg_list",
			Description: "List all packages installed in the workspace with their status",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		// Workspace Info
		{
			Name:        "kl_status",
			Description: "Show workspace status including phase, resources, and access URLs",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "kl_config_get",
			Description: "Get workspace configuration value",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"key": {Type: "string", Description: "Configuration key (e.g., 'display-name', 'auto-stop', 'idle-timeout', 'git.user-name')"},
				},
			},
		},
		{
			Name:        "kl_config_set",
			Description: "Set workspace configuration value",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"key":   {Type: "string", Description: "Configuration key"},
					"value": {Type: "string", Description: "Value to set"},
				},
				Required: []string{"key", "value"},
			},
		},
		// Environment Management
		{
			Name:        "kl_env_list",
			Description: "List all available environments that can be connected to",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "kl_env_connect",
			Description: "Connect the workspace to an environment for service access and intercepts",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"environment": {Type: "string", Description: "Environment name to connect to"},
				},
				Required: []string{"environment"},
			},
		},
		{
			Name:        "kl_env_disconnect",
			Description: "Disconnect the workspace from the currently connected environment",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "kl_env_status",
			Description: "Show the current environment connection status and active intercepts",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		// Service Interception
		{
			Name:        "kl_intercept_list",
			Description: "List all services available for interception in the connected environment",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "kl_intercept_start",
			Description: "Start intercepting a service to redirect its traffic to the workspace",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"service":        {Type: "string", Description: "Service name to intercept"},
					"service_port":   {Type: "string", Description: "Port on the service to intercept (e.g., '80')"},
					"workspace_port": {Type: "string", Description: "Port on the workspace to forward traffic to (e.g., '3000')"},
				},
				Required: []string{"service", "service_port", "workspace_port"},
			},
		},
		{
			Name:        "kl_intercept_stop",
			Description: "Stop intercepting a service",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"service": {Type: "string", Description: "Service name to stop intercepting"},
				},
				Required: []string{"service"},
			},
		},
		{
			Name:        "kl_intercept_status",
			Description: "Show status of active service intercepts",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		// Port Exposure
		{
			Name:        "kl_expose",
			Description: "Expose a workspace port to the internet with a public URL",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"port": {Type: "string", Description: "Port number to expose (1-65535)"},
				},
				Required: []string{"port"},
			},
		},
		{
			Name:        "kl_expose_list",
			Description: "List all exposed ports and their public URLs",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "kl_expose_remove",
			Description: "Remove an exposed port",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"port": {Type: "string", Description: "Port number to unexpose"},
				},
				Required: []string{"port"},
			},
		},
	}
}

func (s *MCPServer) HandleToolCall(ctx context.Context, params ToolCallParams) (*ToolCallResult, error) {
	var result string
	var err error

	switch params.Name {
	// Package Management
	case "kl_pkg_search":
		result, err = s.handlePkgSearch(ctx, params.Arguments)
	case "kl_pkg_add":
		result, err = s.handlePkgAdd(ctx, params.Arguments)
	case "kl_pkg_install":
		result, err = s.handlePkgInstall(ctx, params.Arguments)
	case "kl_pkg_uninstall":
		result, err = s.handlePkgUninstall(ctx, params.Arguments)
	case "kl_pkg_list":
		result, err = s.handlePkgList(ctx)

	// Workspace Info
	case "kl_status":
		result, err = s.handleStatus(ctx)
	case "kl_config_get":
		result, err = s.handleConfigGet(ctx, params.Arguments)
	case "kl_config_set":
		result, err = s.handleConfigSet(ctx, params.Arguments)

	// Environment Management
	case "kl_env_list":
		result, err = s.handleEnvList(ctx)
	case "kl_env_connect":
		result, err = s.handleEnvConnect(ctx, params.Arguments)
	case "kl_env_disconnect":
		result, err = s.handleEnvDisconnect(ctx)
	case "kl_env_status":
		result, err = s.handleEnvStatus(ctx)

	// Service Interception
	case "kl_intercept_list":
		result, err = s.handleInterceptList(ctx)
	case "kl_intercept_start":
		result, err = s.handleInterceptStart(ctx, params.Arguments)
	case "kl_intercept_stop":
		result, err = s.handleInterceptStop(ctx, params.Arguments)
	case "kl_intercept_status":
		result, err = s.handleInterceptStatus(ctx)

	// Port Exposure
	case "kl_expose":
		result, err = s.handleExpose(ctx, params.Arguments)
	case "kl_expose_list":
		result, err = s.handleExposeList(ctx)
	case "kl_expose_remove":
		result, err = s.handleExposeRemove(ctx, params.Arguments)

	default:
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", params.Name)}},
			IsError: true,
		}, nil
	}

	if err != nil {
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil
	}

	return &ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// Package Management Handlers
func (s *MCPServer) handlePkgSearch(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query is required")
	}

	results, err := devbox.SearchPackages(query)
	if err != nil {
		return "", fmt.Errorf("failed to search packages: %w", err)
	}

	if len(results.Packages) == 0 {
		return "No packages found matching the query.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d packages:\n\n", len(results.Packages)))
	for i, pkg := range results.Packages {
		if i >= 20 { // Limit to 20 results
			sb.WriteString(fmt.Sprintf("\n... and %d more results", len(results.Packages)-20))
			break
		}
		version := ""
		if len(pkg.Versions) > 0 {
			version = pkg.Versions[0].Version
		}
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", pkg.Name, version))
		if len(pkg.Versions) > 0 && pkg.Versions[0].Summary != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", pkg.Versions[0].Summary))
		}
	}
	return sb.String(), nil
}

func (s *MCPServer) handlePkgAdd(ctx context.Context, args map[string]interface{}) (string, error) {
	packagesStr, ok := args["packages"].(string)
	if !ok || packagesStr == "" {
		return "", fmt.Errorf("packages is required")
	}

	packageNames := strings.Split(packagesStr, ",")
	for i := range packageNames {
		packageNames[i] = strings.TrimSpace(packageNames[i])
	}

	pkgReq, err := s.client.GetOrCreatePackageRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get package request: %w", err)
	}

	var added []string
	var skipped []string

	for _, pkgName := range packageNames {
		if pkgName == "" {
			continue
		}

		// Check if already installed
		alreadyExists := false
		for _, existing := range pkgReq.Spec.Packages {
			if existing.Name == pkgName {
				alreadyExists = true
				skipped = append(skipped, pkgName)
				break
			}
		}

		if !alreadyExists {
			pkgReq.Spec.Packages = append(pkgReq.Spec.Packages, packagesv1.PackageSpec{
				Name:    pkgName,
				Channel: DefaultNixpkgsChannel,
			})
			added = append(added, pkgName)
		}
	}

	if len(added) > 0 {
		if err := s.client.UpdatePackageRequest(ctx, pkgReq); err != nil {
			return "", fmt.Errorf("failed to update package request: %w", err)
		}
	}

	var sb strings.Builder
	if len(added) > 0 {
		sb.WriteString(fmt.Sprintf("Added packages: %s\n", strings.Join(added, ", ")))
		sb.WriteString("Packages are being installed. Use kl_pkg_list to check status.")
	}
	if len(skipped) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("Already installed: %s", strings.Join(skipped, ", ")))
	}
	if sb.Len() == 0 {
		sb.WriteString("No packages to add.")
	}
	return sb.String(), nil
}

func (s *MCPServer) handlePkgInstall(ctx context.Context, args map[string]interface{}) (string, error) {
	pkgName, ok := args["package"].(string)
	if !ok || pkgName == "" {
		return "", fmt.Errorf("package is required")
	}

	version, _ := args["version"].(string)
	channel, _ := args["channel"].(string)

	pkgReq, err := s.client.GetOrCreatePackageRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get package request: %w", err)
	}

	// Check if already installed
	for _, existing := range pkgReq.Spec.Packages {
		if existing.Name == pkgName {
			return fmt.Sprintf("Package '%s' is already installed.", pkgName), nil
		}
	}

	newPkg := packagesv1.PackageSpec{
		Name: pkgName,
	}

	if channel != "" {
		newPkg.Channel = channel
	} else if version != "" {
		// Resolve version to commit using devbox API
		resolved, err := devbox.ResolvePackageVersion(pkgName, version)
		if err != nil {
			return "", fmt.Errorf("version '%s' not found for package '%s': %w", version, pkgName, err)
		}
		// Get commit hash from the first available system
		for _, sysInfo := range resolved.Systems {
			newPkg.NixpkgsCommit = sysInfo.FlakeInstallable.Ref.Rev
			break
		}
		if newPkg.NixpkgsCommit == "" {
			return "", fmt.Errorf("could not resolve version '%s' for package '%s'", version, pkgName)
		}
	} else {
		newPkg.Channel = DefaultNixpkgsChannel
	}

	pkgReq.Spec.Packages = append(pkgReq.Spec.Packages, newPkg)

	if err := s.client.UpdatePackageRequest(ctx, pkgReq); err != nil {
		return "", fmt.Errorf("failed to update package request: %w", err)
	}

	return fmt.Sprintf("Installing package '%s'. Use kl_pkg_list to check status.", pkgName), nil
}

func (s *MCPServer) handlePkgUninstall(ctx context.Context, args map[string]interface{}) (string, error) {
	pkgName, ok := args["package"].(string)
	if !ok || pkgName == "" {
		return "", fmt.Errorf("package is required")
	}

	pkgReq, err := s.client.GetPackageRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get package request: %w", err)
	}

	found := false
	newPackages := make([]packagesv1.PackageSpec, 0)
	for _, pkg := range pkgReq.Spec.Packages {
		if pkg.Name == pkgName {
			found = true
		} else {
			newPackages = append(newPackages, pkg)
		}
	}

	if !found {
		return fmt.Sprintf("Package '%s' is not installed.", pkgName), nil
	}

	pkgReq.Spec.Packages = newPackages
	if err := s.client.UpdatePackageRequest(ctx, pkgReq); err != nil {
		return "", fmt.Errorf("failed to update package request: %w", err)
	}

	return fmt.Sprintf("Package '%s' has been removed.", pkgName), nil
}

func (s *MCPServer) handlePkgList(ctx context.Context) (string, error) {
	pkgReq, err := s.client.GetPackageRequest(ctx)
	if err != nil {
		return "No packages installed.", nil
	}

	if len(pkgReq.Spec.Packages) == 0 {
		return "No packages installed.", nil
	}

	var sb strings.Builder
	sb.WriteString("Installed packages:\n\n")

	// Create a set of installed package names from status
	installedSet := make(map[string]bool)
	for _, pkgName := range pkgReq.Status.Packages {
		installedSet[pkgName] = true
	}

	for _, pkg := range pkgReq.Spec.Packages {
		status := "pending"
		if pkgReq.Status.Phase == "Ready" && installedSet[pkg.Name] {
			status = "installed"
		} else if pkgReq.Status.Phase == "Failed" && pkgReq.Status.FailedPackage == pkg.Name {
			status = "failed"
		} else if pkgReq.Status.Phase == "Installing" {
			status = "installing"
		}

		source := pkg.Channel
		if source == "" && pkg.NixpkgsCommit != "" {
			if len(pkg.NixpkgsCommit) >= 8 {
				source = fmt.Sprintf("commit:%s", pkg.NixpkgsCommit[:8])
			} else {
				source = fmt.Sprintf("commit:%s", pkg.NixpkgsCommit)
			}
		}
		if source == "" {
			source = "default"
		}

		sb.WriteString(fmt.Sprintf("- %s [%s] (%s)\n", pkg.Name, status, source))
	}

	return sb.String(), nil
}

// Workspace Info Handlers
func (s *MCPServer) handleStatus(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Workspace: %s\n", ws.Name))
	sb.WriteString(fmt.Sprintf("Phase: %s\n", ws.Status.Phase))
	sb.WriteString(fmt.Sprintf("Owner: %s\n", ws.Spec.OwnedBy))

	if ws.Status.Message != "" {
		sb.WriteString(fmt.Sprintf("Status: %s\n", ws.Status.Message))
	}

	// Connected environment
	if ws.Status.ConnectedEnvironment != nil && ws.Status.ConnectedEnvironment.Name != "" {
		sb.WriteString(fmt.Sprintf("\nConnected Environment: %s\n", ws.Status.ConnectedEnvironment.Name))
	}

	// Exposed ports
	if len(ws.Spec.Expose) > 0 {
		sb.WriteString("\nExposed Ports:\n")
		for _, port := range ws.Spec.Expose {
			sb.WriteString(fmt.Sprintf("  - %d\n", port.Port))
		}
	}

	return sb.String(), nil
}

func (s *MCPServer) handleConfigGet(ctx context.Context, args map[string]interface{}) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	key, _ := args["key"].(string)

	if key == "" {
		// Return all config
		var sb strings.Builder
		sb.WriteString("Workspace Configuration:\n")
		sb.WriteString(fmt.Sprintf("  display-name: %s\n", ws.Spec.DisplayName))
		sb.WriteString(fmt.Sprintf("  description: %s\n", ws.Spec.Description))
		sb.WriteString(fmt.Sprintf("  owner: %s\n", ws.Spec.OwnedBy))
		if ws.Spec.Settings != nil {
			sb.WriteString(fmt.Sprintf("  auto-stop: %v\n", ws.Spec.Settings.AutoStop))
			sb.WriteString(fmt.Sprintf("  idle-timeout: %d minutes\n", ws.Spec.Settings.IdleTimeout))
		}
		return sb.String(), nil
	}

	// Get specific key
	switch key {
	case "display-name":
		return ws.Spec.DisplayName, nil
	case "description":
		return ws.Spec.Description, nil
	case "owner":
		return ws.Spec.OwnedBy, nil
	case "auto-stop":
		if ws.Spec.Settings != nil {
			return fmt.Sprintf("%v", ws.Spec.Settings.AutoStop), nil
		}
		return "false", nil
	case "idle-timeout":
		if ws.Spec.Settings != nil {
			return fmt.Sprintf("%d", ws.Spec.Settings.IdleTimeout), nil
		}
		return "not set", nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func (s *MCPServer) handleConfigSet(ctx context.Context, args map[string]interface{}) (string, error) {
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)

	if key == "" || value == "" {
		return "", fmt.Errorf("key and value are required")
	}

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	switch key {
	case "display-name":
		ws.Spec.DisplayName = value
	case "description":
		ws.Spec.Description = value
	default:
		return "", fmt.Errorf("cannot set config key: %s (read-only or unknown)", key)
	}

	if err := s.client.Update(ctx, ws); err != nil {
		return "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return fmt.Sprintf("Configuration '%s' updated to '%s'", key, value), nil
}

// Environment Management Handlers
func (s *MCPServer) handleEnvList(ctx context.Context) (string, error) {
	envList := &environmentsv1.EnvironmentList{}
	if err := s.client.K8sClient.List(ctx, envList); err != nil {
		return "", fmt.Errorf("failed to list environments: %w", err)
	}

	if len(envList.Items) == 0 {
		return "No environments available.", nil
	}

	var sb strings.Builder
	sb.WriteString("Available environments:\n\n")
	for _, env := range envList.Items {
		state := string(env.Status.State)
		if state == "" {
			state = "unknown"
		}
		sb.WriteString(fmt.Sprintf("- %s [%s]\n", env.Name, state))
	}
	return sb.String(), nil
}

func (s *MCPServer) handleEnvConnect(ctx context.Context, args map[string]interface{}) (string, error) {
	envName, ok := args["environment"].(string)
	if !ok || envName == "" {
		return "", fmt.Errorf("environment name is required")
	}

	// Verify environment exists
	env := &environmentsv1.Environment{}
	if err := s.client.K8sClient.Get(ctx, client.ObjectKey{Name: envName}, env); err != nil {
		return "", fmt.Errorf("environment '%s' not found: %w", envName, err)
	}

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	// Set environment connection
	ws.Spec.EnvironmentConnection = &workspacesv1.EnvironmentConnectionSpec{
		EnvironmentRef: corev1.ObjectReference{
			Name:      envName,
			Namespace: env.Namespace,
		},
	}

	if err := s.client.Update(ctx, ws); err != nil {
		return "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return fmt.Sprintf("Connected to environment '%s'. DNS will be configured shortly.", envName), nil
}

func (s *MCPServer) handleEnvDisconnect(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Spec.EnvironmentConnection == nil || ws.Spec.EnvironmentConnection.EnvironmentRef.Name == "" {
		return "Workspace is not connected to any environment.", nil
	}

	envName := ws.Spec.EnvironmentConnection.EnvironmentRef.Name
	ws.Spec.EnvironmentConnection = nil

	if err := s.client.Update(ctx, ws); err != nil {
		return "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return fmt.Sprintf("Disconnected from environment '%s'.", envName), nil
}

func (s *MCPServer) handleEnvStatus(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status.ConnectedEnvironment == nil || ws.Status.ConnectedEnvironment.Name == "" {
		return "Workspace is not connected to any environment.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Connected to: %s\n", ws.Status.ConnectedEnvironment.Name))
	sb.WriteString(fmt.Sprintf("Target Namespace: %s\n", ws.Status.ConnectedEnvironment.TargetNamespace))

	// List available services
	if len(ws.Status.ConnectedEnvironment.AvailableServices) > 0 {
		sb.WriteString("\nAvailable Services:\n")
		for _, svc := range ws.Status.ConnectedEnvironment.AvailableServices {
			sb.WriteString(fmt.Sprintf("  - %s\n", svc))
		}
	}

	return sb.String(), nil
}

// Service Interception Handlers
func (s *MCPServer) handleInterceptList(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status.ConnectedEnvironment == nil || ws.Status.ConnectedEnvironment.Name == "" {
		return "Workspace is not connected to any environment. Connect to an environment first using kl_env_connect.", nil
	}

	targetNs := ws.Status.ConnectedEnvironment.TargetNamespace

	// List compositions in the environment
	compList := &environmentsv1.CompositionList{}
	if err := s.client.K8sClient.List(ctx, compList, client.InNamespace(targetNs)); err != nil {
		return "", fmt.Errorf("failed to list compositions: %w", err)
	}

	if len(compList.Items) == 0 {
		return "No services found in the connected environment.", nil
	}

	var sb strings.Builder
	sb.WriteString("Available services for interception:\n\n")
	for _, comp := range compList.Items {
		sb.WriteString(fmt.Sprintf("Composition: %s\n", comp.Name))
		if comp.Status.Services != nil {
			for _, svc := range comp.Status.Services {
				intercepted := ""
				for _, intercept := range comp.Status.ActiveIntercepts {
					if intercept.ServiceName == svc.Name {
						intercepted = fmt.Sprintf(" [INTERCEPTED by %s]", intercept.WorkspaceName)
						break
					}
				}
				sb.WriteString(fmt.Sprintf("  - %s (ports: %v)%s\n", svc.Name, svc.Ports, intercepted))
			}
		}
	}
	return sb.String(), nil
}

func (s *MCPServer) handleInterceptStart(ctx context.Context, args map[string]interface{}) (string, error) {
	serviceName, ok := args["service"].(string)
	if !ok || serviceName == "" {
		return "", fmt.Errorf("service name is required")
	}
	servicePortStr, ok := args["service_port"].(string)
	if !ok || servicePortStr == "" {
		return "", fmt.Errorf("service_port is required")
	}
	workspacePortStr, ok := args["workspace_port"].(string)
	if !ok || workspacePortStr == "" {
		return "", fmt.Errorf("workspace_port is required")
	}

	var servicePort, workspacePort int
	fmt.Sscanf(servicePortStr, "%d", &servicePort)
	fmt.Sscanf(workspacePortStr, "%d", &workspacePort)

	if servicePort <= 0 || workspacePort <= 0 {
		return "", fmt.Errorf("invalid port numbers")
	}

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status.ConnectedEnvironment == nil || ws.Status.ConnectedEnvironment.Name == "" {
		return "", fmt.Errorf("workspace is not connected to any environment")
	}

	targetNs := ws.Status.ConnectedEnvironment.TargetNamespace

	// Find the composition containing the service
	compList := &environmentsv1.CompositionList{}
	if err := s.client.K8sClient.List(ctx, compList, client.InNamespace(targetNs)); err != nil {
		return "", fmt.Errorf("failed to list compositions: %w", err)
	}

	var targetComp *environmentsv1.Composition
	for i := range compList.Items {
		comp := &compList.Items[i]
		if comp.Status.Services != nil {
			for _, svc := range comp.Status.Services {
				if svc.Name == serviceName {
					targetComp = comp
					break
				}
			}
		}
		if targetComp != nil {
			break
		}
	}

	if targetComp == nil {
		return "", fmt.Errorf("service '%s' not found in any composition", serviceName)
	}

	// Add intercept to composition
	interceptConfig := environmentsv1.ServiceInterceptConfig{
		ServiceName: serviceName,
		Enabled:     true,
		WorkspaceRef: &corev1.ObjectReference{
			Name:      ws.Name,
			Namespace: ws.Namespace,
		},
		PortMappings: []environmentsv1.PortMapping{
			{
				ServicePort:   int32(servicePort),
				WorkspacePort: int32(workspacePort),
				Protocol:      corev1.ProtocolTCP,
			},
		},
	}

	// Check if intercept already exists
	found := false
	for i, existing := range targetComp.Spec.Intercepts {
		if existing.ServiceName == serviceName {
			targetComp.Spec.Intercepts[i] = interceptConfig
			found = true
			break
		}
	}
	if !found {
		targetComp.Spec.Intercepts = append(targetComp.Spec.Intercepts, interceptConfig)
	}

	if err := s.client.K8sClient.Update(ctx, targetComp); err != nil {
		return "", fmt.Errorf("failed to update composition: %w", err)
	}

	return fmt.Sprintf("Intercept started for service '%s'. Traffic on port %d will be forwarded to workspace port %d.", serviceName, servicePort, workspacePort), nil
}

func (s *MCPServer) handleInterceptStop(ctx context.Context, args map[string]interface{}) (string, error) {
	serviceName, ok := args["service"].(string)
	if !ok || serviceName == "" {
		return "", fmt.Errorf("service name is required")
	}

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status.ConnectedEnvironment == nil || ws.Status.ConnectedEnvironment.Name == "" {
		return "", fmt.Errorf("workspace is not connected to any environment")
	}

	targetNs := ws.Status.ConnectedEnvironment.TargetNamespace

	// Find and update composition
	compList := &environmentsv1.CompositionList{}
	if err := s.client.K8sClient.List(ctx, compList, client.InNamespace(targetNs)); err != nil {
		return "", fmt.Errorf("failed to list compositions: %w", err)
	}

	for i := range compList.Items {
		comp := &compList.Items[i]
		for j, intercept := range comp.Spec.Intercepts {
			if intercept.ServiceName == serviceName {
				comp.Spec.Intercepts[j].Enabled = false
				if err := s.client.K8sClient.Update(ctx, comp); err != nil {
					return "", fmt.Errorf("failed to update composition: %w", err)
				}
				return fmt.Sprintf("Intercept stopped for service '%s'.", serviceName), nil
			}
		}
	}

	return fmt.Sprintf("No active intercept found for service '%s'.", serviceName), nil
}

func (s *MCPServer) handleInterceptStatus(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status.ConnectedEnvironment == nil || ws.Status.ConnectedEnvironment.Name == "" {
		return "Workspace is not connected to any environment.", nil
	}

	targetNs := ws.Status.ConnectedEnvironment.TargetNamespace

	compList := &environmentsv1.CompositionList{}
	if err := s.client.K8sClient.List(ctx, compList, client.InNamespace(targetNs)); err != nil {
		return "", fmt.Errorf("failed to list compositions: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Intercept Status:\n\n")
	hasIntercepts := false

	for _, comp := range compList.Items {
		for _, status := range comp.Status.ActiveIntercepts {
			if status.WorkspaceName == ws.Name {
				hasIntercepts = true
				sb.WriteString(fmt.Sprintf("- %s [%s]\n", status.ServiceName, status.Phase))
				if status.Message != "" {
					sb.WriteString(fmt.Sprintf("  Message: %s\n", status.Message))
				}
			}
		}
	}

	if !hasIntercepts {
		sb.WriteString("No active intercepts for this workspace.")
	}

	return sb.String(), nil
}

// Port Exposure Handlers
func (s *MCPServer) handleExpose(ctx context.Context, args map[string]interface{}) (string, error) {
	portStr, ok := args["port"].(string)
	if !ok || portStr == "" {
		return "", fmt.Errorf("port is required")
	}

	var port int
	fmt.Sscanf(portStr, "%d", &port)
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid port number (must be 1-65535)")
	}

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if already exposed
	for _, ep := range ws.Spec.Expose {
		if ep.Port == int32(port) {
			return fmt.Sprintf("Port %d is already exposed.", port), nil
		}
	}

	ws.Spec.Expose = append(ws.Spec.Expose, workspacesv1.ExposedPort{
		Port: int32(port),
	})

	if err := s.client.Update(ctx, ws); err != nil {
		return "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return fmt.Sprintf("Port %d is now exposed. Public URL will be available shortly.", port), nil
}

func (s *MCPServer) handleExposeList(ctx context.Context) (string, error) {
	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	if len(ws.Spec.Expose) == 0 {
		return "No ports are currently exposed.", nil
	}

	var sb strings.Builder
	sb.WriteString("Exposed ports:\n\n")
	for _, ep := range ws.Spec.Expose {
		sb.WriteString(fmt.Sprintf("- Port %d\n", ep.Port))
	}
	return sb.String(), nil
}

func (s *MCPServer) handleExposeRemove(ctx context.Context, args map[string]interface{}) (string, error) {
	portStr, ok := args["port"].(string)
	if !ok || portStr == "" {
		return "", fmt.Errorf("port is required")
	}

	var port int
	fmt.Sscanf(portStr, "%d", &port)

	ws, err := s.client.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace: %w", err)
	}

	found := false
	newPorts := make([]workspacesv1.ExposedPort, 0)
	for _, ep := range ws.Spec.Expose {
		if ep.Port == int32(port) {
			found = true
		} else {
			newPorts = append(newPorts, ep)
		}
	}

	if !found {
		return fmt.Sprintf("Port %d is not exposed.", port), nil
	}

	ws.Spec.Expose = newPorts
	if err := s.client.Update(ctx, ws); err != nil {
		return "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return fmt.Sprintf("Port %d is no longer exposed.", port), nil
}

// Run starts the MCP server
func (s *MCPServer) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		var response MCPResponse
		response.JSONRPC = "2.0"
		response.ID = req.ID

		switch req.Method {
		case "initialize":
			response.Result = InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: ServerCapabilities{
					Tools: &ToolsCapability{},
				},
				ServerInfo: ServerInfo{
					Name:    "kl-mcp",
					Version: "0.1.0",
				},
			}

		case "notifications/initialized":
			// No response needed for notifications
			continue

		case "tools/list":
			response.Result = ToolsListResult{
				Tools: s.GetTools(),
			}

		case "tools/call":
			var params ToolCallParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				response.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			} else {
				result, err := s.HandleToolCall(context.Background(), params)
				if err != nil {
					response.Error = &MCPError{Code: -32603, Message: err.Error()}
				} else {
					response.Result = result
				}
			}

		default:
			response.Error = &MCPError{Code: -32601, Message: "Method not found"}
		}

		encoder.Encode(response)
	}

	return scanner.Err()
}

// Cobra command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes kl commands as tools for AI assistants like Claude.

The server communicates via stdio using JSON-RPC 2.0 protocol.

To use with Claude Desktop, add to your config:
{
  "mcpServers": {
    "kloudlite": {
      "command": "kl",
      "args": ["mcp"]
    }
  }
}`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := NewMCPServer()
		if err != nil {
			return err
		}
		return server.Run()
	},
}

func init() {
	RootCmd.AddCommand(mcpCmd)
}
