import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { ArrowRight, Lock, Plug, Copy, Route, Users } from 'lucide-react'
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

export default function EnvironmentsPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="environments" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-6">Service Infrastructure</p>
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Environments</span>
                  </h1>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-lg text-lg leading-relaxed">
                    Isolated sets of services your application depends on.
                    Databases, caches, APIs—all running and accessible.
                  </p>
                  <div className="mt-10 flex items-center justify-center gap-4">
                    <GetStartedButton size="lg" />
                    <Button variant="outline" size="lg" className="rounded-none" asChild>
                      <Link href="/docs/concepts/environments">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Docker Compose Compatible */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">Compatible With</p>
                  <h2 className="text-foreground mt-3 text-2xl font-bold tracking-[-0.02em]">
                    Docker Compose
                  </h2>
                  <p className="text-muted-foreground mt-2 text-sm">
                    If it runs in Docker, it runs here.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10 bg-foreground/[0.02]">
                  <pre className="text-sm font-mono overflow-x-auto">
                    <code className="text-foreground">
                      <span className="text-muted-foreground"># compose.yaml</span>{'\n'}
                      <span className="text-primary">services</span>:{'\n'}
                      {'  '}<span className="text-foreground">api</span>:{'\n'}
                      {'    '}image: my-api:latest{'\n'}
                      {'    '}ports: [&quot;8080:8080&quot;]{'\n'}
                      {'  '}<span className="text-foreground">db</span>:{'\n'}
                      {'    '}image: postgres:16{'\n'}
                      {'  '}<span className="text-foreground">redis</span>:{'\n'}
                      {'    '}image: redis:alpine
                    </code>
                  </pre>
                </div>
              </div>

              {/* Core Properties */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Lock className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Isolated</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Each environment runs in its own network namespace. No cross-contamination between staging and production.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Plug className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Accessible</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Connect from any workspace to access services by name. No complex networking setup required.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Copy className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Clonable</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Fork environments to get your own isolated copy with the same services and data. Perfect for experimentation.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Route className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Interceptable</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Route service traffic to your workspace for debugging. Test with real requests without deploying.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Use Cases */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">Use Cases</p>
                  <h2 className="text-foreground mt-3 text-2xl font-bold tracking-[-0.02em]">
                    How Teams Use Environments
                  </h2>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <div className="grid sm:grid-cols-2 gap-6">
                    <div className="flex gap-4">
                      <div className="text-muted-foreground/20 text-2xl font-bold font-mono">01</div>
                      <div>
                        <p className="text-foreground font-semibold">Shared Staging</p>
                        <p className="text-muted-foreground text-sm mt-1">Team-wide environment for integration testing</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-muted-foreground/20 text-2xl font-bold font-mono">02</div>
                      <div>
                        <p className="text-foreground font-semibold">Personal Copies</p>
                        <p className="text-muted-foreground text-sm mt-1">Clone for isolated feature development</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-muted-foreground/20 text-2xl font-bold font-mono">03</div>
                      <div>
                        <p className="text-foreground font-semibold">PR Previews</p>
                        <p className="text-muted-foreground text-sm mt-1">Spin up environments per pull request</p>
                      </div>
                    </div>
                    <div className="flex gap-4">
                      <div className="text-muted-foreground/20 text-2xl font-bold font-mono">04</div>
                      <div>
                        <p className="text-foreground font-semibold">Load Testing</p>
                        <p className="text-muted-foreground text-sm mt-1">Isolated environments for performance tests</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Sharing */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Users className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Share with Team</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Make an environment visible to other developers. They can connect their workspaces and access the same services with the same data.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <div className="flex items-start gap-4">
                    <div className="p-2 border border-foreground/10">
                      <Copy className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <h3 className="text-foreground text-lg font-bold tracking-[-0.02em]">Fork for Isolation</h3>
                      <p className="text-muted-foreground mt-2 text-sm leading-relaxed group-hover:text-foreground transition-colors">
                        Fork an environment to get your own copy with the same configuration. Test changes without affecting others.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* CTA Section */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12">
                <Link
                  href="/docs/concepts/environments"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">Learn More</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Read the Docs</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-muted-foreground group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>

                <Link
                  href="/workspaces"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">Next</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Explore Workspaces</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-muted-foreground group-hover:translate-x-1 transition-all" />
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
