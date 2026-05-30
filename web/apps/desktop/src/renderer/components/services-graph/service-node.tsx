import { Handle, Position, type NodeProps } from '@xyflow/react'
import { useState } from 'react'
import { Server, Zap, ArrowRight, HardDrive, FileText, Lock, FolderOpen, ScrollText, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ServicePort {
  port: number
  targetPort: number
  protocol?: string
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
  const [copiedDns, setCopiedDns] = useState(false)

  function copyDns(e: React.MouseEvent) {
    e.stopPropagation()
    navigator.clipboard.writeText(d.dns)
    setCopiedDns(true)
    setTimeout(() => setCopiedDns(false), 1200)
  }

  return (
    <div className={cn(
      'group w-[340px] overflow-hidden rounded-xl border bg-card shadow-sm transition-all duration-200',
      selected
        ? 'border-primary/50 ring-1 ring-primary/20 shadow-lg'
        : hasIntercepts
          ? 'border-amber-500/30 shadow-md shadow-amber-500/5 hover:border-amber-500/50 hover:shadow-lg'
          : 'border-border/50 hover:shadow-md'
    )}>
      {/* Header */}
      <div className="flex items-center gap-3 px-4 py-3">
        <div className={cn(
          'relative flex h-9 w-9 shrink-0 items-center justify-center rounded-lg',
          hasIntercepts
            ? 'bg-amber-500/10 text-amber-500 ring-1 ring-amber-500/20'
            : 'bg-emerald-500/10 text-emerald-500 ring-1 ring-emerald-500/20'
        )}>
          <Server className="h-4 w-4" strokeWidth={1.7} />
          <span className={cn(
            'absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full border-2 border-card',
            hasIntercepts ? 'bg-amber-500' : 'bg-emerald-400'
          )} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="truncate text-[14px] font-semibold text-foreground">{d.name}</p>
            {d.type !== 'ClusterIP' && (
              <span className={cn(
                'shrink-0 rounded-full px-1.5 py-px text-[9px] font-semibold uppercase tracking-[0.14em]',
                d.type === 'LoadBalancer' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' :
                'bg-purple-500/10 text-purple-600 dark:text-purple-400'
              )}>{d.type}</span>
            )}
          </div>
          <div className="mt-1 flex items-center gap-1.5">
            <p className="min-w-0 truncate font-mono text-[11px] font-medium text-muted-foreground/75" title={d.dns}>{d.dns}</p>
            <button
              className="flex h-4 w-4 shrink-0 items-center justify-center rounded text-muted-foreground/45 opacity-0 transition-all hover:bg-accent hover:text-foreground group-hover:opacity-100"
              onClick={copyDns}
              title="Copy service DNS"
            >
              {copiedDns ? <Check className="h-3 w-3 text-emerald-500" /> : <Copy className="h-3 w-3" />}
            </button>
          </div>
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
          <div className="flex items-center justify-between bg-muted/20 px-4 py-1.5">
            <span className="text-[9px] font-semibold uppercase tracking-[0.16em] text-muted-foreground/65">Ports</span>
          </div>
          <div className="flex flex-wrap items-center gap-1.5 px-4 py-2.5">
            {d.ports.map((p, i) => {
              const protocol = p.protocol ?? 'tcp'
              const portKey = `${protocol}-${p.port}`
              const intercepted = !!p.interceptedBy
              const wsName = p.interceptedBy ? d.workspaceMap[p.interceptedBy] : null
              const disabled = hasIntercepts && !intercepted
              return (
                <div
                  key={portKey}
                  className={cn(
                    'relative inline-flex h-6 items-center gap-1.5 rounded-full bg-muted/35 px-2.5 transition-colors ring-1 ring-border/40',
                    intercepted && 'bg-amber-500/[0.08] ring-amber-500/25',
                    disabled && 'bg-muted/20 ring-transparent'
                  )}
                >
                  <span className={cn('font-mono text-[11px] font-semibold leading-none text-foreground/85', disabled && 'opacity-35')}>:{p.port}</span>
                  {intercepted && (
                    <div className="flex max-w-[78px] shrink-0 items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5">
                      <Zap className="h-3 w-3 text-amber-500" strokeWidth={2} />
                      <span className="truncate text-[10px] font-medium text-amber-600 dark:text-amber-400">{wsName}</span>
                    </div>
                  )}
                  <Handle
                    type="source"
                    id={`port-${portKey}`}
                    position={Position.Right}
                    className={cn(
                      '!h-2.5 !w-2.5 !border-2 !border-card !bg-transparent !opacity-0',
                      intercepted && '!bg-amber-500 !opacity-100'
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
          <div className="border-t border-border/50 bg-blue-500/[0.04] px-4 py-1.5">
            <span className="text-[9px] font-semibold uppercase tracking-[0.18em] text-blue-600/75 dark:text-blue-400/80">Mounted volumes</span>
          </div>
          <div className="flex flex-col">
            {d.volumes.map((v) => {
              const Icon = volumeIcons[v.type]
              return (
                <div key={`${v.name}-${v.mountPath}`} className="flex items-center gap-2.5 border-t border-border/30 px-4 py-2">
                  <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-blue-500/10 ring-1 ring-blue-500/15">
                    <Icon className={cn('h-3.5 w-3.5', volumeColors[v.type])} strokeWidth={1.7} />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-1.5">
                      <span className="truncate font-mono text-[11px] font-semibold text-foreground/85">{v.name}</span>
                      <span className="rounded bg-blue-500/10 px-1.5 py-px text-[8px] font-semibold uppercase tracking-wider text-blue-600/70 dark:text-blue-400/75">{v.type}</span>
                    </div>
                    <span className="block truncate font-mono text-[10px] text-muted-foreground/55">{v.mountPath}</span>
                  </div>
                </div>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}
