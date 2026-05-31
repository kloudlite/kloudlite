import { useRef, useEffect } from 'react'
import { ChevronLeft, Server, FileText, Settings, Plus, History, MoreHorizontal, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'
import { useEnvironmentStore } from '@/store/environments'
import { SidebarListItem } from './sidebar-list-item'

const ENV_TABS = [
  { id: 'services', label: 'Services', icon: Server },
  { id: 'configs', label: 'Configs & Secrets', icon: FileText },
  { id: 'snapshots', label: 'Snapshots', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const USER_NAMESPACE = 'wm-karthik-dev'

async function showEnvMenu(envId: string, envName: string) {
  const action = await window.electronAPI.showPopupMenu([
    { label: 'Refresh', id: 'refresh' },
    { label: 'Delete Environment', id: 'delete', danger: true },
  ])
  if (action === 'refresh') {
    await useEnvironmentStore.getState().fetchEnvironments(USER_NAMESPACE)
  }
  if (action === 'delete' && window.confirm(`Delete environment "${envName}"? This cannot be undone.`)) {
    await useEnvironmentStore.getState().deleteEnvironment(USER_NAMESPACE, envName)
  }
}

async function showEnvListMenu(envName: string) {
  const action = await window.electronAPI.showPopupMenu([
    { label: 'Refresh', id: 'refresh' },
    { label: `Delete "${envName}"`, id: 'delete', danger: true },
  ])
  if (action === 'refresh') {
    await useEnvironmentStore.getState().fetchEnvironments(USER_NAMESPACE)
  }
  if (action === 'delete' && window.confirm(`Delete environment "${envName}"? This cannot be undone.`)) {
    await useEnvironmentStore.getState().deleteEnvironment(USER_NAMESPACE, envName)
  }
}

export function SidebarEnvironments() {
  const { selectedEnvId, envActiveTab, selectEnvironment, setEnvActiveTab, clearSelectedEnv, setShowNewEnvDialog } = useModeStore()
  const { environments: envs, loading, error, fetchEnvironments } = useEnvironmentStore()
  const selectedEnv = envs.find((e) => e.id === selectedEnvId)

  // Fetch environments on mount
  const fetchedRef = useRef(false)
  useEffect(() => {
    if (!fetchedRef.current) {
      fetchedRef.current = true
      fetchEnvironments(USER_NAMESPACE)
    }
  }, [fetchEnvironments])

  // Loading state
  if (loading) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-sidebar-foreground/40" />
        <p className="mt-2 text-[12px] text-sidebar-foreground/40">Loading environments...</p>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center px-6">
        <p className="text-center text-[12px] text-red-400">{error}</p>
        <button className="no-drag mt-3 rounded-lg border border-border px-4 py-1.5 text-[11px] text-sidebar-foreground/60" onClick={() => fetchEnvironments(USER_NAMESPACE)}>
          Retry
        </button>
      </div>
    )
  }

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
              onClick={() => showEnvMenu(selectedEnv.id, selectedEnv.name)}
            >
              <MoreHorizontal className="h-4 w-4" />
            </button>
          </div>
          <p className="mt-0.5 pl-[18px] text-[11px] text-sidebar-foreground/40">{selectedEnv.ownedBy || 'unknown'} · {selectedEnv.namespace}</p>
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
          <div className="flex items-center gap-0.5">
            <button
              className="no-drag flex items-center gap-1 rounded-md px-2 py-1 text-[12px] font-medium text-sidebar-foreground/70 transition-colors hover:bg-sidebar-foreground/[0.1] hover:text-sidebar-foreground/90"
              onClick={() => fetchEnvironments(USER_NAMESPACE)}
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </button>
            <button
              className="no-drag flex items-center gap-1 rounded-md px-2 py-1 text-[12px] font-medium text-sidebar-foreground/70 transition-colors hover:bg-sidebar-foreground/[0.1] hover:text-sidebar-foreground/90"
              onClick={() => setShowNewEnvDialog(true)}
            >
              <Plus className="h-3.5 w-3.5" />
              New
            </button>
          </div>
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col gap-0.5 px-3">
          {envs.length === 0 && !loading && (
            <div className="px-3 py-8 text-center text-[12px] text-sidebar-foreground/40">
              No environments yet
            </div>
          )}
          {envs.map((env) => (
            <SidebarListItem
              key={env.id}
              icon={
                <div className={cn(
                  'h-2 w-2 rounded-full',
                  env.status === 'active' ? 'bg-emerald-400' : env.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
                )} />
              }
              label={env.name}
              onClick={() => selectEnvironment(env.id, env.name, env.name)}
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
