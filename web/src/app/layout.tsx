import type { Metadata } from 'next'
import { IBM_Plex_Mono } from 'next/font/google'
import { cn } from '@/lib/utils'
import { ToastProvider } from '@/components/ui/toast-provider'
import { getTheme } from '@/lib/theme-cookie'
import { ThemeScript } from './theme-script'
import './globals.css'
import { Toaster } from '@/components/ui/toaster'

const ibmPlexMono = IBM_Plex_Mono({ 
  weight: ['400', '500', '600', '700'],
  subsets: ['latin'],
  variable: '--font-mono'
})

export const metadata: Metadata = {
  title: 'Kloudlite - Cloud-Native Development Made Simple',
  description: 'Fast, reliable, and consistent development environments. Eliminate local setup complexity and enable seamless collaboration through service intercepts.',
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const theme = await getTheme()
  
  return (
    <html lang="en" className={theme} suppressHydrationWarning>
      <head>
        <ThemeScript theme={theme} />
      </head>
      <body className={cn(ibmPlexMono.className)} suppressHydrationWarning>
        <ToastProvider>
          {children}
          <Toaster />
        </ToastProvider>
        <div id="radix-portal-root" tabIndex={-1} aria-hidden="true" />
      </body>
    </html>
  )
}