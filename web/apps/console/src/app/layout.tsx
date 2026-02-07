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
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* Handle theme before React hydration to prevent flash */}
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                const theme = '${theme}';
                const html = document.documentElement;

                if (theme === 'dark') {
                  html.classList.remove('light');
                  html.classList.add('dark');
                } else if (theme === 'system') {
                  const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
                  html.classList.remove('light', 'dark');
                  html.classList.add(isDark ? 'dark' : 'light');

                  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
                    html.classList.remove('light', 'dark');
                    html.classList.add(e.matches ? 'dark' : 'light');
                  });
                } else {
                  html.classList.remove('dark');
                  html.classList.add('light');
                }
              })();
            `,
          }}
        />
      </head>
      <body className={`${openSans.variable} ${ibmPlexMono.variable} font-sans h-screen overflow-hidden`}>
        <Providers>{children}</Providers>
        <Toaster position="bottom-right" richColors />
      </body>
    </html>
  )
}
