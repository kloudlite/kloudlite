import { ChevronLeft, Server, FileText, Settings, Plus, History } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'

const DUMMY_ENVS = [
  { id: 'env-1', hash: 'a1b2c3', name: 'Staging', owner: 'karthik', status: 'active' as const, services: 4, visibility: 'shared' },
  { id: 'env-2', hash: 'd4e5f6', name: 'Development', owner: 'karthik', status: 'active' as const, services: 2, visibility: 'private' },
  { id: 'env-3', hash: 'g7h8i9', name: 'Production', owner: 'karthik', status: 'active' as const, services: 2, visibility: 'shared' },
  { id: 'env-4', hash: 'j1k2l3', name: 'QA Testing', owner: 'sohail', status: 'inactive' as const, services: 3, visibility: 'open' },
  { id: 'env-5', hash: 'm4n5o6', name: 'Demo', owner: 'karthik', status: 'error' as const, services: 1, visibility: 'private' },
]

const ENV_TABS = [
  { id: 'services', label: 'Services', icon: Server },
  { id: 'configs', label: 'Configs & Secrets', icon: FileText },
  { id: 'snapshots', label: 'Snapshots', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export function SidebarEnvironments() {
  const { selectedEnvId, envActiveTab, selectEnvironment, setEnvActiveTab, clearSelectedEnv, setShowNewEnvDialog } = useModeStore()
  const selectedEnv = DUMMY_ENVS.find((e) => e.id === selectedEnvId)

  // Detail view
  if (selectedEnv) {
    return (
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="shrink-0 px-3 pb-2">
          <button
            className="no-drag flex w-full items-center gap-1.5 rounded-lg px-2 py-1.5 text-[12px] text-sidebar-foreground/60 transition-colors hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/80"
            onClick={clearSelectedEnv}
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            <span>Environments</span>
          </button>
        </div>

        <div className="shrink-0 px-5 pb-3">
          <div className="flex items-center gap-2">
            <div className={cn(
              'h-2.5 w-2.5 shrink-0 rounded-full',
              selectedEnv.status === 'active' ? 'bg-emerald-400' : selectedEnv.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
            )} />
            <h2 className="text-[14px] font-semibold text-sidebar-foreground/90">{selectedEnv.name}</h2>
          </div>
          <p className="mt-0.5 pl-[18px] text-[11px] text-sidebar-foreground/40">{selectedEnv.owner} · {selectedEnv.visibility}</p>
        </div>

        <div className="flex flex-col gap-0.5 px-3">
          {ENV_TABS.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              className={cn(
                'no-drag flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-[12px] font-medium transition-all duration-150',
                envActiveTab === id
                  ? 'bg-sidebar-foreground/[0.1] text-sidebar-foreground/90'
                  : 'text-sidebar-foreground/55 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/75'
              )}
              onClick={() => setEnvActiveTab(id)}
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
      <div className="shrink-0 px-3 pb-2">
        <div className="flex items-center justify-between px-3 pb-1.5">
          <span className="text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/40">
            Environments
          </span>
          <button
            className="no-drag flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium text-sidebar-foreground/50 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
            onClick={() => setShowNewEnvDialog(true)}
          >
            <Plus className="h-3 w-3" />
            New
          </button>
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col gap-0.5 px-3">
          {DUMMY_ENVS.map((env) => (
            <button
              key={env.id}
              className="no-drag flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-[12px] transition-all duration-150 hover:bg-sidebar-foreground/[0.06]"
              onClick={() => selectEnvironment(env.id, env.hash, env.name)}
            >
              <div className={cn(
                'h-2 w-2 shrink-0 rounded-full',
                env.status === 'active' ? 'bg-emerald-400' : env.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
              )} />
              <div className="min-w-0 flex-1">
                <div className="truncate font-medium text-sidebar-foreground/80">{env.name}</div>
                <div className="text-[10px] text-sidebar-foreground/40">{env.owner} · {env.services} services</div>
              </div>
              <span className={cn(
                'shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium',
                env.visibility === 'shared' ? 'bg-sidebar-foreground/10 text-sidebar-foreground/50' :
                env.visibility === 'open' ? 'bg-blue-500/15 text-blue-400/80' :
                'bg-sidebar-foreground/8 text-sidebar-foreground/35'
              )}>
                {env.visibility}
              </span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
