import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';

export interface Workspace {
  metadata: {
    name: string;
    namespace?: string;
    creationTimestamp?: string;
  };
  spec: {
    image?: string;
    resources?: any;
  };
  status?: {
    phase?: string;
    accessUrls?: {
      'code-server'?: string;
      'ttyd'?: string;
      'ssh'?: string;
    };
  };
}

export class KloudliteAPI {
  private client: AxiosInstance;

  constructor() {
    const config = vscode.workspace.getConfiguration('kloudlite');
    const apiUrl = config.get<string>('apiUrl', 'http://localhost:3000');

    this.client = axios.create({
      baseURL: apiUrl,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  async listWorkspaces(): Promise<Workspace[]> {
    try {
      const response = await this.client.get('/api/workspaces');
      return response.data.workspaces || [];
    } catch (error) {
      console.error('Failed to list workspaces:', error);
      throw new Error('Failed to fetch workspaces from Kloudlite API');
    }
  }

  async getWorkspace(name: string): Promise<Workspace> {
    try {
      const response = await this.client.get(`/api/workspaces/${name}`);
      return response.data;
    } catch (error) {
      console.error(`Failed to get workspace ${name}:`, error);
      throw new Error(`Failed to fetch workspace: ${name}`);
    }
  }

  async createWorkspace(workspace: Partial<Workspace>): Promise<Workspace> {
    try {
      const response = await this.client.post('/api/workspaces', workspace);
      return response.data;
    } catch (error) {
      console.error('Failed to create workspace:', error);
      throw new Error('Failed to create workspace');
    }
  }

  async deleteWorkspace(name: string): Promise<void> {
    try {
      await this.client.delete(`/api/workspaces/${name}`);
    } catch (error) {
      console.error(`Failed to delete workspace ${name}:`, error);
      throw new Error(`Failed to delete workspace: ${name}`);
    }
  }
}
