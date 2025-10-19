import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';
import { jwtDecode } from 'jwt-decode';

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
  private connectionToken: string | null = null;

  constructor() {
    console.log('[Kloudlite API] Constructor called');
    try {
      const config = vscode.workspace.getConfiguration('kloudlite');
      const apiUrl = config.get<string>('apiUrl', 'http://localhost:8080');
      console.log('[Kloudlite API] API URL:', apiUrl);

      this.client = axios.create({
        baseURL: apiUrl,
        timeout: 10000,
        headers: {
          'Content-Type': 'application/json',
        },
      });
      console.log('[Kloudlite API] Axios client created');

      // Add request interceptor to include authentication token
      this.client.interceptors.request.use((config) => {
        if (this.connectionToken) {
          config.headers.Authorization = `Bearer ${this.connectionToken}`;
        }
        return config;
      });
      console.log('[Kloudlite API] Request interceptor added');

      // Load saved connection token synchronously
      const savedToken = config.get<string>('connectionToken');
      if (savedToken) {
        this.connectionToken = savedToken;
        console.log('[Kloudlite API] Loaded saved token from settings');
        // We'll validate it lazily on first API call
      }
      console.log('[Kloudlite API] Constructor completed successfully');
    } catch (error) {
      console.error('[Kloudlite API] Constructor error:', error);
      throw error;
    }
  }

  async setConnectionToken(token: string): Promise<void> {
    try {
      // Decode the JWT to extract the apiUrl
      const decoded = jwtDecode<any>(token);

      if (decoded.apiUrl) {
        // Update the API URL from the JWT
        await vscode.workspace.getConfiguration('kloudlite').update(
          'apiUrl',
          decoded.apiUrl,
          vscode.ConfigurationTarget.Global
        );

        // Recreate the axios client with the new base URL
        this.client = axios.create({
          baseURL: decoded.apiUrl,
          timeout: 10000,
          headers: {
            'Content-Type': 'application/json',
          },
        });

        // Re-add the request interceptor
        this.client.interceptors.request.use((config) => {
          if (this.connectionToken) {
            config.headers.Authorization = `Bearer ${this.connectionToken}`;
          }
          return config;
        });

        this.connectionToken = token;

        // Save token to settings
        await vscode.workspace.getConfiguration('kloudlite').update(
          'connectionToken',
          token,
          vscode.ConfigurationTarget.Global
        );

        vscode.window.showInformationMessage(`Connected to ${decoded.apiUrl}`);
      } else {
        // No apiUrl in JWT, use default
        this.connectionToken = token;

        // Save token to settings
        await vscode.workspace.getConfiguration('kloudlite').update(
          'connectionToken',
          token,
          vscode.ConfigurationTarget.Global
        );

        vscode.window.showInformationMessage('Connection token saved successfully');
      }
    } catch (e) {
      console.error('Failed to decode JWT:', e);
      vscode.window.showErrorMessage('Invalid connection token format');
    }
  }

  isAuthenticated(): boolean {
    return this.connectionToken !== null;
  }

  async disconnect(): Promise<void> {
    this.connectionToken = null;
    await vscode.workspace.getConfiguration('kloudlite').update(
      'connectionToken',
      undefined,
      vscode.ConfigurationTarget.Global
    );
    vscode.window.showInformationMessage('Disconnected from Kloudlite');
  }

  async listWorkspaces(): Promise<Workspace[]> {
    try {
      const response = await this.client.get('/api/vscode/workspaces');
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

  async addSSHKey(publicKey: string): Promise<void> {
    try {
      console.log('[Kloudlite API] Adding SSH key...');
      console.log('[Kloudlite API] Public key preview:', publicKey.substring(0, 50) + '...');

      // Get current SSH keys from work machine
      const workMachine = await this.client.get('/api/vscode/work-machines/my');
      console.log('[Kloudlite API] Current work machine:', JSON.stringify(workMachine.data, null, 2));

      const currentKeys = workMachine.data.spec?.sshPublicKeys || [];
      console.log('[Kloudlite API] Current SSH keys count:', currentKeys.length);

      // Check if key already exists
      if (currentKeys.includes(publicKey)) {
        console.log('[Kloudlite API] SSH key already exists');
        return;
      }

      // Add new key
      const updatedKeys = [...currentKeys, publicKey];
      console.log('[Kloudlite API] Updating with keys count:', updatedKeys.length);
      console.log('[Kloudlite API] Request body:', JSON.stringify({ sshPublicKeys: updatedKeys }));

      const response = await this.client.put('/api/vscode/work-machines/my', {
        sshPublicKeys: updatedKeys
      });

      console.log('[Kloudlite API] SSH key added successfully, response:', JSON.stringify(response.data, null, 2));
    } catch (error: any) {
      console.error('[Kloudlite API] Failed to add SSH key:', error);
      console.error('[Kloudlite API] Error response:', error.response?.data);
      console.error('[Kloudlite API] Error status:', error.response?.status);
      throw new Error(`Failed to add SSH key: ${error.response?.data?.error || error.message}`);
    }
  }
}
