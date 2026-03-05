'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button, ScrollArea, VSCodeIcon, JetBrainsIcon, AntigravityIcon, ZedIcon, CursorIcon } from '@kloudlite/ui'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { ArrowRight, Copy, Layers, ArrowLeftRight, Route, Package, Terminal, Camera, Sparkles, Code2 } from 'lucide-react'
import { cn } from '@kloudlite/lib'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      {/* Vertical line */}
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      {/* Horizontal line */}
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

// Grid container like Vercel
function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-7xl', className)}>
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
      {/* Grid lines */}
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Vertical lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />

        {/* Horizontal lines */}
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

      {/* Content */}
      <div className="relative">
        {children}
      </div>
    </div>
  )
}

// Typewriter effect for developer roles
function TypewriterText() {
  const roles = [
    'Frontend Developer',
    'Backend Developer',
    'Full Stack Developer',
    'DevOps Engineer',
    'Platform Engineer',
  ]

  const [currentRoleIndex, setCurrentRoleIndex] = useState(0)
  const [currentText, setCurrentText] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    const currentRole = roles[currentRoleIndex]

    const timeout = setTimeout(() => {
      if (!isDeleting) {
        if (currentText.length < currentRole.length) {
          setCurrentText(currentRole.slice(0, currentText.length + 1))
        } else {
          setTimeout(() => setIsDeleting(true), 2000)
        }
      } else {
        if (currentText.length > 0) {
          setCurrentText(currentText.slice(0, -1))
        } else {
          setIsDeleting(false)
          setCurrentRoleIndex((prev) => (prev + 1) % roles.length)
        }
      }
    }, isDeleting ? 50 : 100)

    return () => clearTimeout(timeout)
  }, [currentText, isDeleting, currentRoleIndex])

  return (
    <span className="text-primary">
      {currentText}
      <span className="animate-pulse">|</span>
    </span>
  )
}

// Testimonials data
const testimonials = [
  {
    quote: "Kloudlite reduced our environment setup time from hours to minutes. Our developers can now focus on shipping features instead of debugging local configs.",
    name: "Sarah Johnson",
    title: "Engineering Lead",
    company: "TechCorp",
    initials: "SJ"
  },
  {
    quote: "The service intercept feature is a game changer. We can test against production services without the risk. It's like magic.",
    name: "Michael Park",
    title: "Senior Developer",
    company: "StartupX",
    initials: "MP"
  },
  {
    quote: "Environment forking changed how we work. Every developer can spin up their own copy and work in parallel. No more waiting.",
    name: "Emily Rodriguez",
    title: "CTO",
    company: "BuildFast",
    initials: "ER"
  },
  {
    quote: "The AI-ready workspaces with Claude Code integration saved us weeks of setup. Our team is shipping faster than ever.",
    name: "David Chen",
    title: "VP Engineering",
    company: "CodeLabs",
    initials: "DC"
  },
  {
    quote: "Nix-based package management ensures every developer has the exact same environment. No more 'works on my machine' issues.",
    name: "Lisa Williams",
    title: "DevOps Lead",
    company: "CloudNative Inc",
    initials: "LW"
  },
  {
    quote: "We moved from local Docker to Kloudlite and our onboarding time dropped from 3 days to 30 minutes. Incredible.",
    name: "James Miller",
    title: "Head of Platform",
    company: "DataStream",
    initials: "JM"
  }
]

// Development workflow visualization
function WorkflowVisualization() {
  const steps = [
    { label: 'Setup', active: false, color: 'gray' },
    { label: 'Code', active: true, color: 'blue' },
    { label: 'Build', active: false, color: 'gray' },
    { label: 'Deploy', active: false, color: 'gray' },
    { label: 'Test', active: true, color: 'green' },
  ]

  return (
    <div className="mt-16">
      {/* Workflow steps */}
      <div className="flex items-center justify-center gap-3 sm:gap-4 flex-wrap">
        {steps.map((step, i) => (
          <div key={step.label} className="flex items-center gap-3 sm:gap-4">
            <div
              className={cn(
                'px-6 py-3 text-sm font-semibold tracking-wide uppercase transition-all',
                step.color === 'blue' && 'bg-primary/5 text-primary border border-primary/20',
                step.color === 'green' && 'bg-success/5 text-success border border-success/20',
                step.color === 'gray' && 'bg-foreground/[0.02] text-muted-foreground border border-foreground/10 line-through opacity-50'
              )}
            >
              {step.label}
            </div>
            {i < steps.length - 1 && (
              <div className="w-6 sm:w-8 h-px bg-foreground/20" />
            )}
          </div>
        ))}
      </div>

      {/* Tagline */}
      <p className="text-muted-foreground text-center mt-12 text-lg font-medium">
        Designed to reduce development loop
      </p>
    </div>
  )
}

