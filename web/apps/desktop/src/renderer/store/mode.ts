import { create } from 'zustand'

export type AppMode = 'environments' | 'workspaces' | 'browse'

interface ModeStore {
  mode: AppMode
  setMode: (mode: AppMode) => void
  // Environment detail view state
  selectedEnvId: string | null
  selectedEnvHash: string | null
  selectedEnvName: string | null
  envActiveTab: string
  selectEnvironment: (id: string, hash: string, name: string) => void
  setEnvActiveTab: (tab: string) => void
  clearSelectedEnv: () => void
  showNewEnvDialog: boolean
  setShowNewEnvDialog: (show: boolean) => void
}

export const useModeStore = create<ModeStore>((set) => ({
  mode: 'environments',
  setMode: (mode) => set({ mode }),
  selectedEnvId: null,
  selectedEnvHash: null,
  selectedEnvName: null,
  envActiveTab: 'services',
  selectEnvironment: (id, hash, name) => set({ selectedEnvId: id, selectedEnvHash: hash, selectedEnvName: name, envActiveTab: 'services' }),
  setEnvActiveTab: (tab) => set({ envActiveTab: tab }),
  clearSelectedEnv: () => set({ selectedEnvId: null, selectedEnvHash: null, selectedEnvName: null, envActiveTab: 'services' }),
  showNewEnvDialog: false,
  setShowNewEnvDialog: (show) => set({ showNewEnvDialog: show }),
}))
