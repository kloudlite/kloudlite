import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Users, Copy } from 'lucide-react'

const tocItems = [
  { id: 'sharing', title: 'Sharing Workspaces' },
  { id: 'cloning', title: 'Cloning Workspaces' },
]

export default function CloningSharingPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Cloning & Sharing
      </h1>

      <p className="text-muted-foreground mb-12 leading-relaxed">
        Share your workspace with teammates or clone it to run parallel copies for different tasks.
      </p>

      {/* Sharing */}
      <section id="sharing" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Users className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Sharing workspaces</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Share your workspace to let teammates access your{' '}
            <strong className="text-foreground">exposed ports</strong> via public URLs. No VPN required
            for shared access.
          </p>

          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">What gets shared</p>
              <p className="text-muted-foreground text-xs m-0">
                Only exposed HTTP ports are accessible. Teammates can view your running services
                through their public URLs.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">What stays private</p>
              <p className="text-muted-foreground text-xs m-0">
                SSH access, terminal, and IDE connections still require VPN. Your files and
                environment remain private.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Cloning */}
      <section id="cloning" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Copy className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Cloning workspaces</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Clone any workspace that is shared with or owned by you to create your own independent copy.
            Works like <strong className="text-foreground">git worktrees</strong> for your
            development environment.
          </p>

          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">AI-assisted coding</p>
              <p className="text-muted-foreground text-xs m-0">
                Run an AI agent in a cloned workspace while you continue working in the original.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Parallel experiments</p>
              <p className="text-muted-foreground text-xs m-0">
                Test different approaches simultaneously without affecting your main workspace.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Team onboarding</p>
              <p className="text-muted-foreground text-xs m-0">
                Clone a configured workspace to quickly onboard new team members with the same setup.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/expose-ports"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Expose Ports</p>
            <p className="text-muted-foreground text-sm m-0">Get public URLs for your services</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/concepts/workspaces"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Workspaces Overview</p>
            <p className="text-muted-foreground text-sm m-0">Learn about workspace features</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
