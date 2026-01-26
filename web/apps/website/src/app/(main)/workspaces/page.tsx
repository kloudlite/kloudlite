'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Terminal, Zap, GitBranch, Shield, Package, ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'
import { PageHeroTitle } from '@/components/page-hero-title'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

// Feature card components (from home page pattern)
function FeatureCardContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn("group relative p-8 lg:p-12 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden", className)}>
      {/* Vertical highlight line - animated on hover */}
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
      <p className="text-muted-foreground mt-3 text-base leading-relaxed transition-colors group-hover:text-foreground">{description}</p>
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <style jsx>{`
        @keyframes pulseTopLeftToRight {
          0% { left: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes pulseRightTopToBottom {
          0% { top: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes pulseBottomRightToLeft {
          0% { right: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { right: 100%; opacity: 0; }
        }
        @keyframes pulseLeftBottomToTop {
          0% { bottom: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { bottom: 100%; opacity: 0; }
        }
        .pulse-top {
          animation: pulseTopLeftToRight 4s ease-in-out infinite;
        }
        .pulse-right {
          animation: pulseRightTopToBottom 4s ease-in-out infinite 1s;
        }
        .pulse-bottom {
          animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s;
        }
        .pulse-left {
          animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s;
        }
      `}</style>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Static border lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

        {/* Animated pulses */}
        <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />

        {/* Corner markers */}
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
                  <PageHeroTitle accentedWord="Workspaces.">
                    Cloud Development
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Full-featured development environments with VS Code, Nix package management, and direct access to your running services.
                  </p>
                  <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                    <GetStartedButton size="lg" className="w-full sm:w-auto" />
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="/docs/concepts/workspaces">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Features Grid */}
              <div className="grid sm:grid-cols-2 lg:grid-cols-3 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Feature Section Header */}
                <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Your code editor in <span className="relative inline-block">
                      <span className="relative z-10">the cloud.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-2xl">
                    Develop in cloud-hosted environments with full access to your services and infrastructure.
                  </p>
                </div>

                {/* Feature 1: Service Intercepts */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Zap className="h-5 w-5" />}
                    title="Service Intercepts"
                    description="Route traffic from any environment service directly to your workspace. Debug with real production data."
                  />
                  <Link href="/blog/service-intercepts" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Feature 2: Workspace Forking */}
                <FeatureCardContainer className="border-b border-foreground/10 lg:border-r">
                  <FeatureCard
                    icon={<GitBranch className="h-5 w-5" />}
                    title="Workspace Forking"
                    description="Clone your workspace for parallel development. Run multiple experiments or AI agents simultaneously."
                  />
                  <Link href="/blog/workspace-forking" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Feature 3: Nix Packages */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r lg:border-r-0">
                  <FeatureCard
                    icon={<Terminal className="h-5 w-5" />}
                    title="Nix Package Management"
                    description="Reproducible package management with Nix. Install any tool or dependency with a single command."
                  />
                  <Link href="/blog/nix-package-management" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Feature 4: Private Network */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Shield className="h-5 w-5" />}
                    title="Private Network Access"
                    description="Secure VPN connection to your environments. Access internal services as if you were on the same network."
                  />
                  <Link href="/blog/private-network-access" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Feature 5: Fast Startup */}
                <FeatureCardContainer className="border-b border-foreground/10 lg:border-r">
                  <FeatureCard
                    icon={<Zap className="h-5 w-5" />}
                    title="Sub-30s Startup"
                    description="Workspace environments ready in under 30 seconds. No waiting, no builds, just code."
                  />
                  <Link href="/blog/sub-30s-startup" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Feature 6: Flexible Resources */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r lg:border-r-0">
                  <FeatureCard
                    icon={<Package className="h-5 w-5" />}
                    title="Flexible Resources"
                    description="Scale from 1 vCPU to 16 vCPU and up to 64GB RAM. Choose the right size for your workload."
                  />
                  <Link href="/blog/flexible-resources" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Learn more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Section Spacer */}
                <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* CLI Showcase Card */}
                <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] flex flex-col lg:flex-row gap-8 items-center">
                  {/* Left: Description */}
                  <div className="flex-1">
                    <div className="text-primary mb-4">
                      <Terminal className="h-8 w-8" />
                    </div>
                    <h3 className="text-foreground text-3xl lg:text-4xl font-bold mb-4">
                      CLI First Experience
                    </h3>
                    <p className="text-muted-foreground text-base lg:text-lg leading-relaxed max-w-xl">
                      Manage everything from your terminal. Connect to environments, add packages, and intercept services with simple commands.
                    </p>
                  </div>

                  {/* Right: Code Example */}
                  <div className="flex-1 w-full">
                    <div className="bg-foreground/[0.03] border border-foreground/10 p-6 rounded-sm">
                      <pre className="text-sm font-mono overflow-x-auto">
                        <code className="text-foreground">
                          <span className="text-muted-foreground"># Connect to environment</span>{'\n'}
                          <span className="text-primary">kl</span> env connect staging{'\n\n'}
                          <span className="text-muted-foreground"># Add packages</span>{'\n'}
                          <span className="text-primary">kl</span> pkg add nodejs go python{'\n\n'}
                          <span className="text-muted-foreground"># Intercept a service</span>{'\n'}
                          <span className="text-primary">kl</span> intercept start api-gateway
                        </code>
                      </pre>
                    </div>
                  </div>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* CTA Section */}
                <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 lg:col-span-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                  <div>
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                      Ready to start coding?
                    </h2>
                    <p className="text-muted-foreground mt-2 text-base lg:text-lg">
                      Create your first workspace in minutes.
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
                      <Link href="/pricing">
                        View Pricing
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
