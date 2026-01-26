import type { Metadata } from 'next'
import { IBM_Plex_Mono, Open_Sans } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from '@kloudlite/ui'
import { getTheme } from '@/lib/theme-server'

const ibmPlexMono = IBM_Plex_Mono({
  weight: ['400', '500', '600', '700'],
  subsets: ['latin'],
  variable: '--font-mono',
})

const openSans = Open_Sans({
  weight: ['400', '500', '600', '700'],
  subsets: ['latin'],
  variable: '--font-sans',
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
      <body className={`${openSans.variable} ${ibmPlexMono.variable} font-sans`}>
        <Providers>{children}</Providers>
        <Toaster />
      </body>
    </html>
  )
}
