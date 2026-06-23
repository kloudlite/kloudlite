import { useEffect, useRef, useState } from 'react'
import {
  ChevronLeft,
  Plus,
  Terminal,
  Package,
  Settings,
  MoreHorizontal,
  GitBranch,
  History,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useModeStore } from '@/store/mode'
import { useWorkspaceStore, type Workspace } from '@/store/workspaces'
import { SidebarListItem } from './sidebar-list-item'
import { Dialog, Button, IconButton } from './ui'

const USER_NAMESPACE = 'wm-karthik-dev'

const statusColors = {
  running: 'bg-emerald-400',
  stopped: 'bg-sidebar-foreground/25',
  failed: 'bg-red-400',
  deleting: 'bg-amber-400',
}

const statusLabels = {
  running: 'Running',
  stopped: 'Stopped',
  failed: 'Failed',
  deleting: 'Deleting',
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

  return window.electronAPI.showPopupMenu(items)
}

export function SidebarWorkspaces() {
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const {
    selectedWsId,
    wsActiveTab,
    selectWorkspace,
    setWsActiveTab,
    clearSelectedWs,
    setShowNewWsDialog,
  } = useModeStore()
  const { workspaces, loading, refreshing, error, fetchWorkspaces, deletingWorkspaces } =
    useWorkspaceStore()
  const selectedWs = workspaces.find((w) => w.id === selectedWsId)
  const isWorkspaceDeleting = (ws: Workspace) =>
    ws.status === 'deleting' || deletingWorkspaces.has(ws.name)

  function DeletingBadge() {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full border border-amber-400/20 bg-amber-400/[0.08] px-2 py-0.5 text-[10px] font-medium uppercase tracking-[0.14em] text-amber-300/90">
        <Loader2 className="h-2.5 w-2.5 animate-spin" />
        Deleting
      </span>
    )
  }

  async function handleDeleteWorkspace(name: string) {
    useWorkspaceStore.getState().deleteWorkspace(USER_NAMESPACE, name)
    setConfirmDelete(null)
  }

  const fetchedRef = useRef(false)
  useEffect(() => {
    if (!fetchedRef.current) {
      fetchedRef.current = true
      fetchWorkspaces(USER_NAMESPACE)
    }
  }, [fetchWorkspaces])

  const deleteDialog = confirmDelete && (
    <Dialog
      title={`Delete "${confirmDelete}"`}
      description="This action cannot be undone."
      onClose={() => setConfirmDelete(null)}
      footer={(close) => (
        <div className="flex w-full gap-2">
          <Button variant="secondary" className="flex-1" onClick={close}>
            Cancel
          </Button>
          <Button
            variant="danger"
            className="flex-1"
            onClick={() => {
              handleDeleteWorkspace(confirmDelete)
              close()
            }}
          >
            Delete
          </Button>
        </div>
      )}
    >
      <div className="flex items-center gap-3 rounded-lg border border-red-500/20 bg-red-500/[0.04] px-4 py-3">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-red-500/10">
          <svg
            className="h-4 w-4 text-red-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M12 9v4m0 4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
          </svg>
        </div>
        <p className="text-[12px] leading-relaxed text-foreground">
          Workspace sessions and attached resources will be permanently removed.
        </p>
      </div>
    </Dialog>
  )

  if (loading) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-sidebar-foreground/40" />
        <p className="mt-2 text-[12px] text-sidebar-foreground/40">Loading workspaces...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center px-6">
        <p className="text-center text-[12px] text-red-400">{error}</p>
        <Button
          variant="secondary"
          size="sm"
          className="mt-3"
          onClick={() => fetchWorkspaces(USER_NAMESPACE)}
        >
          Retry
        </Button>
      </div>
    )
  }

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
            {isWorkspaceDeleting(selectedWs) ? (
              <span className="flex h-2.5 w-2.5 shrink-0 items-center justify-center rounded-full border border-amber-300/40 bg-amber-400/10">
                <Loader2 className="h-2 w-2 animate-spin text-amber-300" />
              </span>
            ) : (
              <div
                className={cn('h-2.5 w-2.5 shrink-0 rounded-full', statusColors[selectedWs.status])}
              />
            )}
            <h2 className="min-w-0 flex-1 truncate text-[14px] font-semibold text-sidebar-foreground/90">
              {selectedWs.name}
            </h2>
            {isWorkspaceDeleting(selectedWs) && <DeletingBadge />}
            <IconButton
              size="sm"
              disabled={isWorkspaceDeleting(selectedWs)}
              onClick={async () => {
                const action = await showWsMenu(selectedWs.name, selectedWs.status)
                if (action === 'delete') setConfirmDelete(selectedWs.name)
              }}
            >
              <MoreHorizontal className="h-4 w-4" />
            </IconButton>
          </div>
          <div className="mt-0.5 flex items-center gap-2 pl-[18px]">
            <span
              className={cn(
                'text-[10px] font-medium',
                selectedWs.status === 'running'
                  ? 'text-emerald-400/80'
                  : selectedWs.status === 'failed'
                    ? 'text-red-400/80'
                    : 'text-sidebar-foreground/35',
              )}
            >
              {statusLabels[selectedWs.status]}
            </span>
            {selectedWs.idle && (
              <span className="rounded px-1 py-px text-[8px] font-medium bg-amber-500/15 text-amber-400/80">
                IDLE
              </span>
            )}
            <span className="text-[10px] text-sidebar-foreground/30">·</span>
            <span className="text-[11px] text-sidebar-foreground/40">
              {selectedWs.environmentName || selectedWs.namespace}
            </span>
          </div>
        </div>

        <div className="flex flex-col gap-0.5 px-3">
          {WS_TABS.map(({ id, label, icon: Icon }) => (
            <SidebarListItem
              key={id}
              icon={<Icon className="h-4 w-4" />}
              label={label}
              active={wsActiveTab === id}
              onClick={() => setWsActiveTab(id)}
            />
          ))}
        </div>
        {deleteDialog}
      </div>
    )
  }

  // List view
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0 px-3">
        <div className="flex items-center justify-between px-3 pb-1.5">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-sidebar-foreground/50">
            Workspaces
          </span>
          <div className="flex items-center gap-0.5">
            <Button variant="ghost" size="sm" onClick={() => fetchWorkspaces(USER_NAMESPACE, true)}>
              <RefreshCw className={cn('h-3.5 w-3.5', refreshing && 'animate-spin')} />
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setShowNewWsDialog(true)}>
              <Plus className="h-3.5 w-3.5" />
              New
            </Button>
          </div>
        </div>
      </div>
      <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto">
        <div className="flex flex-col gap-0.5 px-3">
          {workspaces.length === 0 && !loading && (
            <div className="px-3 py-8 text-center text-[12px] text-sidebar-foreground/40">
              No workspaces yet
            </div>
          )}
          {workspaces.map((ws) => (
            <SidebarListItem
              key={ws.id}
              icon={
                isWorkspaceDeleting(ws) ? (
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-amber-400/[0.08] text-amber-300/90">
                    <Loader2 className="h-3 w-3 animate-spin" />
                  </span>
                ) : (
                  <div className={cn('h-2 w-2 rounded-full', statusColors[ws.status])} />
                )
              }
              label={ws.name}
              right={
                isWorkspaceDeleting(ws) ? (
                  <DeletingBadge />
                ) : ws.idle ? (
                  <span className="rounded px-1.5 py-0.5 text-[9px] font-semibold bg-amber-500/15 text-amber-400/90">
                    IDLE
                  </span>
                ) : undefined
              }
              disabled={isWorkspaceDeleting(ws)}
              onClick={() => selectWorkspace(ws.id, ws.name)}
              onContextMenu={async (e) => {
                e.preventDefault()
                const action = await showWsMenu(ws.name, ws.status)
                if (action === 'delete') setConfirmDelete(ws.name)
              }}
            />
          ))}
        </div>
      </div>
      {deleteDialog}
    </div>
  )
}
