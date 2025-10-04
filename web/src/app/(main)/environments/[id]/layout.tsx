import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { Breadcrumb } from '@/components/breadcrumb'
import { EnvironmentNav } from '../_components/environment-nav'

interface LayoutProps {
  children: React.ReactNode
  params: {
    id: string
  }
}

export default async function EnvironmentLayout({ children, params }: LayoutProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  // Mock environment data - in real app, fetch by ID
  const environment = {
    id: params.id,
    name: params.id === '1' ? 'my-dev-env' : 'feature-auth',
    owner: session.user?.email || 'user@example.com',
    status: 'active' as const,
    created: '2 days ago',
  }

  const breadcrumbItems = [
    { label: 'Environments', href: '/environments' },
    { label: environment.name }
  ]

  return (
    <>
      {/* Environment Header with Info */}
      <div className="bg-white">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb */}
          <div className="py-4">
            <Breadcrumb items={breadcrumbItems} />
          </div>

          {/* Environment Header */}
          <div className="pb-4">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-semibold text-gray-900">{environment.name}</h1>
                <div className="mt-1.5 flex items-center gap-4 text-sm text-gray-600">
                  <span>Owner: {environment.owner}</span>
                  <span>•</span>
                  <span>Created: {environment.created}</span>
                  <span>•</span>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    environment.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {environment.status}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <EnvironmentNav environmentId={params.id} />

      {/* Page Content */}
      <div className="flex-1">
        {children}
      </div>
    </>
  )
}
