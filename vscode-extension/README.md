# Kloudlite Workspace VS Code Extension

Connect to your Kloudlite workspaces directly from Visual Studio Code.

## Features

- **List Workspaces**: View all your Kloudlite workspaces in the sidebar
- **One-Click Connect**: Connect to workspaces via SSH Remote with a single click
- **Status Monitoring**: See real-time status of your workspaces
- **Quick Access**: Open web terminals and VS Code Web directly from the extension
- **SSH Configuration**: Automatically generate SSH config for workspace connections

## Installation

### From Source

1. Clone the repository
2. Navigate to the `vscode-extension` directory
3. Run `npm install`
4. Run `npm run compile`
5. Press F5 to open a new VS Code window with the extension loaded

### From VSIX

1. Download the `.vsix` file
2. In VS Code, go to Extensions
3. Click the `...` menu and select "Install from VSIX..."
4. Select the downloaded `.vsix` file

## Configuration

Configure the extension in VS Code settings:

- `kloudlite.apiUrl`: Kloudlite API URL (default: `http://localhost:3000`)
- `kloudlite.sshPort`: SSH jump host port (default: `2222`)

## Usage

### Connecting to a Workspace

1. Open the Kloudlite sidebar (click the Kloudlite icon in the activity bar)
2. Browse your workspaces
3. Click on a workspace to expand its details
4. Click "Connect via SSH" to connect

The extension will:
- Show you the SSH configuration needed
- Offer to copy it to your clipboard
- Allow you to open the workspace directly in VS Code Remote SSH

### SSH Configuration

When connecting to a workspace, you'll need to add SSH configuration to `~/.ssh/config`:

```
Host workspace-name
  HostName workspace-name
  User kl
  ProxyJump kloudlite@localhost:2222
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
```

The extension provides this configuration automatically - just click "Copy SSH Config" when prompted.

## Commands

Access these commands from the Command Palette (Cmd/Ctrl+Shift+P):

- `Kloudlite: Connect to Workspace` - Connect to a workspace
- `Kloudlite: List Workspaces` - Show all workspaces in quick pick
- `Kloudlite: Refresh Workspaces` - Refresh the workspace list

## Requirements

- Visual Studio Code 1.80.0 or higher
- Kloudlite platform running and accessible
- SSH access to Kloudlite jump host

## Known Issues

- SSH configuration must be manually added to `~/.ssh/config` before connecting
- Workspace must be in "Running" state to connect

## Release Notes

### 0.1.0

Initial release of Kloudlite Workspace extension

- Basic workspace listing
- SSH connection support
- Tree view in sidebar
- Quick pick for workspace selection

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
