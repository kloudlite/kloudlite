import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Link } from '@/components/ui/link'
import { ArrowLeft } from 'lucide-react'

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        <div className="mb-6">
          <Link href="/auth/signup" className="flex items-center gap-2 text-sm mb-6">
            <ArrowLeft className="h-4 w-4" />
            Back to sign up
          </Link>
        </div>

        <Card>
          <CardHeader className="space-y-4 pb-8">
            <CardTitle className="text-3xl font-semibold tracking-tight">
              Terms and Conditions
            </CardTitle>
            <p className="text-sm text-muted-foreground">
              Last updated: {new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
            </p>
          </CardHeader>
          <CardContent className="prose prose-sm max-w-none space-y-6 text-foreground">
            <section className="space-y-3">
              <h2 className="text-xl font-semibold">1. Acceptance of Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                By accessing and using Kloudlite ("Service"), you accept and agree to be bound by the terms and provision of this agreement. If you do not agree to abide by the above, please do not use this service.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">2. Use License</h2>
              <div className="space-y-2 text-muted-foreground">
                <p className="leading-relaxed">
                  Permission is granted to temporarily use Kloudlite for personal, non-commercial transitory viewing only. This is the grant of a license, not a transfer of title, and under this license you may not:
                </p>
                <ul className="list-disc pl-6 space-y-1">
                  <li>modify or copy the materials</li>
                  <li>use the materials for any commercial purpose or for any public display</li>
                  <li>attempt to reverse engineer any software contained on Kloudlite</li>
                  <li>remove any copyright or other proprietary notations from the materials</li>
                </ul>
                <p className="leading-relaxed">
                  This license shall automatically terminate if you violate any of these restrictions and may be terminated by Kloudlite at any time. Upon terminating your viewing of these materials or upon the termination of this license, you must destroy any downloaded materials in your possession whether in electronic or printed format.
                </p>
              </div>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">3. User Accounts</h2>
              <p className="text-muted-foreground leading-relaxed">
                When you create an account with us, you must provide us with information that is accurate, complete, and current at all times. You are responsible for safeguarding the password and for all activities that occur under your account. You must notify us immediately upon becoming aware of any breach of security or unauthorized use of your account.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">4. Privacy Policy</h2>
              <p className="text-muted-foreground leading-relaxed">
                Your use of Kloudlite is also governed by our Privacy Policy. Please review our Privacy Policy, which also governs the Site and informs users of our data collection practices. We respect your privacy and are committed to protecting your personal data.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">5. Prohibited Uses</h2>
              <div className="space-y-2 text-muted-foreground">
                <p className="leading-relaxed">
                  You may not use Kloudlite:
                </p>
                <ul className="list-disc pl-6 space-y-1">
                  <li>For any unlawful purpose or to solicit others to perform unlawful acts</li>
                  <li>To violate any international, federal, provincial, or state regulations, rules, laws, or local ordinances</li>
                  <li>To infringe upon or violate our intellectual property rights or the intellectual property rights of others</li>
                  <li>To harass, abuse, insult, harm, defame, slander, disparage, intimidate, or discriminate</li>
                  <li>To submit false or misleading information</li>
                  <li>To upload or transmit viruses or any other type of malicious code</li>
                </ul>
              </div>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">6. Intellectual Property</h2>
              <p className="text-muted-foreground leading-relaxed">
                The Service and its original content, features, and functionality are and will remain the exclusive property of Kloudlite and its licensors. The Service is protected by copyright, trademark, and other laws. Our trademarks and trade dress may not be used in connection with any product or service without our prior written consent.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">7. Disclaimer</h2>
              <p className="text-muted-foreground leading-relaxed">
                The materials on Kloudlite are provided on an 'as is' basis. Kloudlite makes no warranties, expressed or implied, and hereby disclaims and negates all other warranties including, without limitation, implied warranties or conditions of merchantability, fitness for a particular purpose, or non-infringement of intellectual property or other violation of rights.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">8. Limitations</h2>
              <p className="text-muted-foreground leading-relaxed">
                In no event shall Kloudlite or its suppliers be liable for any damages (including, without limitation, damages for loss of data or profit, or due to business interruption) arising out of the use or inability to use Kloudlite, even if Kloudlite or a Kloudlite authorized representative has been notified orally or in writing of the possibility of such damage. Because some jurisdictions do not allow limitations on implied warranties, or limitations of liability for consequential or incidental damages, these limitations may not apply to you.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">9. Termination</h2>
              <p className="text-muted-foreground leading-relaxed">
                We may terminate or suspend your account and bar access to the Service immediately, without prior notice or liability, under our sole discretion, for any reason whatsoever and without limitation, including but not limited to a breach of the Terms. If you wish to terminate your account, you may simply discontinue using the Service.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">10. Governing Law</h2>
              <p className="text-muted-foreground leading-relaxed">
                These Terms shall be governed and construed in accordance with the laws of the jurisdiction in which Kloudlite operates, without regard to its conflict of law provisions. Our failure to enforce any right or provision of these Terms will not be considered a waiver of those rights.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">11. Changes to Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                Kloudlite reserves the right, at our sole discretion, to modify or replace these Terms at any time. If a revision is material, we will provide at least 30 days notice prior to any new terms taking effect. By continuing to access or use our Service after any revisions become effective, you agree to be bound by the revised terms.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">12. Contact Information</h2>
              <p className="text-muted-foreground leading-relaxed">
                If you have any questions about these Terms and Conditions, please contact us at:
              </p>
              <div className="text-muted-foreground pl-6">
                <p>Kloudlite Support</p>
                <p>Email: legal@kloudlite.io</p>
              </div>
            </section>

            <div className="border-t pt-6 mt-8">
              <p className="text-sm text-muted-foreground text-center">
                By using Kloudlite, you acknowledge that you have read, understood, and agree to be bound by these Terms and Conditions.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}