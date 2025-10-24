import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { APP_MODE } from '@/lib/app-mode'
import { Zap, GitBranch, Package, Terminal } from 'lucide-react'
import { DocsSidebar } from './_components/docs-sidebar'

// Documentation page for website mode
function DocsPage() {
  return (
    <div className="prose prose-slate mx-auto max-w-3xl px-4 pb-16 pt-8 dark:prose-invert sm:px-6 lg:px-8 xl:pr-16">
          {/* Header */}
          <div className="mb-16">
            <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
              Documentation
            </h1>
            <p className="mt-4 text-xl text-muted-foreground">
              Everything you need to get started with Kloudlite
            </p>
          </div>

          {/* Quick Start */}
          <section className="mb-16">
            <h2 className="mb-6 text-3xl font-bold text-foreground">Quick Start</h2>
            <div className="rounded-lg border bg-card p-6">
              <h3 className="mb-4 text-xl font-semibold text-card-foreground">
                Get started with your workspace
              </h3>
              <p className="mb-4 text-muted-foreground">
                The Kloudlite CLI (<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">kl</code>) is pre-installed in all workspaces. Use it to manage your development environment, packages, and service connections.
              </p>
              <ol className="space-y-4">
                <li className="flex items-start gap-3">
                  <span className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
                    1
                  </span>
                  <div>
                    <p className="font-medium text-card-foreground">Check workspace status</p>
                    <div className="mt-2 rounded bg-muted p-3 font-mono text-sm">
                      kl status
                    </div>
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <span className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
                    2
                  </span>
                  <div>
                    <p className="font-medium text-card-foreground">Install packages you need</p>
                    <div className="mt-2 rounded bg-muted p-3 font-mono text-sm">
                      kl pkg add nodejs python3 golang
                    </div>
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <span className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
                    3
                  </span>
                  <div>
                    <p className="font-medium text-card-foreground">Start coding</p>
                    <p className="mt-2 text-sm text-muted-foreground">
                      Your workspace is ready! All tools and packages are instantly available without any build or setup time.
                    </p>
                  </div>
                </li>
              </ol>
            </div>
          </section>

          {/* Key Features */}
          <section className="mb-16">
            <h2 className="mb-6 text-3xl font-bold text-foreground">Key Features</h2>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="rounded-lg border bg-card p-6">
                <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
                  <Zap className="h-6 w-6 text-primary-foreground" />
                </div>
                <h3 className="mb-2 text-xl font-semibold text-card-foreground">
                  Instant Development Environments
                </h3>
                <p className="text-muted-foreground">
                  Spin up fully configured development environments in seconds. No setup, no configuration required.
                </p>
              </div>

              <div className="rounded-lg border bg-card p-6">
                <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
                  <GitBranch className="h-6 w-6 text-primary-foreground" />
                </div>
                <h3 className="mb-2 text-xl font-semibold text-card-foreground">
                  Environment Isolation
                </h3>
                <p className="text-muted-foreground">
                  Each workspace runs in its own isolated environment with dedicated resources and network policies.
                </p>
              </div>

              <div className="rounded-lg border bg-card p-6">
                <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
                  <Package className="h-6 w-6 text-primary-foreground" />
                </div>
                <h3 className="mb-2 text-xl font-semibold text-card-foreground">
                  Package Management
                </h3>
                <p className="text-muted-foreground">
                  Declarative package management powered by Nix. Reproducible environments across your entire team.
                </p>
              </div>

              <div className="rounded-lg border bg-card p-6">
                <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
                  <Terminal className="h-6 w-6 text-primary-foreground" />
                </div>
                <h3 className="mb-2 text-xl font-semibold text-card-foreground">
                  Multiple Access Methods
                </h3>
                <p className="text-muted-foreground">
                  Access your workspace via SSH, VS Code, browser-based IDE, or terminal. Work the way you prefer.
                </p>
              </div>
            </div>
          </section>

          {/* Core Concepts */}
          <section className="mb-16">
            <h2 className="mb-6 text-3xl font-bold text-foreground">Core Concepts</h2>
            <div className="space-y-6">
              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-xl font-semibold text-card-foreground">Workspaces</h3>
                <p className="text-muted-foreground">
                  A workspace is your personal development environment. It includes your code, tools, dependencies, and
                  configuration. Workspaces are ephemeral and can be started or stopped on demand to save resources.
                </p>
              </div>

              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-xl font-semibold text-card-foreground">Environments</h3>
                <p className="text-muted-foreground">
                  Environments represent different stages of your application (development, staging, production).
                  Workspaces can connect to environments to access services and test integrations.
                </p>
              </div>

              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-xl font-semibold text-card-foreground">Service Intercepts</h3>
                <p className="text-muted-foreground">
                  Intercept traffic from production services and route it to your local workspace. Test changes against
                  real data without affecting production.
                </p>
              </div>
            </div>
          </section>

          {/* CLI Commands */}
          <section className="mb-16">
            <h2 className="mb-6 text-3xl font-bold text-foreground">Common CLI Commands</h2>
            <div className="space-y-4 rounded-lg border bg-card p-6">
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl init</div>
                <p className="text-sm text-muted-foreground">Initialize a new workspace configuration</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl start</div>
                <p className="text-sm text-muted-foreground">Start your workspace</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl stop</div>
                <p className="text-sm text-muted-foreground">Stop your workspace</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl status</div>
                <p className="text-sm text-muted-foreground">Check workspace status</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl pkg add &lt;package&gt;</div>
                <p className="text-sm text-muted-foreground">Install a package in your workspace</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl pkg remove &lt;package&gt;</div>
                <p className="text-sm text-muted-foreground">Remove a package from your workspace</p>
              </div>
              <div>
                <div className="mb-1 font-mono text-sm text-primary">kl intercept &lt;service&gt;</div>
                <p className="text-sm text-muted-foreground">Intercept traffic from a service</p>
              </div>
            </div>
          </section>

          {/* Deployment Options */}
          <section className="mb-16">
            <h2 className="mb-6 text-3xl font-bold text-foreground">Deployment Options</h2>
            <div className="grid gap-6 md:grid-cols-3">
              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-lg font-semibold text-card-foreground">AWS</h3>
                <p className="text-sm text-muted-foreground">
                  Deploy Kloudlite on Amazon Web Services using EKS for managed Kubernetes.
                </p>
              </div>
              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-lg font-semibold text-card-foreground">Azure</h3>
                <p className="text-sm text-muted-foreground">
                  Run Kloudlite on Microsoft Azure using AKS for enterprise-grade deployments.
                </p>
              </div>
              <div className="rounded-lg border bg-card p-6">
                <h3 className="mb-3 text-lg font-semibold text-card-foreground">GCP</h3>
                <p className="text-sm text-muted-foreground">
                  Deploy on Google Cloud Platform using GKE for scalable infrastructure.
                </p>
              </div>
            </div>
          </section>

          {/* Support */}
          <section className="mb-16">
            <div className="rounded-lg border bg-card p-8 text-center">
              <h2 className="mb-4 text-2xl font-bold text-card-foreground">Need Help?</h2>
              <p className="mb-6 text-muted-foreground">
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
