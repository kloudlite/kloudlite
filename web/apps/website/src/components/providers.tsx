'use client'

import { Toaster } from '@kloudlite/ui'

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <>
      {children}
      <Toaster />
    </>
  )
}
