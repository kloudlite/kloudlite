import { ChevronLeft, Server, FileText, Settings, Plus, History, MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'
import { SidebarListItem } from './sidebar-list-item'

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

async function showEnvMenu() {
  const action = await window.electronAPI.showPopupMenu([
    { label: 'Fork Environment', id: 'fork' },
    { label: '', id: '', type: 'separator' },
    { label: 'Delete Environment', id: 'delete', danger: true },
  ])
  if (action === 'fork') {
    // TODO: implement fork
  } else if (action === 'delete') {
    // TODO: implement delete
  }
}

async function showEnvListMenu(envName: string) {
  const action = await window.electronAPI.showPopupMenu([
    { label: 'Fork Environment', id: 'fork' },
    { label: '', id: '', type: 'separator' },
    { label: `Delete "${envName}"`, id: 'delete', danger: true },
  ])
  if (action === 'fork') {
    // TODO
  } else if (action === 'delete') {
    // TODO
  }
}

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
            <h2 className="min-w-0 flex-1 truncate text-[14px] font-semibold text-sidebar-foreground/90">{selectedEnv.name}</h2>
            <button
              className="no-drag rounded-md p-1 text-sidebar-foreground/40 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
              onClick={showEnvMenu}
            >
              <MoreHorizontal className="h-4 w-4" />
            </button>
          </div>
          <p className="mt-0.5 pl-[18px] text-[11px] text-sidebar-foreground/40">{selectedEnv.owner} · {selectedEnv.visibility}</p>
        </div>

        <div className="flex flex-col gap-0.5 px-3">
          {ENV_TABS.map(({ id, label, icon: Icon }) => (
            <SidebarListItem
              key={id}
              icon={<Icon className="h-4 w-4" />}
              label={label}
              active={envActiveTab === id}
              onClick={() => setEnvActiveTab(id)}
            />
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
          <span className="text-[11px] font-semibold uppercase tracking-wider text-sidebar-foreground/50">
            Environments
          </span>
          <button
            className="no-drag flex items-center gap-1 rounded-md px-2 py-1 text-[12px] font-medium text-sidebar-foreground/70 transition-colors hover:bg-sidebar-foreground/[0.1] hover:text-sidebar-foreground/90"
            onClick={() => setShowNewEnvDialog(true)}
          >
            <Plus className="h-3.5 w-3.5" />
            New
          </button>
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col gap-0.5 px-3">
          {DUMMY_ENVS.map((env) => (
            <SidebarListItem
              key={env.id}
              icon={
                <div className={cn(
                  'h-2 w-2 rounded-full',
                  env.status === 'active' ? 'bg-emerald-400' : env.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
                )} />
              }
              label={env.name}
              right={<span className="text-[12px] font-medium text-sidebar-foreground/50">{env.services}</span>}
              onClick={() => selectEnvironment(env.id, env.hash, env.name)}
              onContextMenu={(e) => {
                e.preventDefault()
                showEnvListMenu(env.name)
              }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
