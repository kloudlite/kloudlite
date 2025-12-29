import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Server, Clock, Key, Shield } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'features', title: 'Features' },
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
          A workmachine is the <strong className="text-foreground">host machine</strong> where your
          workspaces run. It provides the compute resources, manages packages at the host level,
          and handles workspace lifecycle.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Server className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">What it provides</h3>
          </div>
          <div className="space-y-2 text-muted-foreground text-sm">
            <p className="m-0">
              <strong className="text-foreground">Compute resources</strong> — CPU, memory, and storage for your workspaces
            </p>
            <p className="m-0">
              <strong className="text-foreground">Package cache</strong> — Nix packages installed once, shared across workspaces
            </p>
            <p className="m-0">
              <strong className="text-foreground">Docker runtime</strong> — DIND runtime for building and running containers
            </p>
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Features</h2>

        <div className="space-y-6">
          <div className="grid gap-6 md:grid-cols-3">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Clock className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Auto Stop</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Automatically stop idle workmachines to save costs.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Key className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">SSH Keys</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Manage SSH keys for secure workspace access.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Shield className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">Authorized Keys</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Control who can access your workspaces.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workmachines/auto-stop"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Auto Stop</p>
            <p className="text-muted-foreground text-sm m-0">Configure idle timeout settings</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workmachines/ssh-keys"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">SSH Keys</p>
            <p className="text-muted-foreground text-sm m-0">Manage SSH keys for access</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workmachines/authorized-keys"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Authorized Keys</p>
            <p className="text-muted-foreground text-sm m-0">Control workspace access</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
