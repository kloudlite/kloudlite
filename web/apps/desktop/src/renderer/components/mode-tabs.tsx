import { useRef } from 'react'
import { Layers, Box, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore, type AppMode } from '@/store/mode'

const modes: { id: AppMode; label: string; icon: typeof Layers }[] = [
  { id: 'environments', label: 'Envs', icon: Layers },
  { id: 'workspaces', label: 'Workspaces', icon: Box },
  { id: 'browse', label: 'Browse', icon: Globe },
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

  return (
    <div className="no-drag shrink-0 border-y border-sidebar-foreground/[0.06] px-2 py-2 mb-2.5" onWheel={handleWheel}>
      <div className="grid grid-cols-3">
        {modes.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            className={cn(
              'flex flex-col items-center gap-1 rounded-lg py-2 outline-none transition-all duration-150',
              mode === id
                ? 'text-sidebar-foreground'
                : 'text-sidebar-foreground/30 hover:text-sidebar-foreground/50'
            )}
            onClick={() => setMode(id)}
          >
            <Icon className="h-[18px] w-[18px]" strokeWidth={mode === id ? 2 : 1.5} />
            <span className={cn('text-[10px] tracking-wide', mode === id ? 'font-semibold' : 'font-medium')}>
              {label}
            </span>
          </button>
        ))}
      </div>
    </div>
  )
}
