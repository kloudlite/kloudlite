import { X } from 'lucide-react'
import { cn } from '@kloudlite/lib'
import type { Tab } from '@/store/tabs'

interface TabItemProps {
  tab: Tab
  isActive: boolean
  onSelect: () => void
  onClose: () => void
}

export function TabItem({ tab, isActive, onSelect, onClose }: TabItemProps) {
  return (
    <div
      className={cn(
        'group flex h-9 cursor-pointer items-center gap-2 px-3 text-sm transition-colors',
        isActive
          ? 'bg-accent text-accent-foreground'
          : 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground'
      )}
      onClick={onSelect}
    >
      {tab.favicon ? (
        <img src={tab.favicon} alt="" className="h-4 w-4 shrink-0" />
      ) : (
        <div className="h-4 w-4 shrink-0 rounded-sm bg-muted" />
      )}
      <span className="min-w-0 flex-1 truncate">
        {tab.title || 'New Tab'}
      </span>
      <button
        className="shrink-0 rounded-sm p-0.5 opacity-0 transition-opacity hover:bg-muted group-hover:opacity-100"
        onClick={(e) => {
          e.stopPropagation()
          onClose()
        }}
      >
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}
