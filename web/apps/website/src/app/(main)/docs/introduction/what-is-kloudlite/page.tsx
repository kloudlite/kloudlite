import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Database, Code2, Server, Network } from 'lucide-react'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'core-concepts', title: 'Core Concepts' },
]

export default function WhatIsKloudlitePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>What is Kloudlite?</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-4">
          Kloudlite is a Cloud Development Environment platform designed to shorten the
          development loop and increase developer productivity.
        </p>
        <p className="text-muted-foreground leading-relaxed">
          By connecting cloud-based workspaces directly to application services, developers
          can test code changes instantly without waiting for builds or deployments.
        </p>
      </section>

      <section id="core-concepts" className="mb-16">
        <SectionTitle id="core-concepts">Core Concepts</SectionTitle>

        <div className="space-y-5">
          <FeatureCard
            icon={Network}
            title="Installation"
            description="A cluster of workmachines that enables users to collaborate. Your team's dedicated Kloudlite deployment."
          />

          <FeatureCard
            icon={Server}
            title="Workmachines"
            description={
              <>
                <p className="mb-3">
                  VMs designed to run environments and workspaces. Resizable and configurable
                  based on your needs.
                </p>
                <ul className="space-y-1.5 text-sm">
                  <li>• Auto shutdown when no workspace is active (select cloud providers)</li>
                  <li>• Each workmachine has its own generated SSH key pair</li>
                  <li>• Add authorized keys to access workspaces</li>
                </ul>
              </>
            }
          />

          <div className="grid gap-5 md:grid-cols-2">
            <FeatureCard
              icon={Database}
              title="Environments"
              description="Isolated namespaces containing the services your application depends on—databases, caches, message queues, and other microservices."
            />

            <FeatureCard
              icon={Code2}
              title="Workspaces"
              description="Development containers with your codebase, packages, and IDE. Connect to environments to access services by name."
            />
          </div>
        </div>
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
