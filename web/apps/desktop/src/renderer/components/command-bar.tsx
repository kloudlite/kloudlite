import { useState, useRef, useEffect, useMemo, useCallback, type KeyboardEvent } from 'react'
import { Search, ArrowRight, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTabStore, type Tab } from '@/store/tabs'
import { useEnvironmentStore } from '@/store/environments'

const ENTER_ANIM = 'popover-in 150ms ease-out'
const EXIT_ANIM = 'popover-out 150ms ease-in forwards'

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

// ---------- New Tab (Cmd+T) — service picker ----------

interface NewTabBarProps {
  onNavigate: (url: string) => void
  onClose: () => void
}

export function NewTabBar({ onNavigate, onClose }: NewTabBarProps) {
  const { tabs, activeTabId, setActiveTab, addTab } = useTabStore()
  const selectedEnv = useEnvironmentStore((s) => s.environments.find(e => e.id === s.selectedEnvironmentId))
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [exiting, setExiting] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  // Filter open tabs
  const openTabUrls = new Set(tabs.map((t) => normalizeForDedup(t.url)))
  const filteredOpenTabs = useMemo(() => {
    const q = query.toLowerCase().trim()
    if (!q) return tabs
    return tabs.filter((t) =>
      t.title.toLowerCase().includes(q) ||
      t.url.toLowerCase().includes(q) ||
      (t.siteName && t.siteName.toLowerCase().includes(q))
    )
  }, [query, tabs])

  // Filter available services (not already open)
  const availableServices = useMemo(() => {
    if (!selectedEnv) return []
    const services = selectedEnv.services.filter((s) => !openTabUrls.has(normalizeForDedup(s.vpnUrl)))
    const q = query.toLowerCase().trim()
    if (!q) return services
    return services.filter((s) =>
      s.name.toLowerCase().includes(q) ||
      s.dnsHostname.toLowerCase().includes(q) ||
      String(s.port).includes(q)
    )
  }, [query, selectedEnv, openTabUrls])

  const totalItems = filteredOpenTabs.length + availableServices.length

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      if (selectedIndex < filteredOpenTabs.length) {
        setActiveTab(filteredOpenTabs[selectedIndex].id)
        close()
      } else if (selectedIndex < totalItems) {
        const svc = availableServices[selectedIndex - filteredOpenTabs.length]
        addTab(svc.vpnUrl)
        close()
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
        className="w-full max-w-[580px] overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl shadow-black/30"
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
            placeholder={`Search services${selectedEnv ? ` in ${selectedEnv.name}` : ''}...`}
            spellCheck={false}
          />
        </div>

        {totalItems > 0 && (
          <div ref={listRef} className="max-h-[50vh] overflow-y-auto border-t border-border/30 py-1.5">
            {/* Open tabs */}
            {filteredOpenTabs.map((tab, i) => (
              <div
                key={tab.id}
                className={cn(
                  'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
                  i === selectedIndex ? 'bg-accent/80' : 'hover:bg-accent/50'
                )}
                onClick={() => { setActiveTab(tab.id); close() }}
              >
                {tab.favicon ? (
                  <img src={tab.favicon} alt="" className="h-5 w-5 shrink-0 rounded-sm" />
                ) : (
                  <Globe className="h-5 w-5 shrink-0 text-muted-foreground/40" />
                )}
                <div className="min-w-0 flex-1">
                  <span className="block truncate text-[13px] text-foreground">{tab.title || 'New Tab'}</span>
                  <span className="block truncate text-[11px] text-muted-foreground/60">{extractDomain(tab.url)}</span>
                </div>
                <span className="flex shrink-0 items-center gap-1 text-[11px] text-muted-foreground">
                  Switch to Tab <ArrowRight className="h-3 w-3" />
                </span>
              </div>
            ))}

            {/* Available services */}
            {availableServices.length > 0 && (
              <>
                {filteredOpenTabs.length > 0 && (
                  <div className="mx-4 my-1 flex items-center gap-2 border-t border-border/20 pt-2">
                    <span className="text-[11px] font-medium text-muted-foreground/50">Services</span>
                  </div>
                )}
                {availableServices.map((svc, i) => (
                  <div
                    key={svc.id}
                    className={cn(
                      'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
                      filteredOpenTabs.length + i === selectedIndex ? 'bg-accent/80' : 'hover:bg-accent/50'
                    )}
                    onClick={() => { addTab(svc.vpnUrl); close() }}
                  >
                    <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-emerald-500/10">
                      <div className="h-2 w-2 rounded-full bg-emerald-500" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <span className="block truncate text-[13px] text-foreground">{svc.name}</span>
                      <span className="block truncate text-[11px] text-muted-foreground/60">{svc.dnsHostname}</span>
                    </div>
                    <span className="text-[11px] text-muted-foreground/40">Open</span>
                  </div>
                ))}
              </>
            )}
          </div>
        )}

        {totalItems === 0 && query && (
          <div className="border-t border-border/30 px-5 py-6 text-center text-[13px] text-muted-foreground/50">
            No matching services found
          </div>
        )}
      </div>
    </div>
  )
}
