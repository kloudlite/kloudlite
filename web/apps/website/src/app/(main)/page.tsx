import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import { TypingText } from './_components/typing-text'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'

// Website Landing Page Component
function WebsiteLandingPage() {
  return (
    <div className="bg-background flex min-h-screen flex-col overflow-x-hidden">
      <WebsiteHeader currentPage="home" />

      {/* Hero Section */}
      <main className="flex flex-1 items-center py-12 sm:py-16 w-full">
        <div className="mx-auto w-full max-w-[90rem] px-4 sm:px-6 lg:px-8">
          {/* Main Heading */}
          <div className="text-center">
            <h1 className="text-foreground text-2xl leading-tight font-bold tracking-tight sm:text-4xl md:text-5xl lg:text-6xl xl:text-7xl break-words">
              Platform of Development Environments
            </h1>
            <p className="text-muted-foreground mx-auto mt-4 max-w-3xl text-sm sm:text-base md:text-xl lg:text-2xl break-words">
              No Setup. No build. No Deploy. Just Code.
            </p>

            <p className="text-muted-foreground mx-auto mt-6 text-sm sm:text-base md:text-xl lg:text-2xl">
              For <TypingText />
            </p>

            {/* Visual Flow - Hidden on mobile and tablet */}
            <div className="mx-auto mt-14 w-full hidden md:block">
              <div className="flex flex-row items-center justify-center gap-3 lg:gap-4">
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Setup
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="border-info bg-info/10 text-info rounded-xl border-2 px-6 py-2.5 text-sm font-bold lg:px-9 lg:py-4 lg:text-lg text-center">
                  Code
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Build
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Deploy
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="border-success bg-success/10 text-success rounded-xl border-2 px-6 py-2.5 text-sm font-bold lg:px-9 lg:py-4 lg:text-lg text-center">
                  Test
                </div>
              </div>
            </div>

            <p className="text-muted-foreground mt-8 sm:mt-12 text-sm sm:text-base px-4">
              Designed to reduce development loop
            </p>

            {/* CTAs */}
            <div className="mt-6 flex flex-col items-center justify-center gap-4 sm:flex-row px-4">
              <GetStartedButton
                size="lg"
                className="w-full px-8 text-base font-semibold sm:w-auto"
              />
              <Button
                asChild
                variant="outline"
                size="lg"
                className="w-full px-8 text-base font-semibold sm:w-auto"
              >
                <Link href="/docs">Read Documentation</Link>
              </Button>
            </div>
          </div>
        </div>
      </main>

      <WebsiteFooter />
    </div>
  )
}

// Main page - website mode only shows landing page
export default async function HomePage() {
  return <WebsiteLandingPage />
}
