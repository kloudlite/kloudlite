import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Lock, Plug, Copy, Route, Users } from 'lucide-react'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'sharing-cloning', title: 'Sharing & Cloning' },
]

export default function EnvironmentOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Environments</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          An environment is an <strong className="text-foreground">isolated set of services</strong> that
          developers access from their workspace. Databases, caches, APIs, and any other service your
          application depends on runs in an environment.
        </p>

        <div className="grid gap-5 md:grid-cols-2">
          <FeatureCard
            icon={Lock}
            title="Isolated"
            description="Each environment runs in its own network namespace."
          />
          <FeatureCard
            icon={Plug}
            title="Accessible"
            description="Connect from any workspace to access services by name."
          />
          <FeatureCard
            icon={Copy}
            title="Clonable"
            description="Fork environments to get your own isolated copy."
          />
          <FeatureCard
            icon={Route}
            title="Interceptable"
            description="Route service traffic to your workspace for debugging."
          />
        </div>
      </section>

      <section id="sharing-cloning" className="mb-16">
        <SectionTitle id="sharing-cloning">Sharing & Cloning</SectionTitle>

        <div className="grid gap-5 md:grid-cols-2">
          <FeatureCard
            icon={Users}
            title="Share with team"
            description="Make an environment visible to other developers. They can connect their workspaces and access the same services with the same data."
          />
          <FeatureCard
            icon={Copy}
            title="Fork for isolation"
            description="Fork an environment to get your own copy with the same composition, configuration, and data. Perfect for testing changes without affecting others."
          />
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/environment-internals/services"
          title="Services"
          description="Define services using Docker Compose syntax"
        />
        <NextLinkCard
          href="/docs/environment-internals/configs-secrets"
          title="Configs & Secrets"
          description="Manage environment variables and config files"
        />
      </div>
    </DocsContentLayout>
  )
}
