'use client'

import { Button } from '@/components/ui/button'
import { UserProfileDropdown } from '@/components/teams/user-profile-dropdown'
import { ThemeToggleClient } from '@/components/theme-toggle-client'
import { 
  Search, 
  Bell, 
  HelpCircle,
  Command,
  Terminal,
  Sparkles,
  Menu
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface DashboardHeaderProps {
  teamName: string
  onSidebarToggle?: () => void
}

export function DashboardHeader({ teamName, onSidebarToggle }: DashboardHeaderProps) {
  return (
    <header className="h-16 border-b bg-background/80 backdrop-blur-xl">
      <div className="max-w-screen-2xl mx-auto">
        <div className="flex h-full items-center gap-4 px-6 lg:px-8">
        {/* Mobile Menu Button */}
        {onSidebarToggle && (
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={onSidebarToggle}
          >
            <Menu className="h-5 w-5" />
          </Button>
        )}
        
        {/* Search Bar */}
        <div className="flex-1 max-w-xl">
          <Button 
            variant="outline" 
            className="h-9 w-full max-w-md justify-start gap-3 bg-muted/30 hover:bg-muted/50 border-muted-foreground/10"
            onClick={() => {/* Open command palette */}}
          >
            <Search className="h-4 w-4 text-muted-foreground" />
            <span className="flex-1 text-left text-sm text-muted-foreground font-normal">
              Search or run commands...
            </span>
            <div className="hidden sm:flex items-center gap-1">
              <kbd className="h-5 px-1.5 text-[10px] font-mono bg-background border rounded shadow-sm">
                ⌘
              </kbd>
              <kbd className="h-5 px-1.5 text-[10px] font-mono bg-background border rounded shadow-sm">
                K
              </kbd>
            </div>
          </Button>
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-2">
          {/* Quick Actions */}
          <Button
            variant="ghost"
            size="icon"
            className="relative h-9 w-9"
          >
            <Terminal className="h-4 w-4" />
          </Button>

          {/* AI Assistant */}
          <Button
            variant="ghost"
            size="sm"
            className="hidden lg:flex items-center gap-2 h-9 px-3"
          >
            <Sparkles className="h-4 w-4" />
            <span className="text-sm">AI</span>
          </Button>

          {/* Notifications */}
          <Button variant="ghost" size="icon" className="relative h-9 w-9">
            <Bell className="h-4 w-4" />
            <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-destructive" />
          </Button>

          {/* Help */}
          <Button variant="ghost" size="icon" className="h-9 w-9">
            <HelpCircle className="h-4 w-4" />
          </Button>

          {/* Theme Toggle */}
          <ThemeToggleClient />

          {/* Divider */}
          <div className="mx-2 h-6 w-px bg-border" />

          {/* User Menu */}
          <UserProfileDropdown />
        </div>
      </div>
      </div>
    </header>
  )
}