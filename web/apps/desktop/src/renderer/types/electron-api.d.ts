export {}

declare global {
  interface Window {
    electronAPI: {
      platform: string
      webviewPreload: string
      windowControl: (action: 'close' | 'minimize' | 'maximize') => Promise<void>
      showContextMenu: (webContentsId: number, x: number, y: number) => Promise<void>
      showPopupMenu: (items: { label: string; id: string; type?: string; danger?: boolean }[]) => Promise<string | null>
      openDevTools: (webContentsId: number) => Promise<void>
      onShortcut: (callback: (action: string) => void) => void
      getTheme: () => Promise<'dark' | 'light'>
      onThemeChanged: (callback: (theme: 'dark' | 'light') => void) => void
      onOpenUrlInNewTab: (callback: (url: string) => void) => void
      getCertificate: (url: string) => Promise<any>
      onMCPCommand: (callback: (command: { requestId: string; command: string; args: Record<string, unknown> }) => void) => void
      sendMCPResult: (requestId: string, result: unknown, error: string | null) => void
      listEnvironments: (namespace: string) => Promise<{ items: Record<string, unknown>[]; error?: string }>
      createEnvironment: (namespace: string, name: string, spec: Record<string, unknown>) => Promise<Record<string, unknown>>
      deleteEnvironment: (namespace: string, name: string) => Promise<{ success?: boolean; error?: string }>
      listWorkspaces: (namespace: string) => Promise<{ items: Record<string, unknown>[]; error?: string }>
      createWorkspace: (namespace: string, name: string, spec: Record<string, unknown>) => Promise<Record<string, unknown> & { error?: string }>
      deleteWorkspace: (namespace: string, name: string) => Promise<{ success?: boolean; error?: string }>
      listWorkMachines: () => Promise<{ items: Record<string, unknown>[]; error?: string }>
      listResources: (namespace: string | null, resource: string) => Promise<{ items: Record<string, unknown>[]; error?: string }>
      createResource: (namespace: string | null, resource: string, object: Record<string, unknown>) => Promise<Record<string, unknown> & { error?: string }>
      deleteResource: (namespace: string | null, resource: string, name: string) => Promise<{ success?: boolean; error?: string }>
      patchResource: (namespace: string | null, resource: string, name: string, patch: Record<string, unknown>) => Promise<Record<string, unknown> & { error?: string }>
      debugLog: (msg: string) => void
    }
  }
}
