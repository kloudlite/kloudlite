import { useRef, useEffect } from 'react'
import { PanelLeft, ArrowLeft, ArrowRight, RotateCw } from 'lucide-react'
import ViewPager from 'react-view-pager-touch'
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

  const viewPagerRef = useRef<any>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  // Update ViewPager width when sidebar resizes
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver(() => {
      viewPagerRef.current?.updateWidth?.()
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [])
  const wheelIdleRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const wheelActiveRef = useRef(false)
  const fakeXRef = useRef(0)

  function handlePageSelected(position: number) {
    if (MODES[position]) {
      setMode(MODES[position])
    }
  }

  function handleWheel(e: React.WheelEvent) {
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX) * 1.5) return
    if (Math.abs(e.deltaX) < 1) return

    const vp = viewPagerRef.current
    if (!vp?.el) return

    const rect = vp.el.getBoundingClientRect()
    const centerY = rect.top + rect.height / 2

    // Simulate touchstart on first wheel
    if (!wheelActiveRef.current) {
      wheelActiveRef.current = true
      fakeXRef.current = rect.left + rect.width / 2

      vp.el.dispatchEvent(new MouseEvent('mousedown', {
        clientX: fakeXRef.current,
        clientY: centerY,
        bubbles: true
      }))
    }

    // Simulate drag via mousemove
    fakeXRef.current -= e.deltaX
    document.dispatchEvent(new MouseEvent('mousemove', {
      clientX: fakeXRef.current,
      clientY: centerY,
      bubbles: true
    }))

    // On idle, simulate mouseup to trigger snap
    if (wheelIdleRef.current) clearTimeout(wheelIdleRef.current)
    wheelIdleRef.current = setTimeout(() => {
      wheelActiveRef.current = false
      document.dispatchEvent(new MouseEvent('mouseup', {
        clientX: fakeXRef.current,
        clientY: centerY,
        bubbles: true
      }))
    }, 80)
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

      {/* Swipeable sidebar content */}
      <div ref={containerRef} className="min-h-0 flex-1 overflow-hidden" onWheel={handleWheel}>
        <ViewPager
          ref={viewPagerRef}
          items={MODES}
          currentPage={modeIndex}
          onPageSelected={handlePageSelected}
          renderItem={(item: AppMode, index: number) => {
            const shouldRender = Math.abs(index - modeIndex) <= 1
            return (
              <div className="h-full">
                {shouldRender && item === 'environments' && <SidebarEnvironments onNavigate={onDashboardNavigate} />}
                {shouldRender && item === 'workspaces' && <SidebarWorkspaces />}
                {shouldRender && item === 'browse' && <SidebarBrowse />}
              </div>
            )
          }}
        />
      </div>

      {/* WorkMachine status bar */}
      <WorkMachineBar />
    </div>
  )
}
