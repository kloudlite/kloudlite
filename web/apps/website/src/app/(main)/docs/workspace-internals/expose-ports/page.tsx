import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Globe, Terminal } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'exposing', title: 'Exposing Ports' },
  { id: 'managing', title: 'Managing Exposed Ports' },
]

export default function ExposePortsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Expose Ports
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Expose workspace ports to get <strong className="text-foreground">public URLs</strong> for
          your HTTP services. Share your local development server with teammates or test webhooks
          without deploying.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Globe className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">URL format</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-3">
            Exposed ports get unique URLs with your workspace identifier:
          </p>
          <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
            <pre className="m-0">https://p3000-abc123.subdomain.khost.dev</pre>
          </div>
          <p className="text-muted-foreground text-xs mt-3 m-0">
            The URL includes the port number and a unique hash for your workspace.
          </p>
        </div>
      </section>

      {/* Exposing Ports */}
      <section id="exposing" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Exposing Ports</h2>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Terminal className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">kl expose</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Expose a port to get a public URL. Works with any HTTP service running in your workspace.
          </p>

          <div className="bg-muted p-4 font-mono text-sm overflow-x-auto space-y-2">
            <pre className="m-0">kl expose 3000               # Expose port 3000</pre>
            <pre className="m-0">kl expose 8080               # Expose port 8080</pre>
          </div>
        </div>

        <div className="bg-muted/50 border p-4 mt-6">
          <p className="text-foreground text-sm font-medium mb-1">HTTP only</p>
          <p className="text-muted-foreground text-xs m-0">
            Exposed ports are designed for HTTP/HTTPS services. Other protocols are not supported.
          </p>
        </div>
      </section>

      {/* Managing Exposed Ports */}
      <section id="managing" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Managing Exposed Ports</h2>

        <div className="grid gap-6 md:grid-cols-2">
          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">List exposed ports</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl expose list</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Shows all exposed ports with their public URLs.
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Remove exposed port</h3>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">kl expose remove 3000</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Removes the public URL for the specified port.
            </p>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/workspace-internals/forking-sharing"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Forking Cloning & Sharing Sharing</p>
            <p className="text-muted-foreground text-sm m-0">Share workspaces with your team</p>
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
