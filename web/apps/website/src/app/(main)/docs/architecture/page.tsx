import { Server, Boxes } from 'lucide-react'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArchitectureDiagram } from '@/components/docs/architecture'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'control-node', title: 'Control Node' },
  { id: 'workmachines', title: 'Workmachines' },
]

export default function ArchitecturePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Architecture</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground text-base leading-relaxed">
          Kloudlite consists of two main components: a Control Node that orchestrates
          everything, and Workmachines where development happens.
        </p>
      </section>

      <ArchitectureDiagram />

      <section id="control-node" className="mb-10">
        <FeatureCard
          icon={Server}
          title="Control Node"
          description={
            <>
              <p className="mb-3">
                The central management plane running at{' '}
                <code className="bg-muted px-1.5 py-0.5 font-mono text-xs border border-foreground/10 rounded-sm">{'{subdomain}'}.khost.dev</code>.
              </p>
              <ul className="space-y-1.5 text-sm">
                <li>• Team authentication and access control</li>
                <li>• Workmachine provisioning and orchestration</li>
                <li>• Environment and workspace management</li>
                <li>• Backups and recovery</li>
              </ul>
            </>
          }
        />
      </section>

      <section id="workmachines" className="mb-16">
        <FeatureCard
          icon={Boxes}
          title="Workmachines"
          description={
            <>
              <p className="mb-3">
                VMs where users run their environments and workspaces.
              </p>
              <ul className="space-y-1.5 text-sm">
                <li>• Run multiple environments (services via Docker Compose)</li>
                <li>• Run multiple workspaces (dev containers with IDE access)</li>
                <li>• Packages installed at host level using Nix</li>
                <li>• Shared home directory across workspaces</li>
                <li>• Network isolation per workspace</li>
              </ul>
            </>
          }
        />
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/introduction/getting-started"
          title="Getting Started"
          description="Create your first environment and workspace"
        />
      </div>
    </DocsContentLayout>
  )
}
