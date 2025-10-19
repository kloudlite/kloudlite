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
    // Only show root level - no children
    if (element) {
      return [];
    }

    // Root level - check authentication first
    if (!this.api.isAuthenticated()) {
      // Show authentication prompt
      const authItem = new WorkspaceItem(
        'Add Connection Token',
        'Click to authenticate',
        {} as Workspace,
        vscode.TreeItemCollapsibleState.None
      );
      authItem.command = {
        command: 'kloudlite.setConnectionToken',
        title: 'Add Connection Token'
      };
      authItem.iconPath = new vscode.ThemeIcon('key', new vscode.ThemeColor('terminal.ansiYellow'));
      authItem.contextValue = 'auth-prompt';
      return [authItem];
    }

    // Show workspaces in flat list
    try {
      const workspaces = await this.api.listWorkspaces();
      if (workspaces.length === 0) {
        const emptyItem = new WorkspaceItem(
          'No workspaces found',
          'Create a workspace in the web dashboard',
          {} as Workspace,
          vscode.TreeItemCollapsibleState.None
        );
        emptyItem.iconPath = new vscode.ThemeIcon('info');
        emptyItem.contextValue = 'empty';
        return [emptyItem];
      }

      // Create flat list with connect on click
      return workspaces.map(ws => {
        const item = new WorkspaceItem(
          ws.metadata.name,
          ws.status?.phase || 'Unknown',
          ws,
          vscode.TreeItemCollapsibleState.None // No children, flat list
        );

        // Add click command to connect - pass the workspace object
        item.command = {
          command: 'kloudlite.connectWorkspace',
          title: 'Connect to Workspace',
          arguments: [ws]
        };

        return item;
      });
    } catch (error) {
      // If authentication error, show auth prompt
      if (error instanceof Error && error.message.includes('401')) {
        const authItem = new WorkspaceItem(
          'Authentication Required',
          'Click to add connection token',
          {} as Workspace,
          vscode.TreeItemCollapsibleState.None
        );
        authItem.command = {
          command: 'kloudlite.setConnectionToken',
          title: 'Add Connection Token'
        };
        authItem.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('terminal.ansiRed'));
        authItem.contextValue = 'auth-error';
        return [authItem];
      }
      vscode.window.showErrorMessage(`Failed to load workspaces: ${error}`);
      return [];
    }
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
