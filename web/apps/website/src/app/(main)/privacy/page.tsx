'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { cn } from '@kloudlite/lib'
import Link from 'next/link'

// Cross marker component with pulse animation
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20 animate-pulse" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20 animate-pulse" />
    </div>
  )
}

// GridContainer with clean static borders
function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none">
        {/* Static border lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

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

// Section group container
function SectionGroup({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn("relative p-8 lg:p-12 bg-foreground/[0.015]", className)}>
      {children}
    </div>
  )
}

export default function PrivacyPolicyPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <PageHeroTitle accentedWord="Policy">
                    Privacy
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Last updated: December 2024
                  </p>
                </div>
              </div>

              {/* Content */}
              <div className="border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Spacer */}
                <div className="h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Section Group 1: Introduction & Information Collection */}
                <SectionGroup className="border-b border-foreground/10">
                  <div className="mb-12">
                    <h2 className="text-foreground text-3xl font-bold tracking-tight">Introduction & Data Collection</h2>
                    <div className="mt-2 h-1 w-20 bg-primary"></div>
                  </div>
                  <div className="space-y-10">
                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">1. Introduction</h2>
                      <p className="leading-relaxed">
                        Kloudlite (&quot;we,&quot; &quot;our,&quot; or &quot;us&quot;) is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our cloud development platform and services.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">2. Information We Collect</h2>
                      <h3 className="text-foreground text-lg font-medium mt-6 mb-2">Account Information</h3>
                      <p className="leading-relaxed">
                        When you create an account, we collect your name, email address, and authentication credentials. If you sign up using a third-party service (such as GitHub or Google), we receive information from that service according to their privacy policies.
                      </p>
                      <h3 className="text-foreground text-lg font-medium mt-6 mb-2">Usage Data</h3>
                      <p className="leading-relaxed">
                        We automatically collect information about how you interact with our services, including workspace creation, environment configurations, and feature usage. This helps us improve our platform and provide better support.
                      </p>
                      <h3 className="text-foreground text-lg font-medium mt-6 mb-2">Technical Data</h3>
                      <p className="leading-relaxed">
                        We collect technical information such as IP addresses, browser type, device information, and log data to ensure security and optimize performance.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">3. How We Use Your Information</h2>
                      <ul className="list-disc list-inside space-y-2 leading-relaxed">
                        <li>To provide and maintain our services</li>
                        <li>To authenticate users and manage accounts</li>
                        <li>To communicate with you about updates, security alerts, and support</li>
                        <li>To improve and personalize your experience</li>
                        <li>To detect, prevent, and address technical issues and security threats</li>
                        <li>To comply with legal obligations</li>
                      </ul>
                    </section>
                  </div>
                </SectionGroup>

                {/* Spacer */}
                <div className="h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Section Group 2: Data & Security */}
                <SectionGroup className="border-b border-foreground/10">
                  <div className="mb-12">
                    <h2 className="text-foreground text-3xl font-bold tracking-tight">Data Storage & Security</h2>
                    <div className="mt-2 h-1 w-20 bg-primary"></div>
                  </div>
                  <div className="space-y-10">
                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">4. Data Storage and Security</h2>
                      <p className="leading-relaxed">
                        Your data is stored on secure servers with industry-standard encryption. We implement appropriate technical and organizational measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">5. Data Sharing</h2>
                      <p className="leading-relaxed">
                        We do not sell your personal information. We may share your information with:
                      </p>
                      <ul className="list-disc list-inside space-y-2 leading-relaxed mt-4">
                        <li>Service providers who assist in operating our platform</li>
                        <li>Legal authorities when required by law</li>
                        <li>Business partners with your consent</li>
                      </ul>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">6. Your Rights</h2>
                      <p className="leading-relaxed">
                        You have the right to access, correct, or delete your personal information. You can also request a copy of your data or restrict certain processing activities. To exercise these rights, please contact us at privacy@kloudlite.io.
                      </p>
                    </section>
                  </div>
                </SectionGroup>

                {/* Spacer */}
                <div className="h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-3/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Section Group 3: Cookies & Third-Party */}
                <SectionGroup className="border-b border-foreground/10">
                  <div className="mb-12">
                    <h2 className="text-foreground text-3xl font-bold tracking-tight">Cookies & Third-Party Services</h2>
                    <div className="mt-2 h-1 w-20 bg-primary"></div>
                  </div>
                  <div className="space-y-10">
                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">7. Cookies and Tracking</h2>
                      <p className="leading-relaxed">
                        We use cookies and similar technologies to enhance your experience, analyze usage patterns, and deliver relevant content. You can manage cookie preferences through your browser settings.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">8. Third-Party Services</h2>
                      <p className="leading-relaxed">
                        Our platform may integrate with third-party services. These services have their own privacy policies, and we encourage you to review them. We are not responsible for the privacy practices of third-party services.
                      </p>
                    </section>
                  </div>
                </SectionGroup>

                {/* Spacer */}
                <div className="h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Section Group 4: Children, Changes & Contact */}
                <SectionGroup className="border-b border-foreground/10">
                  <div className="mb-12">
                    <h2 className="text-foreground text-3xl font-bold tracking-tight">Policy Updates & Contact</h2>
                    <div className="mt-2 h-1 w-20 bg-primary"></div>
                  </div>
                  <div className="space-y-10">
                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">9. Children&apos;s Privacy</h2>
                      <p className="leading-relaxed">
                        Our services are not intended for individuals under 16 years of age. We do not knowingly collect personal information from children.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">10. Changes to This Policy</h2>
                      <p className="leading-relaxed">
                        We may update this Privacy Policy from time to time. We will notify you of any material changes by posting the new policy on this page and updating the &quot;Last updated&quot; date.
                      </p>
                    </section>

                    <section>
                      <h2 className="text-foreground text-2xl font-semibold mb-4">11. Contact Us</h2>
                      <p className="leading-relaxed">
                        If you have any questions about this Privacy Policy, please contact us at:
                      </p>
                      <div className="mt-4 p-6 bg-foreground/5 border border-foreground/10">
                        <p className="font-medium text-foreground">Kloudlite</p>
                        <p className="mt-2">415, Floor 4, Shaft-1, Tower-B</p>
                        <p>VRR Fortuna, Carmelaram</p>
                        <p>Janatha Colony, Bangalore</p>
                        <p>Karnataka, India - 560035</p>
                        <p className="mt-4">Email: privacy@kloudlite.io</p>
                      </div>
                    </section>
                  </div>
                </SectionGroup>

                {/* Spacer */}
                <div className="h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Enhanced Contact CTA */}
                <div className="p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <div className="max-w-3xl">
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl mb-6">
                      Have Questions About <span className="relative inline-block">
                        <span className="relative z-10">Privacy?</span>
                        <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                      </span>
                    </h2>
                    <p className="text-muted-foreground text-base lg:text-lg mb-8">
                      We&apos;re committed to protecting your data and privacy. Reach out to our team for any questions or concerns.
                    </p>
                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                      <Button size="lg" asChild>
                        <Link href="/contact">Contact Us</Link>
                      </Button>
                      <Button variant="outline" size="lg" asChild>
                        <Link href="/terms">View Terms of Service</Link>
                      </Button>
                    </div>
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
