"use client"

import { KloudliteLogo } from "@/components/kloudlite-logo"
import { NotificationBell } from "@/components/notification-bell"
import { UserMenu } from "@/components/user-menu"

interface OverviewHeaderProps {
  user: {
    id?: string
    email?: string | null
    name?: string | null
    image?: string | null
  } | undefined
  canManagePlatform?: boolean
}

export function OverviewHeader({ user, canManagePlatform = false }: OverviewHeaderProps) {

  return (
    <header className="sticky-header">
      <div className="container-responsive">
        <div className="flex h-14 items-center gap-4 md:h-16">
          <div className="flex items-center gap-4 md:gap-8">
            {/* Logo */}
            <KloudliteLogo className="h-5 md:h-auto" />
          </div>
          
          {/* Right side - Notifications and User Menu */}
          <div className="flex items-center gap-4 ml-auto">
            {/* Notification Bell */}
            <NotificationBell />
            
            <UserMenu user={user} canManagePlatform={canManagePlatform} />
          </div>
        </div>
      </div>
    </header>
  )
}