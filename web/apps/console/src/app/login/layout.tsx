import { Providers } from '@/components/providers'

export default function RegisterLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return <Providers>{children}</Providers>
}
