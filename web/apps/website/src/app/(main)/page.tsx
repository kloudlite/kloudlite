'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button, ScrollArea, ClaudeCodeIcon, GeminiIcon, OpenCodeIcon, OpenAIIcon, VSCodeIcon, JetBrainsIcon, AntigravityIcon, ZedIcon, CursorIcon } from '@kloudlite/ui'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { ArrowRight, Copy, Layers, ArrowLeftRight, Route, Package, Plug, Terminal } from 'lucide-react'
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
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      {/* Grid lines */}
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Vertical lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />

        {/* Horizontal lines */}
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

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
    <div className="mt-10">
      {/* Workflow steps */}
      <div className="flex items-center justify-center gap-2 sm:gap-3 flex-wrap">
        {steps.map((step, i) => (
          <div key={step.label} className="flex items-center gap-2 sm:gap-3">
            <div
              className={cn(
                'px-5 py-2.5 rounded-none text-sm font-medium transition-all',
                step.color === 'blue' && 'bg-primary/10 text-primary border-2 border-primary/30',
                step.color === 'green' && 'bg-success/10 text-success border-2 border-success/30',
                step.color === 'gray' && 'bg-muted text-muted-foreground border border-border line-through'
              )}
            >
              {step.label}
            </div>
            {i < steps.length - 1 && (
              <div className="w-4 sm:w-6 h-px bg-foreground/20" />
            )}
          </div>
        ))}
      </div>

      {/* Tagline */}
      <p className="text-muted-foreground text-center mt-10 text-base">
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
                <h1 className="text-[2.5rem] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                  <span className="text-foreground">Cloud Development</span>
                  <br />
                  <span className="text-muted-foreground">Environments</span>
                </h1>

                <p className="text-muted-foreground mx-auto mt-6 max-w-lg text-lg leading-relaxed">
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
              <div className="hidden sm:block mt-16 text-center">
                <p className="text-muted-foreground text-xl sm:text-2xl">
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
              <div className="p-8 lg:p-10 border-b border-foreground/10 sm:border-r flex flex-col justify-center bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                  Built for developers
                </h2>
                <p className="text-muted-foreground mt-3 text-base">
                  Focus on code, not ops.
                </p>
              </div>
              <div className="p-8 lg:p-10 border-b border-foreground/10 lg:border-r bg-foreground/[0.015]">
                <FeatureCard
                  icon={<Copy className="h-5 w-5" />}
                  title="Environment Forking"
                  description="Fork entire environments with a single command."
                />
              </div>
              <div className="p-8 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r-0 bg-foreground/[0.015]">
                <FeatureCard
                  icon={<Layers className="h-5 w-5" />}
                  title="Workspace Forking"
                  description="Fork workspaces instantly for parallel work."
                />
              </div>
              {/* Row 2 */}
              <div className="p-8 lg:p-10 sm:border-r border-b border-foreground/10 bg-foreground/[0.015]">
                <FeatureCard
                  icon={<ArrowLeftRight className="h-5 w-5" />}
                  title="Environment Switching"
                  description="Switch between environments seamlessly."
                />
              </div>
              <div className="p-8 lg:p-10 lg:border-r border-b border-foreground/10 bg-foreground/[0.015]">
                <FeatureCard
                  icon={<Route className="h-5 w-5" />}
                  title="Service Intercepts"
                  description="Route environment service traffic to your workspace."
                />
              </div>
              <div className="p-8 lg:p-10 sm:border-r lg:border-r-0 border-b border-foreground/10 bg-foreground/[0.015]">
                <FeatureCard
                  icon={<Package className="h-5 w-5" />}
                  title="Package Management"
                  description="Nix-based reproducible package management."
                />
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* CTA Row - Start building faster */}
              <div className="p-8 lg:p-10 border-b border-foreground/10 sm:col-span-2 lg:col-span-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 lg:h-[180px] bg-foreground/[0.015]">
                <div>
                  <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                    Start building faster
                  </h2>
                  <p className="text-muted-foreground mt-2 text-base">
                    Free with your own infrastructure.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                  <GetStartedButton className="w-full sm:w-auto" />
                  <Button
                    asChild
                    variant="outline"
                    className="w-full sm:w-auto"
                  >
                    <Link href="/pricing">
                      View Pricing
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </Link>
                  </Button>
                </div>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 right-1/3 translate-y-1/2 translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Row - AI Ready Title */}
              <div className="p-6 lg:p-10 border-b border-foreground/10 sm:border-r flex flex-col justify-center lg:h-[180px] bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                  AI Ready Workspaces
                </h2>
                <p className="text-muted-foreground mt-3 text-base">
                  Supports vibecoding sessions.
                </p>
              </div>
              {/* AI Icons - Single row on mobile, individual cells on desktop */}
              <div className="sm:hidden col-span-1 grid grid-cols-5 border-b border-foreground/10 bg-foreground/[0.015]">
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <ClaudeCodeIcon className="h-5 w-5 transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Claude</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <GeminiIcon className="h-5 w-5 transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Gemini</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <OpenCodeIcon className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">OpenCode</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <OpenAIIcon className="h-5 w-5 text-[#10A37F] transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Codex</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <Plug className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">MCP</span>
                </div>
              </div>
              {/* Claude Code */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <ClaudeCodeIcon className="h-8 w-8 transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Claude Code</span>
              </div>
              {/* Gemini CLI */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <GeminiIcon className="h-8 w-8 transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Gemini CLI</span>
              </div>
              {/* OpenCode */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <OpenCodeIcon className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">OpenCode</span>
              </div>
              {/* Codex */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <OpenAIIcon className="h-8 w-8 text-[#10A37F] transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Codex CLI</span>
              </div>
              {/* MCP Integrated */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r-0 flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <Plug className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">MCP Integrated</span>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-2/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* IDE Integration Section */}
              {/* VS Code */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <VSCodeIcon className="h-8 w-8 text-[#007ACC] transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">VS Code</span>
              </div>
              {/* Cursor */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <CursorIcon className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Cursor</span>
              </div>
              {/* Row - IDE Title */}
              <div className="p-6 lg:p-10 border-b border-foreground/10 flex flex-col justify-center lg:h-[180px] bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                  Access from any IDE
                </h2>
                <p className="text-muted-foreground mt-2 lg:mt-3 text-base">
                  Use your favorite editor.
                </p>
              </div>
              {/* IDE Icons - Single row on mobile */}
              <div className="sm:hidden col-span-1 grid grid-cols-5 border-b border-foreground/10 bg-foreground/[0.015]">
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <VSCodeIcon className="h-5 w-5 text-[#007ACC] transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">VS Code</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <CursorIcon className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Cursor</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <JetBrainsIcon className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">JetBrains</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <AntigravityIcon className="h-5 w-5 transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Antigravity</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <ZedIcon className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Zed</span>
                </div>
              </div>
              {/* Row 2 - More IDEs */}
              {/* JetBrains */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <JetBrainsIcon className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">JetBrains</span>
              </div>
              {/* Antigravity */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 sm:border-r lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <AntigravityIcon className="h-8 w-8 transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Antigravity</span>
              </div>
              {/* Zed */}
              <div className="hidden sm:flex p-6 lg:p-10 border-b border-foreground/10 flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <ZedIcon className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Zed</span>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Access Anywhere Section */}
              {/* Row - Access Title */}
              <div className="p-6 lg:p-10 sm:border-r flex flex-col justify-center lg:h-[180px] bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                  Access anywhere
                </h2>
                <p className="text-muted-foreground mt-2 lg:mt-3 text-base">
                  Work from any device.
                </p>
              </div>
              {/* Access Icons - Single row on mobile */}
              <div className="sm:hidden col-span-1 grid grid-cols-2 bg-foreground/[0.015]">
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <Terminal className="h-5 w-5 text-muted-foreground transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">Web Terminal</span>
                </div>
                <div className="p-3 flex flex-col items-center justify-center gap-1.5 group cursor-default transition-colors hover:bg-foreground/[0.03]">
                  <VSCodeIcon className="h-5 w-5 transition-transform group-hover:scale-110" />
                  <span className="text-muted-foreground text-xs font-medium transition-colors group-hover:text-foreground">VS Code Web</span>
                </div>
              </div>
              {/* Web Terminal */}
              <div className="hidden sm:flex p-6 lg:p-10 sm:border-r lg:border-r flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <Terminal className="h-8 w-8 text-muted-foreground transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">Web Terminal</span>
              </div>
              {/* VS Code Web */}
              <div className="hidden sm:flex p-6 lg:p-10 flex-col items-center justify-center gap-2 lg:gap-3 h-[120px] lg:h-[180px] group cursor-default transition-colors bg-foreground/[0.015] hover:bg-foreground/[0.03]">
                <VSCodeIcon className="h-8 w-8 transition-transform group-hover:scale-110" />
                <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">VS Code Web</span>
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

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="group cursor-default">
      <div className="text-muted-foreground mb-3 transition-colors group-hover:text-foreground">{icon}</div>
      <h3 className="text-foreground text-base font-semibold">{title}</h3>
      <p className="text-muted-foreground mt-2 text-base leading-relaxed transition-colors group-hover:text-foreground">{description}</p>
    </div>
  )
}

export default function HomePage() {
  return <WebsiteLandingPage />
}
