import { create } from 'zustand'

export interface Tab {
  id: string
  url: string
  title: string
  favicon: string
  isLoading: boolean
  canGoBack: boolean
  canGoForward: boolean
  siteName?: string
  keywords?: string
  navStack: string[]
  navIndex: number
}

interface TabStore {
  tabs: Tab[]
  activeTabId: string | null
  closedTabs: Tab[]
  addTab: (url?: string) => void
  closeTab: (id: string) => void
  reopenLastTab: () => void
  setActiveTab: (id: string) => void
  updateTab: (id: string, updates: Partial<Tab>) => void
  moveTab: (fromIndex: number, toIndex: number) => void
}

let nextId = 1

function captureTabRects() {
  if (typeof document === 'undefined') return
  document.querySelectorAll('[data-tab-item]').forEach((el) => {
    ;(el as HTMLElement & { __prevRect?: DOMRect }).__prevRect = el.getBoundingClientRect()
  })
}

function createTab(url = 'https://google.com'): Tab {
  return {
    id: String(nextId++),
    url,
    title: url || 'New Tab',
    favicon: '',
    isLoading: false,
    canGoBack: false,
    canGoForward: false,
    navStack: url ? [url] : [],
    navIndex: url ? 0 : -1
  }
}

export const useTabStore = create<TabStore>((set, get) => ({
  tabs: [],
  activeTabId: null,
  closedTabs: [],

  addTab: (url?: string) => {
    captureTabRects()
    const tab = createTab(url)
    set((state) => ({
      tabs: [tab, ...state.tabs],
      activeTabId: tab.id
    }))
  },

  closeTab: (id: string) => {
    const { tabs, activeTabId } = get()
    const index = tabs.findIndex((t) => t.id === id)
    if (index === -1) return

    captureTabRects()
    const closed = tabs.find((t) => t.id === id)
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

    // ponytail: keep last 10 closed tabs for reopen
    const closedTabs = closed
      ? [...get().closedTabs.slice(-9), closed]
      : get().closedTabs

    set({ tabs: newTabs, activeTabId: newActiveId, closedTabs })
  },

  reopenLastTab: () => {
    const { closedTabs } = get()
    if (closedTabs.length === 0) return
    const tab = closedTabs[closedTabs.length - 1]
    set((state) => ({
      closedTabs: state.closedTabs.slice(0, -1),
      tabs: [tab, ...state.tabs],
      activeTabId: tab.id
    }))
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
