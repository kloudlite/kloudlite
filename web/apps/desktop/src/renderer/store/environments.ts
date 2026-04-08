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
  status: 'active' | 'inactive' | 'error'
  services: Service[]
}

interface EnvironmentStore {
  environments: Environment[]
  selectedEnvironmentId: string | null
  setSelectedEnvironment: (id: string) => void
  getSelectedEnvironment: () => Environment | undefined
}

const DUMMY_ENVIRONMENTS: Environment[] = [
  {
    id: 'env-1',
    name: 'Staging',
    slug: 'staging',
    status: 'active',
    services: [
      { id: 'svc-1', name: 'frontend', port: 3000, dnsHostname: 'frontend-a1b2c3.staging.local', vpnUrl: 'http://frontend-a1b2c3.staging.local:3000' },
      { id: 'svc-2', name: 'api-server', port: 8080, dnsHostname: 'api-server-a1b2c3.staging.local', vpnUrl: 'http://api-server-a1b2c3.staging.local:8080' },
      { id: 'svc-3', name: 'redis', port: 6379, dnsHostname: 'redis-a1b2c3.staging.local', vpnUrl: 'http://redis-a1b2c3.staging.local:6379' },
      { id: 'svc-4', name: 'postgres', port: 5432, dnsHostname: 'postgres-a1b2c3.staging.local', vpnUrl: 'http://postgres-a1b2c3.staging.local:5432' },
    ]
  },
  {
    id: 'env-2',
    name: 'Development',
    slug: 'dev',
    status: 'active',
    services: [
      { id: 'svc-5', name: 'web-app', port: 5173, dnsHostname: 'web-app-d4e5f6.dev.local', vpnUrl: 'http://web-app-d4e5f6.dev.local:5173' },
      { id: 'svc-6', name: 'auth-service', port: 9090, dnsHostname: 'auth-d4e5f6.dev.local', vpnUrl: 'http://auth-d4e5f6.dev.local:9090' },
    ]
  },
  {
    id: 'env-3',
    name: 'Production',
    slug: 'prod',
    status: 'active',
    services: [
      { id: 'svc-7', name: 'gateway', port: 443, dnsHostname: 'gateway-g7h8i9.prod.local', vpnUrl: 'https://gateway-g7h8i9.prod.local' },
      { id: 'svc-8', name: 'dashboard', port: 3000, dnsHostname: 'dashboard-g7h8i9.prod.local', vpnUrl: 'http://dashboard-g7h8i9.prod.local:3000' },
    ]
  }
]

export const useEnvironmentStore = create<EnvironmentStore>((set, get) => ({
  environments: DUMMY_ENVIRONMENTS,
  selectedEnvironmentId: DUMMY_ENVIRONMENTS[0].id,

  setSelectedEnvironment: (id) => set({ selectedEnvironmentId: id }),

  getSelectedEnvironment: () => {
    const { environments, selectedEnvironmentId } = get()
    return environments.find((e) => e.id === selectedEnvironmentId)
  }
}))
