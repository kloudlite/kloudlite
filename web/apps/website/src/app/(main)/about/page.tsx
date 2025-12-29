import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Github, Linkedin, Twitter, ArrowRight } from 'lucide-react'
import Link from 'next/link'

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

export default function AboutPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="about" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-24">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground/40">B</span><span className="text-foreground">uilding the</span>{' '}
                    <span className="text-foreground/40">F</span><span className="text-foreground">uture</span>
                    <br />
                    <span className="text-foreground/40">of Development.</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-lg text-lg leading-relaxed">
                    We&apos;re on a mission to eliminate the friction
                    <br />
                    between writing code and seeing it run.
                  </p>
                </div>
              </div>

              {/* Stats Row */}
              <div className="grid grid-cols-2 lg:grid-cols-4 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-6 lg:p-8 border-r border-foreground/10 text-center">
                  <p className="text-foreground text-3xl lg:text-4xl font-bold tracking-tight">2023</p>
                  <p className="text-foreground/40 mt-1 text-sm">Founded</p>
                </div>
                <div className="p-6 lg:p-8 lg:border-r border-foreground/10 text-center">
                  <p className="text-foreground text-3xl lg:text-4xl font-bold tracking-tight">100%</p>
                  <p className="text-foreground/40 mt-1 text-sm">Open Source</p>
                </div>
                <div className="p-6 lg:p-8 border-r border-t lg:border-t-0 border-foreground/10 text-center">
                  <p className="text-foreground text-3xl lg:text-4xl font-bold tracking-tight">10x</p>
                  <p className="text-foreground/40 mt-1 text-sm">Faster Feedback</p>
                </div>
                <div className="p-6 lg:p-8 border-t lg:border-t-0 border-foreground/10 text-center">
                  <p className="text-foreground text-3xl lg:text-4xl font-bold tracking-tight">0</p>
                  <p className="text-foreground/40 mt-1 text-sm">Setup Required</p>
                </div>
              </div>

              {/* Mission Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
                    Our Mission
                  </h2>
                  <p className="text-foreground/50 mt-3 text-base">
                    Why we exist and what drives us.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <p className="text-foreground/70 text-lg leading-relaxed">
                    Developers spend too much time waiting. Waiting for builds, waiting for deployments, waiting for environments. Every context switch costs productivity and focus.
                  </p>
                  <p className="text-foreground/70 text-lg leading-relaxed mt-4">
                    We&apos;re building Kloudlite to eliminate that friction. Cloud development environments that connect directly to your services, so you can write code and see it work instantly.
                  </p>
                </div>
              </div>

              {/* Principles */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Principle 01</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Speed Above All</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                    Every millisecond matters. We obsess over reducing latency in every part of the development loop.
                  </p>
                </div>
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Principle 02</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Zero Configuration</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                    Tools should work out of the box. No complex setup, no infrastructure expertise required.
                  </p>
                </div>
                <div className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Principle 03</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Open by Default</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed group-hover:text-foreground/60 transition-colors">
                    Transparency builds trust. Our core platform is open source and will always remain so.
                  </p>
                </div>
              </div>

              {/* Links Section */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <Link
                  href="https://github.com/kloudlite/kloudlite"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Open Source</p>
                      <h3 className="text-foreground mt-2 text-xl font-bold tracking-[-0.02em]">View on GitHub</h3>
                      <p className="text-foreground/50 mt-1 text-sm group-hover:text-foreground/70 transition-colors">
                        Explore the code, report issues, contribute
                      </p>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>

                <Link
                  href="/contact"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Get in Touch</p>
                      <h3 className="text-foreground mt-2 text-xl font-bold tracking-[-0.02em]">Contact Us</h3>
                      <p className="text-foreground/50 mt-1 text-sm group-hover:text-foreground/70 transition-colors">
                        Questions, partnerships, or just say hello
                      </p>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
              </div>

              {/* Social Links */}
              <div className="p-8 lg:p-10 -mx-6 lg:-mx-12 flex items-center justify-between">
                <p className="text-foreground/40 text-sm">Follow our journey</p>
                <div className="flex gap-2">
                  <a
                    href="https://github.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-foreground/40 hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Github className="h-4 w-4" />
                  </a>
                  <a
                    href="https://twitter.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-foreground/40 hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Twitter className="h-4 w-4" />
                  </a>
                  <a
                    href="https://linkedin.com/company/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-foreground/40 hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Linkedin className="h-4 w-4" />
                  </a>
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
