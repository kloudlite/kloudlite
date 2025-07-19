'use client'

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'

export function SidebarAutoClose() {
  const pathname = usePathname()

  useEffect(() => {
    // Close sidebar when pathname changes (navigation)
    const sidebarToggle = document.getElementById('sidebar-toggle') as HTMLInputElement
    if (sidebarToggle) {
      sidebarToggle.checked = false
    }
  }, [pathname])

  return null
}