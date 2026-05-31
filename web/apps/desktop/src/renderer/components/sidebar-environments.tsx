import { useRef, useEffect, useState } from 'react'
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

function ConfirmDialog({ label, description, onConfirm, onCancel }: { label: string; description: string; onConfirm: () => void; onCancel: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={onCancel}>
      <div
        className="w-full max-w-sm overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl"
        style={{ animation: 'popover-in 150ms ease-out' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4">
          <h3 className="text-[15px] font-semibold text-foreground">{label}</h3>
          <p className="mt-2 text-[12px] text-muted-foreground leading-relaxed">{description}</p>
        </div>
        <div className="flex justify-end gap-2 border-t border-border/30 px-5 py-4">
          <button className="rounded-lg px-4 py-2 text-[12px] font-medium text-muted-foreground transition-colors hover:bg-accent" onClick={onCancel}>Cancel</button>
          <button className="rounded-lg bg-red-500 px-4 py-2 text-[12px] font-medium text-white transition-colors hover:bg-red-600" onClick={onConfirm}>Delete</button>
        </div>
      </div>
    </div>
  )
}

export function SidebarEnvironments() {
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const { selectedEnvId, envActiveTab, selectEnvironment, setEnvActiveTab, clearSelectedEnv, setShowNewEnvDialog } = useModeStore()
  const { environments: envs, loading, refreshing, error, fetchEnvironments } = useEnvironmentStore()
  const selectedEnv = envs.find((e) => e.id === selectedEnvId)

  async function handleDeleteEnv(envName: string) {
    await useEnvironmentStore.getState().deleteEnvironment(USER_NAMESPACE, envName)
    setConfirmDelete(null)
  }

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
              onClick={async () => {
                const action = await window.electronAPI.showPopupMenu([
                  { label: 'Refresh', id: 'refresh' },
                  { label: 'Delete Environment', id: 'delete', danger: true },
                ])
                if (action === 'refresh') fetchEnvironments(USER_NAMESPACE, true)
                if (action === 'delete') setConfirmDelete(selectedEnv.name)
              }}
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
      {confirmDelete && (
        <ConfirmDialog
          label={`Delete "${confirmDelete}"`}
          description="This cannot be undone. All services, configs, and data will be permanently removed."
          onConfirm={() => handleDeleteEnv(confirmDelete)}
          onCancel={() => setConfirmDelete(null)}
        />
      )}
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
              onClick={() => fetchEnvironments(USER_NAMESPACE, true)}
            >
              <RefreshCw className={cn('h-3.5 w-3.5', refreshing && 'animate-spin')} />
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
              onContextMenu={async (e) => {
                e.preventDefault()
                const action = await window.electronAPI.showPopupMenu([
                  { label: 'Refresh', id: 'refresh' },
                  { label: `Delete "${env.name}"`, id: 'delete', danger: true },
                ])
                if (action === 'refresh') fetchEnvironments(USER_NAMESPACE, true)
                if (action === 'delete') setConfirmDelete(env.name)
              }}
            />
          ))}
        </div>
      </div>
      {confirmDelete && (
        <ConfirmDialog
          label={`Delete "${confirmDelete}"`}
          description="This cannot be undone. All services, configs, and data will be permanently removed."
          onConfirm={() => handleDeleteEnv(confirmDelete)}
          onCancel={() => setConfirmDelete(null)}
        />
      )}
    </div>
  )
}
