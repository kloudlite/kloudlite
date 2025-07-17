import { Metadata } from 'next'
import { DocsLayout } from '@/components/docs/docs-layout'

export const metadata: Metadata = {
  title: 'Documentation - Kloudlite',
  description: 'Kloudlite platform documentation and guides',
}

export default function DocsRootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <DocsLayout>
      {children}
    </DocsLayout>
  )
}