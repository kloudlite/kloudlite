import { useRef, useEffect, useState, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { PanelLeft, ArrowLeft, ArrowRight, RotateCw, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore, type AppMode } from '@/store/mode'
import { useTabStore } from '@/store/tabs'
import { ModeTabs } from './mode-tabs'
import { SidebarEnvironments } from './sidebar-environments'
import { SidebarWorkspaces } from './sidebar-workspaces'
import { SidebarBrowse } from './sidebar-browse'
import { TrafficLights } from './traffic-lights'
import { WorkMachineBar } from './workmachine-bar'

const MODES: AppMode[] = ['environments', 'workspaces', 'browse']

interface SidebarProps {
  onNavigate: (url: string) => void
  onDashboardNavigate: (path: string) => void
  onGoBack: () => void
  onGoForward: () => void
  onReload: () => void
  onToggleSidebar: () => void
}

function useLongPress(onLongPress: () => void, ms = 500) {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const firingRef = useRef(false)
  const cbRef = useRef(onLongPress)
  cbRef.current = onLongPress

  const start = useCallback(() => {
    firingRef.current = false
    timerRef.current = setTimeout(() => {
      firingRef.current = true
      cbRef.current()
    }, ms)
  }, [ms])

  const cancel = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = null
  }, [])

  const wasTriggered = useCallback(() => {
    const result = firingRef.current
    firingRef.current = false
    return result
  }, [])

  return { start, cancel, wasTriggered }
}

interface NavPopoverProps {
  direction: 'back' | 'forward'
  items: string[]
  currentUrl: string
  onNavigate: (url: string) => void
  onClose: () => void
  buttonRef: React.RefObject<HTMLButtonElement | null>
}

