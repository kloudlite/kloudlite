# Kloudlite Workspace

Connect to your Kloudlite cloud development workspaces directly from Visual Studio Code with one-click access.

## Features

- **🔐 Secure Authentication**: Token-based authentication with Kloudlite API
- **📋 List Workspaces**: View all your Kloudlite workspaces in the sidebar
- **🚀 One-Click Connect**: Connect to workspaces via SSH Remote with a single click from the sidebar or web dashboard
- **🔗 Deep Link Support**: Click links in your browser to automatically open workspaces in VS Code
- **📊 Status Monitoring**: See real-time status of your workspaces (Running, Pending, Stopped)
- **⚙️ Automatic SSH Setup**: Automatically configures SSH keys and connection settings
- **🔄 Auto-Refresh**: Workspace list updates automatically

## Installation

### From VS Code Marketplace

1. Open VS Code
2. Go to Extensions (Cmd/Ctrl+Shift+X)
3. Search for "Kloudlite Workspace"
4. Click "Install"

### From VSIX (Manual Installation)

1. Download the `.vsix` file from the releases page
2. In VS Code, go to Extensions (Cmd/Ctrl+Shift+X)
3. Click the `...` menu and select "Install from VSIX..."
4. Select the downloaded `.vsix` file

## Configuration

Configure the extension in VS Code settings:

- `kloudlite.apiUrl`: Kloudlite API URL (default: `http://localhost:8080`)
- `kloudlite.connectionToken`: Your connection token (managed by extension, don't edit manually)

## Usage

### Setting Up Authentication

Before using the extension, you need to authenticate:

1. Go to your Kloudlite web dashboard
2. Navigate to your user profile dropdown → "Connection Tokens"
3. Create a new connection token with a descriptive name
4. Copy the generated JWT token
5. In VS Code, open Command Palette (Cmd/Ctrl+Shift+P)
6. Run `Kloudlite: Set Connection Token`
7. Paste the JWT token when prompted
8. The extension will validate and save your token

### Connecting to a Workspace

#### From VS Code Sidebar

1. Open the Kloudlite sidebar (click the Kloudlite icon in the activity bar)
2. Browse your workspaces (requires authentication)
3. Click on any workspace to connect automatically

The extension will:
- Automatically generate and add SSH keys to your workspace
- Configure SSH connection settings
- Open the workspace in VS Code Remote SSH

#### From Web Dashboard (Deep Link)

1. Go to your Kloudlite web dashboard
2. Navigate to a workspace detail page
3. Click the "VS Code Extension" button in the connection options
4. Your browser will open VS Code and automatically connect to the workspace

### SSH Configuration

When connecting to a workspace, you'll need to add SSH configuration to `~/.ssh/config`:

```
Host workspace-name
  HostName workspace-name
  User kl
  ProxyJump kloudlite@jump-host:port
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
```

The extension provides this configuration automatically - just click "Copy SSH Config" when prompted.

## Commands

Access these commands from the Command Palette (Cmd/Ctrl+Shift+P):

- `Kloudlite: Set Connection Token` - Authenticate with your connection token
- `Kloudlite: Disconnect` - Remove saved connection token
- `Kloudlite: Connect to Workspace` - Connect to a workspace
- `Kloudlite: List Workspaces` - Show all workspaces in quick pick
- `Kloudlite: Refresh Workspaces` - Refresh the workspace list

## Security

Connection tokens are:
- Stored securely in VS Code settings
- Used for API authentication via JWT Bearer tokens
- Can be revoked from the Kloudlite web dashboard
- Have configurable expiration (default: 1 year)
- Include connection information (SSH jump host, API URL)

## Requirements

- Visual Studio Code 1.80.0 or higher
- Kloudlite platform running and accessible
- SSH access to Kloudlite jump host
- A valid Kloudlite connection token

## Known Issues

- SSH configuration must be manually added to `~/.ssh/config` before connecting
- Workspace must be in "Running" state to connect

## Release Notes

### 0.1.19

- Deep link support for opening workspaces from web browser
- Automatic SSH key generation and management
- Improved workspace connection flow
- Enhanced error handling and logging
- Fixed SSH configuration for workspaces
- Updated UI with better status indicators

### 0.1.0

Initial release of Kloudlite Workspace extension

- Token-based authentication
- Workspace listing
- SSH connection support
- Tree view in sidebar
- Connection token management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
