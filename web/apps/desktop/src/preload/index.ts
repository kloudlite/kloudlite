import { contextBridge } from 'electron'

export interface ElectronAPI {
  platform: NodeJS.Platform
}

const api: ElectronAPI = {
  platform: process.platform
}

contextBridge.exposeInMainWorld('electronAPI', api)
