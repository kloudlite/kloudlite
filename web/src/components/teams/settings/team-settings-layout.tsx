import { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'
import { TeamSettingsHeader } from './team-settings-header'

interface TeamSettingsLayoutProps {
  children: ReactNode
  teamname: string
  teamDisplayName?: string
  className?: string
}

export function TeamSettingsLayout({
  children,
  teamname,
  teamDisplayName,
  className = ''
}: TeamSettingsLayoutProps) {
  return (
    <div className="h-full flex flex-col">
      {/* Header with breadcrumbs and tabs - Full width */}
      <TeamSettingsHeader 
        teamname={teamname}
        teamDisplayName={teamDisplayName}
      />
      
      {/* Content - Responsive Container with flex-1 to take remaining space */}
      <div className={cn("flex-1 overflow-y-auto", LAYOUT.CONTAINER, LAYOUT.PADDING.PAGE, className)}>
        {children}
      </div>
    </div>
  )
}