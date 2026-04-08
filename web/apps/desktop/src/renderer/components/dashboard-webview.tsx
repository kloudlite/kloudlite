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
          <svg viewBox="0 0 130 131" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-14 w-14 text-primary/15">
            <path d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z" fill="currentColor" opacity="0.5"/>
            <path d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z" fill="currentColor" opacity="0.8"/>
          </svg>
          <p className="mt-5 text-[14px] font-medium text-foreground/40">{currentPath}</p>
          <p className="mt-1 text-[12px] text-muted-foreground/50">Loading dashboard...</p>
        </div>
      )}
    </div>
  )
}