function WebsiteLandingPage() {

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="home" />
        <main>
        {/* Main Content in Grid Container */}
        <div className="px-6 pt-8 lg:px-8 lg:pt-12">
          <GridContainer className="px-6 lg:px-12">
            {/* Hero Section */}
            <div className="py-20 lg:py-28">
              <div className="text-center">
                <PageHeroTitle accentedWord="Environments.">
                  Cloud Development
                </PageHeroTitle>

                <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-base lg:text-lg leading-relaxed">
                  Reduce the development loop. No setup, no builds, no deployments.
                </p>

                <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                  <GetStartedButton size="lg" className="w-full sm:w-auto" />
                  <Button
                    asChild
                    variant="outline"
                    size="lg"
                    className="w-full sm:w-auto"
                  >
                    <Link href="/docs">
                      Documentation
                    </Link>
                  </Button>
                </div>
              </div>

              {/* For Developer typewriter - Hidden on mobile */}
              <div className="hidden sm:block mt-20 text-center">
                <p className="text-muted-foreground text-lg sm:text-xl font-medium">
                  For <TypewriterText />
                </p>
              </div>

              {/* Workflow Visualization - Hidden on mobile */}
              <div className="hidden lg:block">
                <WorkflowVisualization />
              </div>
            </div>

            {/* Features Grid */}
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 border-t border-foreground/10 -mx-6 lg:-mx-12">
              {/* Row 1 - Built for developers */}
              <div className="sm:col-span-2 lg:col-span-1 p-8 lg:p-12 border-l border-b border-foreground/10 lg:border-r flex flex-col justify-center bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-3xl lg:text-4xl">
                  Built for developers
                </h2>
                <p className="text-muted-foreground mt-6 text-base lg:text-lg">
                  Focus on code, not ops.
                </p>
              </div>
              <FeatureCardContainer className="border-l border-b border-foreground/10 sm:border-r lg:border-r" href="/blog/environment-forking">
                <FeatureCard
                  icon={<Copy className="h-7 w-7" />}
                  title="Environment Forking"
                  description="Fork entire environments with a single command."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l border-b border-foreground/10 sm:border-r" href="/blog/workspace-forking">
                <FeatureCard
                  icon={<Layers className="h-7 w-7" />}
                  title="Workspace Forking"
                  description="Fork workspaces instantly for parallel work."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              {/* Row 2 */}
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/environment-switching">
                <FeatureCard
                  icon={<ArrowLeftRight className="h-7 w-7" />}
                  title="Environment Switching"
                  description="Switch contexts between environments without losing any state."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/service-intercepts">
                <FeatureCard
                  icon={<Route className="h-7 w-7" />}
                  title="Service Intercepts"
                  description="Route environment service traffic to your workspace."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r border-b border-foreground/10" href="/blog/nix-package-management">
                <FeatureCard
                  icon={<Package className="h-7 w-7" />}
                  title="Package Management"
                  description="Nix-based reproducible package management."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              {/* Row 3 */}
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/environment-snapshots">
                <FeatureCard
                  icon={<Camera className="h-7 w-7" />}
                  title="Environment Snapshots"
                  description="Capture and restore complete environment states instantly."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/workspace-snapshots">
                <FeatureCard
                  icon={<Camera className="h-7 w-7" />}
                  title="Workspace Snapshots"
                  description="Save and share workspace configurations effortlessly."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <div className="group relative p-8 lg:p-10 bg-background hover:bg-foreground/[0.02] transition-all duration-500 overflow-hidden border-l border-b border-foreground/10 sm:border-r flex items-center justify-center">
                {/* Subtle bottom accent line */}
                <div className="absolute bottom-0 left-0 w-0 h-[1px] bg-primary group-hover:w-full transition-all duration-700 ease-out" />

                <Link
                  href="/docs"
                  className="flex flex-col items-center gap-5 text-center"
                >
                  <div className="text-muted-foreground group-hover:text-primary transition-colors duration-500">
                    <ArrowRight className="h-8 w-8 opacity-60 group-hover:opacity-100 transition-all duration-500 group-hover:translate-x-2" />
                  </div>
                  <div className="space-y-2">
                    <span className="block text-lg font-semibold text-foreground group-hover:text-primary transition-colors duration-500">
                      Explore All Features
                    </span>
                    <span className="block text-sm text-muted-foreground">
                      Discover the complete platform
                    </span>
                  </div>
                </Link>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Toolchain Section Header */}
              <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                <h2 className="text-foreground text-4xl lg:text-4xl xl:text-5xl font-bold tracking-tight">
                  Your entire toolchain, <span className="relative inline-block">
                    <span className="relative z-10">connected.</span>
                    <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                  </span>
                </h2>
                <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-2xl">
                  From AI agents to your favorite IDE, Kloudlite works where you work.
                </p>
              </div>

              {/* AI Ready Workspaces Card */}
              <div className="sm:col-span-2 lg:col-span-1 p-8 lg:p-12 border-b lg:border-r border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] lg:min-h-[500px] flex flex-col">
                {/* Background decorative icon */}
                <div className="absolute top-8 right-8 opacity-10">
                  <Sparkles className="h-32 w-32 text-muted-foreground" />
                </div>

                <div className="relative z-10">
                  <div className="text-primary mb-4">
                    <Sparkles className="h-8 w-8" />
                  </div>
                  <h3 className="text-foreground text-2xl font-bold mb-4">
                    AI Ready Workspaces
                  </h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Built-in support for the next generation of AI coding tools. Supports vibecoding sessions out of the box.
                  </p>
                  <Link href="/blog/ai-ready-workspaces" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-100 md:opacity-0 md:group-hover:opacity-100">
                    Learn more <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>

                {/* AI Tool Tags */}
                <div className="mt-auto relative z-10 flex flex-wrap gap-2">
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">CLAUDE CODE</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">GEMINI CLI</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">OPENCODE</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">CODEX CLI</span>
                </div>
              </div>

              {/* Access from any IDE Card */}
              <div className="sm:col-span-2 lg:col-span-2 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] lg:min-h-[500px] flex flex-col">
                {/* Background decorative icon */}
                <div className="absolute top-8 right-8 opacity-10">
                  <Code2 className="h-32 w-32 text-muted-foreground" />
                </div>

                <div className="relative z-10 mb-8">
                  <div className="text-primary mb-4">
                    <Code2 className="h-8 w-8" />
                  </div>
                  <h3 className="text-foreground text-2xl font-bold mb-4">
                    Access from any IDE
                  </h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Connect your local editor directly to cloud resources. Zero latency, full Intellisense.
                  </p>
                  <Link href="/blog/ide-integration" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-100 md:opacity-0 md:group-hover:opacity-100">
                    Learn more <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>

                {/* IDE Grid */}
                <div className="mt-auto relative z-10 grid grid-cols-2 lg:grid-cols-3 gap-3">
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <VSCodeIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">VS Code</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <CursorIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Cursor</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <JetBrainsIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">JetBrains</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <AntigravityIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Antigravity</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <ZedIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Zed</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <Terminal className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Web Terminal</span>
                  </div>
                </div>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Testimonials Section Header */}
              <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                <h2 className="text-foreground text-4xl lg:text-4xl xl:text-5xl font-bold tracking-tight">
                  Trusted by developers
                </h2>
                <p className="text-muted-foreground mt-4 text-base lg:text-base xl:text-lg max-w-2xl">
                  Teams around the world are building faster with Kloudlite.
                </p>
              </div>

              {/* Testimonials Grid - Rotating Content */}
              {testimonials.slice(0, 3).map((testimonial, index) => (
                <div
                  key={testimonial.name}
                  className={cn(
                    "p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-all duration-300 group cursor-default",
                    "sm:col-span-2 lg:col-span-1",
                    index === 0 && "sm:border-r lg:border-r",
                    index === 1 && "lg:border-r"
                  )}
                >
                  <div className="mb-8 min-h-[140px] flex items-center">
                    <p className="text-foreground text-base leading-relaxed">
                      "{testimonial.quote}"
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <span className="text-primary font-bold text-sm">{testimonial.initials}</span>
                    </div>
                    <div>
                      <p className="text-foreground font-semibold text-sm">{testimonial.name}</p>
                      <p className="text-muted-foreground text-xs">{testimonial.title}, {testimonial.company}</p>
                    </div>
                  </div>
                </div>
              ))}

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* CTA Row - Start building faster */}
              <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 lg:col-span-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                <div>
                  <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-3xl xl:text-4xl">
                    Start building faster
                  </h2>
                  <p className="text-muted-foreground mt-2 text-base">
                    Free with your own infrastructure.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                  <GetStartedButton size="lg" className="w-full sm:w-auto" />
                  <Button
                    asChild
                    variant="outline"
                    size="lg"
                    className="w-full sm:w-auto"
                  >
                    <Link href="/pricing">
                      Pricing
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

