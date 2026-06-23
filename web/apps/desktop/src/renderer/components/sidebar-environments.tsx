import { useRef, useEffect, useState } from 'react'
import { ChevronLeft, Server, FileText, Settings, Plus, History, MoreHorizontal, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'
import { useEnvironmentStore } from '@/store/environments'
import { SidebarListItem } from './sidebar-list-item'
import { Dialog, Button, IconButton } from './ui'

const ENV_TABS = [
  { id: 'services', label: 'Services', icon: Server },
  { id: 'configs', label: 'Configs & Secrets', icon: FileText },
  { id: 'snapshots', label: 'Snapshots', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const USER_NAMESPACE = 'wm-karthik-dev'

export function SidebarEnvironments() {
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const { selectedEnvId, envActiveTab, selectEnvironment, setEnvActiveTab, clearSelectedEnv, setShowNewEnvDialog } = useModeStore()
  const { environments: envs, loading, refreshing, error, fetchEnvironments, deletingEnvs } = useEnvironmentStore()
  const selectedEnv = envs.find((e) => e.id === selectedEnvId)
  const isEnvDeleting = (env: { name: string; status: string }) => env.status === 'deleting' || deletingEnvs.has(env.name)

  function DeletingBadge() {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-amber-400/20 bg-amber-400/[0.08] px-2 py-0.5 text-[10px] font-medium uppercase tracking-[0.14em] text-amber-300/90">
        <Loader2 className="h-2.5 w-2.5 animate-spin" />
        Deleting
      </span>
    )
  }

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
            {isEnvDeleting(selectedEnv) ? (
              <span className="flex h-2.5 w-2.5 shrink-0 items-center justify-center rounded-full border border-amber-300/40 bg-amber-400/10">
                <Loader2 className="h-2 w-2 animate-spin text-amber-300" />
              </span>
            ) : (
              <div className={cn(
                'h-2.5 w-2.5 shrink-0 rounded-full',
                selectedEnv.status === 'active' ? 'bg-emerald-400' : selectedEnv.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
              )} />
            )}
            <h2 className="min-w-0 flex-1 truncate text-[14px] font-semibold text-sidebar-foreground/90">
              {selectedEnv.name}
            </h2>
            {isEnvDeleting(selectedEnv) && <DeletingBadge />}
            <IconButton
              size="sm"
              disabled={isEnvDeleting(selectedEnv)}
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
            </IconButton>
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
        <Dialog
          title={`Delete "${confirmDelete}"`}
          description="This action cannot be undone."
          onClose={() => setConfirmDelete(null)}
          footer={(close) => (
            <div className="flex w-full gap-2">
              <Button variant="secondary" className="flex-1" onClick={close}>Cancel</Button>
              <Button variant="danger" className="flex-1" onClick={() => { handleDeleteEnv(confirmDelete); close() }}>Delete</Button>
            </div>
          )}
        >
          <div className="flex items-center gap-3 rounded-lg border border-red-500/20 bg-red-500/[0.04] px-4 py-3">
            <div className="h-8 w-8 shrink-0 rounded-full bg-red-500/10 flex items-center justify-center">
              <svg className="h-4 w-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 9v4m0 4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
            </div>
            <p className="text-[12px] text-foreground leading-relaxed">All services, configs, and data will be permanently removed.</p>
          </div>
        </Dialog>
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
            <Button variant="ghost" size="sm" onClick={() => fetchEnvironments(USER_NAMESPACE, true)}>
              <RefreshCw className={cn('h-3.5 w-3.5', refreshing && 'animate-spin')} />
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setShowNewEnvDialog(true)}>
              <Plus className="h-3.5 w-3.5" />
              New
            </Button>
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
                isEnvDeleting(env) ? (
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-amber-400/[0.08] text-amber-300/90">
                    <Loader2 className="h-3 w-3 animate-spin" />
                  </span>
                ) : (
                  <div className={cn(
                    'h-2 w-2 rounded-full',
                    env.status === 'active' ? 'bg-emerald-400' : env.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/25'
                  )} />
                )
              }
              label={env.name}
              right={isEnvDeleting(env) ? <DeletingBadge /> : undefined}
              disabled={isEnvDeleting(env)}
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
        <Dialog
          title={`Delete "${confirmDelete}"`}
          description="This action cannot be undone."
          onClose={() => setConfirmDelete(null)}
          footer={(close) => (
            <div className="flex w-full gap-2">
              <Button variant="secondary" className="flex-1" onClick={close}>Cancel</Button>
              <Button variant="danger" className="flex-1" onClick={() => { handleDeleteEnv(confirmDelete); close() }}>Delete</Button>
            </div>
          )}
        >
          <div className="flex items-center gap-3 rounded-lg border border-red-500/20 bg-red-500/[0.04] px-4 py-3">
            <div className="h-8 w-8 shrink-0 rounded-full bg-red-500/10 flex items-center justify-center">
              <svg className="h-4 w-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 9v4m0 4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
            </div>
            <p className="text-[12px] text-foreground leading-relaxed">All services, configs, and data will be permanently removed.</p>
          </div>
        </Dialog>
      )}
    </div>
  )
}
