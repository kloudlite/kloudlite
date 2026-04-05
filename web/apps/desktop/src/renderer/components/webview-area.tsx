import { useEffect, useRef, useCallback } from 'react'
import { useTabStore } from '@/store/tabs'

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
  const { tabs, activeTabId, updateTab } = useTabStore()
  const webviewRefs = useRef<Map<string, WebviewElement>>(new Map())
  const containerRef = useRef<HTMLDivElement>(null)

  const getActiveWebview = useCallback(() => {
    if (!activeTabId) return null
    return webviewRefs.current.get(activeTabId) ?? null
  }, [activeTabId])

  useEffect(() => {
    onHandle({
      navigate: (url: string) => {
        const wv = getActiveWebview()
        if (wv) {
          wv.loadURL(url)
          if (activeTabId) {
            updateTab(activeTabId, { url })
          }
        }
      },
      goBack: () => getActiveWebview()?.goBack(),
      goForward: () => getActiveWebview()?.goForward(),
      reload: () => getActiveWebview()?.reload()
    })
  }, [activeTabId, getActiveWebview, onHandle, updateTab])

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
        }
      }
    }

    for (const tab of tabs) {
      if (!webviewRefs.current.has(tab.id)) {
        const wv = document.createElement('webview') as unknown as WebviewElement
        wv.setAttribute('style', 'width:100%;height:100%;border:none;position:absolute;inset:0;')
        wv.setAttribute('allowpopups', '')

        wv.addEventListener('did-start-loading', () => {
          updateTab(tab.id, { isLoading: true })
        })

        wv.addEventListener('did-stop-loading', () => {
          updateTab(tab.id, {
            isLoading: false,
            url: wv.getURL(),
            title: wv.getTitle() || wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        wv.addEventListener('page-title-updated', ((e: CustomEvent<{ title: string }>) => {
          updateTab(tab.id, { title: (e as any).title || wv.getTitle() })
        }) as EventListener)

        wv.addEventListener('page-favicon-updated', ((e: CustomEvent<{ favicons: string[] }>) => {
          const favicons = (e as any).favicons as string[] | undefined
          if (favicons && favicons.length > 0) {
            updateTab(tab.id, { favicon: favicons[0] })
          }
        }) as EventListener)

        wv.addEventListener('did-navigate', () => {
          updateTab(tab.id, {
            url: wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        wv.addEventListener('did-navigate-in-page', () => {
          updateTab(tab.id, {
            url: wv.getURL(),
            canGoBack: wv.canGoBack(),
            canGoForward: wv.canGoForward()
          })
        })

        if (tab.url) {
          wv.src = tab.url
        }

        container.appendChild(wv as unknown as Node)
        webviewRefs.current.set(tab.id, wv)
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
      {tabs.length === 0 && (
        <div className="flex h-full items-center justify-center text-muted-foreground">
          <p className="text-sm">Press <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-xs">Cmd+T</kbd> to open a new tab</p>
        </div>
      )}
    </div>
  )
}
