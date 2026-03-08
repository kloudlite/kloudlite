'use client'

import { useState, useRef, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@kloudlite/lib'

const tabs = [
  { id: 'organization', label: 'Organization', href: '/installations/settings/organization' },
  { id: 'billing', label: 'Billing', href: '/installations/settings/billing' },
]

export function InstallationSettingsTabs() {
  const pathname = usePathname()
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })
  const tabRefs = useRef<Map<string, HTMLAnchorElement>>(new Map())

  const activeTab = tabs.find((tab) => pathname === tab.href)?.id || 'organization'

  // Update underline position
  useEffect(() => {
    const updatePosition = () => {
      const activeRef = tabRefs.current.get(activeTab)
      if (activeRef) {
        const fullWidth = activeRef.offsetWidth
        const underlineWidth = fullWidth * 0.6 // 60% of tab width
        const leftOffset = activeRef.offsetLeft + (fullWidth - underlineWidth) / 2

        setUnderlineStyle({
          left: leftOffset,
          width: underlineWidth,
        })
      }
    }

    // Small delay to ensure layout is ready
    setTimeout(updatePosition, 10)

    window.addEventListener('resize', updatePosition)
    return () => window.removeEventListener('resize', updatePosition)
  }, [activeTab])

  return (
    <div className="inline-flex gap-1 relative">
      {tabs.map((tab) => (
        <Link
          key={tab.id}
          ref={(el) => {
            if (el) tabRefs.current.set(tab.id, el)
          }}
          href={tab.href}
          className={cn(
            'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
            'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
            activeTab === tab.id
              ? 'text-foreground'
              : 'text-muted-foreground hover:text-foreground'
          )}
        >
          {tab.label}
        </Link>
      ))}

      {/* Animated underline with CSS transition */}
      {underlineStyle.width > 0 && (
        <div
          className="absolute bottom-1 h-[2px] bg-primary transition-all duration-300 ease-out"
          style={{
            left: `${underlineStyle.left}px`,
            width: `${underlineStyle.width}px`,
          }}
        />
      )}
    </div>
  )
}
