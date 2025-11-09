import { KloudliteLogo } from '@/components/kloudlite-logo'
import Link from 'next/link'
import { Mail, MessageSquare, MapPin, Github, Twitter, Linkedin } from 'lucide-react'
import { GetStartedButton } from '@/components/get-started-button'
import ContactForm from './contact-form'

export default function ContactPage() {

  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Navigation Header */}
      <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 border-b backdrop-blur">
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6 lg:gap-8">
            <KloudliteLogo showText={true} linkToHome={true} />
            <div className="hidden items-center gap-6 md:flex">
              <Link
                href="/docs"
                className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
              >
                Pricing
              </Link>
              <Link
                href="/contact"
                className="text-foreground text-sm font-medium"
              >
                Contact
              </Link>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <GetStartedButton size="sm" className="hidden sm:flex" />
          </div>
        </nav>
      </header>

      {/* Contact Section */}
      <main className="flex-1 px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-[90rem]">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
              Get in Touch
            </h1>
            <p className="text-muted-foreground mx-auto mt-4 max-w-2xl text-lg">
              Have a question or feedback? We'd love to hear from you.
            </p>
          </div>

          <div className="mt-16 grid gap-8 lg:grid-cols-3">
            {/* Contact Form */}
            <div className="lg:col-span-2">
              <ContactForm />
            </div>

            {/* Contact Info */}
            <div className="space-y-6">
              {/* Email */}
              <div className="bg-card border-border rounded-lg border p-6">
                <div className="flex items-start gap-4">
                  <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-lg">
                    <Mail className="size-6" />
                  </div>
                  <div>
                    <h3 className="text-foreground font-semibold">Email Us</h3>
                    <p className="text-muted-foreground mt-1 text-sm">
                      Our team is here to help
                    </p>
                    <a
                      href="mailto:hello@kloudlite.io"
                      className="text-primary hover:text-primary/80 mt-2 inline-block text-sm font-medium"
                    >
                      hello@kloudlite.io
                    </a>
                  </div>
                </div>
              </div>

              {/* Support */}
              <div className="bg-card border-border rounded-lg border p-6">
                <div className="flex items-start gap-4">
                  <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-lg">
                    <MessageSquare className="size-6" />
                  </div>
                  <div>
                    <h3 className="text-foreground font-semibold">Live Chat</h3>
                    <p className="text-muted-foreground mt-1 text-sm">
                      Available Monday to Friday
                    </p>
                    <p className="text-muted-foreground mt-1 text-sm">
                      9:00 AM - 6:00 PM IST
                    </p>
                  </div>
                </div>
              </div>

              {/* Office */}
              <div className="bg-card border-border rounded-lg border p-6">
                <div className="flex items-start gap-4">
                  <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-lg">
                    <MapPin className="size-6" />
                  </div>
                  <div>
                    <h3 className="text-foreground font-semibold">Office</h3>
                    <p className="text-muted-foreground mt-1 text-sm">
                      Bangalore, India
                    </p>
                  </div>
                </div>
              </div>

              {/* Social Links */}
              <div className="bg-card border-border rounded-lg border p-6">
                <h3 className="text-foreground font-semibold">Follow Us</h3>
                <div className="mt-4 flex gap-4">
                  <a
                    href="https://github.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:bg-muted flex size-10 items-center justify-center rounded-lg transition-colors"
                  >
                    <Github className="text-muted-foreground size-5" />
                  </a>
                  <a
                    href="https://twitter.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:bg-muted flex size-10 items-center justify-center rounded-lg transition-colors"
                  >
                    <Twitter className="text-muted-foreground size-5" />
                  </a>
                  <a
                    href="https://linkedin.com/company/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:bg-muted flex size-10 items-center justify-center rounded-lg transition-colors"
                  >
                    <Linkedin className="text-muted-foreground size-5" />
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-border border-t py-8">
        <div className="mx-auto max-w-[90rem] px-4 sm:px-6 lg:px-8">
          <div className="text-muted-foreground flex flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-sm">
              © {new Date().getFullYear()} Kloudlite. All rights reserved.
            </p>
            <div className="flex gap-6 text-sm">
              <Link href="/docs" className="hover:text-foreground transition-colors">
                Documentation
              </Link>
              <Link href="/pricing" className="hover:text-foreground transition-colors">
                Pricing
              </Link>
              <Link href="/contact" className="hover:text-foreground transition-colors">
                Contact
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
