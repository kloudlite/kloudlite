import { AuthSessionProvider } from '@/components/console/session-provider'

export default function RegisterLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return <AuthSessionProvider>{children}</AuthSessionProvider>
}
