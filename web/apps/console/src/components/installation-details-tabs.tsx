'use client'

import { CreditCard, Users } from 'lucide-react'
import { NavTabs, type NavTab } from './nav-tabs'

interface InstallationDetailsTabsProps {
  installationId: string
}

export function InstallationDetailsTabs({ installationId }: InstallationDetailsTabsProps) {
  const tabs: NavTab[] = [
    {
      id: 'overview',
      label: 'Overview',
      href: `/installations/${installationId}`,
    },
    {
      id: 'team',
      label: 'Team',
      icon: Users,
      href: `/installations/${installationId}/team`,
    },
    {
      id: 'billing',
      label: 'Billing',
      icon: CreditCard,
      href: `/installations/${installationId}/billing`,
      matchPrefix: true,
    },
  ]

  return <NavTabs tabs={tabs} />
}
