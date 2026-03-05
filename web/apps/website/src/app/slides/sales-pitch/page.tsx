'use client'

import { useState, useEffect, useCallback } from 'react'
import { cn } from '@kloudlite/lib'
import {
  ChevronLeft,
  ChevronRight,
  ChevronUp,
  ChevronDown,
} from 'lucide-react'
import { KloudliteLogo } from '@kloudlite/ui'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

// Animated element wrapper - uses CSS classes for animations
function A({
  children,
  show,
  delay = 0,
  from = 'bottom',
  className,
}: {
  children: React.ReactNode
  show: boolean
  delay?: number
  from?: 'bottom' | 'top' | 'left' | 'right' | 'scale' | 'fade'
  className?: string
}) {
  const baseStyles = 'transition-all duration-[600ms] ease-out'

  const hiddenStyles = {
    bottom: 'opacity-0 translate-y-8',
    top: 'opacity-0 -translate-y-8',
    left: 'opacity-0 -translate-x-8',
    right: 'opacity-0 translate-x-8',
    scale: 'opacity-0 scale-75',
    fade: 'opacity-0',
  }

  const visibleStyles = 'opacity-100 translate-x-0 translate-y-0 scale-100'

  return (
    <div
      className={cn(baseStyles, show ? visibleStyles : hiddenStyles[from], className)}
      style={{ transitionDelay: show ? `${delay}ms` : '0ms' }}
    >
      {children}
    </div>
  )
}

// Position type
type Position = { x: number; y: number }

// Slide component props
interface SlideProps {
  show: boolean
}

// ============================================
// SLIDE COMPONENTS - Fresh Content
// ============================================

// Slide 1: Title
function TitleSlide({ show }: SlideProps) {
  return (
    <div className="text-center">
      <A show={show} delay={0} from="scale">
        <div className="flex justify-center mb-8">
          <KloudliteLogo showText={true} linkToHome={false} className="scale-[2] lg:scale-[2.5]" />
        </div>
      </A>
      <A show={show} delay={200} from="bottom">
        <p className="text-lg lg:text-2xl text-foreground/40 mt-8">
          Cloud Development Environments
        </p>
      </A>
      <A show={show} delay={400} from="bottom">
        <p className="text-xl lg:text-3xl text-foreground/70 mt-4 font-medium">
          Code. Test. Ship. <span className="text-primary">Without the wait.</span>
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <div className="mt-16 flex items-center justify-center gap-8 text-foreground/30 text-sm">
          <span className="flex items-center gap-2">
            <ChevronRight className="h-4 w-4" /> Navigate
          </span>
          <span className="flex items-center gap-2">
            <ChevronDown className="h-4 w-4" /> Details
          </span>
        </div>
      </A>
    </div>
  )
}

// Slide 2: Discovery Questions (vertical stack)
function DiscoveryIntro({ show }: SlideProps) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8">
          Discovery
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-tight mb-6">
          Let's start with<br />
          <span className="text-foreground/40">your current state.</span>
        </h2>
      </A>
      <A show={show} delay={400} from="fade">
        <p className="text-xl text-foreground/50 max-w-lg mx-auto">
          A few questions to understand how your team builds and ships software today.
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <p className="text-foreground/30 text-sm mt-12">
          <ChevronDown className="h-4 w-4 inline mr-1" /> Continue
        </p>
      </A>
    </div>
  )
}

// Reusable question slide component
function QuestionSlide({
  show,
  number,
  question,
  subtext,
  context
}: SlideProps & {
  number: number
  question: string
  subtext: string
  context?: string
}) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8">
          Question {number}
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-3xl lg:text-5xl font-bold tracking-tight mb-6 text-foreground">
          {question}
        </h2>
      </A>
      <A show={show} delay={350} from="fade">
        <p className="text-xl text-foreground/70 mb-8">
          {subtext}
        </p>
      </A>
      {context && (
        <A show={show} delay={550} from="fade">
          <p className="text-primary/70 text-sm max-w-md mx-auto font-medium">
            {context}
          </p>
        </A>
      )}
    </div>
  )
}

function Question1({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={1}
      question="How do new engineers onboard to your codebase?"
      subtext="What does day one look like for a new hire?"
      context="→ How many days until first commit? First production deploy?"
    />
  )
}

function Question2({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={2}
      question="What does your local development setup look like?"
      subtext="Walk me through what runs on a developer's machine."
      context="→ How many services? How long to set up from scratch?"
    />
  )
}

function Question3({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={3}
      question="How do you test changes that touch multiple services?"
      subtext="What's your process for integration testing?"
      context="→ Mocks or real services? How often do mocks drift from reality?"
    />
  )
}

function Question4({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={4}
      question="How do you handle differences between local and CI?"
      subtext="When something works locally but fails in the pipeline."
      context="→ How often does this happen? How long to debug?"
    />
  )
}

function Question5({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={5}
      question="How do engineers share work-in-progress with each other?"
      subtext="Before code is merged or deployed."
      context="→ Can they access each other's running code? Or just review diffs?"
    />
  )
}

function Question6({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={6}
      question="How do non-engineers review features during development?"
      subtext="Product managers, designers, stakeholders."
      context="→ Do they see it live? Or wait for staging deploy?"
    />
  )
}

