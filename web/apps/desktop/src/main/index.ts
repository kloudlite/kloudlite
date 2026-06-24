import { app, BrowserWindow, Menu, MenuItem, nativeImage, nativeTheme, shell, ipcMain, webContents } from 'electron'

// Disable hardware acceleration check — skip GPU init for faster startup on some systems
// app.disableHardwareAcceleration()  // Uncomment if GPU init is slow

// Disable chromium security warnings in dev
process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = 'true'
import { join } from 'path'
import { is } from '@electron-toolkit/utils'
import { startMCPServer, stopMCPServer, receiveMCPResult } from './mcp-server'
import { platformAPI } from './platform-api'
// Lazy-load tls only when needed (speeds up startup)
type PeerCertificate = import('tls').PeerCertificate

// Must be set before app ready for macOS dock label
if (process.platform === 'darwin') {
  app.setName('Kloudlite')
}

let mainWindow: BrowserWindow | null = null
const pendingOpenUrls: string[] = []

function isWebUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

function openWebUrl(url: string): void {
  if (!isWebUrl(url)) return
  const win = mainWindow ?? BrowserWindow.getAllWindows()[0]
  if (!win) {
    pendingOpenUrls.push(url)
    return
  }
  if (win.isMinimized()) win.restore()
  win.show()
  win.focus()
  win.webContents.send('open-url-in-new-tab', url)
}

function drainPendingOpenUrls(): void {
  for (const url of pendingOpenUrls.splice(0)) openWebUrl(url)
}

const gotSingleInstanceLock = app.requestSingleInstanceLock()
if (!gotSingleInstanceLock) app.quit()

app.on('second-instance', (_event, argv) => {
  const url = argv.find(isWebUrl)
  if (url) openWebUrl(url)
})

app.on('open-url', (event, url) => {
  event.preventDefault()
  openWebUrl(url)
})

function sendToMenuWindow(window: Electron.BaseWindow | undefined, channel: string, ...args: string[]): void {
  if (window instanceof BrowserWindow) window.webContents.send(channel, ...args)
}

function setDockIcon(): void {
  if (process.platform !== 'darwin') return
  const icon = nativeImage.createFromPath(join(__dirname, '../../resources/icon.png'))
  if (!icon.isEmpty()) app.dock?.setIcon(icon)
}

function shortcutAction(input: Electron.Input): string | null {
  const key = input.key.toLowerCase()
  const cmdOrCtrl = input.meta || input.control

  if (input.control && key === 'tab') return input.shift ? 'prev-tab' : 'next-tab'
  if (cmdOrCtrl && key === 't') return 'new-tab'
  if (cmdOrCtrl && key === 'l') return 'address-bar'
  if (cmdOrCtrl && key === 'w') return 'close-tab'
  if (cmdOrCtrl && key === 'r') return 'reload'
  if (cmdOrCtrl && key === 's') return 'toggle-sidebar'
  if (cmdOrCtrl && key === '1') return 'mode-1'
  if (cmdOrCtrl && key === '2') return 'mode-2'
  if (cmdOrCtrl && key === '3') return 'mode-3'
  const left = key === 'left' || key === 'arrowleft'
  const right = key === 'right' || key === 'arrowright'

  if (cmdOrCtrl && (key === '[' || left)) return 'go-back'
  if (cmdOrCtrl && (key === ']' || right)) return 'go-forward'
  if (input.alt && !input.meta && !input.control && !input.shift && left) return 'go-back'
  if (input.alt && !input.meta && !input.control && !input.shift && right) return 'go-forward'

  return null
}

function sendShortcut(action: string): void {
  BrowserWindow.getFocusedWindow()?.webContents.send('shortcut', action)
}

function installWindowShortcuts(contents: Electron.WebContents): void {
  contents.on('before-input-event', (event, input) => {
    const action = shortcutAction(input)
    if (!action) return
    event.preventDefault()
    sendShortcut(action)
  })
}

function createWindow(): BrowserWindow {
  const iconPath = process.platform === 'darwin'
    ? join(__dirname, '../../resources/icon.icns')
    : join(__dirname, '../../resources/icon.png')

  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 600,
    minHeight: 400,
    frame: false,
    show: false,
    icon: iconPath,
    backgroundColor: '#1a1b2e',
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      sandbox: false,
      webviewTag: true
    }
  })

  // Maximize immediately before showing — no flash
  mainWindow.maximize()

  mainWindow.once('ready-to-show', () => {
    mainWindow?.show()
    setTimeout(drainPendingOpenUrls, 250)
  })

  mainWindow.on('closed', () => {
    mainWindow = null
  })

  installWindowShortcuts(mainWindow.webContents)

  mainWindow.webContents.setWindowOpenHandler((details) => {
    shell.openExternal(details.url)
    return { action: 'deny' }
  })

  if (is.dev && process.env['ELECTRON_RENDERER_URL']) {
    mainWindow.loadURL(process.env['ELECTRON_RENDERER_URL'])
  } else {
    mainWindow.loadFile(join(__dirname, '../renderer/index.html'))
  }

  return mainWindow
}

