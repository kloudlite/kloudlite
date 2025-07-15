import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-4xl mx-auto px-6 py-12">
        {/* Header */}
        <div className="mb-8">
          <Button variant="ghost" size="sm" asChild>
            <Link href="/">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Home
            </Link>
          </Button>
        </div>

        {/* Content */}
        <div className="prose prose-gray dark:prose-invert max-w-none">
          <h1 className="text-4xl font-bold mb-4">Terms and Conditions</h1>
          <p className="text-muted-foreground mb-8">
            Last updated: {new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">1. Acceptance of Terms</h2>
          <p>
            By accessing and using Kloudlite ("Service"), you accept and agree to be bound by the terms 
            and provision of this agreement. If you do not agree to abide by the above, please do not 
            use this service.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">2. Use License</h2>
          <p>
            Permission is granted to temporarily use Kloudlite for personal, non-commercial transitory 
            viewing only. This is the grant of a license, not a transfer of title, and under this 
            license you may not:
          </p>
          <ul className="list-disc pl-6 mt-2">
            <li>modify or copy the materials</li>
            <li>use the materials for any commercial purpose or for any public display</li>
            <li>attempt to reverse engineer any software contained on Kloudlite</li>
            <li>remove any copyright or other proprietary notations from the materials</li>
          </ul>
          <p className="mt-2">
            This license shall automatically terminate if you violate any of these restrictions and 
            may be terminated by Kloudlite at any time. Upon terminating your viewing of these 
            materials or upon the termination of this license, you must destroy any downloaded 
            materials in your possession whether in electronic or printed format.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">3. User Accounts</h2>
          <p>
            When you create an account with us, you must provide us with information that is accurate, 
            complete, and current at all times. You are responsible for safeguarding the password and 
            for all activities that occur under your account. You must notify us immediately upon 
            becoming aware of any breach of security or unauthorized use of your account.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">4. Privacy Policy</h2>
          <p>
            Your use of Kloudlite is also governed by our Privacy Policy. Please review our Privacy 
            Policy, which also governs the Site and informs users of our data collection practices. 
            We respect your privacy and are committed to protecting your personal data.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">5. Prohibited Uses</h2>
          <p>You may not use Kloudlite:</p>
          <ul className="list-disc pl-6 mt-2">
            <li>For any unlawful purpose or to solicit others to perform unlawful acts</li>
            <li>To violate any international, federal, provincial, or state regulations, rules, laws, or local ordinances</li>
            <li>To infringe upon or violate our intellectual property rights or the intellectual property rights of others</li>
            <li>To harass, abuse, insult, harm, defame, slander, disparage, intimidate, or discriminate</li>
            <li>To submit false or misleading information</li>
            <li>To upload or transmit viruses or any other type of malicious code</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">6. Intellectual Property</h2>
          <p>
            The Service and its original content, features, and functionality are and will remain the 
            exclusive property of Kloudlite and its licensors. The Service is protected by copyright, 
            trademark, and other laws. Our trademarks and trade dress may not be used in connection 
            with any product or service without our prior written consent.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">7. Disclaimer</h2>
          <p>
            The materials on Kloudlite are provided on an 'as is' basis. Kloudlite makes no warranties, 
            expressed or implied, and hereby disclaims and negates all other warranties including, 
            without limitation, implied warranties or conditions of merchantability, fitness for a 
            particular purpose, or non-infringement of intellectual property or other violation of rights.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">8. Limitations</h2>
          <p>
            In no event shall Kloudlite or its suppliers be liable for any damages (including, without 
            limitation, damages for loss of data or profit, or due to business interruption) arising 
            out of the use or inability to use Kloudlite, even if Kloudlite or a Kloudlite authorized 
            representative has been notified orally or in writing of the possibility of such damage. 
            Because some jurisdictions do not allow limitations on implied warranties, or limitations 
            of liability for consequential or incidental damages, these limitations may not apply to you.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">9. Termination</h2>
          <p>
            We may terminate or suspend your account and bar access to the Service immediately, without 
            prior notice or liability, under our sole discretion, for any reason whatsoever and without 
            limitation, including but not limited to a breach of the Terms. If you wish to terminate 
            your account, you may simply discontinue using the Service.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">10. Governing Law</h2>
          <p>
            These Terms shall be governed and construed in accordance with the laws of the jurisdiction 
            in which Kloudlite operates, without regard to its conflict of law provisions. Our failure 
            to enforce any right or provision of these Terms will not be considered a waiver of those rights.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">11. Changes to Terms</h2>
          <p>
            Kloudlite reserves the right, at our sole discretion, to modify or replace these Terms at 
            any time. If a revision is material, we will provide at least 30 days notice prior to any 
            new terms taking effect. By continuing to access or use our Service after any revisions 
            become effective, you agree to be bound by the revised terms.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">12. Contact Information</h2>
          <p>
            If you have any questions about these Terms and Conditions, please contact us at:
          </p>
          <ul className="list-none mt-2">
            <li>Email: legal@kloudlite.io</li>
            <li>Address: Kloudlite, Inc.</li>
          </ul>
        </div>

        {/* Footer */}
        <div className="mt-12 pt-8 border-t border-border">
          <p className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} Kloudlite. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  )
}