import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Mail, Github, Twitter, Linkedin, ArrowRight, MapPin } from 'lucide-react'
import ContactForm from './contact-form'

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

export default function ContactPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="contact" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-24">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground/40">C</span><span className="text-foreground">ontact</span>{' '}
                    <span className="text-foreground/40">U</span><span className="text-foreground">s</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-md text-lg leading-relaxed">
                    Questions, feedback, or partnership inquiries.
                    <br />
                    We&apos;re here to help.
                  </p>
                </div>
              </div>

              {/* Quick Contact Options */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-t border-foreground/10">
                <a
                  href="mailto:hello@kloudlite.io"
                  className="p-8 lg:p-10 border-b lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Email</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">General Inquiries</h3>
                      <p className="text-foreground/50 mt-1 text-sm group-hover:text-foreground/70 transition-colors">
                        hello@kloudlite.io
                      </p>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </a>

                <a
                  href="mailto:support@kloudlite.io"
                  className="p-8 lg:p-10 border-b border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Support</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Technical Help</h3>
                      <p className="text-foreground/50 mt-1 text-sm group-hover:text-foreground/70 transition-colors">
                        support@kloudlite.io
                      </p>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </a>

                <a
                  href="https://github.com/kloudlite/kloudlite"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Open Source</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">GitHub</h3>
                      <p className="text-foreground/50 mt-1 text-sm group-hover:text-foreground/70 transition-colors">
                        Report issues & contribute
                      </p>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </a>

                <div className="p-8 lg:p-10 border-b lg:border-b-0 border-foreground/10">
                  <div className="flex items-start gap-3">
                    <MapPin className="h-5 w-5 text-foreground/40 mt-0.5 flex-shrink-0" />
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Office</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Headquarters</h3>
                      <p className="text-foreground/50 mt-1 text-sm leading-relaxed">
                        415, Floor 4, Shaft-1, Tower-B<br />
                        VRR Fortuna, Carmelaram<br />
                        Janatha Colony, Bangalore<br />
                        Karnataka, India - 560035
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Contact Form Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-between">
                  <div>
                    <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
                      Send a Message
                    </h2>
                    <p className="text-foreground/50 mt-3 text-base">
                      Have something specific to discuss? Fill out the form and we&apos;ll respond within 24 hours.
                    </p>
                  </div>

                  {/* Social Links */}
                  <div className="mt-8 lg:mt-0">
                    <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-4">Follow Us</p>
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
                      <a
                        href="mailto:hello@kloudlite.io"
                        className="p-2.5 border border-foreground/10 text-foreground/40 hover:text-foreground hover:border-foreground/20 transition-colors"
                      >
                        <Mail className="h-4 w-4" />
                      </a>
                    </div>
                  </div>
                </div>

                {/* Form */}
                <div className="lg:col-span-2 p-8 lg:p-10">
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
