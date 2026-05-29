import { useEffect, useRef, useCallback, useState } from 'react'
import { useTabStore } from '@/store/tabs'
import { useHistoryStore, type PageMetadata } from '@/store/history'
import { NavIndicator } from './nav-indicator'
import { EmptyState } from './empty-state'

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
  const addHistoryEntry = useHistoryStore((s) => s.addEntry)
  const updateHistoryMetadata = useHistoryStore((s) => s.updateMetadata)
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
  // Debounce and deduplicate to prevent redirect chains creating multiple tabs
  useEffect(() => {
    let lastUrl = ''
    let lastTime = 0
    window.electronAPI.onOpenUrlInNewTab((url) => {
      const now = Date.now()
      // Skip if same URL within 1 second, or any URL within 300ms
      if ((url === lastUrl && now - lastTime < 1000) || (now - lastTime < 300)) {
        return
      }
      lastUrl = url
      lastTime = now
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
        wv.setAttribute('scrollbounce', '')
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
              navFlashCounter.current++
              setNavFlash(direction)
              setTimeout(() => setNavFlash(null), 500)
            }
          } else if (e.channel === 'page-metadata') {
            const metadata = e.args[0] as PageMetadata
            const currentUrl = wv.getURL()
            if (currentUrl && currentUrl !== 'about:blank') {
              updateHistoryMetadata(currentUrl, metadata)
              updateTab(tabId, {
                siteName: metadata.siteName,
                keywords: metadata.keywords
              })
            }
          }
        }) as EventListener)

        wv.addEventListener('did-start-loading', () => {
          updateTab(tabId, { isLoading: true })
        })

        wv.addEventListener('did-stop-loading', () => {
          const currentUrl = wv.getURL()
          const currentTitle = wv.getTitle() || currentUrl
          updateTab(tabId, {
            isLoading: false,
            url: currentUrl,
            title: currentTitle,
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
          // Record in history
          const tab = useTabStore.getState().tabs.find(t => t.id === tabId)
          addHistoryEntry(currentUrl, currentTitle, tab?.favicon || '')
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

        function onNavigate() {
          const newUrl = wv.getURL()
          if (!newUrl || newUrl === 'about:blank') return

          const tab = useTabStore.getState().tabs.find(t => t.id === tabId)
          if (!tab) return

          const stack = [...tab.navStack]
          let index = tab.navIndex

          if (index > 0 && stack[index - 1] === newUrl) {
            index--
          } else if (index < stack.length - 1 && stack[index + 1] === newUrl) {
            index++
          } else {
            // Look for the URL anywhere in the existing stack
            // (e.g. jumping 2+ steps back/forward via history popover)
            const existingIdx = stack.indexOf(newUrl)
            if (existingIdx >= 0) {
              index = existingIdx
            } else if (stack[index] !== newUrl) {
              // New navigation — truncate forward history
              stack.splice(index + 1)
              stack.push(newUrl)
              index = stack.length - 1
            }
          }

          updateTab(tabId, {
            url: newUrl,
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward(),
            navStack: stack,
            navIndex: index
          })
        }

        wv.addEventListener('did-navigate', onNavigate)
        wv.addEventListener('did-navigate-in-page', onNavigate)

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
    <div ref={containerRef} className="relative h-full w-full">
      <NavIndicator direction={navFlash} />
      {tabs.length === 0 && (
        <EmptyState
          title="Browse Services"
          description={<>Press <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono">Cmd+T</kbd> or select a service from the sidebar to open a tab</>}
          action={{ label: 'New Tab', onClick: () => addTab() }}
        />
      )}
    </div>
  )
}