// Custom app menu — no Cmd+R so our renderer shortcut controls per-tab reload
function createAppMenu(): void {
  const template: Electron.MenuItemConstructorOptions[] = [
    {
      label: app.name,
      submenu: [
        { role: 'about' },
        { type: 'separator' },
        { role: 'services' },
        { type: 'separator' },
        { role: 'hide' },
        { role: 'hideOthers' },
        { role: 'unhide' },
        { type: 'separator' },
        { role: 'quit' }
      ]
    },
    {
      label: 'Edit',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectAll' }
      ]
    },
    {
      label: 'Tab',
      submenu: [
        {
          label: 'New Tab',
          accelerator: 'CmdOrCtrl+T',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'new-tab')
          }
        },
        {
          label: 'Address Bar',
          accelerator: 'CmdOrCtrl+L',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'address-bar')
          }
        },
        {
          label: 'Close Tab',
          accelerator: 'CmdOrCtrl+W',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'close-tab')
          }
        },
        { type: 'separator' },
        {
          label: 'Next Tab',
          accelerator: 'Ctrl+Tab',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'next-tab')
          }
        },
        {
          label: 'Previous Tab',
          accelerator: 'Ctrl+Shift+Tab',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'prev-tab')
          }
        },
        { type: 'separator' },
        {
          label: 'Reload Tab',
          accelerator: 'CmdOrCtrl+R',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'reload')
          }
        },
        {
          label: 'Go Back',
          accelerator: 'CmdOrCtrl+[',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'go-back')
          }
        },
        {
          label: 'Go Back (Arrow)',
          accelerator: 'CmdOrCtrl+Left',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'go-back')
          },
          visible: false
        },
        {
          label: 'Go Forward',
          accelerator: 'CmdOrCtrl+]',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'go-forward')
          }
        },
        {
          label: 'Go Forward (Arrow)',
          accelerator: 'CmdOrCtrl+Right',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'go-forward')
          },
          visible: false
        },
      ]
    },
    {
      label: 'View',
      submenu: [
        {
          label: 'Environments',
          accelerator: 'CmdOrCtrl+1',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'mode-1')
          }
        },
        {
          label: 'Workspaces',
          accelerator: 'CmdOrCtrl+2',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'mode-2')
          }
        },
        {
          label: 'Browse',
          accelerator: 'CmdOrCtrl+3',
          click: (_item, window) => {
            sendToMenuWindow(window, 'shortcut', 'mode-3')
          }
        },
      ]
    },
    {
      label: 'Window',
      submenu: [
        { role: 'minimize' },
        { role: 'zoom' },
        { type: 'separator' },
        { role: 'front' },
        { type: 'separator' },
        { role: 'togglefullscreen' }
      ]
    }
  ]

  if (is.dev) {
    template.push({
      label: 'Developer',
      submenu: [
        {
          label: 'Toggle Shell DevTools',
          accelerator: 'Alt+CmdOrCtrl+I',
          click: (_item, window) => {
            if (window instanceof BrowserWindow) window.webContents.toggleDevTools()
          }
        }
      ]
    })
  }

  Menu.setApplicationMenu(Menu.buildFromTemplate(template))
}



// IPC: window controls (close, minimize, maximize)
ipcMain.handle('window-control', (event, action: string) => {
  const win = BrowserWindow.fromWebContents(event.sender)
  if (!win) return
  if (action === 'close') win.close()
  else if (action === 'minimize') win.minimize()
  else if (action === 'maximize') {
    if (win.isMaximized()) win.unmaximize()
    else win.maximize()
  }
})


// IPC: get current theme
ipcMain.handle('get-theme', () => {
  return nativeTheme.shouldUseDarkColors ? 'dark' : 'light'
})

ipcMain.handle('open-external', (_event, url: string) => shell.openExternal(url))

// Notify renderer when system theme changes
nativeTheme.on('updated', () => {
  const theme = nativeTheme.shouldUseDarkColors ? 'dark' : 'light'
  for (const win of BrowserWindow.getAllWindows()) {
    win.webContents.send('theme-changed', theme)
  }
})

