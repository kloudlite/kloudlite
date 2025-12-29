import Link from 'next/link'
import { Server, Boxes, ArrowRight } from 'lucide-react'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArchitectureDiagram } from '@/components/docs/architecture'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'control-node', title: 'Control Node' },
  { id: 'workmachines', title: 'Workmachines' },
]

export default function ArchitecturePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Architecture
      </h1>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground text-lg leading-relaxed mb-4">
          Kloudlite consists of two main components: a Control Node that orchestrates
          everything, and Workmachines where development happens.
        </p>
      </section>

      <ArchitectureDiagram />

      <section id="control-node" className="mb-8">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Server className="h-5 w-5 text-primary" />
            <h2 className="text-card-foreground text-lg font-semibold m-0">Control Node</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-3">
            The central management plane running at{' '}
            <code className="bg-muted px-1.5 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code>.
          </p>
          <ul className="text-muted-foreground text-sm space-y-1">
            <li>• Team authentication and access control</li>
            <li>• Workmachine provisioning and orchestration</li>
            <li>• Environment and workspace management</li>
            <li>• Backups and recovery</li>
          </ul>
        </div>
      </section>

      <section id="workmachines" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Boxes className="h-5 w-5 text-primary" />
            <h2 className="text-card-foreground text-lg font-semibold m-0">Workmachines</h2>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-3">
            VMs where users run their environments and workspaces.
          </p>
          <ul className="text-muted-foreground text-sm space-y-1">
            <li>• Run multiple environments (services via Docker Compose)</li>
            <li>• Run multiple workspaces (dev containers with IDE access)</li>
            <li>• Packages installed at host level using Nix</li>
            <li>• Shared home directory across workspaces</li>
            <li>• Network isolation per workspace</li>
          </ul>
        </div>
      </section>

      <div className="border-t pt-8">
        <Link
          href="/docs/introduction/getting-started"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Getting Started</p>
            <p className="text-muted-foreground text-sm m-0">Create your first environment and workspace</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
