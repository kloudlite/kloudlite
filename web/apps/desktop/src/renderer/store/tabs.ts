import { create } from 'zustand'

export interface Tab {
  id: string
  url: string
  title: string
  favicon: string
  isLoading: boolean
  canGoBack: boolean
  canGoForward: boolean
}

interface TabStore {
  tabs: Tab[]
  activeTabId: string | null
  addTab: (url?: string) => void
  closeTab: (id: string) => void
  setActiveTab: (id: string) => void
  updateTab: (id: string, updates: Partial<Tab>) => void
  moveTab: (fromIndex: number, toIndex: number) => void
}

let nextId = 1

function createTab(url = ''): Tab {
  return {
    id: String(nextId++),
    url,
    title: url || 'New Tab',
    favicon: '',
    isLoading: false,
    canGoBack: false,
    canGoForward: false
  }
}

export const useTabStore = create<TabStore>((set, get) => ({
  tabs: [],
  activeTabId: null,

  addTab: (url?: string) => {
    const tab = createTab(url)
    set((state) => ({
      tabs: [...state.tabs, tab],
      activeTabId: tab.id
    }))
  },

  closeTab: (id: string) => {
    const { tabs, activeTabId } = get()
    const index = tabs.findIndex((t) => t.id === id)
    if (index === -1) return

    const newTabs = tabs.filter((t) => t.id !== id)

    let newActiveId = activeTabId
    if (activeTabId === id) {
      if (newTabs.length === 0) {
        newActiveId = null
      } else if (index < newTabs.length) {
        newActiveId = newTabs[index].id
      } else {
        newActiveId = newTabs[newTabs.length - 1].id
      }
    }

    set({ tabs: newTabs, activeTabId: newActiveId })
  },

  setActiveTab: (id: string) => {
    set({ activeTabId: id })
  },

  updateTab: (id: string, updates: Partial<Tab>) => {
    set((state) => ({
      tabs: state.tabs.map((t) => (t.id === id ? { ...t, ...updates } : t))
    }))
  },

  moveTab: (fromIndex: number, toIndex: number) => {
    set((state) => {
      const newTabs = [...state.tabs]
      const [moved] = newTabs.splice(fromIndex, 1)
      newTabs.splice(toIndex, 0, moved)
      return { tabs: newTabs }
    })
  }
}))
