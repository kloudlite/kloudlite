import { ArrowLeft } from "lucide-react"
import { type Metadata } from "next"
import Link from "next/link"

import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"

export const metadata: Metadata = {
  title: "Terms of Service",
  description: "Kloudlite Terms of Service - Learn about the terms and conditions for using our development environment platform.",
}

export default function TermsPage() {
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
            <h1 className="text-3xl font-light tracking-tight sm:text-4xl">Terms of Service</h1>
            <p className="mt-4 text-sm text-muted-foreground">Last updated: January 1, 2024</p>
          </header>

          <div className="prose prose-gray dark:prose-invert mx-auto max-w-none">
            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">1. Agreement to Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                By accessing or using Kloudlite's development environment services, you agree to be bound by these Terms of Service and all applicable laws and regulations. If you do not agree with any of these terms, you are prohibited from using or accessing our services.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">2. Use of Service</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                Kloudlite provides cloud-based development environments as a service. You may use our services for lawful purposes only and in accordance with these Terms.
              </p>
              <p className="text-muted-foreground leading-relaxed">
                You agree not to use the service to:
              </p>
              <ul className="list-disc pl-6 mt-2 space-y-2 text-muted-foreground">
                <li>Violate any applicable laws or regulations</li>
                <li>Transmit any harmful or malicious code</li>
                <li>Attempt to gain unauthorized access to any systems</li>
                <li>Interfere with or disrupt the service or servers</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">3. Account Responsibilities</h2>
              <p className="text-muted-foreground leading-relaxed">
                You are responsible for maintaining the confidentiality of your account credentials and for all activities that occur under your account. You agree to notify us immediately of any unauthorized access or use of your account.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">4. Intellectual Property</h2>
              <p className="text-muted-foreground leading-relaxed">
                The service and its original content, features, and functionality are owned by Kloudlite and are protected by international copyright, trademark, patent, trade secret, and other intellectual property laws.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">5. Privacy</h2>
              <p className="text-muted-foreground leading-relaxed">
                Your use of our service is also governed by our Privacy Policy. Please review our Privacy Policy, which also governs the site and informs users of our data collection practices.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">6. Termination</h2>
              <p className="text-muted-foreground leading-relaxed">
                We may terminate or suspend your account and access to the service immediately, without prior notice or liability, for any reason whatsoever, including without limitation if you breach the Terms.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">7. Disclaimer</h2>
              <p className="text-muted-foreground leading-relaxed">
                The service is provided on an "AS IS" and "AS AVAILABLE" basis. Kloudlite disclaims all warranties of any kind, whether express or implied, including but not limited to implied warranties of merchantability, fitness for a particular purpose, and non-infringement.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">8. Limitation of Liability</h2>
              <p className="text-muted-foreground leading-relaxed">
                In no event shall Kloudlite, its directors, employees, partners, agents, suppliers, or affiliates, be liable for any indirect, incidental, special, consequential, or punitive damages resulting from your use of the service.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">9. Changes to Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                We reserve the right to modify or replace these Terms at any time. If a revision is material, we will provide at least 30 days notice prior to any new terms taking effect.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-medium mb-4">10. Contact Information</h2>
              <p className="text-muted-foreground leading-relaxed">
                If you have any questions about these Terms, please contact us at legal@kloudlite.io
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