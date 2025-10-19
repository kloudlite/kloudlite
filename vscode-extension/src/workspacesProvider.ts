import * as vscode from 'vscode';
import { KloudliteAPI, Workspace } from './api';

export class KloudliteWorkspacesProvider implements vscode.TreeDataProvider<WorkspaceItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<WorkspaceItem | undefined | void> = new vscode.EventEmitter<WorkspaceItem | undefined | void>();
  readonly onDidChangeTreeData: vscode.Event<WorkspaceItem | undefined | void> = this._onDidChangeTreeData.event;

  constructor(private api: KloudliteAPI) {}

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: WorkspaceItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: WorkspaceItem): Promise<WorkspaceItem[]> {
    if (!element) {
      // Root level - show workspaces
      try {
        const workspaces = await this.api.listWorkspaces();
        return workspaces.map(ws => new WorkspaceItem(
          ws.metadata.name,
          ws.status?.phase || 'Unknown',
          ws,
          vscode.TreeItemCollapsibleState.Collapsed
        ));
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to load workspaces: ${error}`);
        return [];
      }
    } else {
      // Child level - show workspace details
      return this.getWorkspaceDetails(element.workspace);
    }
  }

  private getWorkspaceDetails(workspace: Workspace): WorkspaceItem[] {
    const items: WorkspaceItem[] = [];

    // Status
    items.push(new WorkspaceItem(
      `Status: ${workspace.status?.phase || 'Unknown'}`,
      '',
      workspace,
      vscode.TreeItemCollapsibleState.None
    ));

    // Access URLs
    if (workspace.status?.accessUrls) {
      const urls = workspace.status.accessUrls;
      if (urls['code-server']) {
        const item = new WorkspaceItem(
          'VS Code Web',
          urls['code-server'],
          workspace,
          vscode.TreeItemCollapsibleState.None
        );
        item.command = {
          command: 'vscode.open',
          title: 'Open in Browser',
          arguments: [vscode.Uri.parse(urls['code-server'])]
        };
        items.push(item);
      }

      if (urls['ttyd']) {
        const item = new WorkspaceItem(
          'Web Terminal',
          urls['ttyd'],
          workspace,
          vscode.TreeItemCollapsibleState.None
        );
        item.command = {
          command: 'vscode.open',
          title: 'Open in Browser',
          arguments: [vscode.Uri.parse(urls['ttyd'])]
        };
        items.push(item);
      }

      if (urls['ssh']) {
        items.push(new WorkspaceItem(
          'SSH Port',
          urls['ssh'],
          workspace,
          vscode.TreeItemCollapsibleState.None
        ));
      }
    }

    // Connect action
    const connectItem = new WorkspaceItem(
      'Connect via SSH',
      'Click to connect',
      workspace,
      vscode.TreeItemCollapsibleState.None
    );
    connectItem.command = {
      command: 'kloudlite.connectWorkspace',
      title: 'Connect to Workspace',
      arguments: [workspace]
    };
    items.push(connectItem);

    return items;
  }
}

class WorkspaceItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly description: string,
    public readonly workspace: Workspace,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState
  ) {
    super(label, collapsibleState);
    this.tooltip = `${this.label}${description ? `: ${description}` : ''}`;
    this.contextValue = 'workspace';

    // Set icon based on status
    if (workspace.status?.phase === 'Running') {
      this.iconPath = new vscode.ThemeIcon('vm-running', new vscode.ThemeColor('terminal.ansiGreen'));
    } else if (workspace.status?.phase === 'Pending') {
      this.iconPath = new vscode.ThemeIcon('loading~spin', new vscode.ThemeColor('terminal.ansiYellow'));
    } else {
      this.iconPath = new vscode.ThemeIcon('vm-outline');
    }
  }
}
