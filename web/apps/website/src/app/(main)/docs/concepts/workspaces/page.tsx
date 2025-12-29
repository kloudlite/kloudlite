import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Package, Network, Route, Globe, Terminal, Container, Users, Copy } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'sharing-cloning', title: 'Sharing & Cloning' },
]

export default function WorkspaceOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Workspaces
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          A workspace is a <strong className="text-foreground">development container</strong> running
          on a workmachine. It provides a complete development environment with pre-installed tools,
          environment connectivity, and multiple access methods.
        </p>

        <div className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Package className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Packages</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Install additional tools using the built-in package manager.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Network className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Environment connection</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Connect to environments and access services by name.
              </p>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Route className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Service intercepts</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Route service traffic to your workspace for debugging.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Globe className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Exposed ports</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Expose HTTP services with public URLs like{' '}
                <code className="text-xs font-mono">p3000-abc.sub.khost.dev</code>.
              </p>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Terminal className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Access methods</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Connect via IDE, SSH, or web terminal. Requires VPN.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Container className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Docker runtime</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Build and run containers locally with Docker DIND.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Sharing & Cloning */}
      <section id="sharing-cloning" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Sharing & Cloning</h2>

        <div className="grid gap-6 md:grid-cols-2">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Users className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Share with team</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Share your workspace with other developers. They can access your exposed ports
              via public URLs without VPN.
            </p>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Copy className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Clone to work independently</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Clone workspaces like git worktrees. Run parallel copies for AI-assisted coding
              or testing different approaches.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/packages"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Packages</p>
            <p className="text-muted-foreground text-sm m-0">Install and manage development tools</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workspace-internals/environment-connection"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environment Connection</p>
            <p className="text-muted-foreground text-sm m-0">Connect to environments and access services</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workspace-internals/intercepts"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Service Intercepts</p>
            <p className="text-muted-foreground text-sm m-0">Route service traffic to your workspace</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
