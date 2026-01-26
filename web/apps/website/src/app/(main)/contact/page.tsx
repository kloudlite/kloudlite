'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Mail, Github, Twitter, Linkedin, ArrowRight, MapPin, MessageSquare, HeadphonesIcon } from 'lucide-react'
import Link from 'next/link'
import { PageHeroTitle } from '@/components/page-hero-title'
import ContactForm from './contact-form'

// Cross marker component with pulse animation
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20 animate-pulse" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20 animate-pulse" />
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

export default function ContactPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="contact" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <PageHeroTitle accentedWord="Us.">
                    Get in Touch with
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Questions about Kloudlite? Technical support? Partnership opportunities? We&apos;re here to help.
                  </p>
                  <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                    <Button variant="default" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <a href="mailto:hello@kloudlite.io">
                        <Mail className="h-4 w-4 mr-2" />
                        Email Us
                      </a>
                    </Button>
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="https://github.com/kloudlite/kloudlite" target="_blank" rel="noopener noreferrer">
                        <Github className="h-4 w-4 mr-2" />
                        GitHub
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Main Content Grid */}
              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Quick Contact Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Quick <span className="relative inline-block">
                      <span className="relative z-10">contact.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-3xl">
                    Choose the best way to reach us based on your needs.
                  </p>
                </div>

                {/* General Inquiries */}
                <a
                  href="mailto:hello@kloudlite.io"
                  className="group relative p-8 lg:p-12 border-b border-foreground/10 sm:border-r bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden"
                >
                  <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
                  <div className="text-primary mb-4 transition-colors">
                    <Mail className="h-6 w-6" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">General Inquiries</h3>
                  <p className="text-muted-foreground text-base leading-relaxed font-medium transition-colors group-hover:text-foreground mb-3">
                    Questions about Kloudlite, pricing, or general information.
                  </p>
                  <p className="text-primary text-sm font-mono">hello@kloudlite.io</p>
                  <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-primary group-hover:translate-x-1 transition-all absolute top-8 right-8" />
                </a>

                {/* Technical Support */}
                <a
                  href="mailto:support@kloudlite.io"
                  className="group relative p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden"
                >
                  <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
                  <div className="text-primary mb-4 transition-colors">
                    <HeadphonesIcon className="h-6 w-6" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Technical Support</h3>
                  <p className="text-muted-foreground text-base leading-relaxed font-medium transition-colors group-hover:text-foreground mb-3">
                    Need help with setup, troubleshooting, or technical issues.
                  </p>
                  <p className="text-primary text-sm font-mono">support@kloudlite.io</p>
                  <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-primary group-hover:translate-x-1 transition-all absolute top-8 right-8" />
                </a>

                {/* Cross Marker between rows */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* GitHub */}
                <a
                  href="https://github.com/kloudlite/kloudlite"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="group relative p-8 lg:p-12 border-b border-foreground/10 sm:border-r bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden"
                >
                  <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
                  <div className="text-primary mb-4 transition-colors">
                    <Github className="h-6 w-6" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Open Source</h3>
                  <p className="text-muted-foreground text-base leading-relaxed font-medium transition-colors group-hover:text-foreground mb-3">
                    Report bugs, request features, or contribute to the codebase.
                  </p>
                  <p className="text-primary text-sm font-mono">github.com/kloudlite</p>
                  <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-primary group-hover:translate-x-1 transition-all absolute top-8 right-8" />
                </a>

                {/* Office Location */}
                <div className="group relative p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden cursor-default">
                  <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
                  <div className="text-primary mb-4 transition-colors">
                    <MapPin className="h-6 w-6" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Headquarters</h3>
                  <p className="text-muted-foreground text-sm leading-relaxed font-medium transition-colors group-hover:text-foreground">
                    415, Floor 4, Shaft-1, Tower-B<br />
                    VRR Fortuna, Carmelaram<br />
                    Janatha Colony, Bangalore<br />
                    Karnataka, India - 560035
                  </p>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-2/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Contact Form Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Send us a <span className="relative inline-block">
                      <span className="relative z-10">message.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-3xl">
                    Fill out the form below and we&apos;ll get back to you within 24 hours.
                  </p>
                </div>

                {/* Form Section - Left Sidebar */}
                <div className="p-8 lg:p-12 border-b border-foreground/10 sm:border-r bg-foreground/[0.015] flex flex-col justify-between">
                  <div>
                    <div className="text-primary mb-4">
                      <MessageSquare className="h-8 w-8" />
                    </div>
                    <h3 className="text-foreground text-xl font-bold mb-3">
                      Let&apos;s discuss your needs
                    </h3>
                    <p className="text-muted-foreground text-base leading-relaxed">
                      Whether you&apos;re exploring Kloudlite for your team, need technical guidance, or want to partner with us, we&apos;re here to help.
                    </p>
                  </div>

                  {/* Social Links */}
                  <div className="mt-8">
                    <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider mb-4">Connect With Us</p>
                    <div className="flex flex-wrap gap-2">
                      <a
                        href="https://github.com/kloudlite"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
                        aria-label="GitHub"
                      >
                        <Github className="h-4 w-4" />
                      </a>
                      <a
                        href="https://twitter.com/kloudlite"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
                        aria-label="Twitter"
                      >
                        <Twitter className="h-4 w-4" />
                      </a>
                      <a
                        href="https://linkedin.com/company/kloudlite"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
                        aria-label="LinkedIn"
                      >
                        <Linkedin className="h-4 w-4" />
                      </a>
                      <a
                        href="mailto:hello@kloudlite.io"
                        className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
                        aria-label="Email"
                      >
                        <Mail className="h-4 w-4" />
                      </a>
                    </div>
                  </div>
                </div>

                {/* Form Section - Contact Form */}
                <div className="p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015]">
                  <ContactForm />
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
