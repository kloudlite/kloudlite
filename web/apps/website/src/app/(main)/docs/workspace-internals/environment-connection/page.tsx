import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Network, Terminal, Database } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'connecting', title: 'Connecting' },
  { id: 'accessing-services', title: 'Accessing Services' },
]

export default function EnvironmentConnectionPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Environment Connection
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Connect your workspace to an environment to{' '}
          <strong className="text-foreground">access services by name</strong>. Once connected,
          services like <code className="bg-muted px-1.5 py-0.5 font-mono text-sm">postgres</code> or{' '}
          <code className="bg-muted px-1.5 py-0.5 font-mono text-sm">redis</code> become directly
          reachable from your workspace.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Network className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">How it works</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed m-0">
            When you connect, your workspace&apos;s DNS is configured to resolve environment
            service names. No port forwarding or IP addresses needed.
          </p>
        </div>
      </section>

      {/* Connecting */}
      <section id="connecting" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Connecting</h2>

        <div className="bg-card border p-6 mb-6">
          <div className="flex items-center gap-3 mb-3">
            <Terminal className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl env connect</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Connect to an environment interactively or by name.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl env connect               # Interactive selection</pre>
            <pre className="m-0">kl env connect my-env        # Connect to specific environment</pre>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Disconnect</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl env disconnect</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Removes active intercepts and clears DNS config.
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Check status</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl env status</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Shows connected environment and available services.
            </p>
          </div>
        </div>
      </section>

      {/* Accessing Services */}
      <section id="accessing-services" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Accessing Services</h2>

        <p className="text-muted-foreground mb-6 text-sm">
          After connecting, use <strong className="text-foreground">service names as hostnames</strong>.
          Always include the port number.
        </p>

        <div className="bg-card border p-6 mb-6">
          <div className="flex items-center gap-3 mb-3">
            <Database className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Examples</h3>
          </div>

          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-medium mb-2">From terminal</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto space-y-1">
                <pre className="m-0">psql -h postgres -p 5432 -U myuser</pre>
                <pre className="m-0">redis-cli -h redis -p 6379</pre>
                <pre className="m-0">curl http://api:8080/health</pre>
              </div>
            </div>

            <div>
              <p className="text-foreground text-sm font-medium mb-2">In application config</p>
              <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
                <pre className="m-0">{`DATABASE_URL=postgresql://user:pass@postgres:5432/myapp
REDIS_URL=redis://redis:6379
API_ENDPOINT=http://api:8080`}</pre>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-muted/50 border p-4">
          <p className="text-foreground text-sm font-medium mb-1">Available service names</p>
          <p className="text-muted-foreground text-xs m-0">
            Run <code className="bg-muted px-1.5 py-0.5 font-mono">kl env status</code> to see
            all services and their ports in the connected environment.
          </p>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/intercepts"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Service Intercepts</p>
            <p className="text-muted-foreground text-sm m-0">Route service traffic to your workspace</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

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
      </div>
    </DocsContentLayout>
  )
}
