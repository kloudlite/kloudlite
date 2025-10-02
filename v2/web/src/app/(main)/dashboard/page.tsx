import { redirect } from 'next/navigation'
import Link from 'next/link'
import { auth } from '@/lib/auth'
import { Button } from '@/components/ui/button'

export default async function Dashboard() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  // Mock overview data
  const stats = {
    environments: 3,
    activeEnvironments: 2,
    workspaces: 4,
    runningWorkspaces: 3,
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
        {/* Page Header */}
        <div className="py-8 border-b bg-white">
          <h1 className="text-2xl font-semibold text-gray-900">Welcome back</h1>
          <p className="text-sm text-gray-600 mt-1.5">{session.user?.email}</p>
        </div>

        {/* Stats Section */}
        <div className="py-8 border-b bg-white">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            <div>
              <p className="text-3xl font-semibold text-gray-900">{stats.environments}</p>
              <p className="text-sm text-gray-600 mt-1">Environments</p>
            </div>
            <div>
              <p className="text-3xl font-semibold text-gray-900">{stats.activeEnvironments}</p>
              <p className="text-sm text-gray-600 mt-1">Active</p>
            </div>
            <div>
              <p className="text-3xl font-semibold text-gray-900">{stats.workspaces}</p>
              <p className="text-sm text-gray-600 mt-1">Workspaces</p>
            </div>
            <div>
              <p className="text-3xl font-semibold text-gray-900">{stats.runningWorkspaces}</p>
              <p className="text-sm text-gray-600 mt-1">Running</p>
            </div>
          </div>
        </div>

        {/* Content Section */}
        <div className="py-8 bg-white">
          <h2 className="text-base font-semibold text-gray-900 mb-6">Resources</h2>
          <div className="grid md:grid-cols-2 gap-4">
            <Link href="/environments" className="group block">
              <div className="p-6 border rounded-lg hover:border-gray-400 transition-colors">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-base font-semibold text-gray-900">Environments</h3>
                  <span className="text-sm text-gray-400 group-hover:text-gray-600">→</span>
                </div>
                <p className="text-sm text-gray-600">Create and manage your development environments</p>
                <p className="text-sm text-gray-400 mt-4">{stats.environments} total</p>
              </div>
            </Link>

            <Link href="/workspaces" className="group block">
              <div className="p-6 border rounded-lg hover:border-gray-400 transition-colors">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-base font-semibold text-gray-900">Workspaces</h3>
                  <span className="text-sm text-gray-400 group-hover:text-gray-600">→</span>
                </div>
                <p className="text-sm text-gray-600">Manage your application workspaces</p>
                <p className="text-sm text-gray-400 mt-4">{stats.workspaces} total</p>
              </div>
            </Link>
          </div>

          {/* Quick Actions */}
          <div className="mt-8 pt-8 border-t">
            <h3 className="text-sm font-medium text-gray-600 mb-4">Quick Actions</h3>
            <div className="flex gap-2">
              <Link href="/environments">
                <Button size="sm">Create Environment</Button>
              </Link>
              <Link href="/workspaces">
                <Button variant="outline" size="sm">Add Workspace</Button>
              </Link>
            </div>
          </div>
        </div>
      </main>
  )
}