import { useEffect, useRef, useCallback, useState } from 'react'
import { useTabStore } from '@/store/tabs'
import { NavIndicator } from './nav-indicator'

declare global {
  interface Window {
    electronAPI: {
      platform: string
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
  }
}

interface WebviewElement extends HTMLElement {
  src: string
  loadURL: (url: string) => Promise<void>
  goBack: () => void
  goForward: () => void
  reload: () => void
  canGoBack: () => boolean
  canGoForward: () => boolean
  getURL: () => string
  getTitle: () => string
  getWebContentsId: () => number
  addEventListener: HTMLElement['addEventListener']
  removeEventListener: HTMLElement['removeEventListener']
}

export interface WebviewAreaHandle {
  navigate: (url: string) => void
  goBack: () => void
  goForward: () => void
  reload: () => void
}

interface WebviewAreaProps {
  onHandle: (handle: WebviewAreaHandle) => void
}

export function WebviewArea({ onHandle }: WebviewAreaProps) {
  const { tabs, activeTabId, updateTab, addTab } = useTabStore()
  const webviewRefs = useRef<Map<string, WebviewElement>>(new Map())
  const readyRefs = useRef<Set<string>>(new Set())
  const containerRef = useRef<HTMLDivElement>(null)
  const [navFlash, setNavFlash] = useState<'back' | 'forward' | null>(null)
  const navFlashCounter = useRef(0)

  const getActiveWebview = useCallback(() => {
    if (!activeTabId) return null
    return webviewRefs.current.get(activeTabId) ?? null
  }, [activeTabId])

  useEffect(() => {
    onHandle({
      navigate: (url: string) => {
        const wv = getActiveWebview()
        if (!wv || !activeTabId) return

        updateTab(activeTabId, { url })

        if (readyRefs.current.has(activeTabId)) {
          wv.loadURL(url)
        } else {
          wv.src = url
        }
      },
      goBack: () => {
        const wv = getActiveWebview()
        if (wv && activeTabId && readyRefs.current.has(activeTabId)) wv.goBack()
      },
      goForward: () => {
        const wv = getActiveWebview()
        if (wv && activeTabId && readyRefs.current.has(activeTabId)) wv.goForward()
      },
      reload: () => {
        const wv = getActiveWebview()
        if (wv && activeTabId && readyRefs.current.has(activeTabId)) wv.reload()
      }
    })
  }, [activeTabId, getActiveWebview, onHandle, updateTab])

  // Listen for new-tab URLs from main process (webview popup interception)
  useEffect(() => {
    window.electronAPI.onOpenUrlInNewTab((url) => {
      addTab(url)
    })
  }, [addTab])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const currentIds = new Set(tabs.map((t) => t.id))
    const existingIds = new Set(webviewRefs.current.keys())

    for (const id of existingIds) {
      if (!currentIds.has(id)) {
        const wv = webviewRefs.current.get(id)
        if (wv) {
          container.removeChild(wv)
          webviewRefs.current.delete(id)
          readyRefs.current.delete(id)
        }
      }
    }

    for (const tab of tabs) {
      if (!webviewRefs.current.has(tab.id)) {
        const wv = document.createElement('webview') as unknown as WebviewElement
        wv.setAttribute('style', 'width:100%;height:100%;border:none;position:absolute;inset:0;')
        wv.setAttribute('allowpopups', '')
        wv.setAttribute('preload', `file://${window.electronAPI.webviewPreload}`)

        const tabId = tab.id

        wv.addEventListener('dom-ready', () => {
          readyRefs.current.add(tabId)
        })

        // IPC messages from webview preload
        wv.addEventListener('ipc-message', ((e: any) => {
          if (e.channel === 'context-menu') {
            const [x, y] = e.args
            const wcId = wv.getWebContentsId()
            window.electronAPI.showContextMenu(wcId, x, y)
          } else if (e.channel === 'swipe-navigate') {
            const direction = e.args[0] as 'back' | 'forward'
            if (readyRefs.current.has(tabId)) {
              if (direction === 'back') wv.goBack()
              else wv.goForward()
              // Trigger flash — use counter to re-trigger even for same direction
              navFlashCounter.current++
              setNavFlash(direction)
              setTimeout(() => setNavFlash(null), 500)
            }
          }
        }) as EventListener)

        wv.addEventListener('did-start-loading', () => {
          updateTab(tabId, { isLoading: true })
        })

        wv.addEventListener('did-stop-loading', () => {
          updateTab(tabId, {
            isLoading: false,
            url: wv.getURL(),
            title: wv.getTitle() || wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        wv.addEventListener('page-title-updated', ((e: any) => {
          updateTab(tabId, { title: e.title || wv.getTitle() })
        }) as EventListener)

        wv.addEventListener('page-favicon-updated', ((e: any) => {
          const favicons = e.favicons as string[] | undefined
          if (favicons && favicons.length > 0) {
            updateTab(tabId, { favicon: favicons[0] })
          }
        }) as EventListener)

        wv.addEventListener('did-navigate', () => {
          updateTab(tabId, {
            url: wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        wv.addEventListener('did-navigate-in-page', () => {
          updateTab(tabId, {
            url: wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        // Set initial src — webview needs a src to initialize
        if (tab.url) {
          wv.src = tab.url
        } else {
          wv.src = 'about:blank'
        }

        container.appendChild(wv as unknown as Node)
        webviewRefs.current.set(tabId, wv)
      }
    }
  }, [tabs, updateTab])

  useEffect(() => {
    for (const [id, wv] of webviewRefs.current) {
      ;(wv as unknown as HTMLElement).style.display = id === activeTabId ? 'flex' : 'none'
    }
  }, [activeTabId])

  return (
    <div ref={containerRef} className="relative flex-1">
      <NavIndicator direction={navFlash} />
      {tabs.length === 0 && (
        <div className="flex h-full items-center justify-center text-muted-foreground">
          <p className="text-sm">Press <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-xs">Cmd+T</kbd> to open a new tab</p>
        </div>
      )}
    </div>
  )
}
