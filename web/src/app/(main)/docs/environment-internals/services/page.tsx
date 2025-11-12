import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Database,
  Container,
  Network,
  FileText,
  Settings,
  CheckCircle2,
  Info,
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'docker-compose', title: 'Docker Compose' },
  { id: 'service-types', title: 'Service Types' },
  { id: 'networking', title: 'Service Networking' },
  { id: 'configuration', title: 'Configuration' },
]

export default function ServicesPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">Services</h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Container className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Services are the applications and infrastructure components that run within your
          environments. They are defined using Docker Compose and can include databases, caches,
          APIs, message queues, and any other containerized services your application needs.
        </p>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Environment Services
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                Services run in the environment&apos;s namespace and are accessible to workspaces
                connected to that environment. Each service gets its own DNS name for easy discovery.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Docker Compose */}
      <section id="docker-compose" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <FileText className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Docker Compose Definition
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Services are defined using standard Docker Compose syntax. Kloudlite processes your
          compose file and deploys the services to the environment.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <p className="text-card-foreground text-sm mb-3 m-0 font-medium">
            Example: Database and Cache Services
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">
{`services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: \${POSTGRES_PASSWORD}
      POSTGRES_DB: myapp
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:`}
            </pre>
          </div>
        </div>
      </section>

      {/* Service Types */}
      <section id="service-types" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Database className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Common Service Types
          </h2>
        </div>

        <div className="grid gap-4 sm:gap-6 mb-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2">
                <Database className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Databases
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0 mb-3">
                  PostgreSQL, MySQL, MongoDB, and other relational or NoSQL databases
                </p>
                <ul className="space-y-1.5 text-xs text-muted-foreground">
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Persistent storage with volumes
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Environment variable configuration
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2">
                <Settings className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Caches & Message Queues
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0 mb-3">
                  Redis, Memcached, RabbitMQ, Kafka for caching and message processing
                </p>
                <ul className="space-y-1.5 text-xs text-muted-foreground">
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Fast data access
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Asynchronous processing
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2">
                <Container className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Application Services
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0 mb-3">
                  APIs, microservices, background workers running your application code
                </p>
                <ul className="space-y-1.5 text-xs text-muted-foreground">
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Multiple replicas support
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500" />
                    Environment-specific configuration
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Networking */}
      <section id="networking" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Service Networking
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Services within an environment can communicate with each other using their service names
          as DNS hostnames. This provides simple service discovery without hardcoding IPs.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Connecting to Services:
          </h4>
          <ul className="space-y-2 m-0">
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>From Workspaces:</strong> Connect using service name and port (e.g.,{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  postgres:5432
                </code>
                )
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Between Services:</strong> Services in the same environment can call each
                other by name
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Port Mapping:</strong> Expose specific ports for service communication
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Configuration */}
      <section id="configuration" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Settings className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Configuration</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Services can be configured using environment variables, volumes, and other Docker Compose
          features. Configuration can reference environment-level configs and secrets.
        </p>

        <div className="bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <CheckCircle2 className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Best Practice
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                Use environment variables for configuration and mount volumes for data persistence.
                Reference configs and secrets defined at the environment level for sensitive data.
              </p>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
