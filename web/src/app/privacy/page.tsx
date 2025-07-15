import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'

export default function PrivacyPolicy() {
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
          <h1 className="text-4xl font-bold mb-4">Privacy Policy</h1>
          <p className="text-muted-foreground mb-8">
            Last updated: {new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">1. Information We Collect</h2>
          <p>
            We collect information you provide directly to us, such as when you create an account, 
            use our services, or contact us for support.
          </p>
          <ul className="list-disc pl-6 mt-2">
            <li>Account information (name, email, organization)</li>
            <li>Usage data and analytics</li>
            <li>Technical information (IP address, browser type, device information)</li>
            <li>Communication preferences</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">2. How We Use Your Information</h2>
          <p>We use the information we collect to:</p>
          <ul className="list-disc pl-6 mt-2">
            <li>Provide, maintain, and improve our services</li>
            <li>Process transactions and send related information</li>
            <li>Send technical notices, updates, and support messages</li>
            <li>Respond to your comments, questions, and requests</li>
            <li>Monitor and analyze trends, usage, and activities</li>
            <li>Detect, investigate, and prevent fraudulent or illegal activities</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">3. Data Security</h2>
          <p>
            We implement appropriate technical and organizational measures to protect your personal 
            information against unauthorized access, alteration, disclosure, or destruction. This includes:
          </p>
          <ul className="list-disc pl-6 mt-2">
            <li>Encryption of data in transit and at rest</li>
            <li>Regular security assessments and audits</li>
            <li>Access controls and authentication mechanisms</li>
            <li>Employee training on data protection</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">4. Your Code and Data</h2>
          <p>
            <strong>We never access your code or application data.</strong> When you use Kloudlite:
          </p>
          <ul className="list-disc pl-6 mt-2">
            <li>Your code remains in your own repositories</li>
            <li>Development environments run in your cloud account</li>
            <li>We only collect metadata necessary to provide our services</li>
            <li>You maintain full ownership and control of your intellectual property</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">5. Data Retention</h2>
          <p>
            We retain your information for as long as necessary to provide our services and fulfill 
            the purposes outlined in this privacy policy. When you delete your account, we will delete 
            or anonymize your personal information within 30 days.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">6. Third-Party Services</h2>
          <p>
            We may share your information with third-party service providers that help us operate 
            our business, such as:
          </p>
          <ul className="list-disc pl-6 mt-2">
            <li>Cloud infrastructure providers (AWS, GCP, Azure)</li>
            <li>Payment processors</li>
            <li>Analytics services</li>
            <li>Customer support tools</li>
          </ul>
          <p className="mt-2">
            These providers are bound by contractual obligations to keep your information confidential 
            and secure.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">7. Your Rights</h2>
          <p>You have the right to:</p>
          <ul className="list-disc pl-6 mt-2">
            <li>Access and receive a copy of your personal data</li>
            <li>Correct inaccurate or incomplete information</li>
            <li>Request deletion of your personal data</li>
            <li>Object to or restrict processing of your data</li>
            <li>Data portability</li>
            <li>Withdraw consent at any time</li>
          </ul>

          <h2 className="text-2xl font-semibold mt-8 mb-4">8. Cookies and Tracking</h2>
          <p>
            We use cookies and similar tracking technologies to track activity on our service and 
            hold certain information. You can instruct your browser to refuse all cookies or to 
            indicate when a cookie is being sent.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">9. Children's Privacy</h2>
          <p>
            Our services are not intended for individuals under the age of 13. We do not knowingly 
            collect personal information from children under 13.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">10. International Data Transfers</h2>
          <p>
            Your information may be transferred to and maintained on computers located outside of 
            your state, province, country, or other governmental jurisdiction where data protection 
            laws may differ. We ensure appropriate safeguards are in place for such transfers.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">11. Changes to This Policy</h2>
          <p>
            We may update our Privacy Policy from time to time. We will notify you of any changes 
            by posting the new Privacy Policy on this page and updating the "Last updated" date.
          </p>

          <h2 className="text-2xl font-semibold mt-8 mb-4">12. Contact Us</h2>
          <p>
            If you have any questions about this Privacy Policy, please contact us at:
          </p>
          <ul className="list-none mt-2">
            <li>Email: privacy@kloudlite.io</li>
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