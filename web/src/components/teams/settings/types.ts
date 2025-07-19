import { ReactNode } from 'react'

export interface TeamSettingsHeaderProps {
  teamname: string
  teamDisplayName?: string
}

export interface TeamSettingsLayoutProps {
  children: ReactNode
  teamname: string
  teamDisplayName?: string
  className?: string
}