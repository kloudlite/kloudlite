import { create } from 'zustand'

export interface Service {
  id: string
  name: string
  port: number
  dnsHostname: string
  vpnUrl: string
}

export interface Environment {
  id: string
  name: string
  slug: string
  status: 'active' | 'inactive' | 'error' | 'deleting'
  services: Service[]
  namespace: string
  ownedBy?: string
}

export const DUMMY_ENVIRONMENTS: Environment[] = [
  {
    id: 'demo-production',
    name: 'demo-production',
    slug: 'demo-production',
    status: 'active',
    namespace: 'wm-demo',
    services: [
      { id: 'prod-api', name: 'api', port: 8080, dnsHostname: 'api.demo.local', vpnUrl: 'https://www.google.com' },
      { id: 'prod-web', name: 'web', port: 3000, dnsHostname: 'web.demo.local', vpnUrl: 'https://news.ycombinator.com' },
    ],
  },
  {
    id: 'demo-staging',
    name: 'demo-staging',
    slug: 'demo-staging',
    status: 'active',
    namespace: 'wm-demo',
    services: [
      { id: 'stage-dashboard', name: 'dashboard', port: 3001, dnsHostname: 'dashboard.demo.local', vpnUrl: 'https://example.com' },
    ],
  },
]

interface EnvironmentStore {
  environments: Environment[]
  selectedEnvironmentId: string | null
  loading: boolean
  refreshing: boolean
  error: string | null
  setSelectedEnvironment: (id: string) => void
  getSelectedEnvironment: () => Environment | undefined
  fetchEnvironments: (namespace: string, silent?: boolean) => Promise<void>
  createEnvironment: (namespace: string, name: string, spec: Record<string, unknown>) => Promise<boolean>
  deleteEnvironment: (namespace: string, name: string) => Promise<boolean>
  deletingEnvs: Set<string>
}

export function mapEnvironmentResource(item: any, namespace: string): Environment {
  const name = item.metadata?.name || item.name || 'unknown'
  const isDeleting = Boolean(item.metadata?.deletionTimestamp)

  return {
    id: name,
    name,
    slug: name,
    status: isDeleting ? 'deleting' : item.spec?.activated ? 'active' : 'inactive',
    services: [],
    namespace: item.metadata?.namespace || namespace,
    ownedBy: item.spec?.ownedBy,
  }
}

export const useEnvironmentStore = create<EnvironmentStore>((set, get) => ({
  environments: [],
  selectedEnvironmentId: null,
  loading: false,
  refreshing: false,
  error: null,
  deletingEnvs: new Set(),

  setSelectedEnvironment: (id) => set({ selectedEnvironmentId: id }),

  getSelectedEnvironment: () => {
    const { environments, selectedEnvironmentId } = get()
    return environments.find((e) => e.id === selectedEnvironmentId)
  },

  fetchEnvironments: async (namespace, silent) => {
    set({ error: null, ...(silent ? { refreshing: true } : { loading: true }) })
    try {
      const result = await window.electronAPI.listEnvironments(namespace)
      if (result.error) {
        set({ error: result.error, loading: false, refreshing: false })
        return
      }
      const envs: Environment[] = (result.items || []).map((item: any) => mapEnvironmentResource(item, namespace))
      set({
        environments: envs,
        loading: false,
        refreshing: false,
        selectedEnvironmentId: envs.length > 0 ? envs[0].id : null,
      })
    } catch (err) {
      set({ error: (err as Error).message, loading: false, refreshing: false })
    }
  },

  createEnvironment: async (namespace, name, spec) => {
    try {
      const fullSpec = {
        ownedBy: 'karthik',
        workmachineName: 'karthik-dev',
        activated: true,
        visibility: 'private',
        ...spec,
      }
      await window.electronAPI.createEnvironment(namespace, name, fullSpec)
      await get().fetchEnvironments(namespace, true)
      return true
    } catch (err) {
      set({ error: (err as Error).message })
      return false
    }
  },

  deleteEnvironment: async (namespace, name) => {
    const prev = get().deletingEnvs
    const next = new Set(prev)
    next.add(name)
    set({ deletingEnvs: next })

    try {
      const result = await window.electronAPI.deleteEnvironment(namespace, name)
      if (result.error) {
        set({ error: result.error })
        return false
      }
      await get().fetchEnvironments(namespace, true)
      return true
    } catch (err) {
      set({ error: (err as Error).message })
      return false
    } finally {
      const cleared = new Set(get().deletingEnvs)
      cleared.delete(name)
      set({ deletingEnvs: cleared })
    }
  },
}))
