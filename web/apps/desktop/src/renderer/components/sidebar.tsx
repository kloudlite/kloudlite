import { Plus, PanelLeft, ArrowLeft, ArrowRight, RotateCw, Link, Check, ShieldCheck, ShieldAlert } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { cn } from '@/lib/utils'
import { useTabStore } from '@/store/tabs'
import { TabItem } from './tab-item'
import { AddressBarOverlay } from './command-bar'
import { CertPopover } from './cert-popover'
import { TrafficLights } from './traffic-lights'

function extractDomain(url: string): string {
  if (!url || url === 'about:blank') return ''
  try {
    const u = new URL(url)
    return u.hostname.replace(/^www\./, '')
  } catch {
    return url
  }
}

interface SidebarProps {
  onNavigate: (url: string) => void
  onGoBack: () => void
  onGoForward: () => void
  onReload: () => void
  onToggleSidebar: () => void
}

export function Sidebar({ onNavigate, onGoBack, onGoForward, onReload, onToggleSidebar }: SidebarProps) {
  const { tabs, activeTabId, closeTab, setActiveTab, moveTab } = useTabStore()
  const activeTab = tabs.find((t) => t.id === activeTabId)
  const [addressBarOpen, setAddressBarOpen] = useState(false)
  const [newTabOpen, setNewTabOpen] = useState(false)
  const [copied, setCopied] = useState(false)
  const [certOpen, setCertOpen] = useState(false)
  const [certAnchorRect, setCertAnchorRect] = useState<DOMRect | null>(null)
  const certBtnRef = useRef<HTMLButtonElement>(null)
  const [anchorRect, setAnchorRect] = useState<DOMRect | null>(null)
  const urlBarRef = useRef<HTMLDivElement>(null)

  function openAddressBar() {
    if (urlBarRef.current) {
      setAnchorRect(urlBarRef.current.getBoundingClientRect())
    }
    setAddressBarOpen(true)
  }

  function openNewTab() {
    window.dispatchEvent(new CustomEvent('open-command-bar'))
  }

  // Listen for address bar and new tab state
  useEffect(() => {
    function handleOpenAddressBar() {
      openAddressBar()
    }
    function handleNewTabOpen() {
      setNewTabOpen(true)
    }
    function handleNewTabClose() {
      setNewTabOpen(false)
    }
    window.addEventListener('open-address-bar', handleOpenAddressBar)
    window.addEventListener('open-command-bar', handleNewTabOpen)
    window.addEventListener('close-command-bar', handleNewTabClose)
    return () => {
      window.removeEventListener('open-address-bar', handleOpenAddressBar)
      window.removeEventListener('open-command-bar', handleNewTabOpen)
      window.removeEventListener('close-command-bar', handleNewTabClose)
    }
  }, [])

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col">
        {/* Top row: traffic lights + sidebar toggle + nav buttons */}
        <div className="drag-region flex h-[52px] shrink-0 items-center justify-between px-4">
          <div className="flex items-center gap-3.5">
            <TrafficLights />
            <button
              className="no-drag rounded-lg p-1.5 text-sidebar-foreground/40 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/70"
              onClick={onToggleSidebar}
            >
              <PanelLeft className="h-[18px] w-[18px]" />
            </button>
          </div>

          <div className="no-drag flex items-center gap-0.5">
            <button
              className={cn(
                'rounded-lg p-1.5 transition-colors',
                activeTab?.canGoBack
                  ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                  : 'text-sidebar-foreground/20'
              )}
              onClick={onGoBack}
              disabled={!activeTab?.canGoBack}
            >
              <ArrowLeft className="h-5 w-5" />
            </button>
            <button
              className={cn(
                'rounded-lg p-1.5 transition-colors',
                activeTab?.canGoForward
                  ? 'text-sidebar-foreground/60 hover:text-sidebar-foreground'
                  : 'text-sidebar-foreground/20'
              )}
              onClick={onGoForward}
              disabled={!activeTab?.canGoForward}
            >
              <ArrowRight className="h-5 w-5" />
            </button>
            <button
              className="rounded-lg p-1.5 text-sidebar-foreground/60 transition-colors hover:text-sidebar-foreground"
              onClick={onReload}
            >
              <RotateCw className={cn('h-[18px] w-[18px]', activeTab?.isLoading && 'animate-spin')} />
            </button>
          </div>
        </div>

        {/* URL bar — frosted glass style */}
        <div className="shrink-0 px-3 pb-3">
          <div
            ref={urlBarRef}
            className="flex w-full items-center rounded-[10px] bg-sidebar-foreground/[0.08] px-3.5 py-1.5 backdrop-blur-sm transition-all duration-150 hover:bg-sidebar-foreground/[0.12]"
          >
            <button
              className="min-w-0 flex-1 text-left"
              onClick={openAddressBar}
            >
              <span className="truncate text-[13px] font-medium text-sidebar-foreground/55">
                {extractDomain(activeTab?.url ?? '') || 'Search or enter URL...'}
              </span>
            </button>
            {activeTab?.url && activeTab.url !== 'about:blank' && (
              <>
                <button
                  className="no-drag ml-2 shrink-0 rounded-md p-0.5 text-sidebar-foreground/40 transition-colors hover:text-sidebar-foreground/70"
                  onClick={(e) => {
                    e.stopPropagation()
                    navigator.clipboard.writeText(activeTab.url)
                    setCopied(true)
                    setTimeout(() => setCopied(false), 1500)
                  }}
                >
                  {copied ? <Check className="h-4 w-4" /> : <Link className="h-4 w-4" />}
                </button>
                <button
                  ref={certBtnRef}
                  className="no-drag ml-1 shrink-0 rounded-md p-0.5 text-sidebar-foreground/40 transition-colors hover:text-sidebar-foreground/70"
                  onClick={(e) => {
                    e.stopPropagation()
                    if (certBtnRef.current) {
                      setCertAnchorRect(certBtnRef.current.getBoundingClientRect())
                    }
                    setCertOpen(!certOpen)
                  }}
                >
                  {activeTab.url.startsWith('https://') ? (
                    <ShieldCheck className="h-4 w-4" />
                  ) : (
                    <ShieldAlert className="h-4 w-4" />
                  )}
                </button>
              </>
            )}
          </div>
        </div>

        {/* New tab button */}
        <div className="shrink-0 px-3 pb-2">
          <button
            className={cn(
              'no-drag flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-[13px] font-medium transition-all duration-150',
              newTabOpen
                ? 'bg-sidebar-foreground/[0.1] text-sidebar-foreground/70'
                : 'text-sidebar-foreground/45 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/70'
            )}
            onClick={openNewTab}
          >
            <Plus className="h-4 w-4" />
            <span>New Tab</span>
          </button>
        </div>

        {/* Tab list */}
        <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto py-1">
          <div className="flex flex-col gap-1">
            {tabs.map((tab, i) => (
              <TabItem
                key={tab.id}
                tab={tab}
                index={i}
                isActive={tab.id === activeTabId}
                onSelect={() => setActiveTab(tab.id)}
                onClose={() => closeTab(tab.id)}
                onMove={moveTab}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Address Bar overlay — anchored to URL bar */}
      {addressBarOpen && (
        <AddressBarOverlay
          anchorRect={anchorRect}
          onNavigate={onNavigate}
          onClose={() => setAddressBarOpen(false)}
        />
      )}

      {/* Certificate popover */}
      {certOpen && activeTab?.url && (
        <CertPopover
          url={activeTab.url}
          anchorRect={certAnchorRect}
          onClose={() => setCertOpen(false)}
        />
      )}
    </>
  )
}
