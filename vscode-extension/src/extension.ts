import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { exec } from 'child_process';
import { promisify } from 'util';
import { KloudliteWorkspacesProvider } from './workspacesProvider';
import { KloudliteAPI } from './api';

const execAsync = promisify(exec);

export function activate(context: vscode.ExtensionContext) {
  console.log('[Kloudlite] Starting activation...');

  try {
    console.log('[Kloudlite] Creating API instance...');
    const api = new KloudliteAPI();
    console.log('[Kloudlite] API instance created');

    console.log('[Kloudlite] Creating workspaces provider...');
    const workspacesProvider = new KloudliteWorkspacesProvider(api);
    console.log('[Kloudlite] Workspaces provider created');

    // Register the tree data provider
    console.log('[Kloudlite] Registering tree data provider...');
    const treeDataProvider = vscode.window.registerTreeDataProvider('kloudliteWorkspaces', workspacesProvider);
    console.log('[Kloudlite] Tree data provider registered');

    // Register commands
    console.log('[Kloudlite] Registering commands...');
    const setConnectionToken = vscode.commands.registerCommand(
      'kloudlite.setConnectionToken',
      async () => {
      try {
        const token = await vscode.window.showInputBox({
          prompt: 'Enter your Kloudlite connection token',
          password: true,
          placeHolder: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
        });

        if (!token) {
          return;
        }

        await api.setConnectionToken(token);
        workspacesProvider.refresh();
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to set connection token: ${error}`);
      }
    }
  );

  const disconnect = vscode.commands.registerCommand(
    'kloudlite.disconnect',
    async () => {
      await api.disconnect();
      workspacesProvider.refresh();
    }
  );

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

        await connectToWorkspace(workspace, api);
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

    // Show welcome command
    const showWelcome = vscode.commands.registerCommand(
      'kloudlite.showWelcome',
      () => {
        showWelcomePage(context, api);
      }
    );

    console.log('[Kloudlite] Registering all subscriptions...');
    context.subscriptions.push(
      treeDataProvider,
      setConnectionToken,
      disconnect,
      connectWorkspace,
      listWorkspaces,
      refreshWorkspaces,
      showWelcome
    );

    console.log('[Kloudlite] Extension activated successfully!');

    // Show welcome page on first activation or if no token is set
    const hasShownWelcome = context.globalState.get('hasShownWelcome', false);
    if (!hasShownWelcome || !api.isAuthenticated()) {
      showWelcomePage(context, api);
      context.globalState.update('hasShownWelcome', true);
    }

    // Register URI handler for deep linking
    // URI format: vscode://kloudlite.kloudlite-workspace/connect?workspace=<name>&namespace=<namespace>
    context.subscriptions.push(
      vscode.window.registerUriHandler({
        handleUri: async (uri: vscode.Uri) => {
          console.log('[Kloudlite] Handling URI:', uri.toString());

          if (uri.path === '/connect') {
            const params = new URLSearchParams(uri.query);
            const workspaceName = params.get('workspace');
            const namespace = params.get('namespace');

            if (!workspaceName) {
              vscode.window.showErrorMessage('Invalid workspace link: missing workspace name');
              return;
            }

            try {
              // Fetch the workspace details
              const workspaces = await api.listWorkspaces();
              const workspace = workspaces.find(ws =>
                ws.metadata.name === workspaceName &&
                (!namespace || ws.metadata.namespace === namespace)
              );

              if (!workspace) {
                vscode.window.showErrorMessage(`Workspace "${workspaceName}" not found`);
                return;
              }

              // Connect to the workspace
              await vscode.commands.executeCommand('kloudlite.connectWorkspace', workspace);
            } catch (error) {
              console.error('[Kloudlite] Failed to connect via URI:', error);
              vscode.window.showErrorMessage(`Failed to connect to workspace: ${error}`);
            }
          }
        }
      })
    );
  } catch (error) {
    const errorMessage = `Kloudlite extension failed to activate: ${error}`;
    console.error('[Kloudlite] Activation error:', error);
    console.error('[Kloudlite] Error stack:', (error as Error).stack);
    vscode.window.showErrorMessage(errorMessage);
    throw error; // Re-throw to see in Extension Host logs
  }
}

function showWelcomePage(context: vscode.ExtensionContext, api: KloudliteAPI) {
  const panel = vscode.window.createWebviewPanel(
    'kloudliteWelcome',
    'Welcome to Kloudlite',
    vscode.ViewColumn.One,
    {
      enableScripts: true
    }
  );

  const isAuthenticated = api.isAuthenticated();

  panel.webview.html = getWelcomePageHtml(isAuthenticated);

  // Handle messages from the webview
  panel.webview.onDidReceiveMessage(
    async (message) => {
      switch (message.command) {
        case 'setConnectionToken':
          await vscode.commands.executeCommand('kloudlite.setConnectionToken');
          panel.webview.html = getWelcomePageHtml(api.isAuthenticated());
          break;
        case 'openDocs':
          vscode.env.openExternal(vscode.Uri.parse('https://docs.kloudlite.io'));
          break;
      }
    },
    undefined,
    context.subscriptions
  );
}

function getWelcomePageHtml(isAuthenticated: boolean): string {
  return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Kloudlite</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
            padding: 20px;
            line-height: 1.6;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        h1 {
            color: var(--vscode-textLink-foreground);
            margin-bottom: 10px;
        }
        .subtitle {
            color: var(--vscode-descriptionForeground);
            margin-bottom: 30px;
        }
        .status {
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 30px;
        }
        .status.authenticated {
            background-color: var(--vscode-inputValidation-infoBackground);
            border: 1px solid var(--vscode-inputValidation-infoBorder);
        }
        .status.not-authenticated {
            background-color: var(--vscode-inputValidation-warningBackground);
            border: 1px solid var(--vscode-inputValidation-warningBorder);
        }
        .section {
            margin-bottom: 30px;
        }
        .section h2 {
            color: var(--vscode-textLink-foreground);
            margin-bottom: 15px;
        }
        .steps {
            list-style: none;
            padding: 0;
            counter-reset: step-counter;
        }
        .steps li {
            counter-increment: step-counter;
            margin-bottom: 20px;
            padding-left: 40px;
            position: relative;
        }
        .steps li::before {
            content: counter(step-counter);
            position: absolute;
            left: 0;
            top: 0;
            width: 30px;
            height: 30px;
            border-radius: 50%;
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
        }
        button {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin-right: 10px;
            margin-top: 10px;
        }
        button:hover {
            background-color: var(--vscode-button-hoverBackground);
        }
        .link-button {
            background-color: transparent;
            color: var(--vscode-textLink-foreground);
            text-decoration: underline;
            padding: 5px 10px;
        }
        .link-button:hover {
            background-color: var(--vscode-list-hoverBackground);
        }
        .feature-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .feature-card {
            padding: 15px;
            border: 1px solid var(--vscode-panel-border);
            border-radius: 5px;
        }
        .feature-card h3 {
            margin-top: 0;
            color: var(--vscode-textLink-foreground);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Welcome to Kloudlite</h1>
        <p class="subtitle">Connect to your cloud development workspaces directly from VS Code</p>

        <div class="status ${isAuthenticated ? 'authenticated' : 'not-authenticated'}">
            ${isAuthenticated ?
                '✅ <strong>Connected!</strong> You\'re ready to access your workspaces.' :
                '⚠️  <strong>Not Connected</strong> - Please add your connection token to get started.'
            }
        </div>

        ${!isAuthenticated ? `
        <div class="section">
            <h2>Getting Started</h2>
            <ol class="steps">
                <li>
                    <strong>Get your connection token</strong><br>
                    Go to the Kloudlite web dashboard and navigate to Connection Tokens page to create a new token.
                </li>
                <li>
                    <strong>Add the token to VS Code</strong><br>
                    Click the button below or use the command palette (Cmd/Ctrl+Shift+P) and search for "Kloudlite: Set Connection Token".
                    <br><button onclick="setConnectionToken()">Add Connection Token</button>
                </li>
                <li>
                    <strong>Connect to a workspace</strong><br>
                    Once authenticated, you'll see your workspaces in the sidebar. Double-click any workspace to connect!
                </li>
            </ol>
        </div>
        ` : `
        <div class="section">
            <h2>What's Next?</h2>
            <ul class="steps">
                <li>
                    <strong>View your workspaces</strong><br>
                    Check the Kloudlite sidebar to see all your available workspaces.
                </li>
                <li>
                    <strong>Connect to a workspace</strong><br>
                    Double-click any workspace or right-click and select "Connect to Workspace".
                </li>
                <li>
                    <strong>Start coding</strong><br>
                    Your workspace will open in a new VS Code window via Remote-SSH.
                </li>
            </ul>
        </div>
        `}

        <div class="section">
            <h2>Features</h2>
            <div class="feature-grid">
                <div class="feature-card">
                    <h3>🔐 Secure SSH</h3>
                    <p>Automatic SSH key management with dedicated Kloudlite keys</p>
                </div>
                <div class="feature-card">
                    <h3>⚡ One-Click Connect</h3>
                    <p>Double-click to connect to any workspace instantly</p>
                </div>
                <div class="feature-card">
                    <h3>🔄 Auto-Sync</h3>
                    <p>SSH keys automatically synced to your workspaces</p>
                </div>
                <div class="feature-card">
                    <h3>📦 Workspace Management</h3>
                    <p>View and manage all your workspaces from the sidebar</p>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>Resources</h2>
            <button class="link-button" onclick="openDocs()">📖 Documentation</button>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        function setConnectionToken() {
            vscode.postMessage({ command: 'setConnectionToken' });
        }

        function openDocs() {
            vscode.postMessage({ command: 'openDocs' });
        }
    </script>
</body>
</html>`;
}

async function ensureKloudliteSSHKey(): Promise<{ privateKeyPath: string; publicKey: string }> {
  const sshDir = path.join(os.homedir(), '.ssh');
  const privateKeyPath = path.join(sshDir, 'kloudlite_rsa');
  const publicKeyPath = path.join(sshDir, 'kloudlite_rsa.pub');

  // Create .ssh directory if it doesn't exist
  if (!fs.existsSync(sshDir)) {
    fs.mkdirSync(sshDir, { mode: 0o700 });
  }

  // Check if Kloudlite SSH key already exists
  if (fs.existsSync(privateKeyPath) && fs.existsSync(publicKeyPath)) {
    const publicKey = fs.readFileSync(publicKeyPath, 'utf8').trim();
    return { privateKeyPath, publicKey };
  }

  // Generate new SSH key pair for Kloudlite
  console.log('[Kloudlite] Generating new SSH key pair...');

  try {
    await execAsync(
      `ssh-keygen -t rsa -b 4096 -f "${privateKeyPath}" -N "" -C "kloudlite-vscode"`
    );

    // Set proper permissions
    fs.chmodSync(privateKeyPath, 0o600);
    fs.chmodSync(publicKeyPath, 0o644);

    const publicKey = fs.readFileSync(publicKeyPath, 'utf8').trim();
    console.log('[Kloudlite] SSH key pair generated successfully');

    return { privateKeyPath, publicKey };
  } catch (error) {
    throw new Error(`Failed to generate SSH key: ${error}`);
  }
}

async function connectToWorkspace(workspace: any, api: KloudliteAPI) {
  // Check if authenticated
  if (!api.isAuthenticated()) {
    const action = await vscode.window.showWarningMessage(
      'Please set your connection token first',
      'Set Connection Token'
    );
    if (action === 'Set Connection Token') {
      await vscode.commands.executeCommand('kloudlite.setConnectionToken');
    }
    return;
  }

  // Validate workspace object
  if (!workspace || !workspace.metadata || !workspace.metadata.name) {
    vscode.window.showErrorMessage('Invalid workspace data');
    return;
  }

  try {
    // Show connecting message
    await vscode.window.withProgress({
      location: vscode.ProgressLocation.Notification,
      title: `Connecting to ${workspace.metadata.name}...`,
      cancellable: false
    }, async (progress) => {
      // Step 1: Ensure Kloudlite SSH key exists
      progress.report({ message: 'Ensuring SSH key...' });
      const { privateKeyPath, publicKey } = await ensureKloudliteSSHKey();

      // Step 2: Add public key to authorized keys
      progress.report({ message: 'Adding SSH key to workspace...' });
      await api.addSSHKey(publicKey);

      // Step 3: Get workspace SSH connection details from workspace spec
      progress.report({ message: 'Getting connection details...' });

      // Extract SSH connection details from workspace
      const jumpHost = workspace.spec?.sshJumpHost || 'localhost';
      const jumpPort = workspace.spec?.sshJumpPort || 2222;
      const jumpUser = workspace.spec?.sshJumpUser || 'kloudlite';
      const workspaceName = workspace.metadata.name;
      const workspaceUser = workspace.spec?.sshUser || 'kl';

      // Step 4: Configure SSH config for this workspace with jump host
      progress.report({ message: 'Configuring SSH...' });
      await configureSSHForWorkspace(
        workspaceName,
        jumpHost,
        jumpPort,
        jumpUser,
        workspaceUser,
        privateKeyPath
      );

      // Step 5: Set Remote-SSH to use Kloudlite config file
      const kloudliteConfigPath = path.join(os.homedir(), '.ssh', 'kloudlite', 'config');
      await vscode.workspace.getConfiguration('remote.SSH').update(
        'configFile',
        kloudliteConfigPath,
        vscode.ConfigurationTarget.Global
      );

      // Step 6: Connect using Remote-SSH
      progress.report({ message: 'Opening remote SSH connection...' });

      // Use the SSH config host alias with proper Remote-SSH URI format
      const hostAlias = `kloudlite-${workspaceName}`;
      // Workspace path follows the pattern: /home/{user}/workspaces/{workspace-name}
      const workspacePath = `/home/${workspaceUser}/workspaces/${workspaceName}`;
      const sshConnectionString = `vscode-remote://ssh-remote+${hostAlias}${workspacePath}`;

      console.log(`[Kloudlite] Connecting to workspace at: ${workspacePath}`);

      // Open a new window connected to the workspace
      await vscode.commands.executeCommand('vscode.openFolder', vscode.Uri.parse(sshConnectionString), true);
    });

  } catch (error) {
    console.error('[Kloudlite] Connection error:', error);
    vscode.window.showErrorMessage(`Failed to connect to workspace: ${error}`);
  }
}

async function configureSSHForWorkspace(
  workspaceName: string,
  jumpHost: string,
  jumpPort: number,
  jumpUser: string,
  workspaceUser: string,
  privateKeyPath: string
): Promise<void> {
  const sshDir = path.join(os.homedir(), '.ssh');
  const kloudliteConfigDir = path.join(sshDir, 'kloudlite');
  const kloudliteConfigPath = path.join(kloudliteConfigDir, 'config');
  const hostAlias = `kloudlite-${workspaceName}`;

  // Create kloudlite config directory if it doesn't exist
  if (!fs.existsSync(kloudliteConfigDir)) {
    fs.mkdirSync(kloudliteConfigDir, { mode: 0o700, recursive: true });
  }

  // SSH config entry for this workspace using ProxyCommand
  // This ensures DNS resolution happens through the jump host
  // The pod hostname has a 'workspace-' prefix
  // Equivalent to: ssh -J kloudlite@localhost:2222 kl@workspace-test-dns-ws
  const podHostname = `workspace-${workspaceName}`;
  const configEntry = `
# Kloudlite Workspace: ${workspaceName}
Host ${hostAlias}
    HostName ${podHostname}
    User ${workspaceUser}
    ProxyCommand ssh -W %h:%p -p ${jumpPort} ${jumpUser}@${jumpHost}
    IdentityFile ${privateKeyPath}
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
`;

  // Read existing Kloudlite SSH config
  let existingConfig = '';
  if (fs.existsSync(kloudliteConfigPath)) {
    existingConfig = fs.readFileSync(kloudliteConfigPath, 'utf8');
  }

  // Check if this workspace is already configured
  const hostPattern = new RegExp(`# Kloudlite Workspace: ${workspaceName}[\\s\\S]*?(?=\\n# Kloudlite Workspace:|\\n\\nHost |$)`, 'g');

  if (hostPattern.test(existingConfig)) {
    // Replace existing config
    existingConfig = existingConfig.replace(hostPattern, configEntry.trim() + '\n');
  } else {
    // Append new config
    existingConfig += '\n' + configEntry;
  }

  // Write updated config to Kloudlite-specific config file
  fs.writeFileSync(kloudliteConfigPath, existingConfig.trim() + '\n', { mode: 0o600 });
  console.log(`[Kloudlite] SSH config updated for workspace: ${workspaceName}`);
  console.log(`[Kloudlite] Config file location: ${kloudliteConfigPath}`);
}

export function deactivate() {}
