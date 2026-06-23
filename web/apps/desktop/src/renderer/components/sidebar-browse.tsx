import { Plus, ChevronDown, Globe } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useTabStore } from '@/store/tabs'
import { useEnvironmentStore, type Environment } from '@/store/environments'
import { TabItem } from './tab-item'

const DUMMY_ENVIRONMENTS: Environment[] = [
  {
    id: 'demo-production',
    name: 'demo-production',
    slug: 'demo-production',
    status: 'active',
    namespace: 'wm-demo',
    services: [
      { id: 'prod-api', name: 'api', port: 8080, dnsHostname: 'api.demo.local', vpnUrl: 'https://www.google.com' },
      { id: 'prod-web', name: 'web', port: 3000, dnsHostname: 'web.demo.local', vpnUrl: 'https://news.ycombinator.com' },
    ],
  },
  {
    id: 'demo-staging',
    name: 'demo-staging',
    slug: 'demo-staging',
    status: 'active',
    namespace: 'wm-demo',
    services: [
      { id: 'stage-dashboard', name: 'dashboard', port: 3001, dnsHostname: 'dashboard.demo.local', vpnUrl: 'https://example.com' },
    ],
  },
]

export function SidebarBrowse() {
  const { tabs, activeTabId, closeTab, setActiveTab, moveTab, addTab } = useTabStore()
  const [envPickerOpen, setEnvPickerOpen] = useState(false)

  const { environments, selectedEnvironmentId, setSelectedEnvironment } = useEnvironmentStore()
  const browseEnvironments = environments.length > 0 ? environments : DUMMY_ENVIRONMENTS
  const selectedEnv = browseEnvironments.find((e) => e.id === selectedEnvironmentId) ?? browseEnvironments[0]

  function openNewTab() {
    window.dispatchEvent(new CustomEvent('open-command-bar'))
  }

  function statusColor(status: Environment['status']) {
    if (status === 'active') return 'bg-emerald-400'
    if (status === 'error') return 'bg-red-400'
    return 'bg-sidebar-foreground/30'
  }

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col">
        {/* Environment picker */}
        <div className="shrink-0 px-3 pb-3">
          <div className="px-3 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/35">
            Environments
          </div>
          <div className="relative">
            <button
              className="no-drag flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium text-sidebar-foreground/75 transition-colors hover:bg-sidebar-foreground/[0.06]"
              onClick={() => setEnvPickerOpen(!envPickerOpen)}
            >
              <Globe className="h-4 w-4 shrink-0 text-sidebar-foreground/45" />
              <span className={cn('h-1.5 w-1.5 rounded-full', selectedEnv ? statusColor(selectedEnv.status) : 'bg-sidebar-foreground/30')} />
              <span className="min-w-0 flex-1 truncate text-left">{selectedEnv?.name || 'Select environment'}</span>
              <span className="text-[11px] text-sidebar-foreground/35">{selectedEnv?.services.length ?? 0} services</span>
              <ChevronDown className={cn('h-3.5 w-3.5 shrink-0 text-sidebar-foreground/45 transition-transform', envPickerOpen && 'rotate-180')} />
            </button>
            {envPickerOpen && (
              <>
              <div className="fixed inset-0 z-10" onClick={() => setEnvPickerOpen(false)} />
              <div className="absolute left-0 right-0 top-full z-20 mt-1 overflow-hidden rounded-lg border border-sidebar-foreground/[0.08] bg-sidebar shadow-xl" style={{ animation: 'dropdown-in 150ms cubic-bezier(0.22,1,0.36,1)' }}>
                {browseEnvironments.map((env) => (
                  <button
                    key={env.id}
                    className={cn(
                      'flex w-full items-center gap-2.5 px-3 py-2 text-left text-[12px] transition-colors',
                      env.id === selectedEnv?.id
                        ? 'bg-sidebar-foreground/[0.1] text-sidebar-foreground'
                        : 'text-sidebar-foreground/65 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/85'
                    )}
                    onClick={() => {
                      setSelectedEnvironment(env.id)
                      setEnvPickerOpen(false)
                    }}
                  >
                    <span className={cn('h-1.5 w-1.5 rounded-full', statusColor(env.status))} />
                    <span className="min-w-0 flex-1 truncate">{env.name}</span>
                    <span className="text-[10px] text-sidebar-foreground/35">{env.services.length} services</span>
                  </button>
                ))}
              </div>
              </>
            )}
          </div>
        </div>

        {/* Service list */}
        {selectedEnv && (
          <div className="shrink-0 px-3 pb-2">
            <div className="px-3 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/35">
              HTTP services
            </div>
            <div className="flex flex-col">
              {selectedEnv.services.map((svc) => (
                <button
                  key={svc.id}
                  className="no-drag flex h-9 w-full items-center gap-3 rounded-lg px-3 text-left text-[13px] text-sidebar-foreground/70 transition-colors hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/90"
                  onClick={() => addTab(svc.vpnUrl)}
                >
                  <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-400/70" />
                  <span className="min-w-0 flex-1 truncate">{svc.name}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Separator */}
        <div className="mx-5 my-2 shrink-0 border-t border-sidebar-foreground/[0.08]" />

        {/* New tab button */}
        <div className="shrink-0 px-3 pb-2">
          <button
            className="no-drag flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-[13px] font-medium text-sidebar-foreground/55 transition-all duration-150 hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/80"
            onClick={openNewTab}
          >
            <Plus className="h-4 w-4" />
            <span>New Tab</span>
          </button>
        </div>

        {/* Tab list */}
        <div className="sidebar-scroll min-h-0 flex-1 overflow-y-auto py-1">
          <div className="flex flex-col">
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

    </>
  )
}
