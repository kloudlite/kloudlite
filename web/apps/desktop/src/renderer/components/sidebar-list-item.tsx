import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface SidebarListItemProps {
  icon?: ReactNode
  label: string
  active?: boolean
  right?: ReactNode
  onClick?: () => void
  onContextMenu?: (e: React.MouseEvent) => void
}

/**
 * Canonical sidebar list item — used for environments, workspaces, services,
 * tabs, and detail view navigation. Ensures consistent font, size, spacing.
 */
export function SidebarListItem({
  icon,
  label,
  active = false,
  right,
  onClick,
  onContextMenu,
}: SidebarListItemProps) {
  return (
    <button
      className={cn(
        'no-drag flex h-10 w-full items-center gap-3 rounded-[10px] px-3 text-left text-[14px] transition-colors duration-150',
        active
          ? 'bg-sidebar-foreground/[0.12] text-sidebar-foreground'
          : 'text-sidebar-foreground/75 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/90'
      )}
      style={{ fontWeight: active ? 500 : 450 }}
      onClick={onClick}
      onContextMenu={onContextMenu}
    >
      {icon && <span className="flex h-5 w-5 shrink-0 items-center justify-center">{icon}</span>}
      <span className="min-w-0 flex-1 truncate">{label}</span>
      {right && <span className="shrink-0">{right}</span>}
    </button>
  )
}
