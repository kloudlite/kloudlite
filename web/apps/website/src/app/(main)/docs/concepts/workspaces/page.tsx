import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Package, Network, Route, Globe, Terminal, Container, Users, Copy } from 'lucide-react'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'sharing-cloning', title: 'Sharing & Cloning' },
]

export default function WorkspaceOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Workspaces</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          A workspace is a <strong className="text-foreground">development container</strong> running
          on a workmachine. It provides a complete development environment with pre-installed tools,
          environment connectivity, and multiple access methods.
        </p>

        <div className="grid gap-5 md:grid-cols-2">
          <FeatureCard
            icon={Package}
            title="Packages"
            description="Install additional tools using the built-in package manager."
          />
          <FeatureCard
            icon={Network}
            title="Environment connection"
            description="Connect to environments and access services by name."
          />
          <FeatureCard
            icon={Route}
            title="Service intercepts"
            description="Route service traffic to your workspace for debugging."
          />
          <FeatureCard
            icon={Globe}
            title="Exposed ports"
            description={
              <>
                Expose HTTP services with public URLs like{' '}
                <code className="text-xs font-mono bg-muted px-1.5 py-0.5 border border-foreground/10 rounded-sm">p3000-abc.sub.khost.dev</code>.
              </>
            }
          />
          <FeatureCard
            icon={Terminal}
            title="Access methods"
            description="Connect via IDE, SSH, or web terminal. Requires VPN."
          />
          <FeatureCard
            icon={Container}
            title="Docker runtime"
            description="Build and run containers locally with Docker DIND."
          />
        </div>
      </section>

      <section id="sharing-cloning" className="mb-16">
        <SectionTitle id="sharing-cloning">Sharing & Cloning</SectionTitle>

        <div className="grid gap-5 md:grid-cols-2">
          <FeatureCard
            icon={Users}
            title="Share with team"
            description="Share your workspace with other developers. They can access your exposed ports via public URLs without VPN."
          />
          <FeatureCard
            icon={Copy}
            title="Clone to work independently"
            description="Fork workspaces like git worktrees. Run parallel copies for AI-assisted coding or testing different approaches."
          />
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/packages"
          title="Packages"
          description="Install and manage development tools"
        />
        <NextLinkCard
          href="/docs/workspace-internals/environment-connection"
          title="Environment Connection"
          description="Connect to environments and access services"
        />
        <NextLinkCard
          href="/docs/workspace-internals/intercepts"
          title="Service Intercepts"
          description="Route service traffic to your workspace"
        />
      </div>
    </DocsContentLayout>
  )
}
