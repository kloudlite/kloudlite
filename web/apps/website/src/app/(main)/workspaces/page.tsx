import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { ArrowRight, Terminal, Zap, GitBranch, Shield } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

export default function WorkspacesPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="workspaces" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-6">Cloud Development</p>
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Workspaces</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-lg text-lg leading-relaxed">
                    Cloud development environments that connect directly to your running services.
                    Write code, test instantly, ship faster.
                  </p>
                  <div className="mt-10 flex items-center justify-center gap-4">
                    <GetStartedButton size="lg" />
                    <Button variant="outline" size="lg" className="rounded-none" asChild>
                      <Link href="/docs/concepts/workspaces">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Key Stats */}
              <div className="grid grid-cols-2 lg:grid-cols-4 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-6 lg:p-8 border-r border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">&lt;30s</p>
                  <p className="text-foreground/40 mt-1 text-xs">Startup Time</p>
                </div>
                <div className="p-6 lg:p-8 lg:border-r border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">16</p>
                  <p className="text-foreground/40 mt-1 text-xs">vCPU Max</p>
                </div>
                <div className="p-6 lg:p-8 border-r border-t lg:border-t-0 border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">64GB</p>
                  <p className="text-foreground/40 mt-1 text-xs">Memory Max</p>
                </div>
                <div className="p-6 lg:p-8 border-t lg:border-t-0 border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">500GB</p>
                  <p className="text-foreground/40 mt-1 text-xs">Storage Max</p>
                </div>
              </div>

              {/* How It Works */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Workflow</p>
                  <h2 className="text-foreground mt-3 text-2xl font-bold tracking-[-0.02em]">
                    How It Works
                  </h2>
                  <p className="text-foreground/50 mt-2 text-sm">
                    From zero to coding in minutes.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <div className="grid sm:grid-cols-2 gap-6">
                    <div className="flex gap-4">
                      <div className="text-foreground/10 text-2xl font-bold font-mono">01</div>
                      <div>
                        <p className="text-foreground font-semibold">Create Workspace</p>
                        <p className="text-foreground/50 text-sm mt-1">Select your repo and machine tier</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-foreground/10 text-2xl font-bold font-mono">02</div>
                      <div>
                        <p className="text-foreground font-semibold">Connect Environment</p>
                        <p className="text-foreground/50 text-sm mt-1">Link to your running services</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-foreground/10 text-2xl font-bold font-mono">03</div>
                      <div>
                        <p className="text-foreground font-semibold">Start Coding</p>
                        <p className="text-foreground/50 text-sm mt-1">Access via browser, SSH, or IDE</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-foreground/10 text-2xl font-bold font-mono">04</div>
                      <div>
                        <p className="text-foreground font-semibold">Intercept Traffic</p>
                        <p className="text-foreground/50 text-sm mt-1">Route live requests to your code</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Core Features */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Zap className="h-5 w-5 text-foreground/40" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Service Intercepts</h3>
                      <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                        Route traffic from any environment service directly to your workspace. Debug with real production data.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <GitBranch className="h-5 w-5 text-foreground/40" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Workspace Forking</h3>
                      <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                        Clone your workspace for parallel development. Run multiple experiments or AI agents simultaneously.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Terminal className="h-5 w-5 text-foreground/40" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Nix Packages</h3>
                      <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                        Reproducible package management with Nix. Install any tool or dependency with a single command.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Shield className="h-5 w-5 text-foreground/40" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Private Network</h3>
                      <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                        Secure VPN connection to your environments. Access internal services as if you were on the same network.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* CLI Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Developer Experience</p>
                  <h2 className="text-foreground mt-3 text-2xl font-bold tracking-[-0.02em]">
                    CLI First
                  </h2>
                  <p className="text-foreground/50 mt-2 text-sm">
                    Everything from your terminal.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10 bg-foreground/[0.02]">
                  <pre className="text-sm font-mono overflow-x-auto">
                    <code className="text-foreground/70">
                      <span className="text-foreground/40"># Connect to environment</span>{'\n'}
                      <span className="text-primary">kl</span> env connect staging{'\n\n'}
                      <span className="text-foreground/40"># Add packages</span>{'\n'}
                      <span className="text-primary">kl</span> pkg add nodejs go python{'\n\n'}
                      <span className="text-foreground/40"># Intercept a service</span>{'\n'}
                      <span className="text-primary">kl</span> intercept start api-gateway
                    </code>
                  </pre>
                </div>
              </div>

              {/* CTA Section */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12">
                <Link
                  href="/docs/concepts/workspaces"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Learn More</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Read the Docs</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>

                <Link
                  href="/pricing"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Pricing</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">View Plans</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
