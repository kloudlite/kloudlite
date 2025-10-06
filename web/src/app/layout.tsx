import type { Metadata } from 'next'
import { IBM_Plex_Mono } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from '@/components/ui/sonner'
import { getTheme } from '@/lib/theme-server'

const ibmPlexMono = IBM_Plex_Mono({
  weight: ['400', '500', '600', '700'],
  subsets: ['latin']
})

export const metadata: Metadata = {
  title: 'Kloudlite',
  description: 'Development environments platform',
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const theme = await getTheme()

  return (
    <html lang="en" className={theme === 'dark' ? 'dark' : ''}>
      <body className={ibmPlexMono.className}>
        <Providers>{children}</Providers>
        <Toaster />
      </body>
    </html>
  )
}