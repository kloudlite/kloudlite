import { useState, useRef, useEffect, useMemo, useCallback, type KeyboardEvent } from 'react'
import { Search, ArrowRight, Clock, Globe, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTabStore, type Tab } from '@/store/tabs'
import { useHistoryStore, type HistoryEntry } from '@/store/history'

const ENTER_ANIM = 'popover-in 150ms ease-out'
const EXIT_ANIM = 'popover-out 150ms ease-in forwards'

function normalizeUrl(input: string): string {
  const trimmed = input.trim()
  if (!trimmed) return ''
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  if (trimmed.includes('.') && !trimmed.includes(' ')) {
    return `https://${trimmed}`
  }
  return `https://www.google.com/search?q=${encodeURIComponent(trimmed)}`
}

function extractDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '')
  } catch {
    return ''
  }
}

function normalizeForDedup(url: string): string {
  try {
    const u = new URL(url)
    return u.origin + u.pathname.replace(/\/+$/, '')
  } catch {
    return url
  }
}

function buildSearchText(parts: (string | undefined)[]): string {
  return parts.filter(Boolean).join(' ').toLowerCase()
}

function filterTabs(tabs: Tab[], query: string): Tab[] {
  const q = query.toLowerCase().trim()
  if (!q) return tabs
  return tabs.filter((t) => {
    const searchable = buildSearchText([t.title, t.url, t.favicon, t.siteName, t.keywords])
    return q.split(/\s+/).every((word) => searchable.includes(word))
  })
}

// ---------- New Tab (Cmd+T) — centered on screen ----------

interface NewTabBarProps {
  onNavigate: (url: string) => void
  onClose: () => void
}

