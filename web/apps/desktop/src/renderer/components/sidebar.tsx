import { useRef, useEffect } from 'react'
import { PanelLeft, ArrowLeft, ArrowRight, RotateCw } from 'lucide-react'
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

export function Sidebar({ onNavigate, onDashboardNavigate, onGoBack, onGoForward, onReload, onToggleSidebar }: SidebarProps) {
  const { mode, setMode } = useModeStore()
  const activeTab = useTabStore((s) => {
    const tab = s.tabs.find((t) => t.id === s.activeTabId)
    return tab
  })

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
            className={cn(
              'rounded-lg p-1.5 transition-colors',
              showNavButtons && activeTab?.canGoBack
                ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                : 'text-sidebar-foreground/20'
            )}
            onClick={onGoBack}
            disabled={!showNavButtons || !activeTab?.canGoBack}
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <button
            className={cn(
              'rounded-lg p-1.5 transition-colors',
              showNavButtons && activeTab?.canGoForward
                ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                : 'text-sidebar-foreground/20'
            )}
            onClick={onGoForward}
            disabled={!showNavButtons || !activeTab?.canGoForward}
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

      {/* Mode tabs — top */}
      <ModeTabs />

      {/* Sidebar content — CSS slide transition */}
      <div className="min-h-0 flex-1 overflow-hidden" onWheel={handleWheel}>
        <div
          className="flex h-full transition-transform duration-250 ease-out"
          style={{
            width: `${MODES.length * 100}%`,
            transform: `translateX(-${modeIndex * (100 / MODES.length)}%)`
          }}
        >
          <div className="h-full overflow-y-auto" style={{ width: `${100 / MODES.length}%` }}>
            <SidebarEnvironments onNavigate={onDashboardNavigate} />
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
