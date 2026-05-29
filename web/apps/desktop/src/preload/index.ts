import 'v8-compile-cache'
import { contextBridge, ipcRenderer } from 'electron'
import { join } from 'path'

const webviewPreloadPath = join(__dirname, './webview.js')

export interface ElectronAPI {
  platform: NodeJS.Platform
  webviewPreload: string
  windowControl: (action: 'close' | 'minimize' | 'maximize') => Promise<void>
  showContextMenu: (webContentsId: number, x: number, y: number) => Promise<void>
  openDevTools: (webContentsId: number) => Promise<void>
  onShortcut: (callback: (action: string) => void) => void
  getTheme: () => Promise<'dark' | 'light'>
  onThemeChanged: (callback: (theme: 'dark' | 'light') => void) => void
  onOpenUrlInNewTab: (callback: (url: string) => void) => void
  getCertificate: (url: string) => Promise<any>
  showPopupMenu: (items: { label: string; id: string; type?: string; danger?: boolean }[]) => Promise<string | null>
  onMCPCommand: (callback: (command: { requestId: string; command: string; args: Record<string, unknown> }) => void) => void
  sendMCPResult: (requestId: string, result: unknown, error: string | null) => void
  // Platform API
  listEnvironments: (namespace: string) => Promise<{ items: Record<string, unknown>[]; error?: string }>
  createEnvironment: (namespace: string, name: string, spec: Record<string, unknown>) => Promise<Record<string, unknown>>
  deleteEnvironment: (namespace: string, name: string) => Promise<{ success?: boolean; error?: string }>
  listWorkspaces: (namespace: string) => Promise<{ items: Record<string, unknown>[]; error?: string }>
  createWorkspace: (namespace: string, name: string, spec: Record<string, unknown>) => Promise<Record<string, unknown>>
  listWorkMachines: () => Promise<{ items: Record<string, unknown>[]; error?: string }>
  patchResource: (namespace: string | null, resource: string, name: string, patch: Record<string, unknown>) => Promise<Record<string, unknown>>
}

const api: ElectronAPI = {
  platform: process.platform,
  webviewPreload: webviewPreloadPath,
  windowControl: (action) => ipcRenderer.invoke('window-control', action),
  showContextMenu: (webContentsId, x, y) => ipcRenderer.invoke('show-context-menu', webContentsId, x, y),
  openDevTools: (webContentsId) => ipcRenderer.invoke('open-devtools', webContentsId),
  onShortcut: (callback) => ipcRenderer.on('shortcut', (_event, action) => callback(action)),
  getTheme: () => ipcRenderer.invoke('get-theme'),
  onThemeChanged: (callback) => ipcRenderer.on('theme-changed', (_event, theme) => callback(theme)),
  onOpenUrlInNewTab: (callback) => ipcRenderer.on('open-url-in-new-tab', (_event, url) => callback(url)),
  getCertificate: (url) => ipcRenderer.invoke('get-certificate', url),
  showPopupMenu: (items) => ipcRenderer.invoke('show-popup-menu', items),
  onMCPCommand: (callback) => ipcRenderer.on('mcp-command', (_event, data) => callback(data)),
  sendMCPResult: (requestId, result, error) => ipcRenderer.send('mcp-browser-result', { requestId, result, error }),
  listEnvironments: (namespace) => ipcRenderer.invoke('api:list-environments', namespace),
  createEnvironment: (namespace, name, spec) => ipcRenderer.invoke('api:create-environment', namespace, name, spec),
  deleteEnvironment: (namespace, name) => ipcRenderer.invoke('api:delete-environment', namespace, name),
  listWorkspaces: (namespace) => ipcRenderer.invoke('api:list-workspaces', namespace),
  createWorkspace: (namespace, name, spec) => ipcRenderer.invoke('api:create-workspace', namespace, name, spec),
  listWorkMachines: () => ipcRenderer.invoke('api:list-workmachines'),
  patchResource: (namespace, resource, name, patch) => ipcRenderer.invoke('api:patch-resource', namespace, resource, name, patch)
}

contextBridge.exposeInMainWorld('electronAPI', api)