function Question7({ show }: SlideProps) {
  return (
    <QuestionSlide
      show={show}
      number={7}
      question="Walk me through a typical change from code to production."
      subtext="The steps between writing code and seeing it live."
      context="→ How many steps? How long end-to-end? Where are the waits?"
    />
  )
}

// Slide 3: The Problem
function ProblemOverview({ show }: SlideProps) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-destructive text-sm font-semibold uppercase tracking-wider mb-8">
          The Problem
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-tight mb-6">
          The development cycle<br />
          <span className="text-foreground/40">is broken.</span>
        </h2>
      </A>
      <A show={show} delay={400} from="fade">
        <p className="text-xl text-foreground/70 max-w-lg mx-auto">
          What you just described has a name. And a cost.
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <p className="text-foreground/30 text-sm mt-12">
          <ChevronDown className="h-4 w-4 inline mr-1" /> Let's quantify it
        </p>
      </A>
    </div>
  )
}

function TheProblem({ show }: SlideProps) {
  const steps = ['Code', 'Commit', 'Push', 'Build', 'Deploy', 'Test']

  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-destructive text-sm font-semibold uppercase tracking-wider mb-6">
          The Inner Loop Tax
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h2 className="text-3xl lg:text-5xl font-bold tracking-tight mb-8">
          Every code change requires
        </h2>
      </A>
      <A show={show} delay={300} from="scale">
        <div className="flex flex-wrap items-center justify-center gap-2 mb-8">
          {steps.map((step, i) => (
            <div key={step} className="flex items-center gap-2">
              <div className={cn(
                "px-4 py-2 border text-sm font-medium",
                step === 'Test' ? "border-primary bg-primary/10 text-primary" : "border-foreground/20"
              )}>
                {step}
              </div>
              {i < steps.length - 1 && (
                <ChevronRight className="h-4 w-4 text-foreground/30" />
              )}
            </div>
          ))}
        </div>
      </A>
      <A show={show} delay={600} from="bottom">
        <p className="text-2xl lg:text-3xl font-bold text-destructive mb-4">
          15-20 minutes before you see results
        </p>
      </A>
      <A show={show} delay={800} from="fade">
        <p className="text-foreground/50 text-lg">
          Found a bug? The entire cycle restarts.
        </p>
      </A>
    </div>
  )
}

// Slide 3: The Scale
function TheScale({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-destructive text-sm font-semibold uppercase tracking-wider mb-8">
          The Scale
        </p>
      </A>
      <A show={show} delay={200} from="bottom">
        <h3 className="text-3xl lg:text-4xl font-bold mb-4 text-foreground/50">
          Engineering teams spend
        </h3>
      </A>
      <A show={show} delay={400} from="scale">
        <h3 className="text-6xl lg:text-8xl font-bold text-destructive">
          50%
        </h3>
      </A>
      <A show={show} delay={600} from="bottom">
        <h3 className="text-3xl lg:text-4xl font-bold mt-4 text-foreground/50">
          of their time not writing code.
        </h3>
      </A>
      <A show={show} delay={900} from="bottom">
        <div className="mt-12 grid grid-cols-3 gap-4 max-w-lg mx-auto">
          <div className="p-4 border border-foreground/10">
            <p className="text-2xl font-bold text-foreground/60">10+</p>
            <p className="text-xs text-foreground/40 mt-1">iterations daily</p>
          </div>
          <div className="p-4 border border-foreground/10">
            <p className="text-2xl font-bold text-foreground/60">20m</p>
            <p className="text-xs text-foreground/40 mt-1">per iteration</p>
          </div>
          <div className="p-4 border border-foreground/10">
            <p className="text-2xl font-bold text-foreground/60">3h+</p>
            <p className="text-xs text-foreground/40 mt-1">lost per day</p>
          </div>
        </div>
      </A>
      <A show={show} delay={1100} from="fade">
        <p className="text-foreground/40 mt-10">Multiply by team size.</p>
      </A>
    </div>
  )
}

// Slide 4: Kloudlite Introduction
function IntroKloudlite({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="scale">
        <div className="flex justify-center mb-8">
          <KloudliteLogo showText={true} linkToHome={false} className="scale-[2] lg:scale-[2.5]" />
        </div>
      </A>
      <A show={show} delay={350} from="bottom">
        <p className="text-lg lg:text-xl text-foreground/50 mt-4">
          Cloud Development Environments
        </p>
      </A>
      <A show={show} delay={550} from="bottom">
        <p className="text-xl lg:text-2xl text-foreground/70 mt-8 max-w-2xl mx-auto">
          A platform designed with laser focus on<br />
          <span className="text-primary font-semibold">reducing the development loop.</span>
        </p>
      </A>
      <A show={show} delay={800} from="fade">
        <p className="text-foreground/50 mt-10">
          From code change to validated result—in seconds, not hours.
        </p>
      </A>
    </div>
  )
}

