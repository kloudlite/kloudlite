# Kloudlite Workspace

You are working inside a Kloudlite workspace - a cloud-based development environment with integrated tools for package management, environment connections, and service interception.

## Available MCP Tools

The `kl mcp` command provides an MCP server with tools you can use to manage this workspace. These tools are available through the MCP protocol:

### Package Management
- **kl_pkg_search** - Search for Nix packages (e.g., nodejs, python3, go)
- **kl_pkg_add** - Add packages to the workspace (comma-separated list)
- **kl_pkg_install** - Install a specific package version
- **kl_pkg_uninstall** - Remove a package
- **kl_pkg_list** - List installed packages with their status

### Workspace Info
- **kl_status** - Show workspace status, phase, and connections
- **kl_config_get** - Get workspace configuration values
- **kl_config_set** - Set workspace configuration (display-name, description)

### Environment Management
- **kl_env_list** - List available environments to connect to
- **kl_env_connect** - Connect workspace to an environment for DNS and intercepts
- **kl_env_disconnect** - Disconnect from the current environment
- **kl_env_status** - Show environment connection status

### Service Interception
- **kl_intercept_list** - List services available for interception
- **kl_intercept_start** - Start intercepting a service (redirect traffic to workspace)
- **kl_intercept_stop** - Stop intercepting a service
- **kl_intercept_status** - Show active intercept status

### Port Exposure
- **kl_expose** - Expose a workspace port to the internet with a public URL
- **kl_expose_list** - List exposed ports and their URLs
- **kl_expose_remove** - Remove an exposed port

## Usage Guidelines

1. **Installing packages**: Use `kl_pkg_add` for quick installation or `kl_pkg_install` for specific versions
2. **Environment connection**: Connect to environments to access their services and enable intercepts
3. **Service interception**: Intercept services to redirect production traffic to your local development
4. **Port exposure**: Expose local ports to share your work or test webhooks

## Workspace Context

- Packages are managed via Nix for reproducibility
- The workspace persists across restarts
- Environment connections enable DNS resolution for services
- Intercepts use SOCAT for traffic forwarding
