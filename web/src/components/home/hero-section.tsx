'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ArrowRight, Terminal, GitBranch, Zap, Settings, Network, Package, ArrowUpDown, Code2, Bot, Users, Briefcase, Shield, Cloud } from 'lucide-react'

const projectTypes = [
  'DevOps',
  'Backend',
  'Full Stack',
  'Frontend',
  'Data Analysis',
  'AI/ML',
  'QA Testing',
  'Microservices',
  'API Development',
  'Mobile Apps'
]

export function HeroSection() {
  const [currentProject, setCurrentProject] = useState(0)
  const [displayedText, setDisplayedText] = useState('')
  const [isTyping, setIsTyping] = useState(true)

  useEffect(() => {
    const project = projectTypes[currentProject]
    
    if (isTyping) {
      if (displayedText.length < project.length) {
        const timeout = setTimeout(() => {
          setDisplayedText(project.slice(0, displayedText.length + 1))
        }, 100)
        return () => clearTimeout(timeout)
      } else {
        const timeout = setTimeout(() => {
          setIsTyping(false)
        }, 2000)
        return () => clearTimeout(timeout)
      }
    } else {
      if (displayedText.length > 0) {
        const timeout = setTimeout(() => {
          setDisplayedText(displayedText.slice(0, -1))
        }, 50)
        return () => clearTimeout(timeout)
      } else {
        setCurrentProject((prev) => (prev + 1) % projectTypes.length)
        setIsTyping(true)
      }
    }
  }, [currentProject, isTyping, displayedText])

  return (
    <section className="relative px-6 py-24 lg:py-32">
      <div className="max-w-6xl mx-auto">
        <div className="max-w-3xl mx-auto text-center">
          {/* Main headline */}
          <h1 className="text-5xl lg:text-6xl font-bold tracking-tight">
            Kloudlite
            <span className="text-primary"> Development Environments</span>
          </h1>
          
          {/* Subheadline */}
          <p className="mt-6 text-xl leading-relaxed">
            <span className="text-2xl font-bold text-foreground">10x faster development</span>
            <br />
            <span className="text-muted-foreground">
              For{' '}
              <span className="text-primary font-semibold">
                {displayedText}
                <span className="animate-pulse">|</span>
              </span>
              {' '}developers
            </span>
          </p>

          {/* Key message */}
          <div className="mt-8">
            <p className="text-sm uppercase tracking-wider text-muted-foreground mb-4">Focus on what matters</p>
            <div className="flex items-center justify-center gap-2">
              <div className="flex items-center gap-2">
                <div className="px-4 py-2 bg-green-500/10 border border-green-500/20 text-green-600 dark:text-green-400 font-mono text-lg">
                  Code
                </div>
                <span className="text-muted-foreground">→</span>
                <div className="px-4 py-2 bg-muted/50 border border-border text-muted-foreground font-mono text-lg">
                  <span className="line-through">Build</span>
                </div>
                <span className="text-muted-foreground">→</span>
                <div className="px-4 py-2 bg-muted/50 border border-border text-muted-foreground font-mono text-lg">
                  <span className="line-through">Deploy</span>
                </div>
                <span className="text-muted-foreground">→</span>
                <div className="px-4 py-2 bg-green-500/10 border border-green-500/20 text-green-600 dark:text-green-400 font-mono text-lg">
                  Test
                </div>
              </div>
            </div>
          </div>

          {/* CTA buttons */}
          <div className="mt-12 flex flex-wrap gap-4 justify-center">
            <Button size="lg" asChild>
              <Link href="/auth/signup">
                Get started
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link href="/docs">
                Documentation
              </Link>
            </Button>
          </div>
        </div>


        {/* Key Features Grid */}
        <div className="mt-32">
          <div className="text-center mb-12 space-y-2">
            <h2 className="text-3xl font-bold">
              Every feature, one purpose:
            </h2>
            <p className="text-2xl text-primary font-semibold">
              Accelerate your development loop
            </p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Feature 1 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Settings className="h-8 w-8 text-primary mb-4 group-hover:rotate-90 transition-transform duration-500" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Zero Config Workspaces</h3>
              <p className="text-sm text-muted-foreground">From git clone to running code in under 30 seconds - no setup scripts, no dependency hell</p>
            </div>

            {/* Feature 2 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Network className="h-8 w-8 text-primary mb-4 group-hover:scale-110 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Connected Environments</h3>
              <p className="text-sm text-muted-foreground">Your code runs in the same network as your services - seamless service discovery</p>
            </div>

            {/* Feature 3 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Package className="h-8 w-8 text-primary mb-4 group-hover:-rotate-12 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Config Management</h3>
              <p className="text-sm text-muted-foreground">Configs and secrets auto-injected - no more .env files or manual setup</p>
            </div>

            {/* Feature 4 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Zap className="h-8 w-8 text-primary mb-4 group-hover:animate-pulse transition-all duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Lightweight Environments</h3>
              <p className="text-sm text-muted-foreground">Spin up 10 environments for the cost of 1 VM - instant start, minimal overhead</p>
            </div>

            {/* Feature 5 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <ArrowUpDown className="h-8 w-8 text-primary mb-4 group-hover:translate-y-1 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Service Intercepts</h3>
              <p className="text-sm text-muted-foreground">Debug production issues locally - intercept any service traffic with one command</p>
            </div>

            {/* Feature 6 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Code2 className="h-8 w-8 text-primary mb-4 group-hover:skew-y-3 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">IDE Integration</h3>
              <p className="text-sm text-muted-foreground">Native extensions for VS Code, JetBrains - debug, test, and deploy without leaving your IDE</p>
            </div>

            {/* Feature 7 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Bot className="h-8 w-8 text-primary mb-4 group-hover:animate-bounce transition-all duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">AI Agent Ready</h3>
              <p className="text-sm text-muted-foreground">Native support for autonomous AI coding agents - Cline, Zencoder, and all major AI tools</p>
            </div>

            {/* Feature 8 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Users className="h-8 w-8 text-primary mb-4 group-hover:scale-105 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Instant Collaboration</h3>
              <p className="text-sm text-muted-foreground">Share a live environment URL - teammates can code together in real-time, no setup needed</p>
            </div>

            {/* Feature 9 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Briefcase className="h-8 w-8 text-primary mb-4 group-hover:rotate-6 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Remote-First Design</h3>
              <p className="text-sm text-muted-foreground">Same performance whether you're in San Francisco or São Paulo - optimized for global teams</p>
            </div>

            {/* Feature 10 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Shield className="h-8 w-8 text-primary mb-4 group-hover:scale-110 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Enterprise Security</h3>
              <p className="text-sm text-muted-foreground">SOC2, HIPAA compliant - encrypted at rest and in transit, with full audit trails</p>
            </div>

            {/* Feature 11 */}
            <div className="group relative p-6 border border-border hover:border-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              <Cloud className="h-8 w-8 text-primary mb-4 group-hover:-translate-y-1 transition-transform duration-300" />
              <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">Your Cloud, Your Control</h3>
              <p className="text-sm text-muted-foreground">Runs in your AWS, GCP, or Azure account - you own the data, we never see your code</p>
            </div>

            {/* CTA Feature */}
            <div className="group relative p-6 border border-primary bg-primary/5 flex flex-col justify-center transition-all duration-300 hover:bg-primary/10 hover:shadow-lg hover:-translate-y-1">
              <h3 className="font-semibold mb-2 text-lg">Start Now. It's Free.</h3>
              <p className="text-sm text-muted-foreground mb-1">No hidden charges</p>
              <p className="text-sm text-muted-foreground mb-4">No lock-in</p>
              <Button size="sm" className="group-hover:scale-105 transition-transform duration-300" asChild>
                <Link href="/auth/signup">
                  Get Started
                  <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}