// Slide 5: Principles
function PrinciplesOverview({ show }: SlideProps) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8">
          Our Approach
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-tight mb-6">
          Built on<br />
          <span className="text-foreground/40">first principles.</span>
        </h2>
      </A>
      <A show={show} delay={400} from="fade">
        <p className="text-xl text-foreground/70 max-w-lg mx-auto">
          Every design decision optimizes for one metric: time from code change to validated result.
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <p className="text-foreground/30 text-sm mt-12">
          <ChevronDown className="h-4 w-4 inline mr-1" /> The principles
        </p>
      </A>
    </div>
  )
}

function Principle1({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Principle 1
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-6">
          Eliminate local setup entirely
        </h3>
      </A>
      <A show={show} delay={300} from="fade">
        <p className="text-xl text-foreground/70 mb-8 max-w-2xl mx-auto">
          Cloud workspaces provision in seconds. No installation guides, no dependency hell, no "works on my machine."
        </p>
      </A>
      <A show={show} delay={500} from="scale">
        <div className="inline-flex items-center gap-6 text-left">
          <div className="p-4 border border-foreground/10">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-1">Before</p>
            <p className="text-2xl font-bold text-destructive/60">Days</p>
            <p className="text-foreground/50 text-sm">to set up environment</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5">
            <p className="text-primary text-xs uppercase tracking-wider mb-1">After</p>
            <p className="text-2xl font-bold text-primary">Seconds</p>
            <p className="text-foreground/70 text-sm">to start coding</p>
          </div>
        </div>
      </A>
    </div>
  )
}

function Principle2({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Principle 2
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-6">
          Connect to real services
        </h3>
      </A>
      <A show={show} delay={300} from="fade">
        <p className="text-xl text-foreground/70 mb-8 max-w-2xl mx-auto">
          No mocks. No emulators. Your workspace connects directly to environment services—databases, APIs, queues.
        </p>
      </A>
      <A show={show} delay={500} from="scale">
        <div className="inline-flex items-center gap-6 text-left">
          <div className="p-4 border border-foreground/10">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-1">Before</p>
            <p className="text-2xl font-bold text-destructive/60">Mocks</p>
            <p className="text-foreground/50 text-sm">that drift from reality</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5">
            <p className="text-primary text-xs uppercase tracking-wider mb-1">After</p>
            <p className="text-2xl font-bold text-primary">Real</p>
            <p className="text-foreground/70 text-sm">services, real behavior</p>
          </div>
        </div>
      </A>
    </div>
  )
}

function Principle3({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Principle 3
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-6">
          Test before you commit
        </h3>
      </A>
      <A show={show} delay={300} from="fade">
        <p className="text-xl text-foreground/70 mb-8 max-w-2xl mx-auto">
          Service intercepts route environment traffic to your workspace. Validate changes against real requests—no deployment needed.
        </p>
      </A>
      <A show={show} delay={500} from="scale">
        <div className="inline-flex items-center gap-6 text-left">
          <div className="p-4 border border-foreground/10">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-1">Before</p>
            <p className="text-2xl font-bold text-destructive/60">Deploy</p>
            <p className="text-foreground/50 text-sm">to see if it works</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5">
            <p className="text-primary text-xs uppercase tracking-wider mb-1">After</p>
            <p className="text-2xl font-bold text-primary">Intercept</p>
            <p className="text-foreground/70 text-sm">test with real traffic</p>
          </div>
        </div>
      </A>
    </div>
  )
}

function Principle4({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Principle 4
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-6">
          Share without shipping
        </h3>
      </A>
      <A show={show} delay={300} from="fade">
        <p className="text-xl text-foreground/70 mb-8 max-w-2xl mx-auto">
          Expose your workspace to teammates, QA, or stakeholders. They interact with your running code—before it's merged.
        </p>
      </A>
      <A show={show} delay={500} from="scale">
        <div className="inline-flex items-center gap-6 text-left">
          <div className="p-4 border border-foreground/10">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-1">Before</p>
            <p className="text-2xl font-bold text-destructive/60">Wait</p>
            <p className="text-foreground/50 text-sm">for staging deploy</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5">
            <p className="text-primary text-xs uppercase tracking-wider mb-1">After</p>
            <p className="text-2xl font-bold text-primary">Share</p>
            <p className="text-foreground/70 text-sm">a URL, get feedback</p>
          </div>
        </div>
      </A>
    </div>
  )
}

// Slide 6: Demo Section
function DemoOverview({ show }: SlideProps) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8">
          Demo
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-tight mb-6">
          Let's see it<br />
          <span className="text-foreground/40">in action.</span>
        </h2>
      </A>
      <A show={show} delay={400} from="fade">
        <p className="text-xl text-foreground/70 max-w-lg mx-auto">
          From zero to testing against real services in minutes.
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <p className="text-foreground/30 text-sm mt-12">
          <ChevronDown className="h-4 w-4 inline mr-1" /> Start demo
        </p>
      </A>
    </div>
  )
}

