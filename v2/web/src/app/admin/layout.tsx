import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { AdminNavigation } from '@/components/admin-navigation'

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const session = await auth()

  // For now, just check if user is logged in
  // TODO: Add proper role checking when available
  if (!session) {
    redirect('/auth/signin')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Admin-specific navigation */}
      <AdminNavigation />

      {/* Main Content */}
      <main className="mx-auto max-w-7xl px-6 py-8">
        {children}
      </main>
    </div>
  )
}