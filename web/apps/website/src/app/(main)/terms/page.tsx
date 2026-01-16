import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'

export default function TermsOfServicePage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader />
        <main>
          <div className="mx-auto max-w-4xl px-6 py-16 lg:px-8 lg:py-24">
            <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
              Terms of Service
            </h1>
            <p className="text-muted-foreground mt-4 text-lg">
              Last updated: December 2024
            </p>

            <div className="mt-12 space-y-10 text-foreground">
              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">1. Agreement to Terms</h2>
                <p className="leading-relaxed">
                  By accessing or using Kloudlite&apos;s cloud development platform and services (&quot;Services&quot;), you agree to be bound by these Terms of Service. If you do not agree to these terms, please do not use our Services.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">2. Description of Services</h2>
                <p className="leading-relaxed">
                  Kloudlite provides cloud development environments, workspaces, and related tools that enable developers to write, test, and deploy code. Our Services include but are not limited to workspace management, environment provisioning, and service integration capabilities.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">3. Account Registration</h2>
                <p className="leading-relaxed">
                  To use our Services, you must create an account. You agree to provide accurate information and keep your account credentials secure. You are responsible for all activities that occur under your account.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">4. Acceptable Use</h2>
                <p className="leading-relaxed mb-4">
                  You agree not to use our Services to:
                </p>
                <ul className="list-disc list-inside space-y-2 leading-relaxed">
                  <li>Violate any applicable laws or regulations</li>
                  <li>Infringe upon intellectual property rights of others</li>
                  <li>Distribute malware, viruses, or harmful code</li>
                  <li>Engage in unauthorized access to systems or data</li>
                  <li>Interfere with or disrupt the integrity of our Services</li>
                  <li>Harvest or collect user information without consent</li>
                  <li>Use the Services for cryptocurrency mining</li>
                  <li>Engage in any activity that could damage our reputation</li>
                </ul>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">5. User Content</h2>
                <p className="leading-relaxed">
                  You retain ownership of any code, data, or content you create using our Services (&quot;User Content&quot;). By using our Services, you grant us a limited license to store and process your User Content solely for the purpose of providing the Services.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">6. Intellectual Property</h2>
                <p className="leading-relaxed">
                  Kloudlite and its licensors own all rights, title, and interest in the Services, including all intellectual property rights. Our platform&apos;s core technology is open source under applicable licenses, which govern your use of that code.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">7. Payment Terms</h2>
                <p className="leading-relaxed">
                  Certain Services may require payment. You agree to pay all fees associated with your selected plan. Fees are non-refundable except as required by law. We reserve the right to change pricing with reasonable notice.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">8. Service Availability</h2>
                <p className="leading-relaxed">
                  We strive to maintain high availability but do not guarantee uninterrupted access to our Services. We may perform maintenance, updates, or modifications that temporarily affect availability. We will provide notice of planned maintenance when possible.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">9. Data and Privacy</h2>
                <p className="leading-relaxed">
                  Your use of our Services is also governed by our Privacy Policy. By using our Services, you consent to the collection and use of information as described in our Privacy Policy.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">10. Termination</h2>
                <p className="leading-relaxed">
                  We may suspend or terminate your access to the Services at any time for violation of these terms or for any reason with reasonable notice. You may terminate your account at any time. Upon termination, your right to use the Services ceases immediately.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">11. Disclaimer of Warranties</h2>
                <p className="leading-relaxed">
                  THE SERVICES ARE PROVIDED &quot;AS IS&quot; AND &quot;AS AVAILABLE&quot; WITHOUT WARRANTIES OF ANY KIND, EXPRESS OR IMPLIED. WE DISCLAIM ALL WARRANTIES INCLUDING MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">12. Limitation of Liability</h2>
                <p className="leading-relaxed">
                  TO THE MAXIMUM EXTENT PERMITTED BY LAW, KLOUDLITE SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES ARISING FROM YOUR USE OF THE SERVICES. OUR TOTAL LIABILITY SHALL NOT EXCEED THE AMOUNT PAID BY YOU IN THE TWELVE MONTHS PRECEDING THE CLAIM.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">13. Indemnification</h2>
                <p className="leading-relaxed">
                  You agree to indemnify and hold harmless Kloudlite, its officers, directors, employees, and agents from any claims, damages, or expenses arising from your use of the Services or violation of these terms.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">14. Governing Law</h2>
                <p className="leading-relaxed">
                  These Terms shall be governed by the laws of India. Any disputes arising from these terms shall be resolved in the courts of Bangalore, Karnataka, India.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">15. Changes to Terms</h2>
                <p className="leading-relaxed">
                  We may modify these Terms at any time. Material changes will be notified via email or through the Services. Continued use of the Services after changes constitutes acceptance of the new terms.
                </p>
              </section>

              <section>
                <h2 className="text-foreground text-2xl font-semibold mb-4">16. Contact Information</h2>
                <p className="leading-relaxed">
                  For questions about these Terms of Service, please contact us at:
                </p>
                <div className="mt-4 p-6 bg-foreground/5 border border-foreground/10">
                  <p className="font-medium text-foreground">Kloudlite</p>
                  <p className="mt-2">415, Floor 4, Shaft-1, Tower-B</p>
                  <p>VRR Fortuna, Carmelaram</p>
                  <p>Janatha Colony, Bangalore</p>
                  <p>Karnataka, India - 560035</p>
                  <p className="mt-4">Email: legal@kloudlite.io</p>
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
