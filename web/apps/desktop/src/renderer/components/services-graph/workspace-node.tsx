import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Box, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface WorkspaceNodeData {
  name: string
  owner: string
  status: 'running' | 'stopped' | 'failed'
  interceptCount: number
  [key: string]: unknown
}

const statusColors = {
  running: 'bg-emerald-400',
  stopped: 'bg-muted-foreground/30',
  failed: 'bg-red-400',
}

const statusLabels = {
  running: 'Running',
  stopped: 'Stopped',
  failed: 'Failed',
}

export function WorkspaceNode({ data, selected }: NodeProps) {
  const d = data as WorkspaceNodeData
  return (
    <div className={cn(
      'min-w-[260px] overflow-hidden rounded-xl border bg-card/95 backdrop-blur-sm transition-all duration-200',
      selected
        ? 'border-primary/50 shadow-lg shadow-primary/10'
        : 'border-blue-500/30 shadow-md shadow-blue-500/5 hover:border-blue-500/50 hover:shadow-blue-500/15'
    )}>
      <Handle
        type="target"
        position={Position.Left}
        className="!h-2.5 !w-2.5 !border-2 !border-card !bg-amber-500"
      />

      {/* Header */}
      <div className="flex items-center gap-2.5 px-3.5 py-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-500">
          <Box className="h-[18px] w-[18px]" strokeWidth={2} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-[14px] font-semibold text-foreground leading-tight">{d.name}</p>
          <div className="mt-1 flex items-center gap-1.5">
            <div className={cn('h-1.5 w-1.5 rounded-full', statusColors[d.status])} />
            <span className="text-[10px] font-medium text-muted-foreground">{statusLabels[d.status]}</span>
            <span className="text-[10px] text-muted-foreground/40">·</span>
            <span className="text-[10px] text-muted-foreground/70">{d.owner}</span>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="flex items-center gap-1.5 border-t border-blue-500/20 bg-blue-500/[0.06] px-3.5 py-2">
        {d.interceptCount > 0 ? (
          <>
            <Zap className="h-3 w-3 text-blue-500" strokeWidth={2.5} />
            <span className="text-[10px] font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400">
              Intercepting
            </span>
            <span className="text-[10px] text-blue-600/70 dark:text-blue-400/70">
              {d.interceptCount} service{d.interceptCount > 1 ? 's' : ''}
            </span>
          </>
        ) : (
          <span className="text-[10px] uppercase tracking-wider text-blue-600/50 dark:text-blue-400/50">
            No intercepts
          </span>
        )}
      </div>
    </div>
  )
}
