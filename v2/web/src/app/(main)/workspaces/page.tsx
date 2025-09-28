import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { WorkspacesList } from '@/components/workspaces-list'

export default async function WorkspacesPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'

  // For demo, assume admin if email ends with @kloudlite.io
  const isAdmin = currentUser.endsWith('@kloudlite.io')

  // Mock data for workspaces - you can add more to test pagination
  const workspaces = [
    {
      id: '1',
      name: 'web-app',
      description: 'Frontend application workspace',
      status: 'active' as const,
      lastActivity: '2 mins ago',
      branch: 'main',
      team: 3,
      environment: 'my-dev-env',
      language: 'TypeScript',
      framework: 'Next.js'
    },
    {
      id: '2',
      name: 'api-server',
      description: 'Backend API service workspace',
      status: 'active' as const,
      lastActivity: '15 mins ago',
      branch: 'develop',
      team: 2,
      environment: 'my-dev-env',
      language: 'Go',
      framework: 'Gin'
    },
    {
      id: '3',
      name: 'frontend',
      description: 'Main frontend workspace for auth feature',
      status: 'active' as const,
      lastActivity: '1 hour ago',
      branch: 'feature/auth',
      team: 4,
      environment: 'feature-auth',
      language: 'React',
      framework: 'Vite'
    },
    {
      id: '4',
      name: 'backend',
      description: 'Backend services for auth feature',
      status: 'active' as const,
      lastActivity: '2 hours ago',
      branch: 'feature/auth',
      team: 3,
      environment: 'feature-auth',
      language: 'Node.js',
      framework: 'Express'
    },
    {
      id: '5',
      name: 'worker',
      description: 'Background job processing workspace',
      status: 'idle' as const,
      lastActivity: '1 day ago',
      branch: 'main',
      team: 1,
      environment: 'feature-auth',
      language: 'Python',
      framework: 'Celery'
    },
    {
      id: '6',
      name: 'api',
      description: 'API service for staging tests',
      status: 'active' as const,
      lastActivity: '3 hours ago',
      branch: 'staging',
      team: 2,
      environment: 'staging-test',
      language: 'Java',
      framework: 'Spring Boot'
    },
    {
      id: '7',
      name: 'mobile-app',
      description: 'React Native mobile application',
      status: 'active' as const,
      lastActivity: '45 mins ago',
      branch: 'develop',
      team: 5,
      environment: 'mobile-dev',
      language: 'TypeScript',
      framework: 'React Native'
    },
    {
      id: '8',
      name: 'analytics',
      description: 'Data analytics service',
      status: 'idle' as const,
      lastActivity: '2 days ago',
      branch: 'main',
      team: 2,
      environment: 'analytics-env',
      language: 'Python',
      framework: 'FastAPI'
    },
    {
      id: '9',
      name: 'ml-service',
      description: 'Machine learning inference service',
      status: 'active' as const,
      lastActivity: '6 hours ago',
      branch: 'feature/model-v2',
      team: 3,
      environment: 'ml-dev',
      language: 'Python',
      framework: 'TensorFlow'
    },
    {
      id: '10',
      name: 'admin-portal',
      description: 'Admin dashboard and management portal',
      status: 'active' as const,
      lastActivity: '30 mins ago',
      branch: 'main',
      team: 4,
      environment: 'admin-env',
      language: 'Vue.js',
      framework: 'Nuxt.js'
    },
    {
      id: '11',
      name: 'payment-gateway',
      description: 'Payment processing and integration service',
      status: 'active' as const,
      lastActivity: '10 mins ago',
      branch: 'feature/stripe-v3',
      team: 3,
      environment: 'feature-billing',
      language: 'Node.js',
      framework: 'NestJS'
    },
    {
      id: '12',
      name: 'notification-hub',
      description: 'Centralized notification service',
      status: 'idle' as const,
      lastActivity: '3 days ago',
      branch: 'main',
      team: 2,
      environment: 'feature-notifications',
      language: 'Go',
      framework: 'Echo'
    },
    {
      id: '13',
      name: 'file-storage',
      description: 'S3-compatible file storage service',
      status: 'active' as const,
      lastActivity: '1 hour ago',
      branch: 'develop',
      team: 4,
      environment: 'my-dev-env',
      language: 'Rust',
      framework: 'Actix'
    },
    {
      id: '14',
      name: 'search-engine',
      description: 'Elasticsearch-based search service',
      status: 'active' as const,
      lastActivity: '25 mins ago',
      branch: 'feature/fuzzy-search',
      team: 5,
      environment: 'feature-search',
      language: 'Java',
      framework: 'Spring Boot'
    },
    {
      id: '15',
      name: 'cache-layer',
      description: 'Redis-based caching service',
      status: 'active' as const,
      lastActivity: '5 mins ago',
      branch: 'main',
      team: 2,
      environment: 'my-dev-env',
      language: 'Go',
      framework: 'Fiber'
    },
    {
      id: '16',
      name: 'auth-service',
      description: 'OAuth2 and JWT authentication service',
      status: 'active' as const,
      lastActivity: '40 mins ago',
      branch: 'feature/2fa',
      team: 6,
      environment: 'feature-auth',
      language: 'Python',
      framework: 'Django'
    },
    {
      id: '17',
      name: 'email-service',
      description: 'Email delivery and template management',
      status: 'idle' as const,
      lastActivity: '1 day ago',
      branch: 'main',
      team: 3,
      environment: 'feature-notifications',
      language: 'Node.js',
      framework: 'Express'
    },
    {
      id: '18',
      name: 'reporting-engine',
      description: 'Business intelligence and reporting',
      status: 'active' as const,
      lastActivity: '2 hours ago',
      branch: 'feature/quarterly-reports',
      team: 4,
      environment: 'analytics-dev',
      language: 'Python',
      framework: 'Flask'
    },
    {
      id: '19',
      name: 'webhook-processor',
      description: 'Webhook processing and delivery service',
      status: 'active' as const,
      lastActivity: '15 mins ago',
      branch: 'develop',
      team: 3,
      environment: 'integration-tests',
      language: 'Go',
      framework: 'Gin'
    },
    {
      id: '20',
      name: 'graphql-gateway',
      description: 'GraphQL API gateway and federation',
      status: 'active' as const,
      lastActivity: '50 mins ago',
      branch: 'feature/federation-v2',
      team: 5,
      environment: 'my-dev-env',
      language: 'TypeScript',
      framework: 'Apollo Server'
    },
    {
      id: '21',
      name: 'monitoring-agent',
      description: 'Application monitoring and metrics collection',
      status: 'active' as const,
      lastActivity: '20 mins ago',
      branch: 'main',
      team: 2,
      environment: 'monitoring-stack',
      language: 'Go',
      framework: 'Native'
    },
    {
      id: '22',
      name: 'cdn-manager',
      description: 'Content delivery network management',
      status: 'idle' as const,
      lastActivity: '4 days ago',
      branch: 'main',
      team: 3,
      environment: 'prod-replica',
      language: 'Node.js',
      framework: 'Fastify'
    },
    {
      id: '23',
      name: 'backup-service',
      description: 'Automated backup and disaster recovery',
      status: 'active' as const,
      lastActivity: '1 hour ago',
      branch: 'feature/incremental-backup',
      team: 4,
      environment: 'security-audit',
      language: 'Python',
      framework: 'AsyncIO'
    },
    {
      id: '24',
      name: 'rate-limiter',
      description: 'API rate limiting and throttling service',
      status: 'active' as const,
      lastActivity: '35 mins ago',
      branch: 'develop',
      team: 3,
      environment: 'my-dev-env',
      language: 'Rust',
      framework: 'Tokio'
    },
    {
      id: '25',
      name: 'queue-manager',
      description: 'Message queue and task management',
      status: 'active' as const,
      lastActivity: '8 mins ago',
      branch: 'main',
      team: 2,
      environment: 'feature-chat',
      language: 'Java',
      framework: 'RabbitMQ'
    },
  ]


  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Title and Filter Section */}
      <div className="mb-8">
        <div className="mb-6">
          <h1 className="text-3xl font-light tracking-tight">Workspaces</h1>
          <p className="text-sm text-gray-600 mt-2">
            Manage your development workspaces and collaborate with your team
          </p>
        </div>

        {/* Workspaces List with Filter */}
        <WorkspacesList
          workspaces={workspaces}
          currentUser={currentUser}
          isAdmin={isAdmin}
        />
      </div>
    </main>
  )
}