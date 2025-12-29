import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Database, Code2, Server, Network, ArrowRight } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'core-concepts', title: 'Core Concepts' },
]

export default function WhatIsKloudlitePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        What is Kloudlite?
      </h1>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground text-lg leading-relaxed mb-4">
          Kloudlite is a Cloud Development Environment platform designed to shorten the
          development loop and increase developer productivity.
        </p>
        <p className="text-muted-foreground leading-relaxed">
          By connecting cloud-based workspaces directly to application services, developers
          can test code changes instantly without waiting for builds or deployments.
        </p>
      </section>

      <section id="core-concepts" className="mb-12">
        <h2 className="text-foreground mb-6 text-2xl font-bold">Core Concepts</h2>

        <div className="space-y-6">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Network className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Installation</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              A cluster of workmachines that enables users to collaborate. Your team's
              dedicated Kloudlite deployment.
            </p>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Server className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Workmachines</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed mb-3">
              VMs designed to run environments and workspaces. Resizable and configurable
              based on your needs.
            </p>
            <ul className="text-muted-foreground text-sm space-y-1">
              <li>• Auto shutdown when no workspace is active (select cloud providers)</li>
              <li>• Each workmachine has its own generated SSH key pair</li>
              <li>• Add authorized keys to access workspaces</li>
            </ul>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Database className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Environments</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Isolated namespaces containing the services your application depends on—databases,
                caches, message queues, and other microservices.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Code2 className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Workspaces</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Development containers with your codebase, packages, and IDE. Connect to
                environments to access services by name.
              </p>
            </div>
          </div>
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
