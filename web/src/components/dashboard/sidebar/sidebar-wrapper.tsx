'use client'

import { Sidebar } from '@/components/ui/sidebar'

interface SidebarWrapperProps {
  children: React.ReactNode
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export function SidebarWrapper({ children, open = false, onOpenChange = () => {} }: SidebarWrapperProps) {
  return (
    <Sidebar
      open={open}
      onOpenChange={onOpenChange}
      variant="default"
      size="default"
      mobileBreakpoint="lg"
      className="h-full !p-0"
    >
      {children}
    </Sidebar>
  )
}