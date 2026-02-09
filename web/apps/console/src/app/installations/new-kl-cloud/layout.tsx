'use client'

import { ReactNode } from 'react'
import { InstallationLayoutShell } from '@/components/installation-layout'

export default function NewKlCloudLayout({ children }: { children: ReactNode }) {
  return <InstallationLayoutShell>{children}</InstallationLayoutShell>
}
