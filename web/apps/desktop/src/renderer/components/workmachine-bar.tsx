import { useState } from 'react'
import { Cpu, Power, ChevronUp, ChevronDown, HardDrive, MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'
import { DotGrid } from './ui/dot-grid'
import { IconButton } from './ui/icon-button'

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
        style={{ maxHeight: expanded ? '220px' : '0px', opacity: expanded ? 1 : 0 }}
      >
        <div className="space-y-3 px-4 pb-4 pt-1">
          {/* Specs row */}
          <div className="flex items-center gap-2 text-[11px] text-sidebar-foreground/60">
            <Cpu className="h-3 w-3" />
            <span>{wm.type}</span>
            <span>·</span>
            <span>Up {wm.uptime}</span>
          </div>

          {/* CPU & Memory side by side */}
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-lg border border-sidebar-foreground/[0.06] bg-sidebar-foreground/[0.02] px-3 py-2.5">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-[11px] font-medium text-sidebar-foreground/70">CPU</span>
                <span className="font-mono text-[13px] font-semibold text-sidebar-foreground/90">{wm.cpu}%</span>
              </div>
              <DotGrid value={wm.cpu} total={10} size={8} />
            </div>
            <div className="rounded-lg border border-sidebar-foreground/[0.06] bg-sidebar-foreground/[0.02] px-3 py-2.5">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-[11px] font-medium text-sidebar-foreground/70">Memory</span>
                <span className="font-mono text-[13px] font-semibold text-sidebar-foreground/90">{wm.memory}%</span>
              </div>
              <DotGrid value={wm.memory} total={10} size={8} />
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-1">
            <IconButton
              variant={wm.status === 'running' ? 'danger' : 'muted'}
              onClick={() => {}}
            >
              <Power className="h-3.5 w-3.5" />
            </IconButton>
            <IconButton
              variant="muted"
              onClick={() => window.electronAPI.showPopupMenu([
                { label: 'Change Machine Type', id: 'type' },
                { label: 'SSH Keys', id: 'ssh' },
                { label: 'Auto-Stop Settings', id: 'autostop' },
                { label: '', id: '', type: 'separator' },
                { label: 'Restart', id: 'restart' },
              ])}
            >
              <MoreHorizontal className="h-3.5 w-3.5" />
            </IconButton>
          </div>
        </div>
      </div>
    </div>
  )
}