function FeatureCardContainer({ children, className, href }: { children: React.ReactNode; className?: string; href?: string }) {
  const content = (
    <>
      {/* Subtle bottom accent line */}
      <div className="absolute bottom-0 left-0 w-0 h-[1px] bg-primary group-hover:w-full transition-all duration-700 ease-out" />

      <div className="relative flex flex-col">
        {children}
      </div>
    </>
  )

  if (href) {
    return (
      <Link href={href} className={cn("group relative p-8 lg:p-10 bg-background hover:bg-foreground/[0.02] transition-all duration-500 overflow-hidden border-b border-foreground/10 cursor-pointer", className)}>
        {content}
      </Link>
    )
  }

  return (
    <div className={cn("group relative p-8 lg:p-10 bg-background hover:bg-foreground/[0.02] transition-all duration-500 overflow-hidden border-b border-foreground/10", className)}>
      {content}
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="flex flex-col h-full space-y-3">
      {/* Icon - minimal, just the icon */}
      <div className="text-muted-foreground group-hover:text-primary transition-colors duration-500">
        <div className="w-8 h-8 opacity-60 group-hover:opacity-100 transition-all duration-500 group-hover:scale-110">
          {icon}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-4">
        <h3 className="text-foreground font-bold text-xl sm:text-2xl tracking-tight leading-tight group-hover:text-primary transition-colors duration-500">
          {title}
        </h3>

        <p className="text-muted-foreground text-sm sm:text-base leading-relaxed">
          {description}
        </p>
      </div>
    </div>
  )
}

export default function HomePage() {
  return <WebsiteLandingPage />
}
