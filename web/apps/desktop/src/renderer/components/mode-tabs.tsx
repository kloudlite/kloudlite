import { useRef } from 'react'
import { cn } from '@/lib/utils'
import { useModeStore, type AppMode } from '@/store/mode'

const modes: { id: AppMode; label: string }[] = [
  { id: 'environments', label: 'Environments' },
  { id: 'workspaces', label: 'Workspaces' },
  { id: 'browse', label: 'Browse' },
]

export function ModeTabs() {
  const { mode, setMode } = useModeStore()
  const accumulatorRef = useRef(0)
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastSwipeRef = useRef(0)

  function navigate(direction: 1 | -1) {
    const now = Date.now()
    if (now - lastSwipeRef.current < 400) return
    lastSwipeRef.current = now

    const currentMode = useModeStore.getState().mode
    const currentIdx = modes.findIndex((m) => m.id === currentMode)
    const nextIdx = currentIdx + direction
    if (nextIdx >= 0 && nextIdx < modes.length) {
      setMode(modes[nextIdx].id)
    }
  }

  function handleWheel(e: React.WheelEvent) {
    // Ignore vertical scrolls
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX) * 1.5) return
    if (Math.abs(e.deltaX) < 1) return

    accumulatorRef.current += e.deltaX

    if (idleTimerRef.current) clearTimeout(idleTimerRef.current)
    idleTimerRef.current = setTimeout(() => {
      if (Math.abs(accumulatorRef.current) > 50) {
        navigate(accumulatorRef.current > 0 ? 1 : -1)
      }
      accumulatorRef.current = 0
    }, 60)
  }

  const activeIdx = modes.findIndex((m) => m.id === mode)

  return (
    <div
      className="no-drag relative shrink-0 px-3 pt-2 pb-3"
      onWheel={handleWheel}
    >
      <div className="relative grid grid-cols-3 gap-0 rounded-[8px] bg-sidebar-foreground/[0.08] p-[3px]">
        {/* Sliding indicator */}
        <div
          className="absolute top-[3px] bottom-[3px] rounded-[6px] bg-sidebar-foreground/[0.15] shadow-[inset_0_1px_0_rgba(255,255,255,0.06)] transition-transform duration-200 ease-out"
          style={{
            width: `calc((100% - 6px) / 3)`,
            left: '3px',
            transform: `translateX(${activeIdx * 100}%)`
          }}
        />
        {modes.map(({ id, label }) => (
          <button
            key={id}
            className={cn(
              'relative z-10 py-[6px] text-center text-[11px] font-semibold tracking-wide transition-colors duration-150',
              mode === id
                ? 'text-sidebar-foreground'
                : 'text-sidebar-foreground/45 hover:text-sidebar-foreground/65'
            )}
            onClick={() => setMode(id)}
          >
            {label}
          </button>
        ))}
      </div>
    </div>
  )
}
