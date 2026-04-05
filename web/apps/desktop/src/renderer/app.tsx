import { useCallback, useEffect, useRef, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { WebviewArea, type WebviewAreaHandle } from '@/components/webview-area'
import { NewTabBar } from '@/components/command-bar'
import { useTabStore } from '@/store/tabs'
import { cn } from '@/lib/utils'

const MIN_SIDEBAR_WIDTH = 200
const MAX_SIDEBAR_WIDTH = 450
const DEFAULT_SIDEBAR_WIDTH = 360

export function App() {
  const handleRef = useRef<WebviewAreaHandle | null>(null)
  const { addTab, closeTab, activeTabId, tabs, setActiveTab } = useTabStore()
  const [sidebarWidth, setSidebarWidth] = useState(DEFAULT_SIDEBAR_WIDTH)
  const [sidebarVisible, setSidebarVisible] = useState(true)
  const [sidebarPeeking, setSidebarPeeking] = useState(false)
  const [isResizing, setIsResizing] = useState(false)
  const [newTabOpen, setNewTabOpen] = useState(false)
  const [showLoading, setShowLoading] = useState(false)
  const [loadingFading, setLoadingFading] = useState(false)
  const loadingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const resizeRef = useRef<{ startX: number; startWidth: number } | null>(null)
  const peekTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const setHandle = useCallback((handle: WebviewAreaHandle) => {
    handleRef.current = handle
  }, [])

  const HIDE_THRESHOLD = 120

  // Resize handlers
  useEffect(() => {
    function onMouseMove(e: MouseEvent) {
      if (!resizeRef.current) return
      e.preventDefault()
      const delta = e.clientX - resizeRef.current.startX
      const rawWidth = resizeRef.current.startWidth + delta

      if (rawWidth < HIDE_THRESHOLD) {
        setSidebarWidth(MIN_SIDEBAR_WIDTH)
        setSidebarVisible(false)
      } else {
        setSidebarVisible(true)
        setSidebarWidth(Math.min(MAX_SIDEBAR_WIDTH, Math.max(MIN_SIDEBAR_WIDTH, rawWidth)))
      }
    }

    function onMouseUp() {
      if (!resizeRef.current) return
      resizeRef.current = null
      setIsResizing(false)
      document.body.style.cursor = ''
    }

    if (isResizing) {
      document.addEventListener('mousemove', onMouseMove)
      document.addEventListener('mouseup', onMouseUp)
    }

    return () => {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
  }, [isResizing])

  function startResize(e: React.MouseEvent) {
    e.preventDefault()
    resizeRef.current = { startX: e.clientX, startWidth: sidebarWidth }
    setIsResizing(true)
    document.body.style.cursor = 'col-resize'
  }

  // Theme detection
  useEffect(() => {
    function applyTheme(theme: 'dark' | 'light') {
      document.documentElement.classList.toggle('dark', theme === 'dark')
    }
    window.electronAPI.getTheme().then(applyTheme)
    window.electronAPI.onThemeChanged(applyTheme)
  }, [])

  // Shortcuts from main process via IPC
  useEffect(() => {
    window.electronAPI.onShortcut((action) => {
      switch (action) {
        case 'new-tab':
          window.dispatchEvent(new CustomEvent('open-command-bar'))
          break
        case 'address-bar':
          window.dispatchEvent(new CustomEvent('open-address-bar'))
          break
        case 'close-tab': {
          const { activeTabId: id } = useTabStore.getState()
          if (id) closeTab(id)
          break
        }
        case 'reload':
          handleRef.current?.reload()
          break
        case 'go-back':
          handleRef.current?.goBack()
          break
        case 'go-forward':
          handleRef.current?.goForward()
          break
        case 'toggle-sidebar':
          setSidebarVisible((v) => !v)
          setSidebarPeeking(false)
          break
        case 'next-tab': {
          const state = useTabStore.getState()
          const idx = state.tabs.findIndex((t) => t.id === state.activeTabId)
          if (idx >= 0 && state.tabs.length > 1) {
            const nextIdx = (idx + 1) % state.tabs.length
            setActiveTab(state.tabs[nextIdx].id)
          }
          break
        }
        case 'prev-tab': {
          const state2 = useTabStore.getState()
          const idx2 = state2.tabs.findIndex((t) => t.id === state2.activeTabId)
          if (idx2 >= 0 && state2.tabs.length > 1) {
            const prevIdx = (idx2 - 1 + state2.tabs.length) % state2.tabs.length
            setActiveTab(state2.tabs[prevIdx].id)
          }
          break
        }
      }
    })
  }, [])

  useEffect(() => {
    if (tabs.length === 0) {
      addTab('https://kloudlite.io')
    }
  }, [])

  // Loading indicator with fade-out
  const isActiveTabLoading = !!tabs.find(t => t.id === activeTabId)?.isLoading
  useEffect(() => {
    if (isActiveTabLoading) {
      if (loadingTimerRef.current) clearTimeout(loadingTimerRef.current)
      setShowLoading(true)
      setLoadingFading(false)
    } else if (showLoading) {
      setLoadingFading(true)
      loadingTimerRef.current = setTimeout(() => {
        setShowLoading(false)
        setLoadingFading(false)
      }, 300)
    }
  }, [isActiveTabLoading])

  // Listen for open-command-bar at App level so it works when sidebar is hidden
  useEffect(() => {
    function handleOpenCommandBar() {
      setNewTabOpen(true)
    }
    window.addEventListener('open-command-bar', handleOpenCommandBar)
    return () => window.removeEventListener('open-command-bar', handleOpenCommandBar)
  }, [])

  // Peek: show sidebar as overlay when cursor hits left edge
  function handleEdgeEnter() {
    if (sidebarVisible) return
    if (peekTimeoutRef.current) clearTimeout(peekTimeoutRef.current)
    setSidebarPeeking(true)
  }

  function handlePeekLeave() {
    if (!sidebarPeeking) return
    peekTimeoutRef.current = setTimeout(() => {
      setSidebarPeeking(false)
    }, 300)
  }

  function handlePeekEnter() {
    if (peekTimeoutRef.current) clearTimeout(peekTimeoutRef.current)
  }

  const showSidebar = sidebarVisible || sidebarPeeking

  return (
    <>
    <div className="flex h-full bg-sidebar">
      {/* Sidebar — normal mode */}
      {sidebarVisible && (
        <div
          className="flex shrink-0 flex-col"
          style={{ width: sidebarWidth, overflow: 'hidden' }}
        >
          <Sidebar
            onNavigate={(url) => handleRef.current?.navigate(url)}
            onGoBack={() => handleRef.current?.goBack()}
            onGoForward={() => handleRef.current?.goForward()}
            onReload={() => handleRef.current?.reload()}
            onToggleSidebar={() => setSidebarVisible(false)}
          />
        </div>
      )}

      {/* Sidebar — peek overlay mode (floating, animated) */}
      {!sidebarVisible && (
        <div
          className="absolute bottom-1 left-1 top-1 z-30 flex flex-col overflow-hidden rounded-[10px] border border-sidebar-foreground/[0.08] bg-sidebar shadow-[0_0_24px_rgba(0,0,0,0.1),0_0_6px_rgba(0,0,0,0.05)]"
          style={{
            width: Math.max(sidebarWidth, 420),
            transform: sidebarPeeking ? 'translateX(0)' : 'translateX(-110%)',
            pointerEvents: sidebarPeeking ? 'auto' : 'none',
            transition: 'transform 250ms ease-in-out'
          }}
          onMouseEnter={handlePeekEnter}
          onMouseLeave={handlePeekLeave}
        >
          <Sidebar
            onNavigate={(url) => handleRef.current?.navigate(url)}
            onGoBack={() => handleRef.current?.goBack()}
            onGoForward={() => handleRef.current?.goForward()}
            onReload={() => handleRef.current?.reload()}
            onToggleSidebar={() => {
              setSidebarVisible(true)
              setSidebarPeeking(false)
            }}
          />
        </div>
      )}

      {/* Resize handle */}
      {sidebarVisible && (
        <div
          className={cn(
            'relative z-10 flex w-[6px] shrink-0 cursor-col-resize items-center justify-center',
            isResizing ? 'bg-sidebar-primary/30' : 'hover:bg-sidebar-primary/15'
          )}
          onMouseDown={startResize}
        />
      )}

      {/* Resize overlay */}
      {isResizing && (
        <div className="fixed inset-0 z-50 cursor-col-resize" />
      )}

      {/* Content area */}
      <div className={cn(
        'relative flex flex-1 flex-col overflow-hidden pb-2.5 pr-2.5 pt-2.5',
        !sidebarVisible && 'pl-2.5'
      )}>
        {/* Left edge hover zone — triggers sidebar peek */}
        {!sidebarVisible && !sidebarPeeking && (
          <div
            className="absolute inset-y-0 left-0 z-20 w-1"
            onMouseEnter={handleEdgeEnter}
          />
        )}

        <div className="relative flex flex-1 flex-col overflow-hidden rounded-[10px] bg-background shadow-[0_0_20px_rgba(0,0,0,0.08),0_0_4px_rgba(0,0,0,0.04)]">
          {showLoading && (
            <div
              className="absolute inset-x-0 top-0 z-10 flex justify-center py-1"
              style={{
                animation: loadingFading ? 'popover-out 300ms ease-in forwards' : 'popover-in 300ms ease-out'
              }}
            >
              <div className="h-1 w-20 rounded-full bg-foreground/25" style={{ animation: 'loading-pill 1.2s ease-in-out infinite' }} />
            </div>
          )}
          <WebviewArea onHandle={setHandle} />
        </div>
      </div>

    </div>

    {/* New Tab overlay — always available, even when sidebar is hidden */}
    {newTabOpen && (
      <NewTabBar
        onNavigate={(url) => handleRef.current?.navigate(url)}
        onClose={() => {
          setNewTabOpen(false)
          window.dispatchEvent(new CustomEvent('close-command-bar'))
        }}
      />
    )}
    </>
  )
}
