import { create } from 'zustand'

export interface Workspace {
  id: string
  name: string
  slug: string
  status: 'running' | 'stopped' | 'failed' | 'deleting'
  namespace: string
  ownedBy?: string
  environmentName?: string
  idle?: boolean
  git?: string
  branch?: string
}

interface WorkspaceStore {
  workspaces: Workspace[]
  loading: boolean
  refreshing: boolean
  error: string | null
  fetchWorkspaces: (namespace: string, silent?: boolean) => Promise<void>
  createWorkspace: (
    namespace: string,
    name: string,
    spec: Record<string, unknown>,
  ) => Promise<boolean>
  deleteWorkspace: (namespace: string, name: string) => Promise<boolean>
  deletingWorkspaces: Set<string>
}

function mapStatus(item: any): Workspace['status'] {
  if (item.metadata?.deletionTimestamp) return 'deleting'
  if (item.status?.phase === 'Failed' || item.status?.state === 'failed') return 'failed'
  if (item.spec?.status === 'suspended' || item.spec?.status === 'archived') return 'stopped'
  return 'running'
}

export function mapWorkspaceResource(item: any, namespace: string): Workspace {
  const name = item.metadata?.name || item.name || 'unknown'

  return {
    id: name,
    name,
    slug: name,
    status: mapStatus(item),
    namespace: item.metadata?.namespace || namespace,
    ownedBy: item.spec?.ownedBy,
    environmentName:
      item.spec?.environmentConnection?.environmentRef?.name ||
      item.spec?.environmentName ||
      item.spec?.targetEnvironmentName,
    idle: Boolean(item.status?.idle),
    git: item.spec?.gitRepository?.url || item.spec?.repository || '',
    branch: item.spec?.gitRepository?.branch || item.spec?.branch || '',
  }
}

export const useWorkspaceStore = create<WorkspaceStore>((set, get) => ({
  workspaces: [],
  loading: false,
  refreshing: false,
  error: null,
  deletingWorkspaces: new Set(),

  fetchWorkspaces: async (namespace, silent) => {
    set({ error: null, ...(silent ? { refreshing: true } : { loading: true }) })
    try {
      const result = await window.electronAPI.listWorkspaces(namespace)
      if (result.error) {
        set({ error: result.error, loading: false, refreshing: false })
        return
      }
      const workspaces: Workspace[] = (result.items || []).map((item: any) =>
        mapWorkspaceResource(item, namespace),
      )
      set({ workspaces, loading: false, refreshing: false })
    } catch (err) {
      set({ error: (err as Error).message, loading: false, refreshing: false })
    }
  },

  createWorkspace: async (namespace, name, spec) => {
    try {
      const fullSpec = {
        displayName: name,
        ownedBy: 'karthik',
        workmachine: 'karthik-dev',
        status: 'active',
        ...spec,
      }
      const result = await window.electronAPI.createWorkspace(namespace, name, fullSpec)
      if (result.error) {
        set({ error: result.error })
        return false
      }
      await get().fetchWorkspaces(namespace, true)
      return true
    } catch (err) {
      set({ error: (err as Error).message })
      return false
    }
  },

  deleteWorkspace: async (namespace, name) => {
    const next = new Set(get().deletingWorkspaces)
    next.add(name)
    set({ deletingWorkspaces: next })

    try {
      const result = await window.electronAPI.deleteWorkspace(namespace, name)
      if (result.error) {
        set({ error: result.error })
        return false
      }
      await get().fetchWorkspaces(namespace, true)
      return true
    } catch (err) {
      set({ error: (err as Error).message })
      return false
    } finally {
      const cleared = new Set(get().deletingWorkspaces)
      cleared.delete(name)
      set({ deletingWorkspaces: cleared })
    }
  },
}))
