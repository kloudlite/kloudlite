import { AdminNavigation } from '@/components/admin-navigation'

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // TODO: Add authentication and role checking when backend is ready

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