function DemoWorkmachine({ show }: SlideProps) {
  const features = [
    { name: 'VS Code Server', desc: 'Full IDE in browser' },
    { name: 'SSH Access', desc: 'Direct terminal connection' },
    { name: 'Persistent Storage', desc: 'Your files survive restarts' },
    { name: 'Auto Suspend', desc: 'Cost optimization built-in' },
  ]

  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Workmachine
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Your cloud development machine
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          A full Linux environment with everything pre-configured.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {features.map((f) => (
            <div key={f.name} className="p-4 border border-foreground/10 text-left">
              <p className="font-semibold text-sm">{f.name}</p>
              <p className="text-foreground/50 text-xs mt-1">{f.desc}</p>
            </div>
          ))}
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm">
          No local installation. No configuration. Ready in seconds.
        </p>
      </A>
    </div>
  )
}

function DemoEnvironment({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Environment Setup
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Define your services
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          An environment is a collection of services your application depends on.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="border border-foreground/10 p-6 inline-block text-left">
          <p className="text-foreground/40 text-xs uppercase tracking-wider mb-3">Environment: staging</p>
          <div className="flex flex-wrap gap-3">
            <div className="px-4 py-2 bg-primary/10 border border-primary/30 text-sm font-medium">api-gateway</div>
            <div className="px-4 py-2 bg-foreground/5 border border-foreground/20 text-sm">postgres</div>
            <div className="px-4 py-2 bg-foreground/5 border border-foreground/20 text-sm">redis</div>
            <div className="px-4 py-2 bg-foreground/5 border border-foreground/20 text-sm">auth-service</div>
            <div className="px-4 py-2 bg-foreground/5 border border-foreground/20 text-sm">queue</div>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm mt-8">
          Import from Kubernetes, Docker Compose, or define manually.
        </p>
      </A>
    </div>
  )
}

function DemoWorkspace({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Workspace Setup
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Create your workspace
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          A development container with your code and tools.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="border border-foreground/10 p-6 inline-block text-left">
          <p className="text-foreground/40 text-xs uppercase tracking-wider mb-4">Create Workspace</p>
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <span className="text-foreground/40 text-sm w-24">Name</span>
              <div className="px-3 py-1.5 bg-foreground/5 border border-foreground/20 text-sm">feature-auth</div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-foreground/40 text-sm w-24">Repository</span>
              <div className="px-3 py-1.5 bg-foreground/5 border border-foreground/20 text-sm font-mono text-xs">github.com/acme/api</div>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-foreground/40 text-sm w-24">Workmachine</span>
              <div className="px-3 py-1.5 bg-primary/10 border border-primary/30 text-sm text-primary">dev-machine-01</div>
            </div>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm mt-8">
          Runs on your workmachine. Access via VS Code or browser.
        </p>
      </A>
    </div>
  )
}

function DemoPackages({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Package Management
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Install any tool instantly
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Nix-powered package management. 80,000+ packages available.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="bg-foreground/5 border border-foreground/10 p-6 font-mono text-sm text-left inline-block">
          <p className="text-foreground/40"># Add packages</p>
          <p><span className="text-foreground/50">$</span> <span className="text-primary">kl</span> pkg add nodejs python go</p>
          <p className="text-foreground/40 mt-3"># Specific versions</p>
          <p><span className="text-foreground/50">$</span> <span className="text-primary">kl</span> pkg add nodejs@20 python@3.11</p>
          <p className="text-foreground/40 mt-3"># List installed</p>
          <p><span className="text-foreground/50">$</span> <span className="text-primary">kl</span> pkg list</p>
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm mt-8">
          Reproducible across all workspaces. No version conflicts.
        </p>
      </A>
    </div>
  )
}

function DemoConnect({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Connect to Environment
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Link workspace to services
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Your workspace can now access all environment services.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="flex flex-col items-center gap-4">
          <div className="px-6 py-3 border border-foreground/20 bg-foreground/5">
            <p className="font-semibold">Workspace</p>
          </div>
          <div className="flex flex-col items-center text-primary">
            <span className="text-xs text-foreground/50 mb-1">kl env connect</span>
            <span className="font-mono text-lg">↓</span>
          </div>
          <div className="px-8 py-4 border-2 border-primary bg-primary/5">
            <p className="text-sm text-primary/70 mb-2">Environment: staging</p>
            <div className="flex gap-2 text-xs">
              <span className="px-2 py-1 bg-background border border-primary/30">postgres</span>
              <span className="px-2 py-1 bg-background border border-primary/30">redis</span>
              <span className="px-2 py-1 bg-background border border-primary/30">auth</span>
            </div>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="scale">
        <div className="bg-foreground/5 border border-foreground/10 p-4 font-mono text-sm inline-block mt-6">
          <span className="text-foreground/50">$</span> <span className="text-primary">kl</span> env connect staging
        </div>
      </A>
    </div>
  )
}

