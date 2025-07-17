'use client'

import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ArrowRight, Settings, Network, Package, Zap, ArrowUpDown, Code2, Bot, Users, Briefcase, Shield, Cloud } from 'lucide-react'
import { 
  Typewriter, 
  SectionHeading, 
  ProcessFlow, 
  FeatureCard, 
  CTACard, 
  SectionDivider 
} from '@/components/marketing/ui'

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

const features = [
  {
    icon: Settings,
    title: 'Zero Config Workspaces',
    description: 'From git clone to running code in under 30 seconds - no setup scripts, no dependency hell',
    iconClassName: 'group-hover:rotate-90 transition-transform duration-500'
  },
  {
    icon: Network,
    title: 'Connected Environments',
    description: 'Your code runs in the same network as your services - seamless service discovery',
    iconClassName: 'group-hover:scale-110 transition-transform duration-300'
  },
  {
    icon: Package,
    title: 'Config Management',
    description: 'Configs and secrets auto-injected - no more .env files or manual setup',
    iconClassName: 'group-hover:-rotate-12 transition-transform duration-300'
  },
  {
    icon: Zap,
    title: 'Lightweight Environments',
    description: 'Spin up 10 environments for the cost of 1 VM - instant start, minimal overhead',
    iconClassName: 'group-hover:animate-pulse transition-all duration-300'
  },
  {
    icon: ArrowUpDown,
    title: 'Service Intercepts',
    description: 'Debug production issues locally - intercept any service traffic with one command',
    iconClassName: 'group-hover:translate-y-1 transition-transform duration-300'
  },
  {
    icon: Code2,
    title: 'IDE Integration',
    description: 'Native extensions for VS Code, JetBrains - debug, test, and deploy without leaving your IDE',
    iconClassName: 'group-hover:skew-y-3 transition-transform duration-300'
  },
  {
    icon: Bot,
    title: 'AI Agent Ready',
    description: 'Native support for autonomous AI coding agents - Cline, Zencoder, and all major AI tools',
    iconClassName: 'group-hover:animate-bounce transition-all duration-300'
  },
  {
    icon: Users,
    title: 'Instant Collaboration',
    description: 'Switch to your teammate\'s environment instantly - share services, debug together, no setup required',
    iconClassName: 'group-hover:scale-105 transition-transform duration-300'
  },
  {
    icon: Briefcase,
    title: 'Remote-First Design',
    description: 'Same performance whether you\'re in San Francisco or São Paulo - optimized for global teams',
    iconClassName: 'group-hover:rotate-6 transition-transform duration-300'
  },
  {
    icon: Shield,
    title: 'Enterprise Security',
    description: 'SOC2, HIPAA compliant - encrypted at rest and in transit, with full audit trails',
    iconClassName: 'group-hover:scale-110 transition-transform duration-300'
  },
  {
    icon: Cloud,
    title: 'Your Cloud, Your Control',
    description: 'Runs in your AWS, GCP, or Azure account - you own the data, we never see your code',
    iconClassName: 'group-hover:-translate-y-1 transition-transform duration-300'
  }
]

export function HeroSection() {

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
              For <Typewriter words={projectTypes} /> developers
            </span>
          </p>

          {/* Key message */}
          <div className="mt-8">
            <p className="text-sm uppercase tracking-wider text-muted-foreground mb-4">Focus on what matters</p>
            <ProcessFlow 
              steps={[
                { label: 'Code', active: true },
                { label: 'Build', active: false, strikethrough: true },
                { label: 'Deploy', active: false, strikethrough: true },
                { label: 'Test', active: true }
              ]}
            />
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


        {/* Section Divider */}
        <SectionDivider />

        {/* Key Features Grid */}
        <div className="">
          <SectionHeading 
            title="Every feature, one purpose:"
            subtitle="Accelerate your development loop"
            className="mb-12"
          />
          
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <FeatureCard
                key={index}
                icon={feature.icon}
                title={feature.title}
                description={feature.description}
                iconClassName={feature.iconClassName}
              />
            ))}
            
            {/* CTA Feature */}
            <CTACard 
              title="Start Now. It's Free."
              subtitle="No hidden charges"
              description="No lock-in"
              buttonText="Get Started"
              buttonHref="/auth/signup"
            />
          </div>
        </div>
      </div>
    </section>
  )
}