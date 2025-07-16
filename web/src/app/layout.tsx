import type { Metadata } from 'next'
import { Roboto_Mono } from 'next/font/google'
import { cn } from '@/lib/utils'
import { ToastProvider } from '@/components/ui/toast-provider'
import { getTheme } from '@/lib/theme-cookie'
import { ThemeScript } from './theme-script'
import { AuthSessionProvider } from '@/components/providers/session-provider'
import './globals.css'
import { Toaster } from '@/components/ui/toaster'

const robotoMono = Roboto_Mono({ 
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
      <body className={cn(robotoMono.variable, robotoMono.className)} suppressHydrationWarning>
        <AuthSessionProvider>
          <ToastProvider>
            {children}
            <Toaster />
          </ToastProvider>
        </AuthSessionProvider>
        <div id="radix-portal-root" tabIndex={-1} aria-hidden="true" />
      </body>
    </html>
  )
}