import { type ReactNode } from 'react'

interface SidebarSectionProps {
  title: string
  action?: ReactNode
  children: ReactNode
}

export function SidebarSection({ title, action, children }: SidebarSectionProps) {
  return (
    <>
      <div className="shrink-0 px-3">
        <div className="flex items-center justify-between px-3 pb-1.5">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-sidebar-foreground/50">
            {title}
          </span>
          {action}
        </div>
      </div>
      {children}
    </>
  )
}