// IPC: open devtools for a webview by its webContents id
ipcMain.handle('open-devtools', (_event, webContentsId: number) => {
  const wc = webContents.fromId(webContentsId)
  if (wc) {
    wc.openDevTools({ mode: 'detach' })
  }
})

// IPC: get TLS certificate for a URL
ipcMain.handle('get-certificate', async (_event, url: string) => {
  try {
    const parsed = new URL(url)
    if (parsed.protocol !== 'https:') return null

    const hostname = parsed.hostname
    const port = parseInt(parsed.port) || 443

    // Lazy-load tls module (speeds up app startup)
    const { connect: tlsConnect } = await import('tls')

    return await new Promise((resolve) => {
      const socket = tlsConnect({ host: hostname, port, servername: hostname, rejectUnauthorized: false }, () => {
        const cert = socket.getPeerCertificate(true)
        if (!cert || !cert.subject) {
          socket.destroy()
          resolve(null)
          return
        }

        function formatCert(c: PeerCertificate) {
          return {
            subject: c.subject,
            issuer: c.issuer,
            validFrom: c.valid_from,
            validTo: c.valid_to,
            fingerprint: c.fingerprint,
            fingerprint256: c.fingerprint256,
            serialNumber: c.serialNumber,
            subjectaltname: c.subjectaltname
          }
        }

        // Build certificate chain
        const chain: ReturnType<typeof formatCert>[] = []
        let current: PeerCertificate | undefined = cert
        const seen = new Set<string>()
        while (current && current.fingerprint256 && !seen.has(current.fingerprint256)) {
          seen.add(current.fingerprint256)
          chain.push(formatCert(current))
          current = (current as any).issuerCertificate
        }

        const authorized = socket.authorized

        socket.destroy()
        resolve({ chain, authorized })
      })

      socket.on('error', () => resolve(null))
      setTimeout(() => { socket.destroy(); resolve(null) }, 5000)
    })
  } catch {
    return null
  }
})

// IPC: show context menu for a webview
ipcMain.handle('show-context-menu', (event, webContentsId: number, x: number, y: number) => {
  const wc = webContentsId ? webContents.fromId(webContentsId) : null
  const window = BrowserWindow.fromWebContents(event.sender)
  if (!window) return

  const menu = new Menu()

  if (wc) {
    menu.append(new MenuItem({
      label: 'Back',
      enabled: wc.canGoBack(),
      click: () => wc.goBack()
    }))
    menu.append(new MenuItem({
      label: 'Forward',
      enabled: wc.canGoForward(),
      click: () => wc.goForward()
    }))
    menu.append(new MenuItem({
      label: 'Reload',
      click: () => wc.reload()
    }))
    menu.append(new MenuItem({ type: 'separator' }))
  }

  menu.append(new MenuItem({ label: 'Cut', role: 'cut' }))
  menu.append(new MenuItem({ label: 'Copy', role: 'copy' }))
  menu.append(new MenuItem({ label: 'Paste', role: 'paste' }))
  menu.append(new MenuItem({ label: 'Select All', role: 'selectAll' }))

  if (wc) {
    menu.append(new MenuItem({ type: 'separator' }))
    menu.append(new MenuItem({
      label: 'Inspect Element',
      click: () => {
        wc.inspectElement(x, y)
      }
    }))
    menu.append(new MenuItem({
      label: 'Open DevTools',
      click: () => {
        wc.openDevTools({ mode: 'detach' })
      }
    }))
  }

  menu.popup({ window })
})

// IPC: show a generic popup menu with custom items
ipcMain.handle('show-popup-menu', (event, items: { label: string; id: string; type?: 'separator' | 'normal'; danger?: boolean }[]) => {
  const window = BrowserWindow.fromWebContents(event.sender)
  if (!window) return null

  return new Promise<string | null>((resolve) => {
    const menu = new Menu()
    for (const item of items) {
      if (item.type === 'separator') {
        menu.append(new MenuItem({ type: 'separator' }))
      } else {
        menu.append(new MenuItem({
          label: item.label,
          click: () => resolve(item.id),
        }))
      }
    }
    menu.on('menu-will-close', () => {
      setTimeout(() => resolve(null), 100)
    })
    menu.popup({ window })
  })
})

// IPC: debug log from renderer
ipcMain.on('debug-log', (_event, ...args: unknown[]) => {
  console.log('[renderer]', ...args)
})

// IPC: receive MCP browser command results from renderer
ipcMain.on('mcp-browser-result', (_event, { requestId, result, error }) => {
  receiveMCPResult(requestId, result, error)
})

