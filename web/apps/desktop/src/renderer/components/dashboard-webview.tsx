import { useEffect, useRef, useState } from 'react'

interface WebviewElement extends HTMLElement {
  src: string
  loadURL: (url: string) => Promise<void>
  goBack: () => void
  goForward: () => void
  reload: () => void
  canGoBack: () => boolean
  canGoForward: () => boolean
}

export interface DashboardWebviewHandle {
  navigate: (url: string) => void
  goBack: () => void
  goForward: () => void
  reload: () => void
}

interface DashboardWebviewProps {
  url: string
  visible: boolean
  onHandle?: (handle: DashboardWebviewHandle) => void
}

export function DashboardWebview({ url, visible, onHandle }: DashboardWebviewProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const webviewRef = useRef<WebviewElement | null>(null)
  const readyRef = useRef(false)
  const createdRef = useRef(false)
  const [currentPath, setCurrentPath] = useState(() => {
    try { return new URL(url).pathname } catch { return url }
  })
  const [loaded, setLoaded] = useState(false)

  // Only create webview when first made visible
  useEffect(() => {
    if (!visible || createdRef.current) return
    const container = containerRef.current
    if (!container) return

    createdRef.current = true
    const wv = document.createElement('webview') as unknown as WebviewElement
    wv.setAttribute('style', 'width:100%;height:100%;border:none;')
    wv.setAttribute('allowpopups', '')

    wv.addEventListener('dom-ready', () => {
      readyRef.current = true
      setLoaded(true)
    })

    wv.src = url
    container.appendChild(wv as unknown as Node)
    webviewRef.current = wv
  }, [visible, url])

  useEffect(() => {
    if (onHandle) {
      onHandle({
        navigate: (navUrl: string) => {
          try { setCurrentPath(new URL(navUrl).pathname) } catch { setCurrentPath(navUrl) }
          if (webviewRef.current && readyRef.current) {
            webviewRef.current.loadURL(navUrl)
          }
        },
        goBack: () => {
          if (webviewRef.current && readyRef.current) webviewRef.current.goBack()
        },
        goForward: () => {
          if (webviewRef.current && readyRef.current) webviewRef.current.goForward()
        },
        reload: () => {
          if (webviewRef.current && readyRef.current) webviewRef.current.reload()
        }
      })
    }
  }, [onHandle])

  return (
    <div className="relative h-full w-full">
      <div ref={containerRef} className="h-full w-full" style={{ display: loaded ? 'block' : 'none' }} />
      {/* Fallback when dashboard isn't loaded */}
      {!loaded && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-background">
          <div className="rounded-xl border border-border/50 bg-card px-8 py-6 text-center shadow-sm">
            <p className="text-[14px] font-medium text-foreground/80">{currentPath}</p>
            <p className="mt-1.5 text-[12px] text-muted-foreground">Dashboard will render here</p>
          </div>
        </div>
      )}
    </div>
  )
}
