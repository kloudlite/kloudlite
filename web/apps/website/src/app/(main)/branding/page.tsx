'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { cn } from '@kloudlite/lib'
import { Download, Check, X } from 'lucide-react'
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

// Animated GridContainer with border pulses
function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <style jsx>{`
        @keyframes pulseTopLeftToRight {
          0% { left: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes pulseRightTopToBottom {
          0% { top: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes pulseBottomRightToLeft {
          0% { right: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { right: 100%; opacity: 0; }
        }
        @keyframes pulseLeftBottomToTop {
          0% { bottom: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { bottom: 100%; opacity: 0; }
        }
        .pulse-top {
          animation: pulseTopLeftToRight 4s ease-in-out infinite;
        }
        .pulse-right {
          animation: pulseRightTopToBottom 4s ease-in-out infinite 1s;
        }
        .pulse-bottom {
          animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s;
        }
        .pulse-left {
          animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s;
        }
      `}</style>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Static border lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

        {/* Animated pulses */}
        <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />

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

// Feature card container
function FeatureCardContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn("relative p-8 lg:p-12 bg-foreground/[0.015]", className)}>
      {children}
    </div>
  )
}

function LogoIcon({ fill = '#2258E5', className }: { fill?: string; className?: string }) {
  return (
    <svg viewBox="0 0 130 131" fill="none" xmlns="http://www.w3.org/2000/svg" className={cn('h-12 w-auto', className)}>
      <path
        d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z"
        fill={fill}
      />
      <path
        d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z"
        fill={fill}
      />
    </svg>
  )
}

function FullLogo({ iconFill = '#2258E5', textFill = '#09090b', className }: { iconFill?: string; textFill?: string; className?: string }) {
  return (
    <svg viewBox="0 0 628 131" fill="none" xmlns="http://www.w3.org/2000/svg" className={cn('h-8 w-auto', className)}>
      <path
        d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z"
        fill={iconFill}
      />
      <path
        d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z"
        fill={iconFill}
      />
      <path d="M164.241 113.166V17.8325H180.841V73.6742L201.591 45.6076H220.333L195.968 78.3597L220.868 113.166H202.126L180.841 83.4467V113.166H164.241Z" fill={textFill} />
      <path d="M588.188 86.6906C588.274 90.651 589.308 93.5352 591.288 95.3432C593.354 97.0652 596.281 97.9261 600.07 97.9261C608.077 97.9261 615.223 97.6678 621.508 97.1513L625.124 96.7638L625.382 109.549C615.481 111.96 606.527 113.165 598.52 113.165C588.791 113.165 581.731 110.582 577.34 105.416C572.949 100.251 570.754 91.8564 570.754 80.2334C570.754 57.0736 580.268 45.4937 599.295 45.4937C618.064 45.4937 627.448 55.2225 627.448 74.6802L626.157 86.6906H588.188ZM610.401 73.5179C610.401 68.3521 609.583 64.7792 607.947 62.7989C606.312 60.7326 603.427 59.6995 599.295 59.6995C595.248 59.6995 592.364 60.7757 590.642 62.9281C589.006 64.9944 588.145 68.5243 588.059 73.5179H610.401Z" fill={textFill} />
      <path d="M560.42 61.7669H544.536V88.2414C544.536 90.8243 544.579 92.6754 544.665 93.7946C544.837 94.8278 545.311 95.7318 546.086 96.5067C546.946 97.2815 548.238 97.669 549.96 97.669L559.775 97.4107L560.55 111.229C554.781 112.521 550.39 113.166 547.377 113.166C539.628 113.166 534.333 111.444 531.492 108C528.651 104.471 527.23 98.0133 527.23 88.6289V61.7669V45.4948V17.8574H544.536V45.4948H560.42V61.7669Z" fill={textFill} />
      <path d="M496.661 113.166V45.4948H513.966V113.166H496.661ZM496.661 35.421V17.8574H513.966V35.421H496.661Z" fill={textFill} />
      <path d="M466.091 113.165L466.091 17.8667H483.396L483.397 113.165H466.091Z" fill={textFill} />
      <path d="M452.826 17.8667L452.826 113.165H435.65V108.904C429.624 111.745 424.415 113.165 420.024 113.165C410.639 113.165 404.096 110.453 400.394 105.029C396.692 99.6052 394.841 91.0387 394.841 79.3296C394.841 67.5345 397.036 58.9679 401.427 53.63C405.904 48.2059 412.62 45.4939 421.574 45.4939C424.329 45.4939 428.16 45.9244 433.067 46.7854L435.521 47.3019L435.521 17.8667H452.826ZM433.713 96.1183L435.521 95.7309V61.7661C430.786 60.9051 426.567 60.4746 422.865 60.4746C415.891 60.4746 412.404 66.6735 412.404 79.0714C412.404 85.7868 413.179 90.5652 414.729 93.4063C416.279 96.2475 418.819 97.6681 422.348 97.6681C425.965 97.6681 429.753 97.1515 433.713 96.1183Z" fill={textFill} />
      <path d="M367.331 45.4937H384.636V113.165H367.46V107.999C361.261 111.443 355.88 113.165 351.317 113.165C342.363 113.165 336.337 110.711 333.237 105.804C330.138 100.81 328.588 92.5021 328.588 80.8791V45.4937H345.893V81.1374C345.893 87.5085 346.41 91.8563 347.443 94.1809C348.476 96.5055 350.973 97.6678 354.933 97.6678C358.721 97.6678 362.295 97.0652 365.652 95.8598L367.331 95.3432V45.4937Z" fill={textFill} />
      <path d="M265.823 54.4046C270.386 48.464 278.006 45.4937 288.682 45.4937C299.358 45.4937 306.977 48.464 311.54 54.4046C316.103 60.2591 318.385 68.5243 318.385 79.2002C318.385 101.844 308.484 113.165 288.682 113.165C268.88 113.165 258.979 101.844 258.979 79.2002C258.979 68.5243 261.26 60.2591 265.823 54.4046ZM279.125 93.7935C280.933 96.893 284.119 98.4427 288.682 98.4427C293.245 98.4427 296.387 96.893 298.109 93.7935C299.917 90.694 300.821 85.8296 300.821 79.2002C300.821 72.5708 299.917 67.7495 298.109 64.7361C296.387 61.7227 293.245 60.2161 288.682 60.2161C284.119 60.2161 280.933 61.7227 279.125 64.7361C277.403 67.7495 276.542 72.5708 276.542 79.2002C276.542 85.8296 277.403 90.694 279.125 93.7935Z" fill={textFill} />
      <path d="M231.468 113.165L231.071 17.8667H248.377L248.774 113.165H231.468Z" fill={textFill} />
    </svg>
  )
}