function DemoIntercept({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Service Intercept
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Route traffic to your code
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Intercept a service. Traffic flows to your workspace instead.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="flex flex-col items-center gap-4">
          <div className="px-8 py-4 border-2 border-primary bg-primary/5">
            <p className="text-sm text-primary/70 mb-2">Environment</p>
            <div className="flex gap-3">
              <div className="px-3 py-1.5 border-2 border-primary bg-primary/20 text-sm">
                <span className="text-primary font-bold">api-gateway</span>
                <span className="text-primary/60 text-xs ml-1">(intercepted)</span>
              </div>
              <div className="px-3 py-1.5 border border-foreground/20 text-sm text-foreground/50">postgres</div>
              <div className="px-3 py-1.5 border border-foreground/20 text-sm text-foreground/50">redis</div>
            </div>
          </div>
          <div className="flex flex-col items-center text-primary">
            <span className="font-mono text-lg">↓ traffic</span>
          </div>
          <div className="px-6 py-3 border-2 border-primary bg-primary/10">
            <p className="text-sm text-primary/70">Your code running in</p>
            <p className="font-bold text-primary">Workspace</p>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="scale">
        <div className="bg-foreground/5 border border-foreground/10 p-4 font-mono text-sm inline-block mt-6">
          <span className="text-foreground/50">$</span> <span className="text-primary">kl</span> intercept start api-gateway --port 3000
        </div>
      </A>
    </div>
  )
}

function DemoCloneEnv({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Fork Environment
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Duplicate any environment
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Create isolated copies for testing, experiments, or new features.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="flex items-center justify-center gap-6">
          <div className="p-4 border border-foreground/20 text-left">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-2">Source</p>
            <p className="font-semibold">staging</p>
            <p className="text-foreground/50 text-xs mt-1">5 services</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5 text-left">
            <p className="text-primary text-xs uppercase tracking-wider mb-2">Clone</p>
            <p className="font-semibold text-primary">feature-auth-v2</p>
            <p className="text-foreground/70 text-xs mt-1">5 services (isolated)</p>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="scale">
        <div className="bg-foreground/5 border border-foreground/10 p-4 font-mono text-sm inline-block mt-8">
          <span className="text-foreground/50">$</span> <span className="text-primary">kl</span> env fork staging --name feature-auth-v2
        </div>
      </A>
      <A show={show} delay={750} from="fade">
        <p className="text-foreground/50 text-sm mt-6">
          Full isolation. Break things without affecting others.
        </p>
      </A>
    </div>
  )
}

function DemoCloneWorkspace({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Demo — Fork Workspace
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Share your exact setup
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Clone a workspace for pair programming, onboarding, or handoffs.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="flex items-center justify-center gap-6">
          <div className="p-4 border border-foreground/20 text-left">
            <p className="text-foreground/40 text-xs uppercase tracking-wider mb-2">Your Workspace</p>
            <p className="font-semibold">alice-feature</p>
            <p className="text-foreground/50 text-xs mt-1">Code + packages + config</p>
          </div>
          <ChevronRight className="h-6 w-6 text-primary" />
          <div className="p-4 border-2 border-primary bg-primary/5 text-left">
            <p className="text-primary text-xs uppercase tracking-wider mb-2">Forked For</p>
            <p className="font-semibold text-primary">bob-review</p>
            <p className="text-foreground/70 text-xs mt-1">Exact same state</p>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="scale">
        <div className="bg-foreground/5 border border-foreground/10 p-4 font-mono text-sm inline-block mt-8">
          <span className="text-foreground/50">$</span> <span className="text-primary">kl</span> workspace fork --name bob-review
        </div>
      </A>
      <A show={show} delay={750} from="fade">
        <p className="text-foreground/50 text-sm mt-6">
          Onboard new engineers in minutes, not days.
        </p>
      </A>
    </div>
  )
}

// Slide 7: AI Features
function AIOverview({ show }: SlideProps) {
  return (
    <div className="w-full max-w-3xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8">
          AI-Powered Development
        </p>
      </A>
      <A show={show} delay={150} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-tight mb-6">
          Built for the<br />
          <span className="text-foreground/40">AI-native workflow.</span>
        </h2>
      </A>
      <A show={show} delay={400} from="fade">
        <p className="text-xl text-foreground/70 max-w-lg mx-auto">
          Your workspace is ready for AI coding agents, automated scans, and intelligent development.
        </p>
      </A>
      <A show={show} delay={700} from="fade">
        <p className="text-foreground/30 text-sm mt-12">
          <ChevronDown className="h-4 w-4 inline mr-1" /> AI features
        </p>
      </A>
    </div>
  )
}

function AIGitWorktree({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          AI Feature — Parallel Workspaces
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Fork workspaces for parallel work
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Each branch gets its own workspace fork. Work on multiple features simultaneously.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="border border-foreground/10 p-6 inline-block text-left">
          <p className="text-foreground/40 text-xs uppercase tracking-wider mb-4">Active Workspace Forks</p>
          <div className="space-y-2">
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-none bg-primary"></div>
              <span className="font-mono text-sm">api-main</span>
              <span className="text-foreground/40 text-xs">branch: main</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-none bg-green-500"></div>
              <span className="font-mono text-sm">api-feature-auth</span>
              <span className="text-foreground/40 text-xs">branch: feature/auth</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-none bg-yellow-500"></div>
              <span className="font-mono text-sm">api-hotfix</span>
              <span className="text-foreground/40 text-xs">branch: hotfix/bug-123</span>
            </div>
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm mt-8">
          AI agents work in isolated forks. No conflicts with your code.
        </p>
      </A>
    </div>
  )
}

