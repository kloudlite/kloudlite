import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { Breadcrumb } from '@/components/breadcrumb'
import { Button } from '@/components/ui/button'
import {
  Terminal,
  Code2,
  Globe,
  ExternalLink,
  Copy,
  Check
} from 'lucide-react'
import { WorkspaceConnectOptions } from './_components/workspace-connect-options'

interface PageProps {
  params: {
    id: string
  }
}

export default async function WorkspaceDetailPage({ params }: PageProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  // Mock workspace data - in real app, fetch by ID
  const workspace = {
    id: params.id,
    name: params.id === '1' ? 'web-app' : 'api-server',
    description: params.id === '1'
      ? 'Frontend application workspace'
      : 'Backend API service workspace',
    owner: session.user?.email || 'user@example.com',
    status: 'active' as const,
    environment: 'my-dev-env',
    branch: 'main',
    language: params.id === '1' ? 'TypeScript' : 'Go',
    framework: params.id === '1' ? 'Next.js' : 'Gin',
    lastActivity: '5 mins ago',
    resources: {
      cpu: '2 cores',
      memory: '4GB',
      storage: '20GB'
    }
  }

  const breadcrumbItems = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: workspace.name }
  ]

  return (
    <>
      {/* Workspace Header with Info */}
      <div className="bg-white border-b">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb */}
          <div className="py-4">
            <Breadcrumb items={breadcrumbItems} />
          </div>

          {/* Workspace Header */}
          <div className="pb-6">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-semibold text-gray-900">{workspace.name}</h1>
                <p className="text-sm text-gray-600 mt-1.5">{workspace.description}</p>
                <div className="mt-3 flex items-center gap-4 text-sm text-gray-600">
                  <span>Environment: {workspace.environment}</span>
                  <span>•</span>
                  <span>Branch: {workspace.branch}</span>
                  <span>•</span>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    workspace.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {workspace.status}
                  </span>
                </div>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" size="sm">
                  Stop Workspace
                </Button>
                <Button variant="outline" size="sm">
                  Settings
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Connection Options - Takes 2/3 width */}
          <div className="lg:col-span-2">
            <WorkspaceConnectOptions workspaceId={workspace.id} />
          </div>

          {/* Workspace Details - Takes 1/3 width */}
          <div className="space-y-6">
            {/* Stack Info */}
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <h3 className="text-sm font-medium text-gray-900 mb-4">Stack Information</h3>
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Language</span>
                  <span className="text-gray-900">{workspace.language}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Framework</span>
                  <span className="text-gray-900">{workspace.framework}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Branch</span>
                  <span className="text-gray-900 font-mono">{workspace.branch}</span>
                </div>
              </div>
            </div>

            {/* Resources */}
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <h3 className="text-sm font-medium text-gray-900 mb-4">Resources</h3>
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">CPU</span>
                  <span className="text-gray-900">{workspace.resources.cpu}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Memory</span>
                  <span className="text-gray-900">{workspace.resources.memory}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Storage</span>
                  <span className="text-gray-900">{workspace.resources.storage}</span>
                </div>
              </div>
            </div>

            {/* Activity */}
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <h3 className="text-sm font-medium text-gray-900 mb-4">Activity</h3>
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Last Activity</span>
                  <span className="text-gray-900">{workspace.lastActivity}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Owner</span>
                  <span className="text-gray-900">{workspace.owner.split('@')[0]}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}