export function NewTabBar({ onNavigate, onClose }: NewTabBarProps) {
  const { tabs, activeTabId, setActiveTab, addTab } = useTabStore()
  const historySearch = useHistoryStore((s) => s.search)
  const clearHistory = useHistoryStore((s) => s.clear)
  const historyEntries = useHistoryStore((s) => s.entries)
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [exiting, setExiting] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  const filteredTabs = filterTabs(tabs, query)
  const openTabUrls = new Set(tabs.map((t) => normalizeForDedup(t.url)))
  const historyResults = useMemo(() =>
    historySearch(query, 6).filter((h) => !openTabUrls.has(normalizeForDedup(h.url))),
    [query, historyEntries]
  )
  const totalItems = filteredTabs.length + historyResults.length

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      if (selectedIndex < filteredTabs.length && filteredTabs.length > 0 && !query.includes('.') && !query.includes('://')) {
        setActiveTab(filteredTabs[selectedIndex].id)
        close()
      } else if (selectedIndex >= filteredTabs.length && selectedIndex < totalItems) {
        const historyItem = historyResults[selectedIndex - filteredTabs.length]
        addTab(historyItem.url)
        close()
      } else {
        const url = normalizeUrl(query)
        if (url) {
          addTab(url)
          close()
        }
      }
    } else if (e.key === 'Escape') {
      close()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, totalItems - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => Math.max(i - 1, 0))
    }
  }

  useEffect(() => {
    if (selectedIndex >= 0 && listRef.current) {
      const item = listRef.current.children[selectedIndex] as HTMLElement
      item?.scrollIntoView({ block: 'nearest' })
    }
  }, [selectedIndex])

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[12vh]" onClick={close}>
      <div
        className="w-full max-w-[680px] overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl shadow-black/30"
        style={{ animation: exiting ? EXIT_ANIM : ENTER_ANIM }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-5 py-4">
          <Search className="h-5 w-5 shrink-0 text-muted-foreground" />
          <input
            ref={inputRef}
            type="text"
            className="w-full bg-transparent text-[16px] text-foreground outline-none placeholder:text-muted-foreground/50"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search or Enter URL..."
            spellCheck={false}
          />
        </div>

        {totalItems > 0 && (
          <div ref={listRef} className="max-h-[50vh] overflow-y-auto border-t border-border/30 py-1.5">
            {filteredTabs.map((tab, i) => (
              <TabRow
                key={tab.id}
                tab={tab}
                isSelected={i === selectedIndex}
                onClick={() => {
                  setActiveTab(tab.id)
                  close()
                }}
              />
            ))}
            {historyResults.length > 0 && (
              <>
                <div className="mx-4 my-1 flex items-center gap-2 border-t border-border/20 pt-2">
                  <span className="text-[11px] font-medium text-muted-foreground/50">History</span>
                  <div className="flex-1" />
                  <button
                    className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] text-muted-foreground/50 transition-colors hover:bg-accent hover:text-muted-foreground"
                    onClick={(e) => {
                      e.stopPropagation()
                      clearHistory()
                    }}
                  >
                    <Trash2 className="h-3 w-3" />
                    Clear
                  </button>
                </div>
                {historyResults.map((entry, i) => (
                  <HistoryRow
                    key={entry.url}
                    entry={entry}
                    isSelected={filteredTabs.length + i === selectedIndex}
                    onClick={() => {
                      addTab(entry.url)
                      close()
                    }}
                  />
                ))}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

// ---------- Address Bar (Cmd+L / click URL) — anchored to URL bar ----------

interface AddressBarOverlayProps {
  onNavigate: (url: string) => void
  onClose: () => void
  anchorRect: DOMRect | null
}

export function AddressBarOverlay({ onNavigate, onClose, anchorRect }: AddressBarOverlayProps) {
  const { tabs, activeTabId, setActiveTab } = useTabStore()
  const historySearch = useHistoryStore((s) => s.search)
  const clearHistory = useHistoryStore((s) => s.clear)
  const historyEntries = useHistoryStore((s) => s.entries)
  const activeTab = tabs.find((t) => t.id === activeTabId)
  const initialUrl = activeTab?.url === 'about:blank' ? '' : (activeTab?.url ?? '')
  const [query, setQuery] = useState(initialUrl)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [exiting, setExiting] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  const filteredTabs = filterTabs(tabs, query)
  const openTabUrls = new Set(tabs.map((t) => normalizeForDedup(t.url)))
  const historyResults = useMemo(() =>
    historySearch(query, 5).filter((h) => !openTabUrls.has(normalizeForDedup(h.url))),
    [query, historyEntries]
  )
  const totalItems = filteredTabs.length + historyResults.length

  useEffect(() => {
    setTimeout(() => {
      inputRef.current?.focus()
      inputRef.current?.select()
    }, 0)
  }, [])

  useEffect(() => {
    setSelectedIndex(-1)
  }, [query])

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      if (selectedIndex >= 0 && selectedIndex < filteredTabs.length) {
        setActiveTab(filteredTabs[selectedIndex].id)
        close()
      } else if (selectedIndex >= filteredTabs.length && selectedIndex < totalItems) {
        const historyItem = historyResults[selectedIndex - filteredTabs.length]
        onNavigate(historyItem.url)
        close()
      } else {
        const url = normalizeUrl(query)
        if (url) {
          onNavigate(url)
          close()
        }
      }
    } else if (e.key === 'Escape') {
      close()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, totalItems - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => Math.max(i - 1, -1))
    }
  }

  useEffect(() => {
    if (selectedIndex >= 0 && listRef.current) {
      const item = listRef.current.children[selectedIndex] as HTMLElement
      item?.scrollIntoView({ block: 'nearest' })
    }
  }, [selectedIndex])

  const style: React.CSSProperties = anchorRect
    ? { top: anchorRect.top, left: anchorRect.left, width: 'fit-content', minWidth: anchorRect.width, maxWidth: Math.min(640, window.innerWidth - anchorRect.left - 16) }
    : { top: 52, left: 12, width: 'fit-content', minWidth: 320, maxWidth: 640 }

  return (
    <div className="fixed inset-0 z-50" onClick={close}>
      <div
        className="fixed overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl shadow-black/30"
        style={{ ...style, animation: exiting ? EXIT_ANIM : ENTER_ANIM }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3">
          {activeTab?.favicon ? (
            <img src={activeTab.favicon} alt="" className="h-4 w-4 shrink-0 rounded-sm" />
          ) : (
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
          )}
          <input
            ref={inputRef}
            type="text"
            className="min-w-0 flex-1 bg-transparent text-[14px] text-foreground outline-none placeholder:text-muted-foreground/50"
            size={Math.max(20, query.length + 2)}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search or enter URL..."
            spellCheck={false}
          />
        </div>

        {totalItems > 0 && (
          <div ref={listRef} className="max-h-[45vh] overflow-y-auto border-t border-border/30 py-1.5">
            {filteredTabs.map((tab, i) => (
              <TabRow
                key={tab.id}
                tab={tab}
                isSelected={i === selectedIndex}
                onClick={() => {
                  setActiveTab(tab.id)
                  close()
                }}
              />
            ))}
            {historyResults.length > 0 && (
              <>
                <div className="mx-4 my-1 flex items-center gap-2 border-t border-border/20 pt-2">
                  <span className="text-[11px] font-medium text-muted-foreground/50">History</span>
                  <div className="flex-1" />
                  <button
                    className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] text-muted-foreground/50 transition-colors hover:bg-accent hover:text-muted-foreground"
                    onClick={(e) => {
                      e.stopPropagation()
                      clearHistory()
                    }}
                  >
                    <Trash2 className="h-3 w-3" />
                    Clear
                  </button>
                </div>
            {historyResults.map((entry, i) => (
              <HistoryRow
                key={entry.url}
                entry={entry}
                isSelected={filteredTabs.length + i === selectedIndex}
                onClick={() => {
                  onNavigate(entry.url)
                  close()
                }}
              />
            ))}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

// ---------- Shared tab row ----------

function TabRow({ tab, isSelected, onClick }: {
  tab: Tab
  isSelected: boolean
  onClick: () => void
}) {
  return (
    <div
      className={cn(
        'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
        isSelected ? 'bg-accent/80' : 'hover:bg-accent/50'
      )}
      onClick={onClick}
    >
      {tab.favicon ? (
        <img src={tab.favicon} alt="" className="h-5 w-5 shrink-0 rounded-sm" />
      ) : (
        <div className="h-5 w-5 shrink-0 rounded-full bg-muted-foreground/20" />
      )}
      <div className="min-w-0 flex-1">
        <span className="block truncate text-[13px] text-foreground">
          {tab.title || 'New Tab'}
        </span>
        <span className="block truncate text-[11px] text-muted-foreground/60">
          {extractDomain(tab.url)}
        </span>
      </div>
      <span className="flex shrink-0 items-center gap-1 text-[11px] text-muted-foreground">
        Switch to Tab
        <ArrowRight className="h-3 w-3" />
      </span>
    </div>
  )
}

// ---------- History row ----------

function HistoryRow({ entry, isSelected, onClick }: {
  entry: HistoryEntry
  isSelected: boolean
  onClick: () => void
}) {
  return (
    <div
      className={cn(
        'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
        isSelected ? 'bg-accent/80' : 'hover:bg-accent/50'
      )}
      onClick={onClick}
    >
      {entry.favicon ? (
        <img src={entry.favicon} alt="" className="h-5 w-5 shrink-0 rounded-sm" />
      ) : (
        <Globe className="h-5 w-5 shrink-0 text-muted-foreground/40" />
      )}
      <div className="min-w-0 flex-1">
        <span className="block truncate text-[13px] text-foreground">
          {entry.title}
        </span>
        <span className="block truncate text-[11px] text-muted-foreground/60">
          {extractDomain(entry.url)}
        </span>
      </div>
      <span className="flex shrink-0 items-center gap-1 text-[11px] text-muted-foreground">
        <Clock className="h-3 w-3" />
        History
      </span>
    </div>
  )
}
