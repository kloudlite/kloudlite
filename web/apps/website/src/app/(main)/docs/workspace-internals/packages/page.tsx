import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Package, Terminal } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'commands', title: 'Commands' },
]

export default function PackageManagementPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Package Management
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Install development tools using <strong className="text-foreground">Nix packages</strong>.
          Packages are installed at the host level and shared across workspaces.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Package className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Quick start</h3>
          </div>
          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl pkg add                   # Interactive mode</pre>
            <pre className="m-0">kl pkg add nodejs python     # Add multiple packages</pre>
            <pre className="m-0">kl pkg list                  # List installed</pre>
          </div>
        </div>
      </section>

      {/* Commands */}
      <section id="commands" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Commands</h2>

        <div className="space-y-6">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Terminal className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Search packages</h3>
            </div>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl pkg search nodejs</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Search the Nix package registry for available packages.
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Add packages</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto space-y-2">
              <pre className="m-0">kl pkg add                   # Interactive with fuzzy search</pre>
              <pre className="m-0">kl pkg add git vim curl      # Add by name</pre>
              <pre className="m-0">kl pkg add nodejs --exact    # Pin to exact version</pre>
            </div>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Install with version control</h3>
            <p className="text-muted-foreground text-sm mb-3">
              Specify exact versions, channels, or commits for reproducibility.
            </p>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto space-y-2">
              <pre className="m-0">kl pkg install nodejs --version 20.0.0</pre>
              <pre className="m-0">kl pkg install python --channel nixos-24.05</pre>
              <pre className="m-0">kl pkg install curl --commit abc123def</pre>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <h3 className="text-card-foreground text-lg font-semibold mb-3">List installed</h3>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg list</pre>
              </div>
              <p className="text-muted-foreground text-xs mt-3 m-0">
                Shows packages with their source and status.
              </p>
            </div>

            <div className="bg-card border p-6">
              <h3 className="text-card-foreground text-lg font-semibold mb-3">Remove package</h3>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg uninstall vim</pre>
              </div>
              <p className="text-muted-foreground text-xs mt-3 m-0">
                Removes from your workspace config only.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/cli"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">CLI Reference</p>
            <p className="text-muted-foreground text-sm m-0">Full kl command reference</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workspace-internals/environment-connection"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environment Connection</p>
            <p className="text-muted-foreground text-sm m-0">Connect to environments</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