export default function BrandingPage() {
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
                  <PageHeroTitle accentedWord="Guidelines">
                    Brand
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Resources and guidelines for using the Kloudlite brand assets consistently across all platforms.
                  </p>
                </div>
              </div>

              {/* Content */}
              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Logo Section Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Brand <span className="relative inline-block">
                      <span className="relative z-10">Assets.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-3xl">
                    Download and use our logo, icon, and color palette to maintain consistency across all brand materials.
                  </p>
                </div>

                {/* Logo Section */}
                <FeatureCardContainer className="sm:col-span-2 border-b border-foreground/10">
                  <h3 className="text-foreground text-xl font-semibold tracking-tight mb-2">Logo</h3>
                  <p className="text-muted-foreground text-sm mb-6">The primary Kloudlite wordmark for general use.</p>
                  <div className="grid grid-cols-2 border border-foreground/10">
                    <div className="flex items-center justify-center p-12 bg-white">
                      <FullLogo iconFill="#2258E5" textFill="#09090b" className="h-10" />
                    </div>
                    <div className="flex items-center justify-center p-12 bg-neutral-950 border-l border-foreground/10">
                      <FullLogo iconFill="#2258E5" textFill="#ffffff" className="h-10" />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 text-center border-x border-b border-foreground/10">
                    <p className="py-3 text-xs text-muted-foreground font-medium border-r border-foreground/10">Light Background</p>
                    <p className="py-3 text-xs text-muted-foreground font-medium">Dark Background</p>
                  </div>
                </FeatureCardContainer>

                {/* Spacer */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Icon Section */}
                <FeatureCardContainer className="sm:col-span-2 border-b border-foreground/10">
                  <h3 className="text-foreground text-xl font-semibold tracking-tight mb-2">Icon</h3>
                  <p className="text-muted-foreground text-sm mb-6">The logomark for favicons, app icons, and compact spaces.</p>
                  <div className="grid grid-cols-4 border border-foreground/10">
                    <div className="flex items-center justify-center p-8 aspect-square bg-white">
                      <LogoIcon fill="#2258E5" />
                    </div>
                    <div className="flex items-center justify-center p-8 aspect-square bg-neutral-950 border-l border-foreground/10">
                      <LogoIcon fill="#2258E5" />
                    </div>
                    <div className="flex items-center justify-center p-8 aspect-square bg-neutral-950 border-l border-foreground/10">
                      <LogoIcon fill="#ffffff" />
                    </div>
                    <div className="flex items-center justify-center p-8 aspect-square bg-white border-l border-foreground/10">
                      <LogoIcon fill="#09090b" />
                    </div>
                  </div>
                  <div className="grid grid-cols-4 text-center border-x border-b border-foreground/10">
                    <p className="py-3 text-xs text-muted-foreground font-medium border-r border-foreground/10">Primary</p>
                    <p className="py-3 text-xs text-muted-foreground font-medium border-r border-foreground/10">Primary</p>
                    <p className="py-3 text-xs text-muted-foreground font-medium border-r border-foreground/10">White</p>
                    <p className="py-3 text-xs text-muted-foreground font-medium">Black</p>
                  </div>
                </FeatureCardContainer>

                {/* Spacer */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-3/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Colors Section */}
                <FeatureCardContainer className="sm:col-span-2 border-b border-foreground/10">
                  <h3 className="text-foreground text-xl font-semibold tracking-tight mb-2">Colors</h3>
                  <p className="text-muted-foreground text-sm mb-6">The official brand color palette.</p>
                  <div className="grid grid-cols-4 border border-foreground/10">
                    <div className="flex flex-col">
                      <div className="aspect-[4/3] bg-[#2258E5]" />
                      <div className="p-4 border-t border-foreground/10">
                        <p className="text-foreground text-sm font-medium">Primary</p>
                        <p className="text-muted-foreground text-xs font-mono mt-1">#2258E5</p>
                      </div>
                    </div>
                    <div className="flex flex-col border-l border-foreground/10">
                      <div className="aspect-[4/3] bg-[#09090b]" />
                      <div className="p-4 border-t border-foreground/10">
                        <p className="text-foreground text-sm font-medium">Black</p>
                        <p className="text-muted-foreground text-xs font-mono mt-1">#09090B</p>
                      </div>
                    </div>
                    <div className="flex flex-col border-l border-foreground/10">
                      <div className="aspect-[4/3] bg-white border-b border-foreground/10" />
                      <div className="p-4">
                        <p className="text-foreground text-sm font-medium">White</p>
                        <p className="text-muted-foreground text-xs font-mono mt-1">#FFFFFF</p>
                      </div>
                    </div>
                    <div className="flex flex-col border-l border-foreground/10">
                      <div className="aspect-[4/3] bg-[#71717a]" />
                      <div className="p-4 border-t border-foreground/10">
                        <p className="text-foreground text-sm font-medium">Gray</p>
                        <p className="text-muted-foreground text-xs font-mono mt-1">#71717A</p>
                      </div>
                    </div>
                  </div>
                </FeatureCardContainer>

                {/* Spacer */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Typography Section */}
                <FeatureCardContainer className="sm:col-span-2 border-b border-foreground/10">
                  <h3 className="text-foreground text-xl font-semibold tracking-tight mb-2">Typography</h3>
                  <p className="text-muted-foreground text-sm mb-6">The typefaces used across Kloudlite products.</p>
                  <div className="grid grid-cols-2 border border-foreground/10">
                    <div className="p-8">
                      <p className="text-muted-foreground text-xs font-medium uppercase tracking-wider mb-4">Primary</p>
                      <p className="text-foreground text-4xl font-semibold tracking-tight">Inter</p>
                      <p className="text-muted-foreground mt-3 text-sm">Used for headings, body text, and UI elements.</p>
                    </div>
                    <div className="p-8 border-l border-foreground/10">
                      <p className="text-muted-foreground text-xs font-medium uppercase tracking-wider mb-4">Monospace</p>
                      <p className="text-foreground text-3xl font-mono">JetBrains Mono</p>
                      <p className="text-muted-foreground mt-3 text-sm">Used for code snippets and technical content.</p>
                    </div>
                  </div>
                </FeatureCardContainer>

                {/* Spacer */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Usage Guidelines */}
                <FeatureCardContainer className="sm:col-span-2 border-b border-foreground/10">
                  <h3 className="text-foreground text-xl font-semibold tracking-tight mb-2">Usage Guidelines</h3>
                  <p className="text-muted-foreground text-sm mb-6">Best practices for using the Kloudlite brand.</p>
                  <div className="grid grid-cols-2 border border-foreground/10">
                    <div className="p-8">
                      <p className="text-green-600 dark:text-green-500 text-xs font-medium uppercase tracking-wider flex items-center gap-2 mb-6">
                        <Check className="h-4 w-4" /> Do
                      </p>
                      <ul className="space-y-3 text-muted-foreground text-sm">
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Use adequate clear space around the logo
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Maintain original proportions when scaling
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Use on high contrast backgrounds
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Use official brand colors only
                        </li>
                      </ul>
                    </div>
                    <div className="p-8 border-l border-foreground/10">
                      <p className="text-red-600 dark:text-red-500 text-xs font-medium uppercase tracking-wider flex items-center gap-2 mb-6">
                        <X className="h-4 w-4" /> Don&apos;t
                      </p>
                      <ul className="space-y-3 text-muted-foreground text-sm">
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Stretch or distort the logo
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Change or alter the logo colors
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Add shadows, gradients, or effects
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="text-muted-foreground/50">—</span>
                          Place on busy or low-contrast backgrounds
                        </li>
                      </ul>
                    </div>
                  </div>
                </FeatureCardContainer>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-3/4 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Download CTA */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                  <div>
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                      Download Brand Kit
                    </h2>
                    <p className="text-muted-foreground mt-2 text-base lg:text-lg">
                      Get logos and icons in SVG format with complete usage guidelines.
                    </p>
                  </div>
                  <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                    <Button size="lg" asChild>
                      <a href="/kloudlite-brand-kit.zip" download>
                        <Download className="h-4 w-4" />
                        Download Brand Kit
                      </a>
                    </Button>
                    <Button variant="outline" size="lg" asChild>
                      <Link href="/contact">Contact Us</Link>
                    </Button>
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
