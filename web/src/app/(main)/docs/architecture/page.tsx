import { Button } from '@/components/ui/button'
import Link from 'next/link'
import {
  Server,
  Boxes,
  Database,
  Code2,
  Users,
  Shield,
  GitBranch,
  Lock,
  Layers,
  ArrowRight,
  Info,
  CheckCircle2,
} from 'lucide-react'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArchitectureDiagram, AnimatedFeatureCard } from '@/components/docs/architecture'

const tocItems = [
  { id: 'control-node', title: 'Control Node' },
  { id: 'workmachines', title: 'Workmachines' },
  { id: 'environments', title: 'Environments', level: 2 },
  { id: 'workspaces', title: 'Workspaces', level: 2 },
  { id: 'key-principles', title: 'Key Principles' },
]

export default function ArchitecturePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
        {/* Header */}
        <div className="mb-12 sm:mb-16">
          <h1 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl break-words leading-tight sm:leading-tight">
            Architecture
          </h1>
          <p className="text-muted-foreground mt-4 text-base sm:text-lg lg:text-xl leading-relaxed">
            Understanding Kloudlite&apos;s architecture: Control Node and Workmachines
          </p>
        </div>

      {/* Architecture Diagram - Flowchart Style */}
      <ArchitectureDiagram />

      {/* Control Node */}
      <section id="control-node" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Server className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Control Node
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          The Control Node is the heart of your installation, running at{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">
            {'{subdomain}'}.khost.dev
          </code>
          . It&apos;s a dedicated VM that orchestrates everything within your installation.
        </p>

        <h3 className="text-foreground mb-4 text-xl font-semibold">Core Responsibilities</h3>
        <div className="grid gap-4 sm:gap-6 mb-6">
          <AnimatedFeatureCard delay={0}>
            <div className="bg-card rounded-lg border p-4 sm:p-6">
              <div className="flex items-start gap-4">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Users className="text-primary h-6 w-6" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                    Team Access Management
                  </h4>
                  <p className="text-muted-foreground text-sm leading-relaxed m-0">
                    Handles team member authentication, OAuth integration (GitHub, Google,
                    Microsoft), and role-based access control for your entire team
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.1}>
            <div className="bg-card rounded-lg border p-4 sm:p-6">
              <div className="flex items-start gap-4">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Boxes className="text-primary h-6 w-6" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                    Workmachine Orchestration
                  </h4>
                  <p className="text-muted-foreground text-sm leading-relaxed m-0">
                    Provisions, configures, and manages VM instances for team members. Handles
                    scaling, health monitoring, and resource allocation
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.2}>
            <div className="bg-card rounded-lg border p-4 sm:p-6">
              <div className="flex items-start gap-4">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Database className="text-primary h-6 w-6" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                    Centralized Management
                  </h4>
                  <p className="text-muted-foreground text-sm leading-relaxed m-0">
                    All environments, workspaces, configurations, and team resources are managed
                    centrally through the Control Node&apos;s web interface
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.3}>
            <div className="bg-card rounded-lg border p-4 sm:p-6">
              <div className="flex items-start gap-4">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Shield className="text-primary h-6 w-6" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                    Backups & Recovery
                  </h4>
                  <p className="text-muted-foreground text-sm leading-relaxed m-0">
                    Automated backups of managed state including configurations, team settings, and metadata.
                    Point-in-time recovery for control plane data (excludes environment and workspace states)
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>
        </div>
      </section>

      {/* Workmachines */}
      <section id="workmachines" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Boxes className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Workmachines (User VMs)
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Workmachines are individual VM instances where users actually run their development work.
          Each workmachine is isolated and contains two main components:
        </p>

        {/* Environments */}
        <div id="environments" className="mb-8">
          <h3 className="text-foreground mb-4 text-xl font-semibold flex items-center gap-2">
            <Database className="text-primary h-6 w-6" />
            Environments
          </h3>

          <p className="text-muted-foreground mb-4 leading-relaxed">
            Isolated spaces where your application services run - think of them as different stages
            like development, staging, or production.
          </p>

          <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
              What runs in Environments:
            </h4>
            <ul className="space-y-2 m-0">
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Services via Docker Compose:</strong> Databases (PostgreSQL, MongoDB,
                  MySQL), caches (Redis), message queues, APIs
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Configuration:</strong> Environment variables, config files, secrets
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Network Isolation:</strong> Each environment has its own network namespace
                  with service discovery
                </span>
              </li>
            </ul>
          </div>
        </div>

        {/* Workspaces */}
        <div id="workspaces" className="mb-6">
          <h3 className="text-foreground mb-4 text-xl font-semibold flex items-center gap-2">
            <Code2 className="text-primary h-6 w-6" />
            Workspaces
          </h3>

          <p className="text-muted-foreground mb-4 leading-relaxed">
            Isolated development containers on workmachines. Each workspace provides network isolation,
            manages package access, and controls environment connectivity while sharing host-level resources.
          </p>

          <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
              What&apos;s in a Workspace:
            </h4>
            <ul className="space-y-2 m-0">
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Multiple Access Methods:</strong> VS Code Web, SSH (for desktop IDEs like VS Code, Cursor, IntelliJ), web terminal, and AI assistants (Claude Code, etc.)
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Network Isolation:</strong> Each workspace controls its own network namespace, providing isolation and security
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Package Management:</strong> Packages installed and persisted at workmachine host level using Nix, made available in workspace PATH based on each workspace&apos;s configuration
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Environment Connection:</strong> Network namespace switches to access environment services by name
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Shared Home Directory:</strong> Home folder (~) shared across all workspaces on the workmachine, tool configurations persist
                </span>
              </li>
              <li className="flex items-start gap-2 text-sm">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span className="text-muted-foreground">
                  <strong>Workspace Code Storage:</strong> Each workspace&apos;s code stored in <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">~/workspaces/[workspace-name]</code>
                </span>
              </li>
            </ul>
          </div>
        </div>

        <div className="bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Workspace-Environment Connection
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                Workspaces can connect to environments to access services. For example, your
                workspace can connect to a &quot;development&quot; environment to access its
                PostgreSQL database at <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">postgres:5432</code>
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Key Architectural Principles */}
      <section id="key-principles" className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-6 text-2xl sm:text-3xl font-bold">
          Key Architectural Principles
        </h2>

        <div className="grid gap-4 sm:gap-6">
          <AnimatedFeatureCard delay={0}>
            <div className="bg-card rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Lock className="text-primary h-5 w-5" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground font-semibold mb-1 m-0">
                    Data Isolation & Security
                  </h4>
                  <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                    Your data never leaves your installation. Each workmachine is isolated with its
                    own network namespaces and security boundaries
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.1}>
            <div className="bg-card rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Layers className="text-primary h-5 w-5" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground font-semibold mb-1 m-0">
                    Resource Efficiency
                  </h4>
                  <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                    Environments can be deactivated when not in use. Workspaces can be suspended.
                    Workmachines scale based on team needs
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.2}>
            <div className="bg-card rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Users className="text-primary h-5 w-5" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground font-semibold mb-1 m-0">
                    Team Collaboration
                  </h4>
                  <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                    Team members share environments and can discover each other&apos;s resources
                    within the installation scope
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>

          <AnimatedFeatureCard delay={0.3}>
            <div className="bg-card rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <div className="bg-primary/10 rounded-lg p-2">
                  <GitBranch className="text-primary h-5 w-5" />
                </div>
                <div className="flex-1">
                  <h4 className="text-card-foreground font-semibold mb-1 m-0">
                    Regional Deployment
                  </h4>
                  <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                    Deploy in your preferred cloud region (AWS, GCP, Azure) close to your team for
                    low-latency access
                  </p>
                </div>
              </div>
            </div>
          </AnimatedFeatureCard>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-6 text-2xl sm:text-3xl font-bold">Next Steps</h2>
        <div className="grid gap-4 sm:gap-6 md:grid-cols-2">
          <Link
            href="/docs/introduction/installation"
            className="bg-card rounded-lg border p-4 sm:p-6 hover:border-primary transition-colors no-underline block"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-lg font-semibold m-0">
                  Installation Guide
                </h3>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Learn how to create your installation and set up the Control Node
                </p>
              </div>
              <ArrowRight className="text-muted-foreground h-5 w-5 flex-shrink-0 mt-1" />
            </div>
          </Link>

          <Link
            href="/docs/introduction/getting-started"
            className="bg-card rounded-lg border p-4 sm:p-6 hover:border-primary transition-colors no-underline block"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-lg font-semibold m-0">
                  Getting Started
                </h3>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Create your first workmachine, environment, and workspace
                </p>
              </div>
              <ArrowRight className="text-muted-foreground h-5 w-5 flex-shrink-0 mt-1" />
            </div>
          </Link>
        </div>
      </section>
    </DocsContentLayout>
  )
}
