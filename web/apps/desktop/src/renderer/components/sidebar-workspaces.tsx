import { cn } from '@/lib/utils'

const DUMMY_WORKSPACES = [
  { id: 'ws-1', name: 'api-dev', owner: 'karthik', status: 'running' as const, env: 'Staging', idle: false },
  { id: 'ws-2', name: 'frontend-dev', owner: 'karthik', status: 'running' as const, env: 'Development', idle: true },
  { id: 'ws-3', name: 'debug-session', owner: 'sohail', status: 'stopped' as const, env: 'Staging', idle: false },
  { id: 'ws-4', name: 'migration-test', owner: 'karthik', status: 'running' as const, env: 'QA Testing', idle: false },
  { id: 'ws-5', name: 'hotfix-auth', owner: 'karthik', status: 'failed' as const, env: 'Production', idle: false },
]

const statusColors = {
  running: 'bg-emerald-400',
  stopped: 'bg-sidebar-foreground/25',
  failed: 'bg-red-400',
}

const statusLabels = {
  running: 'Running',
  stopped: 'Stopped',
  failed: 'Failed',
}

export function SidebarWorkspaces() {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0 px-3 pb-2">
        <div className="text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/40 px-3 pb-1.5">
          Workspaces
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col gap-0.5 px-3">
          {DUMMY_WORKSPACES.map((ws) => (
            <div
              key={ws.id}
              className="flex cursor-default items-center gap-2.5 rounded-lg px-3 py-2 text-[12px] transition-all duration-150 hover:bg-sidebar-foreground/[0.06]"
            >
              <div className={cn('h-2 w-2 shrink-0 rounded-full', statusColors[ws.status])} />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-1.5">
                  <span className="truncate font-medium text-sidebar-foreground/80">{ws.name}</span>
                  {ws.idle && (
                    <span className="shrink-0 rounded px-1 py-px text-[8px] font-medium bg-amber-500/15 text-amber-400/80">IDLE</span>
                  )}
                </div>
                <div className="text-[10px] text-sidebar-foreground/40">{ws.owner} · {ws.env}</div>
              </div>
              <span className={cn(
                'shrink-0 text-[10px] font-medium',
                ws.status === 'running' ? 'text-emerald-400/70' :
                ws.status === 'failed' ? 'text-red-400/70' :
                'text-sidebar-foreground/30'
              )}>
                {statusLabels[ws.status]}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