// IPC: Platform API — list environments
ipcMain.handle('api:list-environments', async (_event, namespace: string) => {
  try {
    const result = await platformAPI.listEnvironments(namespace)
    const count = result.items?.length || 0
    const names = (result.items || []).map((i: any) => i.metadata?.name).join(', ')
    // Log composeStatus details for each environment
    for (const item of (result.items || []) as any[]) {
      const name = item.metadata?.name
      const cs = item.status?.composeStatus
      const svcs = cs?.services || []
      const svcNames = svcs.map((s: any) => s.name).join(',')
      console.log(`[api] env "${name}": composeStatus.services=[${svcNames}] (${svcs.length})`)
    }
    console.log(`[api] list-environments: ${count} items: ${names}`)
    return { items: result.items }
  } catch (err) {
    console.error(`[api] list-environments error:`, (err as Error).message)
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — create environment
ipcMain.handle('api:create-environment', async (_event, namespace: string, name: string, spec: Record<string, unknown>) => {
  try {
    return await platformAPI.createEnvironment(namespace, name, spec)
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — delete environment
ipcMain.handle('api:delete-environment', async (_event, namespace: string, name: string) => {
  try {
    await platformAPI.deleteEnvironment(namespace, name)
    return { success: true }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — list workspaces
ipcMain.handle('api:list-workspaces', async (_event, namespace: string) => {
  try {
    const result = await platformAPI.listWorkspaces(namespace)
    return { items: result.items }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — list generic resources
ipcMain.handle('api:list-resources', async (_event, namespace: string | null, resource: string) => {
  try {
    const result = await platformAPI.listResources(namespace, resource)
    return { items: result.items }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — create generic resource
ipcMain.handle('api:create-resource', async (_event, namespace: string | null, resource: string, object: Record<string, unknown>) => {
  try {
    return await platformAPI.createResource(namespace, resource, object)
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — delete generic resource
ipcMain.handle('api:delete-resource', async (_event, namespace: string | null, resource: string, name: string) => {
  try {
    await platformAPI.deleteResource(namespace, resource, name)
    return { success: true }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — create workspace
ipcMain.handle('api:create-workspace', async (_event, namespace: string, name: string, spec: Record<string, unknown>) => {
  try {
    return await platformAPI.createWorkspace(namespace, name, spec)
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — delete workspace
ipcMain.handle('api:delete-workspace', async (_event, namespace: string, name: string) => {
  try {
    await platformAPI.deleteWorkspace(namespace, name)
    return { success: true }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — list work machines
ipcMain.handle('api:list-workmachines', async () => {
  try {
    const result = await platformAPI.listWorkMachines()
    return { items: result.items }
  } catch (err) {
    return { error: (err as Error).message }
  }
})

// IPC: Platform API — patch resource (e.g., update compose on environment)
ipcMain.handle('api:patch-resource', async (_event, namespace: string | null, resource: string, name: string, patch: Record<string, unknown>) => {
  try {
    console.log(`[api] PATCH ${resource}/${name} called`)
    const result = await platformAPI.patchResource(namespace, resource, name, patch)
    const resultSvcs = (result as any)?.status?.composeStatus?.services || []
    console.log(`[api] PATCH response composeStatus.services=[${resultSvcs.map((s: any) => s.name).join(',')}]`)
    return result
  } catch (err) {
    console.error(`[api] PATCH error:`, (err as Error).message)
    return { error: (err as Error).message }
  }
})

// Handle new-window for webview guests — prevent popups, navigate in app instead
app.on('web-contents-created', (_event, contents) => {
  if (contents.getType() === 'webview') {
    installWindowShortcuts(contents)
    contents.setWindowOpenHandler(({ url, disposition }) => {
      const win = BrowserWindow.getAllWindows()[0]
      if (win) {
        if (disposition === 'new-window' || disposition === 'foreground-tab' || disposition === 'background-tab') {
          // target=_blank or Cmd+click → new tab
          win.webContents.send('open-url-in-new-tab', url)
        } else {
          // Same-window navigation — load in the webview itself
          contents.loadURL(url)
        }
      }
      return { action: 'deny' }
    })
  }
})

app.whenReady().then(() => {
  setDockIcon()

  // Critical path: create window ASAP (everything else is deferred)
  const mainWindow = createWindow()
  const launchUrl = process.argv.find(isWebUrl)
  if (launchUrl) pendingOpenUrls.push(launchUrl)

  // Defer non-critical work until after window is shown
  mainWindow.once('ready-to-show', () => {
    // Create menu (not needed for first paint)
    createAppMenu()

    setDockIcon()

    // Start browser MCP server for AI control
    startMCPServer()
  })

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  stopMCPServer()
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', () => {
  stopMCPServer()
})
