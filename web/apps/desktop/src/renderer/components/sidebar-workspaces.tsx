import { ChevronLeft, Plus, Terminal, Package, Settings, MoreHorizontal, GitBranch, History } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'

const DUMMY_WORKSPACES = [
  { id: 'ws-1', name: 'api-dev', owner: 'karthik', status: 'running' as const, env: 'Staging', idle: false, git: 'github.com/kloudlite/api', branch: 'main' },
  { id: 'ws-2', name: 'frontend-dev', owner: 'karthik', status: 'running' as const, env: 'Development', idle: true, git: 'github.com/kloudlite/web', branch: 'feat/ui' },
  { id: 'ws-3', name: 'debug-session', owner: 'sohail', status: 'stopped' as const, env: 'Staging', idle: false, git: '', branch: '' },
  { id: 'ws-4', name: 'migration-test', owner: 'karthik', status: 'running' as const, env: 'QA Testing', idle: false, git: 'github.com/kloudlite/api', branch: 'fix/migration' },
  { id: 'ws-5', name: 'hotfix-auth', owner: 'karthik', status: 'failed' as const, env: 'Production', idle: false, git: 'github.com/kloudlite/api', branch: 'hotfix/auth' },
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

const WS_TABS = [
  { id: 'connect', label: 'Connect', icon: Terminal },
  { id: 'packages', label: 'Packages', icon: Package },
  { id: 'git', label: 'Git', icon: GitBranch },
  { id: 'snapshots', label: 'Snapshots', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

async function showWsMenu(wsName: string, status: string) {
  const items: { label: string; id: string; type?: string }[] = []
  if (status === 'running') {
    items.push({ label: 'Suspend Workspace', id: 'suspend' })
  } else if (status === 'stopped') {
    items.push({ label: 'Activate Workspace', id: 'activate' })
  }
  items.push({ label: 'Fork Workspace', id: 'fork' })
  items.push({ label: '', id: '', type: 'separator' })
  items.push({ label: `Delete "${wsName}"`, id: 'delete' })

  await window.electronAPI.showPopupMenu(items)
}

export function SidebarWorkspaces() {
  const { selectedWsId, wsActiveTab, selectWorkspace, setWsActiveTab, clearSelectedWs, setShowNewWsDialog } = useModeStore()
  const selectedWs = DUMMY_WORKSPACES.find((w) => w.id === selectedWsId)

  // Detail view
  if (selectedWs) {
    return (
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="shrink-0 px-3 pb-2">
          <button
            className="no-drag flex w-full items-center gap-1.5 rounded-lg px-2 py-1.5 text-[12px] text-sidebar-foreground/60 transition-colors hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/80"
            onClick={clearSelectedWs}
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            <span>Workspaces</span>
          </button>
        </div>

        <div className="shrink-0 px-5 pb-3">
          <div className="flex items-center gap-2">
            <div className={cn('h-2.5 w-2.5 shrink-0 rounded-full', statusColors[selectedWs.status])} />
            <h2 className="min-w-0 flex-1 truncate text-[14px] font-semibold text-sidebar-foreground/90">{selectedWs.name}</h2>
            <button
              className="no-drag rounded-md p-1 text-sidebar-foreground/40 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
              onClick={() => showWsMenu(selectedWs.name, selectedWs.status)}
            >
              <MoreHorizontal className="h-4 w-4" />
            </button>
          </div>
          <div className="mt-0.5 flex items-center gap-2 pl-[18px]">
            <span className={cn(
              'text-[10px] font-medium',
              selectedWs.status === 'running' ? 'text-emerald-400/80' :
              selectedWs.status === 'failed' ? 'text-red-400/80' :
              'text-sidebar-foreground/35'
            )}>
              {statusLabels[selectedWs.status]}
            </span>
            {selectedWs.idle && (
              <span className="rounded px-1 py-px text-[8px] font-medium bg-amber-500/15 text-amber-400/80">IDLE</span>
            )}
            <span className="text-[10px] text-sidebar-foreground/30">·</span>
            <span className="text-[11px] text-sidebar-foreground/40">{selectedWs.env}</span>
          </div>
        </div>

        <div className="flex flex-col gap-0.5 px-3">
          {WS_TABS.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              className={cn(
                'no-drag flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-[13px] transition-all duration-150',
                wsActiveTab === id
                  ? 'bg-sidebar-foreground/[0.1] text-sidebar-foreground/90'
                  : 'text-sidebar-foreground/55 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/75'
              )}
              onClick={() => setWsActiveTab(id)}
            >
              <Icon className="h-4 w-4" />
              <span>{label}</span>
            </button>
          ))}
        </div>
      </div>
    )
  }

  // List view
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0 px-3">
        <div className="flex items-center justify-between px-3 pb-1.5">
          <span className="text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/40">
            Workspaces
          </span>
          <button
            className="no-drag flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium text-sidebar-foreground/50 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
            onClick={() => setShowNewWsDialog(true)}
          >
            <Plus className="h-3 w-3" />
            New
          </button>
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col px-3">
          {DUMMY_WORKSPACES.map((ws) => (
            <button
              key={ws.id}
              className="no-drag flex h-9 w-full items-center gap-2.5 rounded-[10px] px-3 text-left text-[13px] transition-all duration-150 hover:bg-sidebar-foreground/[0.06]"
              onClick={() => selectWorkspace(ws.id, ws.name)}
              onContextMenu={(e) => {
                e.preventDefault()
                showWsMenu(ws.name, ws.status)
              }}
            >
              <div className={cn('h-[6px] w-[6px] shrink-0 rounded-full', statusColors[ws.status])} />
              <span className="min-w-0 flex-1 truncate text-sidebar-foreground/75">{ws.name}</span>
              {ws.idle && (
                <span className="shrink-0 rounded px-1 py-px text-[8px] font-medium bg-amber-500/15 text-amber-400/80">IDLE</span>
              )}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
