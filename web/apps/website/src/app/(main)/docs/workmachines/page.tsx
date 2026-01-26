import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Power, SlidersHorizontal, Package, Container, Clock } from 'lucide-react'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'management', title: 'Management' },
]

export default function WorkmachinesPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Workmachines</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          A workmachine is a <strong className="text-foreground">virtual machine</strong> that runs
          your workspaces and environments. It serves as the gateway for accessing all your development
          resources in Kloudlite.
        </p>

        <div className="grid gap-5 md:grid-cols-2">
          <FeatureCard
            icon={Power}
            title="Start & Stop"
            description="Start and stop VMs on demand. Idle machines can be stopped automatically to optimize costs."
          />
          <FeatureCard
            icon={SlidersHorizontal}
            title="Flexible Sizing"
            description="Change CPU and RAM allocation based on workload needs. Scale up for heavy builds, scale down for light work."
          />
          <FeatureCard
            icon={Package}
            title="Package Cache"
            description="Nix packages installed once at the host level, shared across all workspaces on the machine."
          />
          <FeatureCard
            icon={Container}
            title="Docker Runtime"
            description="Docker-in-Docker runtime for building and running containers within your workspaces."
          />
        </div>
      </section>

      <section id="management" className="mb-16">
        <SectionTitle id="management">Management</SectionTitle>

        <div className="space-y-5">
          <FeatureCard
            icon={Power}
            title="Start / Stop"
            description="Start or stop your workmachine manually from the dashboard. Stopped machines do not consume compute resources."
          />
          <FeatureCard
            icon={Clock}
            title="Auto Stop"
            description="Configure idle timeout to automatically stop machines after a period of inactivity."
          />
          <FeatureCard
            icon={SlidersHorizontal}
            title="Change Shape"
            description="Resize CPU and RAM allocation anytime. Scale up for demanding workloads, scale down to reduce costs."
          />
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/workmachines/access"
          title="Access"
          description="VPN, SSH keys, and IDE integration"
        />
      </div>
    </DocsContentLayout>
  )
}
