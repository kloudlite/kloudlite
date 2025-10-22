# Change Log

All notable changes to the Kloudlite Workspace extension will be documented in this file.

## [0.1.20] - 2025-01-19

### Fixed
- Republished with correct extension code including all latest features
- Ensured deep link support is properly included

## [0.1.19] - 2025-01-19

### Added
- Deep link support for opening workspaces from web browser (`vscode://` protocol handler)
- Automatic SSH key generation and management
- Enhanced workspace connection flow with automatic SSH configuration
- URI handler for seamless browser-to-VS Code workspace connections
- Improved logging throughout the extension for better debugging

### Changed
- Updated workspace path pattern to use `/home/kl/workspaces/{workspace_name}`
- Improved SSH connection configuration with proper jump host setup
- Enhanced error messages for better user feedback
- Updated workspace icons based on status (Running, Pending, Stopped)

### Fixed
- Fixed workspace hostname to include `workspace-` prefix for DNS resolution
- Fixed SSH URI format for VS Code Remote-SSH connections
- Improved workspace data validation before connection attempts
- Fixed authentication token validation and error handling

## [0.1.0] - 2024-10-01

### Added
- Initial release of Kloudlite Workspace extension
- Token-based authentication with Kloudlite API
- Workspace listing in sidebar tree view
- SSH connection support for workspaces
- Connection token management commands
- Quick pick menu for workspace selection
- Real-time workspace status monitoring
- Automatic token validation and refresh
