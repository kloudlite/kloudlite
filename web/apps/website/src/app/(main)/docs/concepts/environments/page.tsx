import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Lock, Plug, Copy, Route, Users } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'sharing-cloning', title: 'Sharing & Cloning' },
]

export default function EnvironmentOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Environments
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          An environment is an <strong className="text-foreground">isolated set of services</strong> that
          developers access from their workspace. Databases, caches, APIs, and any other service your
          application depends on runs in an environment.
        </p>

        <div className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Lock className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Isolated</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Each environment runs in its own network namespace.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Plug className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Accessible</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Connect from any workspace to access services by name.
              </p>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Copy className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Clonable</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Clone environments to get your own isolated copy.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Route className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Interceptable</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Route service traffic to your workspace for debugging.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Sharing & Cloning */}
      <section id="sharing-cloning" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Sharing & Cloning</h2>

        <div className="grid gap-6 md:grid-cols-2">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Users className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Share with team</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Make an environment visible to other developers. They can connect their workspaces
              and access the same services with the same data.
            </p>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Copy className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Clone for isolation</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Clone an environment to get your own copy with the same composition, configuration,
              and data. Perfect for testing changes without affecting others.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/environment-internals/services"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Services</p>
            <p className="text-muted-foreground text-sm m-0">Define services using Docker Compose syntax</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/environment-internals/configs-secrets"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Configs & Secrets</p>
            <p className="text-muted-foreground text-sm m-0">Manage environment variables and config files</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
