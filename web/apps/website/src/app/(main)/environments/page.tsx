import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Lock, Plug, Copy, Route, Users, Server, Activity, Zap } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'
import { PageHeroTitle } from '@/components/page-hero-title'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20 animate-pulse" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20 animate-pulse" />
    </div>
  )
}

// Feature card components
function FeatureCardContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn("group relative p-8 lg:p-12 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden", className)}>
      <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
      {children}
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="cursor-default">
      <div className="text-muted-foreground mb-4 transition-colors group-hover:text-primary">{icon}</div>
      <h3 className="text-foreground text-lg font-bold">{title}</h3>
      <p className="text-muted-foreground mt-3 text-base leading-relaxed font-medium transition-colors group-hover:text-foreground">{description}</p>
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
                  <PageHeroTitle accentedWord="Environments.">
                    Service Infrastructure
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Isolated sets of services your application depends on. Deploy with Docker Compose, connect from workspaces, and intercept traffic.
                  </p>
                  <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                    <GetStartedButton size="lg" className="w-full sm:w-auto" />
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="/docs/concepts/environments">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Features Grid */}
              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Section Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Isolated service <span className="relative inline-block">
                      <span className="relative z-10">infrastructure.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-2xl">
                    Deploy and manage your application services in isolated environments with full Docker Compose compatibility.
                  </p>
                </div>

                {/* Feature 1 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Server className="h-5 w-5" />}
                    title="Docker Compose Compatible"
                    description="Use your existing compose files. If it runs in Docker, it runs here with zero modifications."
                  />
                </FeatureCardContainer>

                {/* Feature 2 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Lock className="h-5 w-5" />}
                    title="Network Isolation"
                    description="Each environment runs in its own network namespace. No cross-contamination between staging and production."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 1 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 3 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Plug className="h-5 w-5" />}
                    title="Connect from Workspaces"
                    description="Connect any workspace to access services by name. DNS resolution and routing handled automatically."
                  />
                </FeatureCardContainer>

                {/* Feature 4 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Copy className="h-5 w-5" />}
                    title="Clone & Fork"
                    description="Create copies of environments for isolated testing. Fork production to debug without affecting users."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 2 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 5 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Route className="h-5 w-5" />}
                    title="Service Intercepts"
                    description="Route traffic from any service to your workspace. Debug with real requests without redeploying."
                  />
                </FeatureCardContainer>

                {/* Feature 6 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Users className="h-5 w-5" />}
                    title="Team Collaboration"
                    description="Share environments with your team. Multiple developers can connect and work with the same services."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 3 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-2/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 7 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Activity className="h-5 w-5" />}
                    title="Live Monitoring"
                    description="Monitor service health, logs, and metrics in real-time. Get instant visibility into your environment's status."
                  />
                </FeatureCardContainer>

                {/* Feature 8 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Zap className="h-5 w-5" />}
                    title="Instant Deployment"
                    description="Deploy changes instantly without waiting. Push new images and see them running in seconds, not minutes."
                  />
                </FeatureCardContainer>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Docker Compose Showcase */}
                <div className="sm:col-span-2 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] flex flex-col lg:flex-row gap-8 items-center">
                  <div className="flex-1">
                    <div className="text-primary mb-4">
                      <Server className="h-8 w-8" />
                    </div>
                    <h3 className="text-foreground text-3xl lg:text-4xl font-bold mb-4">
                      Docker Compose Ready
                    </h3>
                    <p className="text-muted-foreground text-base lg:text-lg leading-relaxed max-w-xl">
                      Bring your existing Docker Compose files. Define services, networks, and volumes—everything you need to run your stack.
                    </p>
                  </div>

                  <div className="flex-1 w-full">
                    <div className="bg-foreground/[0.03] border border-foreground/10 p-6 rounded-sm">
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
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* CTA Section */}
                <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                  <div>
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                      Ready to deploy your services?
                    </h2>
                    <p className="text-muted-foreground mt-2 text-base lg:text-lg">
                      Create your first environment in minutes.
                    </p>
                  </div>
                  <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                    <GetStartedButton size="lg" className="w-full sm:w-auto" />
                    <Button
                      asChild
                      variant="outline"
                      size="lg"
                      className="w-full sm:w-auto rounded-none"
                    >
                      <Link href="/docs/concepts/environments">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
