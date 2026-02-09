'use client'

import { ReactNode } from 'react'
import { InstallationLayoutShell } from '@/components/installation-layout'

export default function NewByocLayout({ children }: { children: ReactNode }) {
  return <InstallationLayoutShell>{children}</InstallationLayoutShell>
}