function AIAgents({ show }: SlideProps) {
  const agents = [
    { name: 'Claude Code', desc: 'Anthropic AI coding assistant' },
    { name: 'Cursor', desc: 'AI-first code editor' },
    { name: 'GitHub Copilot', desc: 'Code completions & chat' },
    { name: 'Custom Agents', desc: 'Your own AI workflows' },
  ]

  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          AI Feature — Agent Integration
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          AI agents run in your workspace
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Full filesystem access, real services, actual execution—not sandboxed simulations.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {agents.map((agent) => (
            <div key={agent.name} className="p-4 border border-foreground/10 text-left hover:border-primary/30 transition-colors">
              <p className="font-semibold text-sm">{agent.name}</p>
              <p className="text-foreground/50 text-xs mt-1">{agent.desc}</p>
            </div>
          ))}
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm">
          Agents test against real environments. No mocked responses.
        </p>
      </A>
    </div>
  )
}

function AICodeScans({ show }: SlideProps) {
  const scans = [
    { name: 'Security', status: 'passed', count: '0 issues' },
    { name: 'Dependencies', status: 'warning', count: '2 outdated' },
    { name: 'Code Quality', status: 'passed', count: 'A rating' },
    { name: 'License', status: 'passed', count: 'Compliant' },
  ]

  return (
    <div className="w-full max-w-4xl px-8 text-center">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          AI Feature — Code Scans
        </p>
      </A>
      <A show={show} delay={100} from="bottom">
        <h3 className="text-3xl lg:text-5xl font-bold tracking-tight mb-4">
          Automated code analysis
        </h3>
      </A>
      <A show={show} delay={250} from="fade">
        <p className="text-lg text-foreground/70 mb-8">
          Security vulnerabilities, dependency updates, code quality—scanned continuously.
        </p>
      </A>
      <A show={show} delay={400} from="scale">
        <div className="border border-foreground/10 p-6 inline-block text-left">
          <p className="text-foreground/40 text-xs uppercase tracking-wider mb-4">Scan Results</p>
          <div className="space-y-3">
            {scans.map((scan) => (
              <div key={scan.name} className="flex items-center justify-between gap-8">
                <span className="text-sm">{scan.name}</span>
                <div className="flex items-center gap-2">
                  <span className={`text-xs ${scan.status === 'passed' ? 'text-green-500' : 'text-yellow-500'}`}>
                    {scan.count}
                  </span>
                  <div className={`w-2 h-2 rounded-none ${scan.status === 'passed' ? 'bg-green-500' : 'bg-yellow-500'}`}></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </A>
      <A show={show} delay={600} from="fade">
        <p className="text-foreground/50 text-sm mt-8">
          Fix issues before they reach production.
        </p>
      </A>
    </div>
  )
}

// Impact slide - Before/After comparison
function TheImpact({ show }: SlideProps) {
  return (
    <div className="w-full max-w-5xl px-8">
      <A show={show} delay={0} from="fade">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-8 text-center">
          Measurable Impact
        </p>
      </A>
      <div className="grid lg:grid-cols-2 gap-8">
        {/* Before */}
        <A show={show} delay={150} from="left">
          <div className="border border-destructive/20 p-8 h-full">
            <p className="text-destructive text-xs font-semibold uppercase tracking-wider mb-4">Current State</p>
            <div className="space-y-6">
              <div>
                <p className="text-4xl font-bold text-destructive/60">20 min</p>
                <p className="text-foreground/40 text-sm">per feedback cycle</p>
              </div>
              <div className="space-y-2 text-foreground/50">
                <p>Deployment required for validation</p>
                <p>Isolated development workflows</p>
                <p>CI/CD pipeline dependencies</p>
              </div>
            </div>
          </div>
        </A>
        {/* After */}
        <A show={show} delay={300} from="right">
          <div className="border-2 border-primary bg-primary/5 p-8 h-full">
            <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-4">With Kloudlite</p>
            <div className="space-y-6">
              <div>
                <p className="text-4xl font-bold text-primary">2 seconds</p>
                <p className="text-foreground/40 text-sm">to validate changes</p>
              </div>
              <div className="space-y-2 text-foreground/70">
                <p>Pre-commit validation</p>
                <p>Real-time collaboration</p>
                <p>Deployment confidence</p>
              </div>
            </div>
          </div>
        </A>
      </div>
      <A show={show} delay={600} from="bottom">
        <div className="grid grid-cols-3 gap-6 max-w-lg mx-auto mt-10">
          <div className="text-center">
            <p className="text-2xl font-bold text-primary">10x</p>
            <p className="text-foreground/40 text-xs">faster iterations</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-primary">3h+</p>
            <p className="text-foreground/40 text-xs">recovered daily</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-primary">100%</p>
            <p className="text-foreground/40 text-xs">pre-commit coverage</p>
          </div>
        </div>
      </A>
    </div>
  )
}

function PricingOverview({ show }: SlideProps) {
  return (
    <div className="w-full max-w-4xl px-8">
      <A show={show} delay={0} from="bottom">
        <p className="text-primary text-sm font-semibold uppercase tracking-wider mb-4">
          Pricing
        </p>
      </A>
      <A show={show} delay={80} from="bottom">
        <h2 className="text-4xl lg:text-6xl font-bold tracking-[-0.03em] mb-12">
          Simple, transparent
        </h2>
      </A>
      <div className="grid lg:grid-cols-2 gap-6">
        <A show={show} delay={160} from="left">
          <div className="border border-foreground/10 p-8 h-full">
            <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Open Source</p>
            <h3 className="text-2xl font-bold mt-3">Self-Hosted</h3>
            <div className="mt-4">
              <span className="text-6xl font-bold">$0</span>
            </div>
            <p className="text-foreground/50 mt-3">Deploy on your infrastructure</p>
          </div>
        </A>
        <A show={show} delay={260} from="right">
          <div className="border-2 border-primary/30 bg-primary/[0.02] p-8 h-full">
            <p className="text-primary text-xs font-semibold uppercase tracking-wider">Fully Managed</p>
            <h3 className="text-2xl font-bold mt-3">Cloud</h3>
            <div className="mt-4">
              <span className="text-6xl font-bold">$29</span>
              <span className="text-foreground/40 text-xl">/mo</span>
            </div>
            <p className="text-foreground/50 mt-3">+ per-user tier pricing</p>
          </div>
        </A>
      </div>
      <A show={show} delay={450} from="fade">
        <p className="text-foreground/40 text-sm mt-8 text-center">
          <ChevronDown className="h-4 w-4 inline mr-1" /> View tier details
        </p>
      </A>
    </div>
  )
}

function PricingTiers({ show }: SlideProps) {
  const tiers = [
    { name: 'Tier 1', price: 29, specs: '8 vCPUs, 16GB RAM', hours: '160 hrs/mo' },
    { name: 'Tier 2', price: 49, specs: '12 vCPUs, 32GB RAM', hours: '160 hrs/mo' },
    { name: 'Tier 3', price: 89, specs: '16 vCPUs, 64GB RAM', hours: '160 hrs/mo' },
  ]

  return (
    <div className="w-full max-w-5xl px-8">
      <A show={show} delay={0} from="bottom">
        <h3 className="text-3xl font-bold mb-10 text-center">Cloud Tiers</h3>
      </A>
      <div className="grid lg:grid-cols-3 gap-6">
        {tiers.map((tier) => (
          <A key={tier.name} show={show} delay={100 + tiers.indexOf(tier) * 120} from="bottom">
            <div className="border border-foreground/10 p-8 text-center">
              <p className="text-primary text-xs font-semibold uppercase tracking-wider">{tier.name}</p>
              <div className="mt-4">
                <span className="text-5xl font-bold">${tier.price}</span>
                <span className="text-foreground/40">/user/mo</span>
              </div>
              <div className="mt-6 space-y-2 text-foreground/60">
                <p>{tier.specs}</p>
                <p>{tier.hours}</p>
              </div>
            </div>
          </A>
        ))}
      </div>
    </div>
  )
}

function CTASlide({ show }: SlideProps) {
  return (
    <div className="text-center">
      <A show={show} delay={0} from="scale">
        <div className="flex justify-center mb-8">
          <KloudliteLogo showText={true} linkToHome={false} className="scale-[2] lg:scale-[2.5]" />
        </div>
      </A>
      <A show={show} delay={100} from="bottom">
        <p className="text-2xl lg:text-3xl text-foreground/50 mb-2">
          Validate before commit.
        </p>
      </A>
      <A show={show} delay={200} from="scale">
        <p className="text-4xl lg:text-6xl font-bold text-primary mb-8">
          Accelerate delivery.
        </p>
      </A>
      <A show={show} delay={400} from="bottom">
        <p className="text-lg text-foreground/40 mb-12 max-w-md mx-auto">
          Eliminate development cycle latency.<br />
          Enable seamless team collaboration.
        </p>
      </A>
      <div className="flex gap-4 justify-center">
        <A show={show} delay={550} from="left">
          <div className="px-10 py-4 bg-primary text-primary-foreground font-semibold text-lg cursor-pointer hover:opacity-90 transition-opacity">
            Get Started
          </div>
        </A>
        <A show={show} delay={650} from="right">
          <div className="px-10 py-4 border border-foreground/20 text-foreground/60 font-semibold text-lg cursor-pointer hover:border-foreground/40 transition-colors">
            Contact Sales
          </div>
        </A>
      </div>
      <A show={show} delay={800} from="fade">
        <p className="text-foreground/30 text-sm mt-10">kloudlite.io</p>
      </A>
    </div>
  )
}

// Slide grid definition
const slideGrid: React.FC<SlideProps>[][] = [
  [TitleSlide],                                           // 1. Title
  [DiscoveryIntro, Question1, Question2, Question3, Question4, Question5, Question6, Question7], // 2. Discovery questions (vertical)
  [ProblemOverview, TheProblem, TheScale],                 // 3. Problem overview + details
  [IntroKloudlite],                                       // 4. Kloudlite Intro
  [PrinciplesOverview, Principle1, Principle2, Principle3, Principle4], // 5. Principles
  [DemoOverview, DemoWorkmachine, DemoEnvironment, DemoWorkspace, DemoPackages, DemoConnect, DemoIntercept, DemoCloneEnv, DemoCloneWorkspace], // 6. Demo
  [AIOverview, AIGitWorktree, AIAgents, AICodeScans],     // 7. AI Features
  [TheImpact],                                            // 8. Before/After Impact
  [PricingOverview, PricingTiers],                        // 8. Pricing
  [CTASlide],                                             // 9. CTA
]

// Slide container that handles mount animation
function SlideContainer({
  SlideComponent,
  slideKey,
}: {
  SlideComponent: React.FC<SlideProps>
  slideKey: string
}) {
  const [show, setShow] = useState(false)

  useEffect(() => {
    // Reset and trigger animation on slide change
    setShow(false)
    const timer = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        setShow(true)
      })
    })
    return () => cancelAnimationFrame(timer)
  }, [slideKey])

  return <SlideComponent show={show} />
}

