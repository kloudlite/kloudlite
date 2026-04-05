import { useState, useRef, useEffect, type KeyboardEvent } from 'react'
import { Search, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTabStore, type Tab } from '@/store/tabs'

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

function filterTabs(tabs: Tab[], query: string): Tab[] {
  const q = query.toLowerCase().trim()
  if (!q) return tabs
  return tabs.filter((t) =>
    t.title.toLowerCase().includes(q) ||
    t.url.toLowerCase().includes(q)
  )
}

// ---------- New Tab (Cmd+T) — centered on screen ----------

interface NewTabBarProps {
  onNavigate: (url: string) => void
  onClose: () => void
}

export function NewTabBar({ onNavigate, onClose }: NewTabBarProps) {
  const { tabs, activeTabId, setActiveTab, addTab } = useTabStore()
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const filteredTabs = filterTabs(tabs, query)

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      if (filteredTabs.length > 0 && selectedIndex >= 0 && selectedIndex < filteredTabs.length && !query.includes('.') && !query.includes('://')) {
        setActiveTab(filteredTabs[selectedIndex].id)
        onClose()
      } else {
        const url = normalizeUrl(query)
        if (url) {
          addTab(url)
          onClose()
        }
      }
    } else if (e.key === 'Escape') {
      onClose()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, filteredTabs.length - 1))
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
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[12vh]" onClick={onClose}>
      <div
        className="w-full max-w-[680px] overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl shadow-black/30"
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

        {filteredTabs.length > 0 && (
          <div ref={listRef} className="max-h-[50vh] overflow-y-auto border-t border-border/30 py-1.5">
            {filteredTabs.map((tab, i) => (
              <TabRow
                key={tab.id}
                tab={tab}
                isSelected={i === selectedIndex}
                onClick={() => {
                  setActiveTab(tab.id)
                  onClose()
                }}
              />
            ))}
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
  const activeTab = tabs.find((t) => t.id === activeTabId)
  const initialUrl = activeTab?.url === 'about:blank' ? '' : (activeTab?.url ?? '')
  const [query, setQuery] = useState(initialUrl)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const filteredTabs = filterTabs(tabs, query)

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
        onClose()
      } else {
        const url = normalizeUrl(query)
        if (url) {
          onNavigate(url)
          onClose()
        }
      }
    } else if (e.key === 'Escape') {
      onClose()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, filteredTabs.length - 1))
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
    <div className="fixed inset-0 z-50" onClick={onClose}>
      <div
        className="fixed overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl shadow-black/30"
        style={style}
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

        {filteredTabs.length > 0 && (
          <div ref={listRef} className="max-h-[45vh] overflow-y-auto border-t border-border/30 py-1.5">
            {filteredTabs.map((tab, i) => (
              <TabRow
                key={tab.id}
                tab={tab}
                isSelected={i === selectedIndex}
                onClick={() => {
                  setActiveTab(tab.id)
                  onClose()
                }}
              />
            ))}
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
