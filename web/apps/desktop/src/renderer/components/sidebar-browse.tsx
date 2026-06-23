import { Plus } from 'lucide-react'
import { useTabStore } from '@/store/tabs'
import { TabItem } from './tab-item'

export function SidebarBrowse() {
  const { tabs, activeTabId, closeTab, setActiveTab, moveTab } = useTabStore()

  function openNewTab() {
    window.dispatchEvent(new CustomEvent('open-command-bar'))
  }

  return (
    <>
      <div className="flex min-h-0 flex-1 flex-col">
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
