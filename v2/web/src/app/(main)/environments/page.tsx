import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { EnvironmentsList } from '@/components/environments-list'
import { environmentService } from '@/lib/services/environment.service'
import { environmentToUIModel } from '@/types/environment'

export default async function EnvironmentsPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'test-user'

  // Fetch real environments from API
  let allEnvironments = []
  try {
    const response = await environmentService.listEnvironments(currentUser)
    allEnvironments = response.environments.map(env => {
      let owner = env.spec.labels?.['kloudlite.io/owned-by'] || 'unknown'

      // Try to get and decode the email if available
      const encodedEmail = env.spec.labels?.['kloudlite.io/owner-email']
      if (encodedEmail) {
        try {
          // Decode base64 URL-encoded email
          owner = Buffer.from(encodedEmail, 'base64').toString('utf-8')
        } catch (e) {
          // Fall back to username if decoding fails
          owner = env.spec.labels?.['kloudlite.io/owned-by'] || 'unknown'
        }
      }

      return environmentToUIModel(env, owner)
    })
  } catch (error) {
    console.error('Failed to fetch environments:', error)
    // Use empty array on error
    allEnvironments = []
  }

  // Commented out mock data - keeping for reference
  /*
  const allEnvironments = [
    {
      id: '1',
      name: 'my-dev-env',
      owner: currentUser,
      status: 'active' as const,
      created: '2 days ago',
      services: 3,
      configs: 5,
      secrets: 8,
      workspaces: ['web-app', 'api-server'],
      lastDeployed: '5 mins ago'
    },
    {
      id: '2',
      name: 'feature-auth',
      owner: currentUser,
      status: 'active' as const,
      created: '1 week ago',
      services: 4,
      configs: 7,
      secrets: 12,
      workspaces: ['frontend', 'backend', 'worker'],
      lastDeployed: '2 hours ago'
    },
    {
      id: '3',
      name: 'test-deployment',
      owner: 'alice@team.com',
      status: 'inactive' as const,
      created: '2 weeks ago',
      services: 0,
      configs: 3,
      secrets: 4,
      workspaces: [],
      lastDeployed: 'Never'
    },
    {
      id: '4',
      name: 'staging-test',
      owner: currentUser,
      status: 'active' as const,
      created: '3 weeks ago',
      services: 2,
      configs: 4,
      secrets: 6,
      workspaces: ['api'],
      lastDeployed: '1 day ago'
    },
    {
      id: '5',
      name: 'prod-replica',
      owner: 'bob@team.com',
      status: 'active' as const,
      created: '4 days ago',
      services: 5,
      configs: 8,
      secrets: 10,
      workspaces: ['microservice-a', 'microservice-b'],
      lastDeployed: '30 mins ago'
    },
    {
      id: '6',
      name: 'feature-payments',
      owner: 'charlie@team.com',
      status: 'active' as const,
      created: '5 days ago',
      services: 3,
      configs: 6,
      secrets: 7,
      workspaces: ['payments-api', 'payments-ui'],
      lastDeployed: '3 hours ago'
    },
    {
      id: '7',
      name: 'feature-search',
      owner: currentUser,
      status: 'inactive' as const,
      created: '1 month ago',
      services: 2,
      configs: 3,
      secrets: 5,
      workspaces: ['search-service'],
      lastDeployed: '2 weeks ago'
    },
    {
      id: '8',
      name: 'ml-training',
      owner: 'david@team.com',
      status: 'active' as const,
      created: '6 days ago',
      services: 3,
      configs: 8,
      secrets: 6,
      workspaces: ['ml-pipeline', 'data-processor'],
      lastDeployed: '1 hour ago'
    },
    {
      id: '9',
      name: 'mobile-backend',
      owner: currentUser,
      status: 'active' as const,
      created: '2 weeks ago',
      services: 5,
      configs: 10,
      secrets: 15,
      workspaces: ['mobile-api', 'push-service', 'sync-worker'],
      lastDeployed: '20 mins ago'
    },
    {
      id: '10',
      name: 'analytics-dev',
      owner: 'eve@team.com',
      status: 'active' as const,
      created: '10 days ago',
      services: 4,
      configs: 7,
      secrets: 9,
      workspaces: ['analytics-api', 'report-generator'],
      lastDeployed: '4 hours ago'
    },
    {
      id: '11',
      name: 'feature-notifications',
      owner: currentUser,
      status: 'active' as const,
      created: '3 days ago',
      services: 2,
      configs: 4,
      secrets: 5,
      workspaces: ['notification-service', 'email-worker'],
      lastDeployed: '45 mins ago'
    },
    {
      id: '12',
      name: 'performance-testing',
      owner: 'frank@team.com',
      status: 'inactive' as const,
      created: '2 months ago',
      services: 6,
      configs: 12,
      secrets: 8,
      workspaces: ['load-tester', 'metrics-collector'],
      lastDeployed: '3 weeks ago'
    },
    {
      id: '13',
      name: 'feature-dashboard',
      owner: currentUser,
      status: 'active' as const,
      created: '1 week ago',
      services: 3,
      configs: 6,
      secrets: 7,
      workspaces: ['dashboard-ui', 'dashboard-api'],
      lastDeployed: '10 mins ago'
    },
    {
      id: '14',
      name: 'integration-tests',
      owner: 'grace@team.com',
      status: 'active' as const,
      created: '4 days ago',
      services: 4,
      configs: 8,
      secrets: 10,
      workspaces: ['test-runner', 'test-db'],
      lastDeployed: '2 hours ago'
    },
    {
      id: '15',
      name: 'feature-export',
      owner: currentUser,
      status: 'inactive' as const,
      created: '3 weeks ago',
      services: 2,
      configs: 3,
      secrets: 4,
      workspaces: ['export-service'],
      lastDeployed: '5 days ago'
    },
    {
      id: '16',
      name: 'security-audit',
      owner: 'henry@team.com',
      status: 'active' as const,
      created: '5 days ago',
      services: 3,
      configs: 7,
      secrets: 12,
      workspaces: ['security-scanner', 'vuln-db'],
      lastDeployed: '30 mins ago'
    },
    {
      id: '17',
      name: 'feature-chat',
      owner: currentUser,
      status: 'active' as const,
      created: '2 weeks ago',
      services: 4,
      configs: 9,
      secrets: 11,
      workspaces: ['chat-server', 'websocket-gateway', 'chat-ui'],
      lastDeployed: '15 mins ago'
    },
    {
      id: '18',
      name: 'data-migration',
      owner: 'iris@team.com',
      status: 'inactive' as const,
      created: '1 month ago',
      services: 2,
      configs: 5,
      secrets: 6,
      workspaces: ['migration-tool'],
      lastDeployed: '2 weeks ago'
    },
    {
      id: '19',
      name: 'feature-billing',
      owner: currentUser,
      status: 'active' as const,
      created: '6 days ago',
      services: 5,
      configs: 10,
      secrets: 14,
      workspaces: ['billing-api', 'payment-processor', 'invoice-generator'],
      lastDeployed: '1 hour ago'
    },
    {
      id: '20',
      name: 'monitoring-stack',
      owner: 'jack@team.com',
      status: 'active' as const,
      created: '2 weeks ago',
      services: 6,
      configs: 15,
      secrets: 8,
      workspaces: ['prometheus', 'grafana', 'alertmanager'],
      lastDeployed: '3 hours ago'
    },
  ]
  */

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
        {/* Title and Filter Section */}
        <div className="mb-8">
          <div className="mb-6">
            <h1 className="text-3xl font-light tracking-tight">Environments</h1>
            <p className="text-sm text-gray-600 mt-2">
              Manage development environments across your team
            </p>
          </div>

          {/* Environments List with Filter */}
          <EnvironmentsList
            environments={allEnvironments}
            currentUser={currentUser}
          />
        </div>
    </main>
  )
}