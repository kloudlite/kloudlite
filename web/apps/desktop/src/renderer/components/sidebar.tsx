import { Plus } from 'lucide-react'
import { ScrollArea } from '@kloudlite/ui'
import { useTabStore } from '@/store/tabs'
import { TabItem } from './tab-item'

export function Sidebar() {
  const { tabs, activeTabId, addTab, closeTab, setActiveTab } = useTabStore()

  return (
    <div className="flex h-full w-[220px] shrink-0 flex-col border-r border-border bg-sidebar text-sidebar-foreground">
      {/* Drag region for macOS traffic lights */}
      <div className="drag-region flex h-12 items-center justify-between px-3 pt-1">
        <span className="pl-16 text-xs font-medium text-muted-foreground">Tabs</span>
        <button
          className="no-drag rounded-sm p-1 text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          onClick={() => addTab()}
        >
          <Plus className="h-4 w-4" />
        </button>
      </div>

      {/* Tab list */}
      <ScrollArea className="flex-1">
        <div className="flex flex-col py-1">
          {tabs.map((tab) => (
            <TabItem
              key={tab.id}
              tab={tab}
              isActive={tab.id === activeTabId}
              onSelect={() => setActiveTab(tab.id)}
              onClose={() => closeTab(tab.id)}
            />
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
