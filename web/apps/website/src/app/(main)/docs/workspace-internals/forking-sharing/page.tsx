import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { Users, Copy } from 'lucide-react'

const tocItems = [
  { id: 'sharing', title: 'Sharing Workspaces' },
  { id: 'forking', title: 'Forking Workspaces' },
]

export default function ForkingSharingPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Forking & Sharing</PageTitle>

      <p className="text-muted-foreground mb-12 leading-relaxed">
        Share your workspace with teammates or fork it to run parallel copies for different tasks.
      </p>

      <section id="sharing" className="mb-12">
        <SectionTitle id="sharing">Sharing Workspaces</SectionTitle>

        <FeatureCard
          icon={Users}
          title="Share with teammates"
          description={
            <div>
              <p className="mb-4">
                Share your workspace to let teammates access your{' '}
                <strong className="text-foreground">exposed ports</strong> via public URLs. No VPN required
                for shared access.
              </p>
              <div className="space-y-3">
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">What gets shared</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Only exposed HTTP ports are accessible. Teammates can view your running services
                    through their public URLs.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">What stays private</p>
                  <p className="text-muted-foreground text-sm m-0">
                    SSH access, terminal, and IDE connections still require VPN. Your files and
                    environment remain private.
                  </p>
                </div>
              </div>
            </div>
          }
        />
      </section>

      <section id="forking" className="mb-16">
        <SectionTitle id="forking">Forking Workspaces</SectionTitle>

        <FeatureCard
          icon={Copy}
          title="Fork for parallel development"
          description={
            <div>
              <p className="mb-4">
                Fork any workspace that is shared with or owned by you to create your own independent copy.
                Works like <strong className="text-foreground">git worktrees</strong> for your
                development environment.
              </p>
              <div className="space-y-3">
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">AI-assisted coding</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Run an AI agent in a forked workspace while you continue working in the original.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Parallel experiments</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Test different approaches simultaneously without affecting your main workspace.
                  </p>
                </div>
                <div>
                  <p className="text-foreground font-semibold text-sm mb-1">Team onboarding</p>
                  <p className="text-muted-foreground text-sm m-0">
                    Fork a configured workspace to quickly onboard new team members with the same setup.
                  </p>
                </div>
              </div>
            </div>
          }
        />
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/expose-ports"
          title="Expose Ports"
          description="Get public URLs for your services"
        />
        <NextLinkCard
          href="/docs/concepts/workspaces"
          title="Workspaces Overview"
          description="Learn about workspace features"
        />
      </div>
    </DocsContentLayout>
  )
}
