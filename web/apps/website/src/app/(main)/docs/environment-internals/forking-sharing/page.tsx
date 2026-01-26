import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { Users, Copy } from 'lucide-react'

const tocItems = [
  { id: 'sharing', title: 'Sharing Environments' },
  { id: 'forking', title: 'Forking Environments' },
]

export default function EnvironmentForkingSharingPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Forking & Sharing</PageTitle>

      <p className="text-muted-foreground mb-12 leading-relaxed">
        Share environments with your team or fork them to create isolated copies for testing.
      </p>

      <section id="sharing" className="mb-12">
        <SectionTitle id="sharing">Sharing Environments</SectionTitle>

        <FeatureCard
          icon={Users}
          title="Share with teammates"
          description={
            <div>
              <p className="mb-4">
                Share an environment to let teammates{' '}
                <strong className="text-foreground">connect their workspaces</strong> and access the
                same services with the same data.
              </p>
              <div className="space-y-3">
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Shared access</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Team members can connect to the environment and access all services by name.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Same data</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Everyone works with the same databases, caches, and service state.
                  </p>
                </div>
              </div>
            </div>
          }
        />
      </section>

      <section id="forking" className="mb-16">
        <SectionTitle id="forking">Forking Environments</SectionTitle>

        <FeatureCard
          icon={Copy}
          title="Fork for isolated testing"
          description={
            <div>
              <p className="mb-4">
                Fork any environment that is shared with or owned by you to create your own{' '}
                <strong className="text-foreground">isolated copy</strong> with the same composition
                and configuration.
              </p>
              <div className="space-y-3">
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Test changes safely</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Modify services, update configs, or test migrations without affecting the original.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Feature development</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Create dedicated environments for feature branches or experiments.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Reproduced issues</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Fork a production-like environment to debug issues in isolation.
                  </p>
                </div>
              </div>
            </div>
          }
        />
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/concepts/environments"
          title="Environments Overview"
          description="Learn about environment features"
        />
        <NextLinkCard
          href="/docs/environment-internals/services"
          title="Services"
          description="Define services in your environment"
        />
      </div>
    </DocsContentLayout>
  )
}
