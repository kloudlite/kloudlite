import { ArrowLeft } from "lucide-react"
import { type Metadata } from "next"
import Link from "next/link"

import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "Kloudlite Privacy Policy - Learn how we collect, use, and protect your personal information.",
}

export default function PrivacyPage() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="sticky-header">
        <div className="container mx-auto flex h-14 items-center justify-between px-6">
          <Link href="/" className="flex items-center transition-opacity hover:opacity-80">
            <svg height="20" viewBox="0 0 628 131" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-5 w-auto">
              <path d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z" className="fill-primary"></path>
              <path d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z" className="fill-primary"></path>
              <path d="M164.241 113.166V17.8325H180.841V73.6742L201.591 45.6076H220.333L195.968 78.3597L220.868 113.166H202.126L180.841 83.4467V113.166H164.241Z" className="fill-foreground"></path>
              <path d="M588.188 86.6906C588.274 90.651 589.308 93.5352 591.288 95.3432C593.354 97.0652 596.281 97.9261 600.07 97.9261C608.077 97.9261 615.223 97.6678 621.508 97.1513L625.124 96.7638L625.382 109.549C615.481 111.96 606.527 113.165 598.52 113.165C588.791 113.165 581.731 110.582 577.34 105.416C572.949 100.251 570.754 91.8564 570.754 80.2334C570.754 57.0736 580.268 45.4937 599.295 45.4937C618.064 45.4937 627.448 55.2225 627.448 74.6802L626.157 86.6906H588.188ZM610.401 73.5179C610.401 68.3521 609.583 64.7792 607.947 62.7989C606.312 60.7326 603.427 59.6995 599.295 59.6995C595.248 59.6995 592.364 60.7757 590.642 62.9281C589.006 64.9944 588.145 68.5243 588.059 73.5179H610.401Z" className="fill-foreground"></path>
              <path d="M560.42 61.7669H544.536V88.2414C544.536 90.8243 544.579 92.6754 544.665 93.7946C544.837 94.8278 545.311 95.7318 546.086 96.5067C546.946 97.2815 548.238 97.669 549.96 97.669L559.775 97.4107L560.55 111.229C554.781 112.521 550.39 113.166 547.377 113.166C539.628 113.166 534.333 111.444 531.492 108C528.651 104.471 527.23 98.0133 527.23 88.6289V61.7669V45.4948V17.8574H544.536V45.4948H560.42V61.7669Z" className="fill-foreground"></path>
              <path d="M496.661 113.166V45.4948H513.966V113.166H496.661ZM496.661 35.421V17.8574H513.966V35.421H496.661Z" className="fill-foreground"></path>
              <path d="M466.091 113.165L466.091 17.8667H483.396L483.397 113.165H466.091Z" className="fill-foreground"></path>
              <path d="M452.826 17.8667L452.826 113.165H435.65V108.904C429.624 111.745 424.415 113.165 420.024 113.165C410.639 113.165 404.096 110.453 400.394 105.029C396.692 99.6052 394.841 91.0387 394.841 79.3296C394.841 67.5345 397.036 58.9679 401.427 53.63C405.904 48.2059 412.62 45.4939 421.574 45.4939C424.329 45.4939 428.16 45.9244 433.067 46.7854L435.521 47.3019L435.521 17.8667H452.826ZM433.713 96.1183L435.521 95.7309V61.7661C430.786 60.9051 426.567 60.4746 422.865 60.4746C415.891 60.4746 412.404 66.6735 412.404 79.0714C412.404 85.7868 413.179 90.5652 414.729 93.4063C416.279 96.2475 418.819 97.6681 422.348 97.6681C425.965 97.6681 429.753 97.1515 433.713 96.1183Z" className="fill-foreground"></path>
              <path d="M367.331 45.4937H384.636V113.165H367.46V107.999C361.261 111.443 355.88 113.165 351.317 113.165C342.363 113.165 336.337 110.711 333.237 105.804C330.138 100.81 328.588 92.5021 328.588 80.8791V45.4937H345.893V81.1374C345.893 87.5085 346.41 91.8563 347.443 94.1809C348.476 96.5055 350.973 97.6678 354.933 97.6678C358.721 97.6678 362.295 97.0652 365.652 95.8598L367.331 95.3432V45.4937Z" className="fill-foreground"></path>
              <path d="M265.823 54.4046C270.386 48.464 278.006 45.4937 288.682 45.4937C299.358 45.4937 306.977 48.464 311.54 54.4046C316.103 60.2591 318.385 68.5243 318.385 79.2002C318.385 101.844 308.484 113.165 288.682 113.165C268.88 113.165 258.979 101.844 258.979 79.2002C258.979 68.5243 261.26 60.2591 265.823 54.4046ZM279.125 93.7935C280.933 96.893 284.119 98.4427 288.682 98.4427C293.245 98.4427 296.387 96.893 298.109 93.7935C299.917 90.694 300.821 85.8296 300.821 79.2002C300.821 72.5708 299.917 67.7495 298.109 64.7361C296.387 61.7227 293.245 60.2161 288.682 60.2161C284.119 60.2161 280.933 61.7227 279.125 64.7361C277.403 67.7495 276.542 72.5708 276.542 79.2002C276.542 85.8296 277.403 90.694 279.125 93.7935Z" className="fill-foreground"></path>
              <path d="M231.468 113.165L231.071 17.8667H248.377L248.774 113.165H231.468Z" className="fill-foreground"></path>
            </svg>
          </Link>
          
          <div className="flex items-center gap-4">
            <ThemeToggle />
            <Button variant="ghost" size="sm" asChild>
              <Link href="/">
                <ArrowLeft className="mr-2 h-3.5 w-3.5" />
                Back to home
              </Link>
            </Button>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1">
        <article className="container mx-auto max-w-4xl px-6 py-16">
          <header className="mb-12 text-center">
            <h1 className="text-3xl font-light tracking-tight sm:text-4xl">Privacy Policy</h1>
            <p className="mt-4 text-sm text-muted-foreground">Last updated: January 1, 2024</p>
          </header>

          <div className="prose prose-gray dark:prose-invert mx-auto max-w-none">
            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">1. Introduction</h2>
              <p className="text-muted-foreground leading-relaxed">
                Kloudlite ("we," "our," or "us") is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our development environment services.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">2. Information We Collect</h2>
              <h3 className="text-lg font-medium mb-3 mt-6">Personal Information</h3>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We may collect personal information that you provide directly to us, including:
              </p>
              <ul className="list-disc pl-6 space-y-2 text-muted-foreground">
                <li>Name and contact information</li>
                <li>Email address</li>
                <li>Account credentials</li>
                <li>Payment information</li>
                <li>Company information</li>
              </ul>

              <h3 className="text-lg font-medium mb-3 mt-6">Usage Information</h3>
              <p className="text-muted-foreground leading-relaxed">
                We automatically collect certain information about your device and usage patterns:
              </p>
              <ul className="list-disc pl-6 space-y-2 text-muted-foreground">
                <li>IP address and device information</li>
                <li>Browser type and version</li>
                <li>Usage data and analytics</li>
                <li>Performance metrics</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">3. How We Use Your Information</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We use the information we collect for the following purposes:
              </p>
              <ul className="list-disc pl-6 space-y-2 text-muted-foreground">
                <li>Provide and maintain our services</li>
                <li>Process transactions and send related information</li>
                <li>Send administrative information and updates</li>
                <li>Respond to customer service requests</li>
                <li>Monitor and analyze usage patterns</li>
                <li>Improve our services and develop new features</li>
                <li>Detect and prevent fraudulent activities</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">4. Data Sharing and Disclosure</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We may share your information in the following circumstances:
              </p>
              <ul className="list-disc pl-6 space-y-2 text-muted-foreground">
                <li>With service providers who assist in our operations</li>
                <li>To comply with legal obligations</li>
                <li>To protect our rights and property</li>
                <li>With your consent or at your direction</li>
                <li>In connection with a business transaction (merger, acquisition, etc.)</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">5. Data Security</h2>
              <p className="text-muted-foreground leading-relaxed">
                We implement appropriate technical and organizational measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction. However, no method of transmission over the Internet is 100% secure.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">6. Data Retention</h2>
              <p className="text-muted-foreground leading-relaxed">
                We retain your personal information for as long as necessary to fulfill the purposes outlined in this Privacy Policy, unless a longer retention period is required or permitted by law.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">7. Your Rights</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                Depending on your location, you may have certain rights regarding your personal information:
              </p>
              <ul className="list-disc pl-6 space-y-2 text-muted-foreground">
                <li>Access and receive a copy of your data</li>
                <li>Correct or update your information</li>
                <li>Delete your personal information</li>
                <li>Object to or restrict processing</li>
                <li>Data portability</li>
                <li>Withdraw consent</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">8. International Data Transfers</h2>
              <p className="text-muted-foreground leading-relaxed">
                Your information may be transferred to and processed in countries other than your country of residence. We ensure appropriate safeguards are in place to protect your information in accordance with this Privacy Policy.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">9. Children's Privacy</h2>
              <p className="text-muted-foreground leading-relaxed">
                Our services are not intended for individuals under the age of 16. We do not knowingly collect personal information from children under 16.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">10. Changes to This Policy</h2>
              <p className="text-muted-foreground leading-relaxed">
                We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the "Last updated" date.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">11. Contact Us</h2>
              <p className="text-muted-foreground leading-relaxed">
                If you have any questions about this Privacy Policy or our privacy practices, please contact us at:
              </p>
              <p className="text-muted-foreground mt-4">
                Email: privacy@kloudlite.io<br />
                Address: Kloudlite Inc.<br />
                [Your business address]
              </p>
            </section>
          </div>
        </article>
      </main>

      {/* Footer */}
      <footer className="border-t">
        <div className="container mx-auto px-6 py-6">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-sm text-muted-foreground">© 2024 Kloudlite. All rights reserved.</p>
            <nav className="flex items-center gap-6 text-sm">
              <Link href="/legal/privacy" className="text-muted-foreground hover:text-foreground transition-colors">
                Privacy Policy
              </Link>
              <Link href="/legal/terms" className="text-muted-foreground hover:text-foreground transition-colors">
                Terms of Service
              </Link>
            </nav>
          </div>
        </div>
      </footer>
    </div>
  )
}