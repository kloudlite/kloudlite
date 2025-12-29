import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Terminal, Package, Network, Route, Settings, Globe, Container, Sparkles } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'pkg', title: 'kl pkg' },
  { id: 'env', title: 'kl env' },
  { id: 'intercept', title: 'kl intercept' },
  { id: 'status', title: 'kl status' },
  { id: 'config', title: 'kl config' },
  { id: 'expose', title: 'kl expose' },
  { id: 'image', title: 'kl image' },
  { id: 'other', title: 'Other Commands' },
]

export default function CLIReferencePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        CLI Reference
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          The <code className="bg-muted px-1.5 py-0.5 font-mono text-sm">kl</code> CLI manages
          your workspace from within the dev container. Most commands support{' '}
          <strong className="text-foreground">interactive mode</strong> when run without arguments.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Terminal className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Quick Start</h3>
          </div>
          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl status                    # Show workspace info</pre>
            <pre className="m-0">kl pkg add nodejs python     # Install packages</pre>
            <pre className="m-0">kl env connect               # Connect to environment (interactive)</pre>
            <pre className="m-0">kl intercept start           # Intercept a service (interactive)</pre>
            <pre className="m-0">kl expose 3000               # Expose port with public URL</pre>
          </div>
        </div>
      </section>

      {/* Package Commands */}
      <section id="pkg" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Package className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl pkg</h3>
            <span className="text-muted-foreground text-xs font-mono">aliases: package, p</span>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Manage Nix packages in your workspace.
          </p>

          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-medium mb-2">Search packages</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg search &lt;query&gt;        # Search available packages</pre>
                <pre className="m-0">kl p s nodejs                # Short form</pre>
              </div>
            </div>

            <div>
              <p className="text-foreground text-sm font-medium mb-2">Add packages</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg add                   # Interactive mode with fuzzy search</pre>
                <pre className="m-0">kl pkg add nodejs python     # Add multiple packages</pre>
                <pre className="m-0">kl pkg add nodejs --exact    # Pin to exact version (slower)</pre>
              </div>
            </div>

            <div>
              <p className="text-foreground text-sm font-medium mb-2">Install with version control</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg install nodejs --version 20.0.0</pre>
                <pre className="m-0">kl pkg install python --channel nixos-24.05</pre>
                <pre className="m-0">kl pkg install curl --commit abc123def</pre>
              </div>
            </div>

            <div>
              <p className="text-foreground text-sm font-medium mb-2">Manage installed</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">kl pkg list                  # List installed packages</pre>
                <pre className="m-0">kl pkg uninstall vim         # Remove package</pre>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Environment Commands */}
      <section id="env" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Network className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl env</h3>
            <span className="text-muted-foreground text-xs font-mono">aliases: environment, e</span>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Connect to environments and access services by name.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl env connect               # Interactive environment selection</pre>
            <pre className="m-0">kl env connect my-env        # Connect to specific environment</pre>
            <pre className="m-0">kl env disconnect            # Disconnect (removes intercepts)</pre>
            <pre className="m-0">kl env status                # Show connection status</pre>
          </div>
        </div>
      </section>

      {/* Intercept Commands */}
      <section id="intercept" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Route className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl intercept</h3>
            <span className="text-muted-foreground text-xs font-mono">aliases: int, i</span>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Route service traffic to your workspace for debugging.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl intercept start           # Interactive service & port selection</pre>
            <pre className="m-0">kl intercept start api       # Start intercepting specific service</pre>
            <pre className="m-0">kl intercept stop api        # Stop intercepting</pre>
            <pre className="m-0">kl intercept list            # List active intercepts</pre>
            <pre className="m-0">kl intercept status          # Show intercept status</pre>
          </div>
        </div>
      </section>

      {/* Status Command */}
      <section id="status" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Terminal className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl status</h3>
            <span className="text-muted-foreground text-xs font-mono">aliases: st, s</span>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Display workspace information including phase, resources, access URLs, and timing.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto">
            <pre className="m-0">kl status                    # Show workspace info</pre>
          </div>

          <div className="mt-4 grid sm:grid-cols-2 gap-3">
            <div className="bg-muted/50 p-3 text-sm">
              <p className="text-foreground font-medium mb-1">Shows:</p>
              <ul className="text-muted-foreground text-xs space-y-1 m-0 pl-4">
                <li>Workspace name & owner</li>
                <li>Current phase</li>
                <li>Resource usage (CPU, Memory, Storage)</li>
              </ul>
            </div>
            <div className="bg-muted/50 p-3 text-sm">
              <p className="text-foreground font-medium mb-1">Also includes:</p>
              <ul className="text-muted-foreground text-xs space-y-1 m-0 pl-4">
                <li>Access URLs (VS Code, SSH)</li>
                <li>Start time & last activity</li>
                <li>Active connections count</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Config Commands */}
      <section id="config" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Settings className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl config</h3>
            <span className="text-muted-foreground text-xs font-mono">aliases: cfg, c</span>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            View and update workspace configuration.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2 mb-4">
            <pre className="m-0">kl config get                # Show all configuration</pre>
            <pre className="m-0">kl config get display-name   # Get specific value</pre>
            <pre className="m-0">kl config set display-name "My Workspace"</pre>
            <pre className="m-0">kl config set git.user-email "dev@example.com"</pre>
          </div>

          <div className="bg-muted/50 p-3 text-sm">
            <p className="text-foreground font-medium mb-2">Available keys:</p>
            <div className="grid sm:grid-cols-2 gap-x-4 gap-y-1 text-xs font-mono text-muted-foreground">
              <span>display-name</span>
              <span>git.user-name</span>
              <span>description</span>
              <span>git.user-email</span>
              <span>auto-stop</span>
              <span>git.default-branch</span>
              <span>idle-timeout</span>
              <span>env.&lt;VAR_NAME&gt;</span>
              <span>max-runtime</span>
              <span>vscode-version</span>
            </div>
          </div>
        </div>
      </section>

      {/* Expose Commands */}
      <section id="expose" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Globe className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl expose</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Expose workspace ports with public URLs for HTTP services.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2 mb-4">
            <pre className="m-0">kl expose 3000               # Expose port 3000</pre>
            <pre className="m-0">kl expose list               # List exposed ports</pre>
            <pre className="m-0">kl expose remove 3000        # Remove exposed port</pre>
          </div>

          <p className="text-muted-foreground text-xs">
            Generates URLs like{' '}
            <code className="bg-muted px-1.5 py-0.5 font-mono">https://p3000-abc.subdomain.khost.dev</code>
          </p>
        </div>
      </section>

      {/* Image Commands */}
      <section id="image" className="mb-12">
        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Container className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl image</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Build and push container images to Kloudlite registry.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl image push myapp          # Build and push image</pre>
            <pre className="m-0">kl image push -f Dockerfile.prod myapp</pre>
            <pre className="m-0">kl image push --build-arg VERSION=1.0 myapp</pre>
          </div>
        </div>
      </section>

      {/* Other Commands */}
      <section id="other" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Other Commands</h2>

        <div className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Sparkles className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">kl mcp</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Start MCP server for AI assistants. Exposes 29 tools for Claude and other AI tools.
              </p>
            </div>

            <div className="bg-card border p-6">
              <div className="flex items-center gap-3 mb-3">
                <Terminal className="h-5 w-5 text-primary" />
                <h3 className="text-card-foreground text-lg font-semibold m-0">kl completion</h3>
              </div>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                Generate shell completions for bash, zsh, fish, or powershell.
              </p>
            </div>
          </div>

          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Terminal className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">kl version</h3>
              <span className="text-muted-foreground text-xs font-mono">alias: v</span>
            </div>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Display CLI version information.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/packages"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Packages</p>
            <p className="text-muted-foreground text-sm m-0">Learn more about package management</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workspace-internals/environment-connection"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environment Connection</p>
            <p className="text-muted-foreground text-sm m-0">Learn more about connecting to environments</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/workspace-internals/intercepts"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Service Intercepts</p>
            <p className="text-muted-foreground text-sm m-0">Learn more about service interception</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