function NavHistoryPopover({ direction, items, currentUrl, onNavigate, onClose, buttonRef }: NavPopoverProps) {
  const popoverRef = useRef<HTMLDivElement>(null)
  const [closing, setClosing] = useState(false)

  function close() {
    if (closing) return
    setClosing(true)
    setTimeout(onClose, 140)
  }

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        close()
      }
    }

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') close()
    }

    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [buttonRef, closing])

  if (items.length === 0) return null

  const displayItems = direction === 'back' ? items.slice().reverse() : items

  return createPortal(
    <>
    <div className="fixed inset-0 z-[999]" onMouseDown={close} />
    <div
      ref={popoverRef}
      className="fixed z-[1000] w-72 overflow-hidden rounded-xl border border-sidebar-border bg-sidebar shadow-2xl"
      style={{ 
        top: buttonRef.current ? buttonRef.current.getBoundingClientRect().bottom + 4 : 0,
        left: buttonRef.current ? buttonRef.current.getBoundingClientRect().left : 0,
        maxHeight: 'min(300px, 50vh)',
        transformOrigin: 'top left',
        animation: `${closing ? 'dropdown-out 140ms ease-in forwards' : 'dropdown-in 180ms cubic-bezier(0.22,1,0.36,1)'}`
      }}
    >
      <div className="overflow-y-auto py-1">
        {displayItems.map((url, i) => {
          const isCurrent = url === currentUrl
          return (
            <div
              key={`${url}-${i}`}
              className={cn(
                'flex cursor-pointer items-center gap-2.5 px-3 py-2 text-[12px] transition-colors',
                isCurrent
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
              )}
              onClick={() => { onNavigate(url); close() }}
            >
              <Globe className="h-3.5 w-3.5 shrink-0 text-sidebar-foreground/40" />
              <div className="min-w-0 flex-1">
                <span className="block truncate">{url.replace(/^https?:\/\//, '')}</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
    </>,
    document.body
  )
}

export function Sidebar({ onNavigate, onDashboardNavigate, onGoBack, onGoForward, onReload, onToggleSidebar }: SidebarProps) {
  const { mode, setMode } = useModeStore()
  const activeTab = useTabStore((s) => {
    const tab = s.tabs.find((t) => t.id === s.activeTabId)
    return tab
  })

  const [navPopover, setNavPopover] = useState<'back' | 'forward' | null>(null)
  const backRef = useRef<HTMLButtonElement>(null)
  const forwardRef = useRef<HTMLButtonElement>(null)
  const backLongPress = useLongPress(() => setNavPopover('back'))
  const forwardLongPress = useLongPress(() => setNavPopover('forward'))

  const showNavButtons = mode === 'browse' && activeTab
  const modeIndex = MODES.indexOf(mode)

  // Swipe detection for mode switching
  const accRef = useRef(0)
  const idleRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lockRef = useRef(false)

  function handleWheel(e: React.WheelEvent) {
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX) * 1.5) return
    if (Math.abs(e.deltaX) < 1 || lockRef.current) return
    accRef.current += e.deltaX
    if (idleRef.current) clearTimeout(idleRef.current)
    idleRef.current = setTimeout(() => {
      if (Math.abs(accRef.current) > 50) {
        const dir = accRef.current > 0 ? 1 : -1
        const idx = MODES.indexOf(useModeStore.getState().mode)
        const next = idx + dir
        if (next >= 0 && next < MODES.length) {
          lockRef.current = true
          setMode(MODES[next])
          setTimeout(() => { lockRef.current = false }, 500)
        }
      }
      accRef.current = 0
    }, 60)
  }

  const backItems = activeTab && navPopover === 'back'
    ? activeTab.navStack.slice(0, activeTab.navIndex)
    : []
  const forwardItems = activeTab && navPopover === 'forward'
    ? activeTab.navStack.slice(activeTab.navIndex + 1)
    : []

  function handleBackClick() {
    if (backLongPress.wasTriggered()) return
    onGoBack()
  }

  function handleForwardClick() {
    if (forwardLongPress.wasTriggered()) return
    onGoForward()
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* Top row: traffic lights + sidebar toggle + nav buttons */}
      <div className="drag-region flex h-[52px] shrink-0 items-center justify-between px-4">
        <div className="flex items-center gap-3.5">
          <TrafficLights />
          <button
            className="no-drag rounded-lg p-1.5 text-sidebar-foreground/40 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
            onClick={onToggleSidebar}
          >
            <PanelLeft className="h-[18px] w-[18px]" />
          </button>
        </div>

        <div className="no-drag flex items-center gap-0.5">
          <button
            ref={backRef}
            className={cn(
              'rounded-lg p-1.5 transition-colors',
              showNavButtons && activeTab?.canGoBack
                ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                : 'text-sidebar-foreground/20'
            )}
            onClick={handleBackClick}
            disabled={!showNavButtons || !activeTab?.canGoBack}
            onMouseDown={backLongPress.start}
            onMouseUp={backLongPress.cancel}
            onMouseLeave={backLongPress.cancel}
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <button
            ref={forwardRef}
            className={cn(
              'rounded-lg p-1.5 transition-colors',
              showNavButtons && activeTab?.canGoForward
                ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                : 'text-sidebar-foreground/20'
            )}
            onClick={handleForwardClick}
            disabled={!showNavButtons || !activeTab?.canGoForward}
            onMouseDown={forwardLongPress.start}
            onMouseUp={forwardLongPress.cancel}
            onMouseLeave={forwardLongPress.cancel}
          >
            <ArrowRight className="h-5 w-5" />
          </button>
          <button
            className="rounded-lg p-1.5 text-sidebar-foreground/60 transition-colors hover:text-sidebar-foreground"
            onClick={onReload}
          >
            <RotateCw className={cn('h-[18px] w-[18px]', showNavButtons && activeTab?.isLoading && 'animate-spin')} />
          </button>
        </div>
      </div>

      {/* Navigation history popover */}
      {navPopover && activeTab && (
        <NavHistoryPopover
          direction={navPopover}
          items={navPopover === 'back' ? backItems : forwardItems}
          currentUrl={activeTab.url}
          onNavigate={(url) => onNavigate(url)}
          onClose={() => setNavPopover(null)}
          buttonRef={navPopover === 'back' ? backRef : forwardRef}
        />
      )}

      {/* Mode tabs — top */}
      <ModeTabs />

      {/* Sidebar content — CSS slide transition */}
      <div className="min-h-0 flex-1 overflow-hidden" onWheel={handleWheel}>
        <div
          className="flex h-full transition-transform duration-[280ms] ease-out"
          style={{
            width: `${MODES.length * 100}%`,
            transform: `translateX(-${modeIndex * (100 / MODES.length)}%)`
          }}
        >
          <div className="h-full overflow-y-auto" style={{ width: `${100 / MODES.length}%` }}>
            <SidebarEnvironments />
          </div>
          <div className="h-full overflow-y-auto" style={{ width: `${100 / MODES.length}%` }}>
            <SidebarWorkspaces />
          </div>
          <div className="h-full overflow-y-auto" style={{ width: `${100 / MODES.length}%` }}>
            <SidebarBrowse />
          </div>
        </div>
      </div>

      {/* WorkMachine status bar */}
      <WorkMachineBar />
    </div>
  )
}
