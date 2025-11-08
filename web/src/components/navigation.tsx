'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { signOutAction } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import { ChevronDown, User, LogOut, Shield, Home, Cloud, Monitor, Key, Package } from 'lucide-react'
import { KloudliteLogo } from './kloudlite-logo'
import { ThemeSwitcher } from './theme-switcher'

interface NavigationProps {
  email?: string
  displayName?: string
  isSuperAdmin?: boolean
  isAdmin?: boolean
  userRoles?: string[]
  hasWorkMachine?: boolean
}

export function Navigation({
  email,
  displayName,
  isSuperAdmin,
  isAdmin,
  userRoles: _userRoles = [],
  hasWorkMachine = false,
}: NavigationProps) {
  const pathname = usePathname()

  const navItems = [
    { href: '/', label: 'Home', icon: Home },
    { href: '/environments', label: 'Environments', icon: Cloud, requiresWorkMachine: true },
    { href: '/workspaces', label: 'Workspaces', icon: Monitor, requiresWorkMachine: true },
    { href: '#', label: 'Artifacts', icon: Package, comingSoon: true },
  ]

  return (
    <header className="bg-background border-b">
      <div className="mx-auto max-w-7xl px-6">
        <div className="flex h-16 items-center justify-between">
          {/* Logo / Brand */}
          <div className="flex items-center gap-8">
            <KloudliteLogo className="text-lg font-medium" />

            {/* Main Navigation */}
            <nav className="hidden items-center gap-1 md:flex">
              {navItems.map((item) => {
                // Check if current path is the item's path or a sub-path
                const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)
                const Icon = item.icon
                const isComingSoon = item.comingSoon || false
                const isDisabled = (item.requiresWorkMachine && !hasWorkMachine) || isComingSoon

                const content = (
                  <>
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    <span className="whitespace-nowrap">{item.label}</span>
                    {isComingSoon && (
                      <Badge variant="secondary" className="ml-1.5 text-[10px] px-1.5 py-0.5 font-normal whitespace-nowrap">
                        Coming Soon
                      </Badge>
                    )}
                  </>
                )

                if (isDisabled) {
                  return (
                    <div
                      key={item.label}
                      className="flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground cursor-not-allowed opacity-60"
                    >
                      {content}
                    </div>
                  )
                }

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors whitespace-nowrap ${
                      isActive
                        ? 'bg-accent text-accent-foreground font-semibold'
                        : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
                    }`}
                  >
                    {content}
                  </Link>
                )
              })}
            </nav>
          </div>

          {/* User Dropdown & Theme Switcher */}
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="gap-1">
                  <User className="h-4 w-4" />
                  <span className="hidden sm:inline">{displayName || 'User'}</span>
                  <ChevronDown className="h-3 w-3" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-semibold">{displayName || 'User'}</p>
                    <p className="text-muted-foreground text-xs">{email}</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link href="/connection-tokens" className="cursor-pointer">
                    <Key className="mr-2 h-4 w-4" />
                    Connection Tokens
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                {(isAdmin || isSuperAdmin) && (
                  <>
                    <DropdownMenuItem asChild>
                      <Link href="/admin" className="cursor-pointer">
                        <Shield className="mr-2 h-4 w-4" />
                        Administration
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                  </>
                )}
                <form action={signOutAction}>
                  <DropdownMenuItem variant="destructive" asChild>
                    <button type="submit" className="w-full">
                      <LogOut className="mr-2 h-4 w-4" />
                      Sign out
                    </button>
                  </DropdownMenuItem>
                </form>
              </DropdownMenuContent>
            </DropdownMenu>
            <ThemeSwitcher />
          </div>
        </div>
      </div>
    </header>
  )
}
