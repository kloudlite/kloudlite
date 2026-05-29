import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface PageMetadata {
  description: string
  siteName: string
  keywords: string
  author: string
  type: string
}

export interface HistoryEntry {
  url: string
  title: string
  favicon: string
  visitedAt: number
  visitCount: number
  metadata?: PageMetadata
}

interface HistoryStore {
  entries: HistoryEntry[]
  addEntry: (url: string, title: string, favicon: string) => void
  updateMetadata: (url: string, metadata: PageMetadata) => void
  search: (query: string, limit?: number) => HistoryEntry[]
  clear: () => void
}

const MAX_HISTORY = 5000

function normalizeUrl(url: string): string {
  try {
    const u = new URL(url)
    return u.origin + u.pathname.replace(/\/+$/, '')
  } catch {
    return url
  }
}

export const useHistoryStore = create<HistoryStore>()(
  persist(
    (set, get) => ({
      entries: [],

      addEntry: (url: string, title: string, favicon: string) => {
        if (!url || url === 'about:blank') return
        const normalized = normalizeUrl(url)
        set((state) => {
          const existing = state.entries.findIndex((e) => normalizeUrl(e.url) === normalized)
          let newEntries: HistoryEntry[]

          if (existing >= 0) {
            // Update existing entry — bump visit count and timestamp
            const entry = state.entries[existing]
            newEntries = [
              { ...entry, title: title || entry.title, favicon: favicon || entry.favicon, visitedAt: Date.now(), visitCount: entry.visitCount + 1 },
              ...state.entries.slice(0, existing),
              ...state.entries.slice(existing + 1)
            ]
          } else {
            newEntries = [
              { url, title: title || url, favicon, visitedAt: Date.now(), visitCount: 1 },
              ...state.entries
            ]
          }

          // Cap history size
          if (newEntries.length > MAX_HISTORY) {
            newEntries = newEntries.slice(0, MAX_HISTORY)
          }

          return { entries: newEntries }
        })
      },

      updateMetadata: (url: string, metadata: PageMetadata) => {
        set((state) => ({
          entries: state.entries.map((e) =>
            e.url === url ? { ...e, metadata } : e
          )
        }))
      },

      search: (query: string, limit = 8) => {
        const q = query.toLowerCase().trim()
        if (!q) {
          const recent: HistoryEntry[] = []
          const seenRecent = new Set<string>()
          for (const entry of get().entries) {
            if (recent.length >= limit) break
            const norm = normalizeUrl(entry.url)
            if (seenRecent.has(norm)) continue
            seenRecent.add(norm)
            recent.push(entry)
          }
          return recent
        }
        const words = q.split(/\s+/)
        const results: HistoryEntry[] = []
        const seen = new Set<string>()
        for (const entry of get().entries) {
          if (results.length >= limit) break
          const norm = normalizeUrl(entry.url)
          if (seen.has(norm)) continue
          seen.add(norm)
          const meta = entry.metadata
          const searchable = [
            entry.title, entry.url, entry.favicon,
            meta?.siteName, meta?.description, meta?.keywords, meta?.author
          ].filter(Boolean).join(' ').toLowerCase()
          if (words.every((w) => searchable.includes(w))) {
            results.push(entry)
          }
        }
        return results
      },

      clear: () => set({ entries: [] })
    }),
    {
      name: 'kloudlite-browser-history'
    }
  )
)
