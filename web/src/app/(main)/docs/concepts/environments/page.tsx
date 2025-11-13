import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Database,
  Container,
  Settings,
  Network,
  Lock,
  Zap,
  CheckCircle2,
  Info,
  Users,
  FileText,
  Key,
} from 'lucide-react'

const tocItems = [
  { id: 'what-is-environment', title: 'What is an Environment?' },
  { id: 'key-features', title: 'Key Features' },
  { id: 'architecture', title: 'Environment Architecture' },
  { id: 'services', title: 'Services' },
  { id: 'configuration', title: 'Configuration & Secrets' },
  { id: 'networking', title: 'Networking' },
  { id: 'use-cases', title: 'Use Cases' },
]

export default function EnvironmentOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        Environment Overview
      </h1>

      {/* What is an Environment? */}
      <section id="what-is-environment" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Database className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            What is an Environment?
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          An environment is an isolated space where your application services run. Think of them as
          different stages like development, staging, or production environments. Each environment
          has its own set of services, configurations, and network isolation.
        </p>

        <div className="bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-950 dark:to-cyan-950 rounded-lg border-2 border-blue-300 dark:border-blue-700 p-4 sm:p-6 mb-6">
          <h4 className="text-blue-900 dark:text-blue-100 font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5" />
            Isolated Service Container
          </h4>
          <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed mb-3">
            Environments run on workmachines as isolated containers with their own network namespace.
            This provides:
          </p>
          <ul className="text-blue-800 dark:text-blue-200 text-sm space-y-1 m-0 list-disc list-inside">
            <li>Isolated services per environment</li>
            <li>Separate configuration for each stage</li>
            <li>Service discovery via DNS</li>
            <li>Network isolation and security</li>
          </ul>
        </div>

        <div className="grid gap-4 sm:gap-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Zap className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Quick Setup
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Define services using Docker Compose and deploy instantly. No complex
                  infrastructure setup required.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Lock className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Isolated & Secure
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Each environment runs in its own network namespace, preventing conflicts and
                  ensuring security between environments.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Key Features */}
      <section id="key-features" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <CheckCircle2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Key Features</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Container className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Docker Compose Services
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Define all your services using standard Docker Compose syntax. Databases, caches,
                  APIs, message queues - anything that runs in a container can be deployed to your
                  environment.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Settings className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Configuration Management
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Manage environment variables, configuration files, and secrets at the environment
                  level. All services can reference these configs, making it easy to update
                  configuration without modifying service definitions.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Network className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Service Discovery
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Services within an environment can communicate using simple service names as DNS
                  hostnames. Connect to your database using{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                    postgres:5432
                  </code>{' '}
                  instead of hardcoding IPs.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Users className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Workspace Connectivity
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Workspaces can connect to environments to access their services. When connected,
                  workspaces can access all environment services by name, making development and
                  testing seamless.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Architecture */}
      <section id="architecture" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Database className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Environment Architecture
          </h2>
        </div>

        {/* Environment Architecture Diagram */}
        <div className="bg-gradient-to-br from-slate-50 to-amber-50/30 dark:from-slate-900/50 dark:to-slate-800/50 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 sm:p-12 mb-8">
          <div className="max-w-4xl mx-auto">
            {/* Main Environment Container */}
            <div className="relative">
              <div className="absolute -inset-1 bg-gradient-to-r from-amber-600 to-orange-600 rounded-2xl blur opacity-25"></div>
              <div className="relative bg-white dark:bg-slate-950 rounded-2xl border-2 border-amber-500 shadow-xl p-8">
                <div className="absolute -top-4 left-6 bg-amber-600 text-white rounded-lg px-4 py-1.5 text-xs font-bold uppercase tracking-wide shadow-lg">
                  Environment Namespace
                </div>

                <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6 mt-2">
                  {/* Services */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <Container className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">
                      <Link href="/docs/environment-internals/services" className="hover:text-blue-600 dark:hover:text-blue-400">
                        Services
                      </Link>
                    </div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Docker Compose
                    </div>
                  </div>

                  {/* Configuration */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <Settings className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">
                      <Link href="/docs/environment-internals/configs-secrets" className="hover:text-purple-600 dark:hover:text-purple-400">
                        Configs & Secrets
                      </Link>
                    </div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Environment variables
                    </div>
                  </div>

                  {/* Network */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <Network className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">Network</div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Isolated namespace
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Learn More Links */}
            <div className="mt-6">
              <div className="flex flex-wrap gap-6 justify-center items-center text-sm">
                <Link href="/docs/environment-internals/services" className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 font-medium">
                  <Container className="h-4 w-4" />
                  Services
                </Link>
                <Link href="/docs/environment-internals/configs-secrets" className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 font-medium">
                  <Key className="h-4 w-4" />
                  Configs & Secrets
                </Link>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Environment Components
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Each environment runs on a workmachine with isolated networking:
          </p>
          <ul className="text-muted-foreground text-sm space-y-2">
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Services:</strong> Containerized applications defined via Docker Compose
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Configuration:</strong> Environment-level configs, secrets, and variables
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Networking:</strong> Isolated network namespace with DNS-based service discovery
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Storage:</strong> Persistent volumes for data storage
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Services */}
      <section id="services" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Container className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Services</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Services are the containerized applications that run in your environment. They are defined
          using standard Docker Compose syntax and can include databases, caches, APIs, and any other
          containerized services your application needs.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Common Service Types:
          </h4>
          <div className="grid sm:grid-cols-2 gap-4">
            <ul className="text-muted-foreground text-sm space-y-2 m-0">
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Databases (PostgreSQL, MongoDB, MySQL)
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Caches (Redis, Memcached)
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Message Queues (RabbitMQ, Kafka)
              </li>
            </ul>
            <ul className="text-muted-foreground text-sm space-y-2 m-0">
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Application APIs and services
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Background workers
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                Third-party tools
              </li>
            </ul>
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Learn More
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                See the{' '}
                <Link href="/docs/environment-internals/services" className="underline hover:no-underline">
                  Services documentation
                </Link>{' '}
                for detailed information on defining and managing services.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Configuration */}
      <section id="configuration" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Settings className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Configuration & Secrets
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Manage configuration and sensitive data at the environment level. All services in an
          environment can reference these configs and secrets, making it easy to update configuration
          without modifying service definitions.
        </p>

        <div className="grid gap-4 sm:gap-6 mb-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Key className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Environment Variables
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Key-value pairs for simple configuration like database URLs, API keys, and feature
                  flags.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <FileText className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Config Files
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Upload complete configuration files that services can mount and use.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Lock className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Secrets Management
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Securely store sensitive data with encryption at rest. Secrets are only decrypted
                  when injected into services.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Learn More
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                See the{' '}
                <Link href="/docs/environment-internals/configs-secrets" className="underline hover:no-underline">
                  Configs & Secrets documentation
                </Link>{' '}
                for detailed information on managing configuration and secrets.
              </p>
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
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Networking</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Each environment has its own isolated network namespace. Services within an environment can
          communicate with each other using service names as DNS hostnames, providing simple service
          discovery without hardcoding IPs.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Network Features:
          </h4>
          <ul className="text-muted-foreground text-sm space-y-2 m-0">
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Service Discovery:</strong> Access services by name (e.g.,{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  postgres:5432
                </code>
                ,{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">redis:6379</code>)
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Network Isolation:</strong> Each environment has its own isolated network
                namespace
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Workspace Connectivity:</strong> Workspaces can connect to access environment
                services
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Service Intercepts:</strong> Route service traffic to workspaces for debugging
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Use Cases */}
      <section id="use-cases" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Zap className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Use Cases</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Users className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Team Collaboration
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Create shared environments that your entire team can access. Everyone connects to the
                  same services, ensuring consistent data and behavior across the team.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Container className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Isolated Testing
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Spin up isolated environments for testing features without affecting other team
                  members. Test integration with services in a controlled environment.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Database className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Multiple Stages
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Create separate environments for different stages of development. Have dedicated
                  environments for testing, staging, or demo purposes with their own configurations.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Settings className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Service Development
                </h4>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Develop and test microservices in an environment that mirrors production. Connect
                  workspaces to test service integration and debug with real traffic using service
                  intercepts.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
