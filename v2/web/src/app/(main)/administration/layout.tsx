import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'

export default async function AdministrationLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const session = await auth()

  // Only allow super_admin and admin users
  if (!session?.user?.platformRole || !['super_admin', 'admin'].includes(session.user.platformRole)) {
    redirect('/')
  }

  return <>{children}</>
}