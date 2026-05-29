import { useState } from 'react'
import { Cpu, Power, ChevronUp, ChevronDown, HardDrive } from 'lucide-react'
import { cn } from '@/lib/utils'

// Dummy workmachine state
const WORKMACHINE = {
  name: "Karthik's WorkMachine",
  status: 'running' as 'running' | 'stopped' | 'starting' | 'error',
  type: '4 vCPU · 8 GB',
  cpu: 42,
  memory: 67,
  uptime: '3d 14h',
}

const statusConfig = {
  running: { color: 'bg-emerald-400', label: 'Running', textColor: 'text-emerald-400' },
  stopped: { color: 'bg-sidebar-foreground/25', label: 'Stopped', textColor: 'text-sidebar-foreground/60' },
  starting: { color: 'bg-amber-400', label: 'Starting', textColor: 'text-amber-400' },
  error: { color: 'bg-red-400', label: 'Error', textColor: 'text-red-400' },
}

export function WorkMachineBar() {
  const [expanded, setExpanded] = useState(false)
  const wm = WORKMACHINE
  const config = statusConfig[wm.status]

  return (
    <div className="shrink-0 border-t border-sidebar-foreground/[0.06]">
      {/* Collapsed bar */}
      <button
        className="no-drag flex w-full items-center gap-2.5 px-4 py-3 transition-colors hover:bg-sidebar-foreground/[0.04]"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="relative">
          <HardDrive className="h-4 w-4 text-sidebar-foreground/70" />
          <div className={cn('absolute -bottom-0.5 -right-0.5 h-[6px] w-[6px] rounded-full', config.color)} />
        </div>
        <div className="min-w-0 flex-1 text-left">
          <span className="block truncate text-[13px] font-medium text-sidebar-foreground/90">WorkMachine</span>
        </div>
        <span className={cn('text-[11px] font-medium', config.textColor)}>{config.label}</span>
        {expanded ? <ChevronDown className="h-3 w-3 text-sidebar-foreground/30" /> : <ChevronUp className="h-3 w-3 text-sidebar-foreground/30" />}
      </button>

      {/* Expanded details */}
      <div
        className="overflow-hidden transition-all duration-200 ease-out"
        style={{ maxHeight: expanded ? '200px' : '0px', opacity: expanded ? 1 : 0 }}
      >
        <div className="px-4 pb-4 pt-1">
          {/* Specs */}
          <div className="mb-3.5 flex items-center gap-2 text-[11px] text-sidebar-foreground/60">
            <Cpu className="h-3 w-3" />
            <span>{wm.type}</span>
            <span>·</span>
            <span>Up {wm.uptime}</span>
          </div>

          {/* CPU bar */}
          <div className="mb-2.5">
            <div className="flex items-center justify-between text-[11px]">
              <span className="text-sidebar-foreground/60">CPU</span>
              <span className="font-medium text-sidebar-foreground/80">{wm.cpu}%</span>
            </div>
            <div className="mt-0.5 h-1.5 overflow-hidden rounded-full bg-sidebar-foreground/[0.08]">
              <div
                className={cn('h-full rounded-full transition-all', wm.cpu > 80 ? 'bg-red-400' : wm.cpu > 60 ? 'bg-amber-400' : 'bg-emerald-400')}
                style={{ width: `${wm.cpu}%` }}
              />
            </div>
          </div>

          {/* Memory bar */}
          <div className="mb-4">
            <div className="flex items-center justify-between text-[11px]">
              <span className="text-sidebar-foreground/60">Memory</span>
              <span className="font-medium text-sidebar-foreground/80">{wm.memory}%</span>
            </div>
            <div className="mt-0.5 h-1.5 overflow-hidden rounded-full bg-sidebar-foreground/[0.08]">
              <div
                className={cn('h-full rounded-full transition-all', wm.memory > 80 ? 'bg-red-400' : wm.memory > 60 ? 'bg-amber-400' : 'bg-emerald-400')}
                style={{ width: `${wm.memory}%` }}
              />
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2">
            <button
              className={cn(
                'flex flex-1 items-center justify-center gap-1.5 rounded-lg py-2 text-[12px] font-medium transition-colors',
                wm.status === 'running'
                  ? 'bg-red-500/10 text-red-400 hover:bg-red-500/20'
                  : 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'
              )}
            >
              <Power className="h-3 w-3" />
              {wm.status === 'running' ? 'Stop' : 'Start'}
            </button>
            <button
              className="flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-sidebar-foreground/[0.06] py-1.5 text-[11px] font-medium text-sidebar-foreground/70 transition-colors hover:bg-sidebar-foreground/[0.1]"
              onClick={() => window.electronAPI.showPopupMenu([
                { label: 'Change Machine Type', id: 'type' },
                { label: 'SSH Keys', id: 'ssh' },
                { label: 'Auto-Stop Settings', id: 'autostop' },
                { label: '', id: '', type: 'separator' },
                { label: 'Restart', id: 'restart' },
              ])}
            >
              Settings
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
