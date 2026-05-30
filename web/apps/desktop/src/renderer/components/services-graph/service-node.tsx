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
      'group min-w-[300px] overflow-hidden rounded-xl border bg-card shadow-sm transition-all duration-200',
      selected
        ? 'border-primary/50 ring-1 ring-primary/20 shadow-lg'
        : hasIntercepts
          ? 'border-amber-500/30 shadow-md shadow-amber-500/5 hover:border-amber-500/50 hover:shadow-lg'
          : 'border-border/50 hover:shadow-md'
    )}>
      {/* Header */}
      <div className="flex items-center gap-3 px-4 py-3">
        <div className={cn(
          'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg',
          hasIntercepts
            ? 'bg-amber-500/10 text-amber-500 ring-1 ring-amber-500/20'
            : 'bg-emerald-500/10 text-emerald-500 ring-1 ring-emerald-500/20'
        )}>
          <Server className="h-5 w-5" strokeWidth={1.5} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <p className="truncate text-[14px] font-semibold text-foreground">{d.name}</p>
            <span className={cn(
              'shrink-0 rounded-full px-1.5 py-px text-[9px] font-medium uppercase tracking-wider',
              d.type === 'LoadBalancer' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' :
              d.type === 'NodePort' ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400' :
              'bg-muted text-muted-foreground/70'
            )}>{d.type === 'ClusterIP' ? 'Internal' : d.type}</span>
          </div>
          <p className="mt-0.5 truncate font-mono text-[10px] text-muted-foreground/60">{d.dns}</p>
        </div>
        <button
          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground/40 opacity-0 transition-all hover:bg-accent hover:text-foreground group-hover:opacity-100"
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
      {d.ports.length > 0 && (
        <div className="border-t border-border/50">
          <div className="bg-muted/30 px-4 py-1.5">
            <span className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground/70">Ports</span>
          </div>
          <div className="flex flex-col">
            {d.ports.map((p, i) => {
              const intercepted = !!p.interceptedBy
              const wsName = p.interceptedBy ? d.workspaceMap[p.interceptedBy] : null
              const disabled = hasIntercepts && !intercepted
              return (
                <div
                  key={p.port}
                  className={cn(
                    'relative flex items-center gap-2.5 border-t border-border/30 px-4 py-2 transition-colors',
                    intercepted && 'border-amber-500/20 bg-amber-500/[0.06]',
                    disabled && 'bg-muted/20',
                    i === 0 && 'border-t-0'
                  )}
                >
                  <div className={cn(
                    'flex h-2 w-2 shrink-0 rounded-full',
                    intercepted ? 'bg-amber-500 shadow-[0_0_6px] shadow-amber-500/30' :
                    disabled ? 'bg-muted-foreground/20' :
                    'bg-emerald-400 shadow-[0_0_6px] shadow-emerald-400/30'
                  )} />
                  <div className={cn(
                    'flex items-center gap-1.5 rounded-md border border-border/50 bg-background px-2 py-0.5',
                    disabled && 'opacity-30'
                  )}>
                    <span className="font-mono text-[12px] font-semibold text-foreground/90">:{p.port}</span>
                    <ArrowRight className="h-3 w-3 text-muted-foreground/30" strokeWidth={2} />
                    <span className="font-mono text-[11px] text-muted-foreground">:{p.targetPort}</span>
                  </div>
                  <span className="ml-auto text-[10px] text-muted-foreground/60">{p.protocol}</span>
                  {intercepted && (
                    <div className="flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5">
                      <Zap className="h-3 w-3 text-amber-500" strokeWidth={2} />
                      <span className="text-[10px] font-medium text-amber-600 dark:text-amber-400">{wsName}</span>
                    </div>
                  )}
                  <Handle
                    type="source"
                    id={`port-${p.port}`}
                    position={Position.Right}
                    className={cn(
                      '!h-3 !w-3 !border-2 !border-card !bg-muted-foreground/30',
                      intercepted && '!bg-amber-500'
                    )}
                    style={{ right: -6 }}
                    isConnectable={false}
                  />
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Volumes */}
      {d.volumes.length > 0 && (
        <>
          <div className="border-t border-border/50 bg-accent/20 px-4 py-1.5">
            <span className="text-[9px] font-semibold uppercase tracking-wider text-muted-foreground/70">Volumes</span>
          </div>
          <div className="flex flex-col">
            {d.volumes.map((v) => {
              const Icon = volumeIcons[v.type]
              return (
                <div key={v.name} className="flex items-center gap-2.5 border-t border-border/30 px-4 py-2">
                  <Icon className={cn('h-3.5 w-3.5 shrink-0', volumeColors[v.type])} strokeWidth={1.5} />
                  <span className="font-mono text-[11px] font-medium text-foreground/80">{v.name}</span>
                  <ArrowRight className="h-3 w-3 text-muted-foreground/30" strokeWidth={2} />
                  <span className="truncate font-mono text-[10px] text-muted-foreground/60">{v.mountPath}</span>
                </div>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}
