import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Users, Copy } from 'lucide-react'

const tocItems = [
  { id: 'sharing', title: 'Sharing Environments' },
  { id: 'forking', title: 'Forking Environments' },
]

export default function EnvironmentForkingSharingPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Forking & Sharing
      </h1>

      <p className="text-muted-foreground mb-12 leading-relaxed">
        Share environments with your team or fork them to create isolated copies for testing.
      </p>

      {/* Sharing */}
      <section id="sharing" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Users className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Sharing environments</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Share an environment to let teammates{' '}
            <strong className="text-foreground">connect their workspaces</strong> and access the
            same services with the same data.
          </p>

          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Shared access</p>
              <p className="text-muted-foreground text-xs m-0">
                Team members can connect to the environment and access all services by name.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Same data</p>
              <p className="text-muted-foreground text-xs m-0">
                Everyone works with the same databases, caches, and service state.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Forking */}
      <section id="forking" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Copy className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Forking environments</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Fork any environment that is shared with or owned by you to create your own{' '}
            <strong className="text-foreground">isolated copy</strong> with the same composition
            and configuration.
          </p>

          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Test changes safely</p>
              <p className="text-muted-foreground text-xs m-0">
                Modify services, update configs, or test migrations without affecting the original.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Feature development</p>
              <p className="text-muted-foreground text-xs m-0">
                Create dedicated environments for feature branches or experiments.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Reproduced issues</p>
              <p className="text-muted-foreground text-xs m-0">
                Fork a production-like environment to debug issues in isolation.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/concepts/environments"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environments Overview</p>
            <p className="text-muted-foreground text-sm m-0">Learn about environment features</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/environment-internals/services"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Services</p>
            <p className="text-muted-foreground text-sm m-0">Define services in your environment</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
