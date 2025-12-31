import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Power, SlidersHorizontal, Package, Container, Clock } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'management', title: 'Management' },
]

export default function WorkmachinesPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Workmachines
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          A workmachine is a <strong className="text-foreground">virtual machine</strong> that runs
          your workspaces and environments. It serves as the gateway for accessing all your development
          resources in Kloudlite.
        </p>

        <div className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Power className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Start & Stop</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Start and stop VMs on demand. Idle machines can be stopped automatically to optimize costs.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <SlidersHorizontal className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Flexible Sizing</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Change CPU and RAM allocation based on workload needs. Scale up for heavy builds, scale down for light work.
              </p>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Package className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Package Cache</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Nix packages installed once at the host level, shared across all workspaces on the machine.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Container className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Docker Runtime</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Docker-in-Docker runtime for building and running containers within your workspaces.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Management */}
      <section id="management" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Management</h2>

        <div className="space-y-4">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Power className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Start / Stop</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Start or stop your workmachine manually from the dashboard. Stopped machines do not consume compute resources.
            </p>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Clock className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Auto Stop</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Configure idle timeout to automatically stop machines after a period of inactivity.
            </p>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <SlidersHorizontal className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Change Shape</h3>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Resize CPU and RAM allocation anytime. Scale up for demanding workloads, scale down to reduce costs.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workmachines/access"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Access</p>
            <p className="text-muted-foreground text-sm m-0">VPN, SSH keys, and IDE integration</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