export default function SalesPitchPage() {
  const [position, setPosition] = useState<Position>({ x: 0, y: 0 })

  const navigate = useCallback((newX: number, newY: number) => {
    if (newX < 0 || newX >= slideGrid.length) return
    if (newY < 0 || newY >= slideGrid[newX].length) return
    if (newX === position.x && newY === position.y) return
    setPosition({ x: newX, y: newY })
  }, [position])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowRight':
        case ' ':
          e.preventDefault()
          navigate(position.x + 1, 0)
          break
        case 'ArrowLeft':
          e.preventDefault()
          navigate(position.x - 1, 0)
          break
        case 'ArrowDown':
          e.preventDefault()
          navigate(position.x, position.y + 1)
          break
        case 'ArrowUp':
          e.preventDefault()
          navigate(position.x, position.y - 1)
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [navigate, position])

  const currentColumn = slideGrid[position.x]
  const hasUp = position.y > 0
  const hasDown = position.y < currentColumn.length - 1
  const hasLeft = position.x > 0
  const hasRight = position.x < slideGrid.length - 1
  const currentDepth = currentColumn.length

  const CurrentSlide = slideGrid[position.x][position.y]
  const slideKey = `${position.x}-${position.y}`

  return (
    <div className="h-screen bg-background text-foreground overflow-hidden relative">
      {/* Frame */}
      <div className="absolute inset-6 lg:inset-12 pointer-events-none">
        <div className="absolute inset-0 border border-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2" />
      </div>

      {/* Header */}
      <div className="absolute top-2 left-2 lg:top-4 lg:left-4 z-20">
        <KloudliteLogo showText={false} linkToHome={false} />
      </div>

      {/* Position indicator */}
      <div className="absolute top-2 right-2 lg:top-4 lg:right-4 z-20 flex items-center gap-1.5">
        {slideGrid.map((col, colIndex) => (
          <div key={`col-${colIndex}`} className="flex flex-col gap-1">
            {col.map((_, rowIndex) => (
              <button
                key={`col-${colIndex}-row-${rowIndex}`}
                onClick={() => navigate(colIndex, rowIndex)}
                className={cn(
                  'w-2 h-2 rounded-none transition-all duration-300',
                  colIndex === position.x && rowIndex === position.y
                    ? 'bg-primary scale-125'
                    : 'bg-foreground/20 hover:bg-foreground/40'
                )}
              />
            ))}
          </div>
        ))}
      </div>

      {/* Slide content */}
      <div className="absolute inset-6 lg:inset-12 flex items-center justify-center overflow-hidden">
        <SlideContainer
          key={slideKey}
          SlideComponent={CurrentSlide}
          slideKey={slideKey}
        />
      </div>

      {/* Navigation arrows */}
      <div className="absolute bottom-2 lg:bottom-4 left-1/2 -translate-x-1/2 z-20 flex items-center gap-3">
        <button
          onClick={() => navigate(position.x - 1, 0)}
          disabled={!hasLeft}
          className={cn(
            'p-2 transition-all duration-200',
            hasLeft ? 'hover:text-primary' : 'opacity-20 cursor-not-allowed'
          )}
        >
          <ChevronLeft className="h-5 w-5" />
        </button>

        {(hasUp || hasDown) && (
          <div className="flex items-center gap-1 px-2">
            <button
              onClick={() => navigate(position.x, position.y - 1)}
              disabled={!hasUp}
              className={cn(
                'p-1 transition-all duration-200',
                hasUp ? 'hover:text-primary' : 'opacity-20 cursor-not-allowed'
              )}
            >
              <ChevronUp className="h-4 w-4" />
            </button>
            <span className="text-xs text-foreground/40 w-8 text-center">
              {position.y + 1}/{currentDepth}
            </span>
            <button
              onClick={() => navigate(position.x, position.y + 1)}
              disabled={!hasDown}
              className={cn(
                'p-1 transition-all duration-200',
                hasDown ? 'hover:text-primary' : 'opacity-20 cursor-not-allowed'
              )}
            >
              <ChevronDown className="h-4 w-4" />
            </button>
          </div>
        )}

        <button
          onClick={() => navigate(position.x + 1, 0)}
          disabled={!hasRight}
          className={cn(
            'p-2 transition-all duration-200',
            hasRight ? 'hover:text-primary' : 'opacity-20 cursor-not-allowed'
          )}
        >
          <ChevronRight className="h-5 w-5" />
        </button>
      </div>
    </div>
  )
}
