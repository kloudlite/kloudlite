import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { APP_MODE } from '@/lib/app-mode'
import { Zap, GitBranch, Package, Terminal } from 'lucide-react'

// Documentation page for website mode
function DocsPage() {
  return (
    <div className="prose prose-slate dark:prose-invert mx-auto max-w-3xl px-4 pt-8 pb-16 sm:px-6 lg:px-8 xl:pr-16">
      {/* Header */}
      <div className="mb-16">
        <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
          Documentation
        </h1>
        <p className="text-muted-foreground mt-4 text-xl">
          Everything you need to get started with Kloudlite
        </p>
      </div>

      {/* Quick Start */}
      <section className="mb-16">
        <h2 className="text-foreground mb-6 text-3xl font-bold">Quick Start</h2>
        <div className="bg-card rounded-lg border p-6">
          <h3 className="text-card-foreground mb-4 text-xl font-semibold">
            Get started with your workspace
          </h3>
          <p className="text-muted-foreground mb-4">
            The Kloudlite CLI (
            <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl</code>) is
            pre-installed in all workspaces. Use it to manage your development environment,
            packages, and service connections.
          </p>
          <ol className="space-y-4">
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                1
              </span>
              <div>
                <p className="text-card-foreground font-medium">Check workspace status</p>
                <div className="bg-muted mt-2 rounded p-3 font-mono text-sm">kl status</div>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                2
              </span>
              <div>
                <p className="text-card-foreground font-medium">Install packages you need</p>
                <div className="bg-muted mt-2 rounded p-3 font-mono text-sm">
                  kl pkg add nodejs python3 golang
                </div>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                3
              </span>
              <div>
                <p className="text-card-foreground font-medium">Start coding</p>
                <p className="text-muted-foreground mt-2 text-sm">
                  Your workspace is ready! All tools and packages are instantly available without
                  any build or setup time.
                </p>
              </div>
            </li>
          </ol>
        </div>
      </section>

      {/* Key Features */}
      <section className="mb-16">
        <h2 className="text-foreground mb-6 text-3xl font-bold">Key Features</h2>
        <div className="grid gap-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-6">
            <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-lg">
              <Zap className="text-primary-foreground h-6 w-6" />
            </div>
            <h3 className="text-card-foreground mb-2 text-xl font-semibold">
              Instant Development Environments
            </h3>
            <p className="text-muted-foreground">
              Spin up fully configured development environments in seconds. No setup, no
              configuration required.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-lg">
              <GitBranch className="text-primary-foreground h-6 w-6" />
            </div>
            <h3 className="text-card-foreground mb-2 text-xl font-semibold">
              Environment Isolation
            </h3>
            <p className="text-muted-foreground">
              Each workspace runs in its own isolated environment with dedicated resources and
              network policies.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-lg">
              <Package className="text-primary-foreground h-6 w-6" />
            </div>
            <h3 className="text-card-foreground mb-2 text-xl font-semibold">Package Management</h3>
            <p className="text-muted-foreground">
              Declarative package management powered by Nix. Reproducible environments across your
              entire team.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-lg">
              <Terminal className="text-primary-foreground h-6 w-6" />
            </div>
            <h3 className="text-card-foreground mb-2 text-xl font-semibold">
              Multiple Access Methods
            </h3>
            <p className="text-muted-foreground">
              Access your workspace via SSH, VS Code, browser-based IDE, or terminal. Work the way
              you prefer.
            </p>
          </div>
        </div>
      </section>

      {/* Core Concepts */}
      <section className="mb-16">
        <h2 className="text-foreground mb-6 text-3xl font-bold">Core Concepts</h2>
        <div className="space-y-6">
          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-xl font-semibold">Workspaces</h3>
            <p className="text-muted-foreground">
              A workspace is your personal development environment. It includes your code, tools,
              dependencies, and configuration. Workspaces are ephemeral and can be started or
              stopped on demand to save resources.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-xl font-semibold">Environments</h3>
            <p className="text-muted-foreground">
              Environments represent different stages of your application (development, staging,
              production). Workspaces can connect to environments to access services and test
              integrations.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-xl font-semibold">Service Intercepts</h3>
            <p className="text-muted-foreground">
              Intercept traffic from production services and route it to your local workspace. Test
              changes against real data without affecting production.
            </p>
          </div>
        </div>
      </section>

      {/* CLI Commands */}
      <section className="mb-16">
        <h2 className="text-foreground mb-6 text-3xl font-bold">Common CLI Commands</h2>
        <div className="bg-card space-y-4 rounded-lg border p-6">
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl init</div>
            <p className="text-muted-foreground text-sm">
              Initialize a new workspace configuration
            </p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl start</div>
            <p className="text-muted-foreground text-sm">Start your workspace</p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl stop</div>
            <p className="text-muted-foreground text-sm">Stop your workspace</p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl status</div>
            <p className="text-muted-foreground text-sm">Check workspace status</p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl pkg add &lt;package&gt;</div>
            <p className="text-muted-foreground text-sm">Install a package in your workspace</p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl pkg remove &lt;package&gt;</div>
            <p className="text-muted-foreground text-sm">Remove a package from your workspace</p>
          </div>
          <div>
            <div className="text-primary mb-1 font-mono text-sm">kl intercept &lt;service&gt;</div>
            <p className="text-muted-foreground text-sm">Intercept traffic from a service</p>
          </div>
        </div>
      </section>

      {/* Deployment Options */}
      <section className="mb-16">
        <h2 className="text-foreground mb-6 text-3xl font-bold">Deployment Options</h2>
        <div className="grid gap-6 md:grid-cols-3">
          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold">AWS</h3>
            <p className="text-muted-foreground text-sm">
              Deploy Kloudlite on Amazon Web Services using EKS for managed Kubernetes.
            </p>
          </div>
          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold">Azure</h3>
            <p className="text-muted-foreground text-sm">
              Run Kloudlite on Microsoft Azure using AKS for enterprise-grade deployments.
            </p>
          </div>
          <div className="bg-card rounded-lg border p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold">GCP</h3>
            <p className="text-muted-foreground text-sm">
              Deploy on Google Cloud Platform using GKE for scalable infrastructure.
            </p>
          </div>
        </div>
      </section>

      {/* Support */}
      <section className="mb-16">
        <div className="bg-card rounded-lg border p-8 text-center">
          <h2 className="text-card-foreground mb-4 text-2xl font-bold">Need Help?</h2>
          <p className="text-muted-foreground mb-6">
            Our team is here to help you get the most out of Kloudlite
          </p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button asChild variant="outline" size="lg">
              <Link href="https://github.com/kloudlite/kloudlite">View on GitHub</Link>
            </Button>
            <Button asChild size="lg">
              <Link href="/contact">Contact Support</Link>
            </Button>
          </div>
        </div>
      </section>
    </div>
  )
}

export default function Page() {
  // Only show docs page in website mode
  if (APP_MODE === 'website') {
    return <DocsPage />
  }

  // Redirect to home for other modes
  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Documentation page is only available in website mode.</p>
    </div>
  )
}
