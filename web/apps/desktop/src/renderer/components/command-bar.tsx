import { useState, useRef, useEffect, useMemo, useCallback, type KeyboardEvent } from 'react'
import { Search, ArrowRight, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTabStore } from '@/store/tabs'
import { useEnvironmentStore } from '@/store/environments'
import { useHistoryStore } from '@/store/history'

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
  if (!url) return ''
  try {
    const u = new URL(url)
    return u.origin + u.pathname.replace(/\/+$/, '')
  } catch {
    return url
  }
}

function urlForQuery(query: string): string {
  const raw = query.trim()
  const hasScheme = /^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(raw)
  const looksLikeHost = !raw.includes(' ') && raw.includes('.')
  if (hasScheme) return raw
  if (looksLikeHost) return `https://${raw}`
  return `https://www.google.com/search?q=${encodeURIComponent(raw)}`
}

interface NewTabBarProps {
  onNavigate: (url: string) => void
  onClose: () => void
}

export function NewTabBar({ onClose }: NewTabBarProps) {
  const { tabs, setActiveTab, addTab } = useTabStore()
  const environments = useEnvironmentStore((s) => s.environments)
  const searchHistory = useHistoryStore((s) => s.search)
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [exiting, setExiting] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  const q = query.trim()
  const openTabUrls = useMemo(() => new Set(tabs.map((t) => normalizeForDedup(t.url))), [tabs])

  const filteredOpenTabs = useMemo(() => {
    const term = q.toLowerCase()
    if (!term) return tabs
    return tabs.filter((t) =>
      t.title.toLowerCase().includes(term) ||
      t.url.toLowerCase().includes(term) ||
      (t.siteName && t.siteName.toLowerCase().includes(term))
    )
  }, [q, tabs])

  const availableServices = useMemo(() => {
    const services = environments.flatMap((env) =>
      env.services.map((svc) => ({ ...svc, envName: env.name }))
    ).filter((s) => !openTabUrls.has(normalizeForDedup(s.vpnUrl)))
    const term = q.toLowerCase()
    if (!term) return services
    return services.filter((s) =>
      s.name.toLowerCase().includes(term) ||
      s.envName.toLowerCase().includes(term) ||
      s.dnsHostname.toLowerCase().includes(term) ||
      String(s.port).includes(term)
    )
  }, [q, environments, openTabUrls])

  const serviceUrls = useMemo(() => new Set(availableServices.map((s) => normalizeForDedup(s.vpnUrl))), [availableServices])
  const historyEntries = useMemo(() => {
    return searchHistory(q, 6).filter((entry) => {
      const url = normalizeForDedup(entry.url)
      return url && !openTabUrls.has(url) && !serviceUrls.has(url)
    })
  }, [q, searchHistory, openTabUrls, serviceUrls])

  const hasQueryAction = q.length > 0
  const openTabStart = hasQueryAction ? 1 : 0
  const historyStart = openTabStart + filteredOpenTabs.length
  const servicesStart = historyStart + historyEntries.length
  const totalItems = servicesStart + availableServices.length

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  function openUrl(url: string) {
    addTab(url)
    close()
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      if (hasQueryAction && selectedIndex === 0) {
        openUrl(urlForQuery(q))
      } else if (selectedIndex >= openTabStart && selectedIndex < historyStart) {
        setActiveTab(filteredOpenTabs[selectedIndex - openTabStart].id)
        close()
      } else if (selectedIndex >= historyStart && selectedIndex < servicesStart) {
        openUrl(historyEntries[selectedIndex - historyStart].url)
      } else if (selectedIndex >= servicesStart && selectedIndex < totalItems) {
        openUrl(availableServices[selectedIndex - servicesStart].vpnUrl)
      } else {
        addTab()
        close()
      }
    } else if (e.key === 'Escape') {
      close()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, Math.max(totalItems - 1, 0)))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => Math.max(i - 1, 0))
    }
  }

  useEffect(() => {
    const item = listRef.current?.querySelector(`[data-result-index="${selectedIndex}"]`) as HTMLElement | null
    item?.scrollIntoView({ block: 'nearest' })
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
            placeholder="Search services..."
            spellCheck={false}
          />
        </div>

        {totalItems > 0 && (
          <div ref={listRef} className="max-h-[50vh] overflow-y-auto border-t border-border/30 py-1.5">
            {hasQueryAction && (
              <div
                data-result-index={0}
                className={cn(
                  'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
                  selectedIndex === 0 ? 'bg-accent/80' : 'hover:bg-accent/50'
                )}
                onClick={() => openUrl(urlForQuery(q))}
              >
                <Search className="h-5 w-5 shrink-0 text-muted-foreground/50" />
                <div className="min-w-0 flex-1">
                  <span className="block truncate text-[13px] text-foreground">{q}</span>
                  <span className="block truncate text-[11px] text-muted-foreground/60">{q.includes(' ') || !q.includes('.') ? 'Google Search' : 'Open URL'}</span>
                </div>
              </div>
            )}

            {filteredOpenTabs.map((tab, i) => {
              const index = openTabStart + i
              return (
                <div
                  key={tab.id}
                  data-result-index={index}
                  className={cn(
                    'group mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
                    index === selectedIndex ? 'bg-accent/80' : 'hover:bg-accent/50'
                  )}
                  onClick={() => { setActiveTab(tab.id); close() }}
                >
                  {tab.favicon ? <img src={tab.favicon} alt="" className="h-5 w-5 shrink-0 rounded-sm" /> : <Globe className="h-5 w-5 shrink-0 text-muted-foreground/40" />}
                  <div className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] text-foreground">{tab.title || 'New Tab'}</span>
                    <span className="block truncate text-[11px] text-muted-foreground/60">{extractDomain(tab.url)}</span>
                  </div>
                  <span className={cn('flex shrink-0 items-center gap-1 text-[11px] text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100', index === selectedIndex && 'opacity-100')}>
                    Switch to Tab <ArrowRight className="h-3 w-3" />
                  </span>
                </div>
              )
            })}

            {historyEntries.length > 0 && (
              <div className={cn('mx-4 mb-1 flex items-center gap-2', (hasQueryAction || filteredOpenTabs.length > 0) && 'mt-1 border-t border-border/20 pt-2')}>
                <span className="text-[10px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/45">History</span>
              </div>
            )}
            {historyEntries.map((entry, i) => {
              const index = historyStart + i
              return (
                <div
                  key={entry.url}
                  data-result-index={index}
                  className={cn(
                    'mx-1.5 flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 transition-colors',
                    index === selectedIndex ? 'bg-accent/80' : 'hover:bg-accent/50'
                  )}
                  onClick={() => openUrl(entry.url)}
                >
                  {entry.favicon ? <img src={entry.favicon} alt="" className="h-5 w-5 shrink-0 rounded-sm" /> : <Globe className="h-5 w-5 shrink-0 text-muted-foreground/40" />}
                  <div className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] text-foreground">{entry.title || extractDomain(entry.url)}</span>
                    <span className="block truncate text-[11px] text-muted-foreground/60">{extractDomain(entry.url)}</span>
                  </div>
                </div>
              )
            })}

            {availableServices.length > 0 && (
              <div className={cn('mx-4 mb-1 flex items-center gap-2', (hasQueryAction || filteredOpenTabs.length > 0 || historyEntries.length > 0) && 'mt-1 border-t border-border/20 pt-2')}>
                <span className="text-[10px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/45">HTTP services</span>
              </div>
            )}
            {availableServices.map((svc, i) => {
              const index = servicesStart + i
              return (
                <div
                  key={`${svc.envName}-${svc.id}`}
                  data-result-index={index}
                  className={cn(
                    'mx-1.5 flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 transition-colors',
                    index === selectedIndex ? 'bg-accent/75' : 'hover:bg-accent/45'
                  )}
                  onClick={() => openUrl(svc.vpnUrl)}
                >
                  <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-emerald-500/10"><span className="h-2 w-2 rounded-full bg-emerald-500" /></span>
                  <div className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] font-medium text-foreground">{svc.envName}/{svc.name}</span>
                    <span className="block truncate text-[11px] text-muted-foreground/55">{svc.dnsHostname}</span>
                  </div>
                  <span className="rounded-md bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">:{svc.port}</span>
                </div>
              )
            })}
          </div>
        )}

        {totalItems === 0 && (
          <div className="flex cursor-pointer items-center gap-3 border-t border-border/30 px-5 py-4 text-[13px] text-muted-foreground/50 transition-colors hover:bg-accent/30" onClick={() => { addTab(); close() }}>
            <Globe className="h-4 w-4" />
            <span>New Tab</span>
            <span className="ml-auto font-mono text-[10px] text-muted-foreground/30">⏎</span>
          </div>
        )}
      </div>
    </div>
  )
}
