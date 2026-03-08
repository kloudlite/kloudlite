'use client'

import { CreditCard } from 'lucide-react'
import { NavTabs, type NavTab } from './nav-tabs'

interface InstallationDetailsTabsProps {
  installationId: string
  cloudProvider?: string
}

export function InstallationDetailsTabs({ installationId, cloudProvider }: InstallationDetailsTabsProps) {
  const tabs: NavTab[] = [
    {
      id: 'overview',
      label: 'Overview',
      href: `/installations/${installationId}`,
    },
  ]

  if (cloudProvider === 'oci') {
    tabs.push({
      id: 'billing',
      label: 'Billing',
      icon: CreditCard,
      href: `/installations/${installationId}/billing`,
    })
  }

  return <NavTabs tabs={tabs} />
}
