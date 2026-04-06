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
  const accRef = useRef(0)
  const idleRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lockRef = useRef(false)

  function handleWheel(e: React.WheelEvent) {
    if (Math.abs(e.deltaY) > Math.abs(e.deltaX) * 1.5) return
    if (Math.abs(e.deltaX) < 1 || lockRef.current) return
    accRef.current += e.deltaX
    if (idleRef.current) clearTimeout(idleRef.current)
    idleRef.current = setTimeout(() => {
      if (Math.abs(accRef.current) > 30) {
        const dir = accRef.current > 0 ? 1 : -1
        const idx = modes.findIndex((m) => m.id === useModeStore.getState().mode)
        const next = idx + dir
        if (next >= 0 && next < modes.length) {
          lockRef.current = true
          setMode(modes[next].id)
          setTimeout(() => { lockRef.current = false }, 500)
        }
      }
      accRef.current = 0
    }, 50)
  }

  const activeIdx = modes.findIndex((m) => m.id === mode)

  return (
    <div className="no-drag shrink-0 px-3 pt-1 pb-2" onWheel={handleWheel}>
      <div className="flex items-center rounded-lg bg-sidebar-foreground/[0.06] p-[2px]">
        {modes.map(({ id, label }) => (
          <button
            key={id}
            className={cn(
              'relative flex-1 rounded-md py-1.5 text-center text-[11px] font-medium transition-all duration-200',
              mode === id
                ? 'bg-sidebar-foreground/[0.12] text-sidebar-foreground shadow-sm'
                : 'text-sidebar-foreground/40 hover:text-sidebar-foreground/60'
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
