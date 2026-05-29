import { WebContentsView, BrowserWindow, ipcMain } from 'electron'
import { join } from 'path'

const PADDING = 10
const RESIZE_HANDLE = 6

export class TabManager {
  private tabs: Map<string, WebContentsView> = new Map()
  private tabOrder: string[] = []
  private activeTabId: string | null = null
  private sidebarWidth = 360
  private sidebarVisible = true
  private window: BrowserWindow
  private nextId = 1
  private overlayVisible = false

  constructor(window: BrowserWindow) {
    this.window = window
    this.setupIpcListeners()
  }

  private setupIpcListeners(): void {
    ipcMain.on('wv-context-menu', (event) => {
      const webContentsId = event.sender.id
      this.window.webContents.send('show-context-menu', webContentsId)
    })

    ipcMain.on('wv-swipe-navigate', (event, direction: 'back' | 'forward') => {
      const tabId = this.findTabByWebContentsId(event.sender.id)
      if (!tabId) return

      const view = this.tabs.get(tabId)
      if (!view) return

      if (direction === 'back') {
        view.webContents.goBack()
      } else {
        view.webContents.goForward()
      }

      this.window.webContents.send('swipe-navigate', tabId, direction)
    })

    ipcMain.on('wv-page-metadata', (event, metadata: unknown) => {
      const tabId = this.findTabByWebContentsId(event.sender.id)
      if (tabId) {
        this.window.webContents.send('tab-metadata', tabId, metadata)
      }
    })
  }

  private findTabByWebContentsId(webContentsId: number): string | null {
    for (const [tabId, view] of this.tabs) {
      if (view.webContents.id === webContentsId) {
        return tabId
      }
    }
    return null
  }

  createTab(url?: string): string {
    const tabId = String(this.nextId++)
    const targetUrl = url || 'about:blank'

    const view = new WebContentsView({
      webPreferences: {
        preload: join(__dirname, '../preload/webview.js'),
        sandbox: true,
        contextIsolation: true,
      },
    })

    view.setBorderRadius(10)

    const wc = view.webContents

    wc.on('did-start-loading', () => {
      this.window.webContents.send('tab-updated', tabId, { loading: true })
    })

    wc.on('did-stop-loading', () => {
      this.window.webContents.send('tab-updated', tabId, { loading: false })
    })

    wc.on('page-title-updated', (_event, title) => {
      this.window.webContents.send('tab-updated', tabId, { title })
    })

    wc.on('page-favicon-updated', (_event, favicons) => {
      this.window.webContents.send('tab-updated', tabId, {
        favicon: favicons[0],
      })
    })

    wc.on('did-navigate', (_event, navigatedUrl) => {
      this.window.webContents.send('tab-updated', tabId, {
        url: navigatedUrl,
        canGoBack: wc.canGoBack(),
        canGoForward: wc.canGoForward(),
      })
    })

    wc.on('did-navigate-in-page', (_event, navigatedUrl) => {
      this.window.webContents.send('tab-updated', tabId, {
        url: navigatedUrl,
        canGoBack: wc.canGoBack(),
        canGoForward: wc.canGoForward(),
      })
    })

    wc.setWindowOpenHandler(({ url: openUrl, disposition }) => {
      if (
        disposition === 'foreground-tab' ||
        disposition === 'background-tab' ||
        disposition === 'new-window'
      ) {
        this.createTab(openUrl)
      } else {
        wc.loadURL(openUrl)
      }
      return { action: 'deny' }
    })

    wc.loadURL(targetUrl)

    this.tabs.set(tabId, view)
    this.tabOrder.push(tabId)

    this.window.webContents.send('tab-created', tabId, targetUrl)

    this.window.contentView.addChildView(view)
    this.activateTab(tabId)

    return tabId
  }

  closeTab(id: string): void {
    const view = this.tabs.get(id)
    if (!view) return

    this.window.contentView.removeChildView(view)
    view.webContents.close()
    this.tabs.delete(id)

    const index = this.tabOrder.indexOf(id)
    if (index !== -1) {
      this.tabOrder.splice(index, 1)
    }

    this.window.webContents.send('tab-closed', id)

    if (this.activeTabId === id) {
      if (this.tabOrder.length === 0) {
        this.activeTabId = null
      } else {
        const nextIndex = Math.min(index, this.tabOrder.length - 1)
        this.activateTab(this.tabOrder[nextIndex])
      }
    }
  }

  activateTab(id: string): void {
    const view = this.tabs.get(id)
    if (!view) return

    this.activeTabId = id

    for (const [tabId, tabView] of this.tabs) {
      if (tabId === id) {
        if (!this.overlayVisible) {
          tabView.setVisible(true)
        }
      } else {
        tabView.setVisible(false)
      }
    }

    this.applyBounds()
    this.window.webContents.send('tab-activated', id)
  }

  navigate(url: string): void {
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (view) {
      view.webContents.loadURL(url)
    }
  }

  goBack(): void {
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (view && view.webContents.canGoBack()) {
      view.webContents.goBack()
    }
  }

  goForward(): void {
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (view && view.webContents.canGoForward()) {
      view.webContents.goForward()
    }
  }

  reload(): void {
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (view) {
      view.webContents.reload()
    }
  }

  updateBounds(sidebarWidth?: number, sidebarVisible?: boolean): void {
    if (sidebarWidth !== undefined) {
      this.sidebarWidth = sidebarWidth
    }
    if (sidebarVisible !== undefined) {
      this.sidebarVisible = sidebarVisible
    }
    this.applyBounds()
  }

  private applyBounds(): void {
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (!view) return

    const [contentWidth, contentHeight] = this.window.getContentSize()

    const x = this.sidebarVisible
      ? this.sidebarWidth + RESIZE_HANDLE
      : PADDING
    const y = PADDING
    const width = contentWidth - x - PADDING
    const height = contentHeight - y - PADDING

    if (width > 0 && height > 0) {
      view.setBounds({ x, y, width, height })
    }
  }

  setOverlayVisible(visible: boolean): void {
    this.overlayVisible = visible
    if (!this.activeTabId) return
    const view = this.tabs.get(this.activeTabId)
    if (view) {
      view.setVisible(!visible)
    }
  }

  moveTab(fromIndex: number, toIndex: number): void {
    if (
      fromIndex < 0 ||
      fromIndex >= this.tabOrder.length ||
      toIndex < 0 ||
      toIndex >= this.tabOrder.length
    ) {
      return
    }

    const [tabId] = this.tabOrder.splice(fromIndex, 1)
    this.tabOrder.splice(toIndex, 0, tabId)
  }

  getTabIds(): string[] {
    return [...this.tabOrder]
  }
}
