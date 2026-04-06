import { Plus, ChevronDown, Globe } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useTabStore } from '@/store/tabs'
import { useEnvironmentStore } from '@/store/environments'
import { TabItem } from './tab-item'

export function SidebarBrowse() {
  const { tabs, activeTabId, closeTab, setActiveTab, moveTab, addTab } = useTabStore()
  const [envPickerOpen, setEnvPickerOpen] = useState(false)

  const { environments, selectedEnvironmentId, setSelectedEnvironment } = useEnvironmentStore()
  const selectedEnv = environments.find((e) => e.id === selectedEnvironmentId)

  function openNewTab() {
    window.dispatchEvent(new CustomEvent('open-command-bar'))
  }

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col">
        {/* Environment picker */}
        <div className="shrink-0 px-3 pb-2">
          <div className="relative">
            <button
              className="no-drag flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[12px] font-medium text-sidebar-foreground/70 transition-all duration-150 hover:bg-sidebar-foreground/[0.06]"
              onClick={() => setEnvPickerOpen(!envPickerOpen)}
            >
              <Globe className="h-3.5 w-3.5 shrink-0" />
              <span className="flex-1 truncate text-left">{selectedEnv?.name || 'Select Environment'}</span>
              <ChevronDown className={cn('h-3.5 w-3.5 transition-transform', envPickerOpen && 'rotate-180')} />
            </button>
            {envPickerOpen && (
              <>
              <div className="fixed inset-0 z-10" onClick={() => setEnvPickerOpen(false)} />
              <div className="absolute left-0 right-0 top-full z-20 mt-1 overflow-hidden rounded-lg border border-sidebar-foreground/[0.08] bg-sidebar shadow-lg" style={{ animation: 'popover-in 150ms ease-out' }}>
                {environments.map((env) => (
                  <button
                    key={env.id}
                    className={cn(
                      'flex w-full items-center gap-2 px-3 py-2 text-[12px] transition-colors',
                      env.id === selectedEnvironmentId
                        ? 'bg-sidebar-foreground/[0.1] text-sidebar-foreground/90 font-medium'
                        : 'text-sidebar-foreground/60 hover:bg-sidebar-foreground/[0.06]'
                    )}
                    onClick={() => {
                      setSelectedEnvironment(env.id)
                      setEnvPickerOpen(false)
                    }}
                  >
                    <div className={cn(
                      'h-1.5 w-1.5 rounded-full',
                      env.status === 'active' ? 'bg-emerald-400' : env.status === 'error' ? 'bg-red-400' : 'bg-sidebar-foreground/30'
                    )} />
                    <span>{env.name}</span>
                    <span className="ml-auto text-[10px] text-sidebar-foreground/40">{env.services.length} services</span>
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
            <div className="text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/40 px-3 pb-1.5">
              Exposed Web Services
            </div>
            <div className="flex flex-col gap-0.5">
              {selectedEnv.services.map((svc) => (
                <button
                  key={svc.id}
                  className="no-drag flex w-full items-center gap-2 rounded-lg px-3 py-1.5 text-left text-[12px] text-sidebar-foreground/65 transition-all duration-150 hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/90"
                  onClick={() => addTab(svc.vpnUrl)}
                >
                  <div className="h-2 w-2 shrink-0 rounded-full bg-emerald-400/60" />
                  <span className="flex-1 truncate">{svc.name}</span>
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

    </>
  )
}
