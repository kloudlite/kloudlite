import type { Metadata } from 'next'
import { IBM_Plex_Mono, Open_Sans } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from 'sonner'
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
  title: 'kloudlite / console',
  description: 'Development environments platform',
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const theme = await getTheme()

  return (
    <html lang="en" className={theme === 'dark' ? 'dark' : ''} suppressHydrationWarning>
      <head>
        {/* Only apply system preference on client if theme is 'system' */}
        {theme === 'system' && (
          <script
            dangerouslySetInnerHTML={{
              __html: `
                if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
                  document.documentElement.classList.add('dark');
                }
              `,
            }}
          />
        )}
      </head>
      <body className={`${openSans.variable} ${ibmPlexMono.variable} font-sans h-screen overflow-hidden`}>
        <Providers>{children}</Providers>
        <Toaster position="bottom-right" richColors />
      </body>
    </html>
  )
}
