import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Route, Terminal } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'starting', title: 'Starting Intercepts' },
  { id: 'managing', title: 'Managing Intercepts' },
]

export default function ServiceInterceptsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Service Intercepts
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Intercept a service to <strong className="text-foreground">route its traffic to your workspace</strong>.
          Requests that would go to the service in the environment are redirected to your local code instead.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Route className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">How it works</h3>
          </div>
          <div className="space-y-3 text-muted-foreground text-sm">
            <p className="m-0">
              <strong className="text-foreground">1.</strong> Start an intercept for a service (e.g., <code className="bg-muted px-1.5 py-0.5 font-mono text-xs">api</code>)
            </p>
            <p className="m-0">
              <strong className="text-foreground">2.</strong> All traffic to that service is redirected to your workspace
            </p>
            <p className="m-0">
              <strong className="text-foreground">3.</strong> Debug with real requests, then stop the intercept to restore normal flow
            </p>
          </div>
        </div>

        <div className="bg-muted/50 border p-4 mt-6">
          <p className="text-foreground text-sm font-medium mb-1">Prerequisite</p>
          <p className="text-muted-foreground text-xs m-0">
            You must first connect to an environment with{' '}
            <code className="bg-muted px-1.5 py-0.5 font-mono">kl env connect</code> before intercepting services.
          </p>
        </div>
      </section>

      {/* Starting Intercepts */}
      <section id="starting" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Starting Intercepts</h2>

        <div className="bg-card border p-6 mb-6">
          <div className="flex items-center gap-3 mb-3">
            <Terminal className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl intercept start</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Start intercepting a service. You&apos;ll be prompted to configure port mapping.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl intercept start           # Interactive service selection</pre>
            <pre className="m-0">kl intercept start api       # Intercept specific service</pre>
          </div>
        </div>

        <div className="bg-card border p-6">
          <h3 className="text-card-foreground text-lg font-semibold mb-3">Port mapping</h3>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Map the service port to your workspace port where your local server runs.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto">
            <pre className="m-0 text-muted-foreground"># Service port 8080 → Workspace port 3000</pre>
            <pre className="m-0 text-muted-foreground"># Traffic to api:8080 now reaches localhost:3000</pre>
          </div>
        </div>
      </section>

      {/* Managing Intercepts */}
      <section id="managing" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Managing Intercepts</h2>

        <div className="space-y-6">
          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">List active intercepts</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl intercept list</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Shows service name, phase (Active/Pending/Failed), and port mappings.
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Check status</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl intercept status api</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Detailed info including workspace pod, port mappings, and start time.
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Stop intercept</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto space-y-2">
              <pre className="m-0">kl intercept stop            # Interactive selection</pre>
              <pre className="m-0">kl intercept stop api        # Stop specific service</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Traffic routes back to the original service within seconds.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/environment-connection"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environment Connection</p>
            <p className="text-muted-foreground text-sm m-0">Connect to environments first</p>
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
