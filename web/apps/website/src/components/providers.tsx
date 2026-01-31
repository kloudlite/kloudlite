'use client'

import { ThemeProvider } from 'next-themes'
import { Toaster } from '@kloudlite/ui'

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem={true}
      enableColorScheme
      storageKey="theme"
    >
      {children}
      <Toaster />
    </ThemeProvider>
  )
}
