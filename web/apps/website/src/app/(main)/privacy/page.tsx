import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'

export default function PrivacyPolicyPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader />
        <main>
          <div className="mx-auto max-w-4xl px-6 py-16 lg:px-8 lg:py-24">
            <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
              Privacy Policy
            </h1>
            <p className="text-foreground/50 mt-4 text-lg">
              Last updated: December 2024
            </p>

            <div className="mt-12 space-y-10 text-foreground/70">
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
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
