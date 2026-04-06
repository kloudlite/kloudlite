import { app, BrowserWindow, globalShortcut, Menu, MenuItem, nativeImage, nativeTheme, shell, ipcMain, webContents } from 'electron'
import { join } from 'path'
import { connect as tlsConnect, type PeerCertificate } from 'tls'
import { is } from '@electron-toolkit/utils'

// Must be set before app ready for macOS dock label
if (process.platform === 'darwin') {
  app.setName('Kloudlite')
}

function createWindow(): BrowserWindow {
  const iconPath = process.platform === 'darwin'
    ? join(__dirname, '../../resources/icon.icns')
    : join(__dirname, '../../resources/icon.png')

  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 600,
    minHeight: 400,
    frame: false,
    icon: iconPath,
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      sandbox: false,
      webviewTag: true
    }
  })

  mainWindow.on('ready-to-show', () => {
    mainWindow.maximize()
    mainWindow.show()
  })


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
            window?.webContents.send('shortcut', 'new-tab')
          }
        },
        {
          label: 'Address Bar',
          accelerator: 'CmdOrCtrl+L',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'address-bar')
          }
        },
        {
          label: 'Close Tab',
          accelerator: 'CmdOrCtrl+W',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'close-tab')
          }
        },
        { type: 'separator' },
        {
          label: 'Reload Tab',
          accelerator: 'CmdOrCtrl+R',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'reload')
          }
        },
        {
          label: 'Go Back',
          accelerator: 'CmdOrCtrl+[',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'go-back')
          }
        },
        {
          label: 'Go Back (Arrow)',
          accelerator: 'CmdOrCtrl+Left',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'go-back')
          },
          visible: false
        },
        {
          label: 'Go Forward',
          accelerator: 'CmdOrCtrl+]',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'go-forward')
          }
        },
        {
          label: 'Go Forward (Arrow)',
          accelerator: 'CmdOrCtrl+Right',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'go-forward')
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
            window?.webContents.send('shortcut', 'mode-1')
          }
        },
        {
          label: 'Workspaces',
          accelerator: 'CmdOrCtrl+2',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'mode-2')
          }
        },
        {
          label: 'Browse',
          accelerator: 'CmdOrCtrl+3',
          click: (_item, window) => {
            window?.webContents.send('shortcut', 'mode-3')
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
            window?.webContents.toggleDevTools()
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

// Handle new-window for webview guests — prevent popups, navigate in app instead
app.on('web-contents-created', (_event, contents) => {
  if (contents.getType() === 'webview') {
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
  // Set dock icon (works in dev mode too)
  if (process.platform === 'darwin') {
    const iconPath = join(__dirname, '../../resources/icon.png')
    const icon = nativeImage.createFromPath(iconPath)
    if (!icon.isEmpty()) {
      app.dock.setIcon(icon)
    }
  }

  createAppMenu()
  createWindow()

  // globalShortcut for keys that menu accelerators can't reliably capture
  globalShortcut.register('Ctrl+Tab', () => {
    const win = BrowserWindow.getFocusedWindow()
    win?.webContents.send('shortcut', 'next-tab')
  })
  globalShortcut.register('Ctrl+Shift+Tab', () => {
    const win = BrowserWindow.getFocusedWindow()
    win?.webContents.send('shortcut', 'prev-tab')
  })
  globalShortcut.register('CommandOrControl+S', () => {
    const win = BrowserWindow.getFocusedWindow()
    win?.webContents.send('shortcut', 'toggle-sidebar')
  })

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
