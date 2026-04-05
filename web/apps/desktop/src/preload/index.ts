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
  getCertificate: (url) => ipcRenderer.invoke('get-certificate', url)
}

contextBridge.exposeInMainWorld('electronAPI', api)
