import * as vscode from 'vscode';
import { KloudliteWorkspacesProvider } from './workspacesProvider';
import { KloudliteAPI } from './api';

export function activate(context: vscode.ExtensionContext) {
  console.log('Kloudlite extension is now active');

  const api = new KloudliteAPI();
  const workspacesProvider = new KloudliteWorkspacesProvider(api);

  // Register the tree data provider
  vscode.window.registerTreeDataProvider('kloudliteWorkspaces', workspacesProvider);

  // Register commands
  const connectWorkspace = vscode.commands.registerCommand(
    'kloudlite.connectWorkspace',
    async (workspace?: any) => {
      try {
        if (!workspace) {
          // Show quick pick if no workspace provided
          const workspaces = await api.listWorkspaces();
          const items = workspaces.map(ws => ({
            label: ws.metadata.name,
            description: ws.status?.phase || 'Unknown',
            workspace: ws
          }));

          const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'Select a workspace to connect'
          });

          if (!selected) {
            return;
          }
          workspace = selected.workspace;
        }

        await connectToWorkspace(workspace);
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to connect: ${error}`);
      }
    }
  );

  const listWorkspaces = vscode.commands.registerCommand(
    'kloudlite.listWorkspaces',
    async () => {
      try {
        const workspaces = await api.listWorkspaces();
        const items = workspaces.map(ws => ({
          label: ws.metadata.name,
          description: ws.status?.phase || 'Unknown',
          workspace: ws
        }));

        const selected = await vscode.window.showQuickPick(items, {
          placeHolder: 'Select a workspace to view details'
        });

        if (selected) {
          vscode.window.showInformationMessage(
            `Workspace: ${selected.workspace.metadata.name}\nStatus: ${selected.workspace.status?.phase}`
          );
        }
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to list workspaces: ${error}`);
      }
    }
  );

  const refreshWorkspaces = vscode.commands.registerCommand(
    'kloudlite.refreshWorkspaces',
    () => {
      workspacesProvider.refresh();
    }
  );

  context.subscriptions.push(connectWorkspace, listWorkspaces, refreshWorkspaces);
}

async function connectToWorkspace(workspace: any) {
  const config = vscode.workspace.getConfiguration('kloudlite');
  const sshPort = config.get<number>('sshPort', 2222);

  const workspaceName = workspace.metadata.name;
  const jumpHost = `kloudlite@localhost:${sshPort}`;
  const targetHost = `kl@${workspaceName}`;
  const workspaceDir = `/home/kl/workspaces/${workspaceName}`;

  // Generate SSH config
  const sshConfig = `Host ${workspaceName}
  HostName ${workspaceName}
  User kl
  ProxyJump ${jumpHost}
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null`;

  // Show SSH config to user
  const action = await vscode.window.showInformationMessage(
    `To connect to workspace "${workspaceName}", add this SSH config to ~/.ssh/config`,
    'Copy SSH Config',
    'Open Workspace'
  );

  if (action === 'Copy SSH Config') {
    await vscode.env.clipboard.writeText(sshConfig);
    vscode.window.showInformationMessage('SSH config copied to clipboard');
  } else if (action === 'Open Workspace') {
    // Use VS Code Remote SSH to connect
    const remoteUri = vscode.Uri.parse(`vscode-remote://ssh-remote+${targetHost}${workspaceDir}`);
    await vscode.commands.executeCommand('vscode.openFolder', remoteUri, { forceNewWindow: true });
  }
}

export function deactivate() {}
