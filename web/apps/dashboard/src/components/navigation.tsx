'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  ThemeSwitcher,
  KloudliteLogo,
} from '@kloudlite/ui'
import { LayoutDashboard, Server, Cloud, Monitor, Package } from 'lucide-react'
import { VPNStatusIndicator } from './vpn-status-indicator'
import { UserProfileDropdown } from './user-profile-dropdown'

interface NavigationProps {
  email?: string
  displayName?: string
  isSuperAdmin?: boolean
  isAdmin?: boolean
  userRoles?: string[]
  hasWorkMachine?: boolean
  isWorkMachineRunning?: boolean
  setThemeCookie?: (theme: 'light' | 'dark' | 'system') => Promise<void>
}

export function Navigation({
  email,
  displayName,
  isSuperAdmin,
  isAdmin,
  userRoles: _userRoles = [],
  hasWorkMachine = false,
  isWorkMachineRunning = false,
  setThemeCookie,
}: NavigationProps) {
  const pathname = usePathname()
  const [mounted, setMounted] = useState(false)

  // Only render dropdown components after mounting to avoid hydration mismatch
  useEffect(() => {
    setMounted(true)
  }, [])

  const navItems = [
    { href: '/dashboard', label: 'Overview', icon: LayoutDashboard },
    { href: '/', label: 'Workmachine', icon: Server },
    { href: '/environments', label: 'Environments', icon: Cloud, requiresWorkMachine: true },
    { href: '/workspaces', label: 'Workspaces', icon: Monitor, requiresWorkMachine: true },
    { href: '/artifacts', label: 'Artifacts', icon: Package },
  ]

  return (
    <header className="sticky top-0 z-50 bg-background/95 backdrop-blur-sm border-b border-foreground/10 transition-colors duration-200">
      <div className="mx-auto max-w-7xl px-6">
        <div className="flex h-16 items-center justify-between">
          {/* Logo / Brand */}
          <div className="flex items-center gap-8">
            <KloudliteLogo className="text-lg font-semibold" />

            {/* Main Navigation */}
            <nav className="hidden items-center gap-8 md:flex">
              {navItems.map((item) => {
                // Check if current path is the item's path or a sub-path
                // Special handling for root "/" to avoid matching all paths
                const isActive = item.href === '/'
                  ? pathname === '/'
                  : pathname === item.href || pathname.startsWith(`${item.href}/`)
                const isDisabled = item.requiresWorkMachine && !hasWorkMachine

                if (isDisabled) {
                  return (
                    <div
                      key={item.label}
                      className="text-sm font-semibold text-foreground/30 cursor-not-allowed pb-1"
                      title="Create a work machine first"
                    >
                      {item.label}
                    </div>
                  )
                }

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`relative text-sm font-semibold transition-colors pb-1 after:absolute after:bottom-0 after:left-0 after:h-[2px] after:w-0 after:bg-primary after:transition-all after:duration-300 hover:after:w-full ${
                      isActive
                        ? 'text-foreground after:w-full'
                        : 'text-foreground/50 hover:text-foreground'
                    }`}
                  >
                    {item.label}
                  </Link>
                )
              })}
            </nav>
          </div>

          {/* VPN Status, Theme Switcher & User Dropdown */}
          <div className="flex items-center gap-2">
            {mounted && (
              <>
                {/* VPN Status with enhanced styling */}
                <div className="hidden sm:block">
                  <VPNStatusIndicator isWorkMachineRunning={isWorkMachineRunning} />
                </div>

                {/* Theme Switcher */}
                <ThemeSwitcher setThemeCookie={setThemeCookie} />

                {/* Divider */}
                <div className="h-6 w-px bg-border/60" />

                {/* User Profile */}
                <UserProfileDropdown
                  email={email}
                  displayName={displayName}
                  isAdmin={isAdmin}
                  isSuperAdmin={isSuperAdmin}
                />
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
