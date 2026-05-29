import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Server, Zap, ArrowRight, HardDrive, FileText, Lock, FolderOpen, ScrollText } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ServicePort {
  port: number
  targetPort: number
  interceptedBy?: string
}

export interface ServiceVolume {
  name: string
  mountPath: string
  type: 'persistent' | 'config' | 'secret' | 'host'
}

export interface ServiceNodeData {
  name: string
  dns: string
  type: 'ClusterIP' | 'LoadBalancer' | 'NodePort'
  ports: ServicePort[]
  volumes: ServiceVolume[]
  interceptedCount: number
  workspaceMap: Record<string, string>
  [key: string]: unknown
}

const volumeIcons = {
  persistent: HardDrive,
  config: FileText,
  secret: Lock,
  host: FolderOpen,
}

const volumeColors = {
  persistent: 'text-blue-500',
  config: 'text-emerald-500',
  secret: 'text-amber-500',
  host: 'text-purple-500',
}

export function ServiceNode({ data, selected }: NodeProps) {
  const d = data as ServiceNodeData
  const hasIntercepts = d.interceptedCount > 0

  return (
    <div className={cn(
      'group min-w-[300px] overflow-hidden rounded-xl border bg-card/95 backdrop-blur-sm transition-all duration-200',
      selected
        ? 'border-primary/50 shadow-lg shadow-primary/10'
        : hasIntercepts
          ? 'border-amber-500/30 shadow-md shadow-amber-500/5 hover:border-amber-500/50'
          : 'border-border/60 shadow-sm hover:border-border hover:shadow-md'
    )}>
      {/* Header */}
      <div className="flex items-center gap-2.5 px-3.5 py-3">
        <div className={cn(
          'flex h-9 w-9 shrink-0 items-center justify-center rounded-lg',
          hasIntercepts ? 'bg-amber-500/10 text-amber-500' : 'bg-emerald-500/10 text-emerald-500'
        )}>
          <Server className="h-[18px] w-[18px]" strokeWidth={2} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-[14px] font-semibold text-foreground leading-tight">{d.name}</p>
          <p className="mt-0.5 truncate font-mono text-[10px] text-muted-foreground/70">{d.dns}</p>
        </div>
        <button
          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground/50 opacity-0 transition-all hover:bg-accent hover:text-foreground group-hover:opacity-100"
          onClick={(e) => {
            e.stopPropagation()
            window.dispatchEvent(new CustomEvent('open-service-logs', { detail: { name: d.name } }))
          }}
          title="View logs"
        >
          <ScrollText className="h-4 w-4" />
        </button>
      </div>

      {/* Ports */}
      <div className="flex flex-col">
        {d.ports.map((p, i) => {
          const intercepted = !!p.interceptedBy
          const wsName = p.interceptedBy ? d.workspaceMap[p.interceptedBy] : null
          // When any port is intercepted, the workload stops — disabled ports are unreachable
          const disabled = hasIntercepts && !intercepted
          return (
            <div
              key={p.port}
              className={cn(
                'relative flex items-center gap-2 border-t px-3.5 py-2 transition-colors',
                intercepted
                  ? 'border-amber-500/20 bg-amber-500/[0.06]'
                  : disabled
                    ? 'border-border/30 bg-muted/10'
                    : 'border-border/30',
                i === 0 && 'border-t-border/40'
              )}
            >
              <span className={cn(
                'flex h-1.5 w-1.5 rounded-full',
                intercepted
                  ? 'bg-amber-500'
                  : disabled
                    ? 'bg-muted-foreground/25'
                    : 'bg-emerald-400/70'
              )} />
              <span className={cn(
                'font-mono text-[12px] font-medium',
                disabled ? 'text-muted-foreground/40 line-through' : 'text-foreground/85'
              )}>:{p.port}</span>
              <ArrowRight className={cn(
                'h-2.5 w-2.5',
                disabled ? 'text-muted-foreground/20' : 'text-muted-foreground/40'
              )} strokeWidth={2.5} />
              <span className={cn(
                'font-mono text-[11px]',
                disabled ? 'text-muted-foreground/30 line-through' : 'text-muted-foreground'
              )}>:{p.targetPort}</span>
              {intercepted && (
                <>
                  <Zap className="ml-auto h-3 w-3 text-amber-500" strokeWidth={2.5} />
                  <span className="text-[10px] font-medium text-amber-600 dark:text-amber-400">{wsName}</span>
                </>
              )}
              {disabled && (
                <span className="ml-auto rounded-sm bg-muted/40 px-1.5 py-px text-[9px] font-medium uppercase tracking-wider text-muted-foreground/50">
                  Unreachable
                </span>
              )}

              <Handle
                type="source"
                id={`port-${p.port}`}
                position={Position.Right}
                className={cn(
                  '!h-2.5 !w-2.5 !border-2 !border-card',
                  intercepted ? '!bg-amber-500' : '!opacity-0'
                )}
                style={{ right: -5 }}
                isConnectable={false}
              />
            </div>
          )
        })}
      </div>

      {/* Volumes */}
      {d.volumes.length > 0 && (
        <>
          <div className="border-t border-border/40 bg-accent/20 px-3.5 py-1.5">
            <p className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground/60">Volumes</p>
          </div>
          <div className="flex flex-col">
            {d.volumes.map((v) => {
              const Icon = volumeIcons[v.type]
              return (
                <div key={v.name} className="flex items-center gap-2 border-t border-border/30 px-3.5 py-1.5">
                  <Icon className={cn('h-3 w-3 shrink-0', volumeColors[v.type])} strokeWidth={2} />
                  <span className="font-mono text-[11px] font-medium text-foreground/80">{v.name}</span>
                  <ArrowRight className="h-2.5 w-2.5 text-muted-foreground/40" strokeWidth={2.5} />
                  <span className="truncate font-mono text-[10px] text-muted-foreground/70">{v.mountPath}</span>
                </div>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}
