import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Download, Check, X } from 'lucide-react'

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

function LogoIcon({ fill = '#4f46e5' }: { fill?: string }) {
  return (
    <svg viewBox="0 0 130 131" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-12 w-auto">
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

function FullLogo({ iconFill = '#4f46e5', textFill = '#09090b' }: { iconFill?: string; textFill?: string }) {
  return (
    <svg viewBox="0 0 628 131" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-8 w-auto">
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

function ColorCard({ name, hex, className }: { name: string; hex: string; className?: string }) {
  return (
    <div className="group">
      <div className={cn('h-28 rounded-lg border border-foreground/20', className)} />
      <div className="mt-3 flex items-center justify-between">
        <div>
          <p className="text-foreground text-sm font-medium">{name}</p>
          <p className="text-foreground/50 text-xs font-mono">{hex}</p>
        </div>
      </div>
    </div>
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
              <div className="py-16 lg:py-20 border-b border-foreground/10">
                <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
                  Brand Guidelines
                </h1>
                <p className="text-foreground/60 mt-4 max-w-xl text-lg leading-relaxed">
                  Resources and guidelines for using the Kloudlite brand consistently across all platforms and materials.
                </p>
              </div>

              {/* Logo Section */}
              <div className="py-12 border-b border-foreground/10">
                <div className="flex items-center justify-between mb-8">
                  <div>
                    <h2 className="text-foreground text-2xl font-bold tracking-tight">Logo</h2>
                    <p className="text-foreground/50 mt-1 text-sm">Primary brand mark</p>
                  </div>
                </div>

                <div className="grid md:grid-cols-2 gap-6">
                  {/* Light Background */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-white p-12 flex items-center justify-center min-h-[180px]">
                      <FullLogo iconFill="#4f46e5" textFill="#09090b" />
                    </div>
                    <div className="bg-neutral-100 px-4 py-3 border-t border-neutral-200">
                      <p className="text-neutral-600 text-sm font-medium">Light Background</p>
                    </div>
                  </div>

                  {/* Dark Background */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-neutral-950 p-12 flex items-center justify-center min-h-[180px]">
                      <FullLogo iconFill="#4f46e5" textFill="#ffffff" />
                    </div>
                    <div className="bg-neutral-900 px-4 py-3 border-t border-neutral-800">
                      <p className="text-neutral-400 text-sm font-medium">Dark Background</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Icon Section */}
              <div className="py-12 border-b border-foreground/10">
                <div className="flex items-center justify-between mb-8">
                  <div>
                    <h2 className="text-foreground text-2xl font-bold tracking-tight">Icon</h2>
                    <p className="text-foreground/50 mt-1 text-sm">For app icons, favicons, and small spaces</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {/* Primary on Light */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-white p-6 flex items-center justify-center aspect-square">
                      <LogoIcon fill="#4f46e5" />
                    </div>
                    <div className="bg-neutral-100 px-3 py-2 border-t border-neutral-200">
                      <p className="text-neutral-600 text-xs font-medium">Primary / Light</p>
                    </div>
                  </div>

                  {/* Primary on Dark */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-neutral-950 p-6 flex items-center justify-center aspect-square">
                      <LogoIcon fill="#4f46e5" />
                    </div>
                    <div className="bg-neutral-900 px-3 py-2 border-t border-neutral-800">
                      <p className="text-neutral-400 text-xs font-medium">Primary / Dark</p>
                    </div>
                  </div>

                  {/* White on Dark */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-neutral-950 p-6 flex items-center justify-center aspect-square">
                      <LogoIcon fill="#ffffff" />
                    </div>
                    <div className="bg-neutral-900 px-3 py-2 border-t border-neutral-800">
                      <p className="text-neutral-400 text-xs font-medium">White / Dark</p>
                    </div>
                  </div>

                  {/* Black on Light */}
                  <div className="rounded-lg overflow-hidden border border-foreground/20">
                    <div className="bg-white p-6 flex items-center justify-center aspect-square">
                      <LogoIcon fill="#09090b" />
                    </div>
                    <div className="bg-neutral-100 px-3 py-2 border-t border-neutral-200">
                      <p className="text-neutral-600 text-xs font-medium">Black / Light</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Colors Section */}
              <div className="py-12 border-b border-foreground/10">
                <div className="mb-8">
                  <h2 className="text-foreground text-2xl font-bold tracking-tight">Colors</h2>
                  <p className="text-foreground/50 mt-1 text-sm">Brand color palette</p>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <ColorCard name="Primary" hex="#4F46E5" className="bg-[#4f46e5]" />
                  <ColorCard name="Black" hex="#09090B" className="bg-[#09090b]" />
                  <ColorCard name="White" hex="#FFFFFF" className="bg-white" />
                  <ColorCard name="Gray" hex="#71717A" className="bg-[#71717a]" />
                </div>
              </div>

              {/* Typography Section */}
              <div className="py-12 border-b border-foreground/10">
                <div className="mb-8">
                  <h2 className="text-foreground text-2xl font-bold tracking-tight">Typography</h2>
                  <p className="text-foreground/50 mt-1 text-sm">Font families and styles</p>
                </div>

                <div className="space-y-4">
                  <div className="p-6 rounded-lg border border-foreground/10">
                    <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-3">Primary Font</p>
                    <p className="text-foreground text-4xl font-bold tracking-tight">Inter</p>
                    <p className="text-foreground/60 mt-2 text-sm">Used for headings, body text, and UI elements</p>
                  </div>
                  <div className="p-6 rounded-lg border border-foreground/10">
                    <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-3">Monospace Font</p>
                    <p className="text-foreground text-3xl font-mono">JetBrains Mono</p>
                    <p className="text-foreground/60 mt-2 text-sm">Used for code snippets and technical content</p>
                  </div>
                </div>
              </div>

              {/* Usage Guidelines */}
              <div className="py-12 border-b border-foreground/10">
                <div className="mb-8">
                  <h2 className="text-foreground text-2xl font-bold tracking-tight">Usage Guidelines</h2>
                  <p className="text-foreground/50 mt-1 text-sm">Best practices for brand consistency</p>
                </div>

                <div className="grid md:grid-cols-2 gap-4">
                  <div className="p-6 rounded-lg border border-foreground/10">
                    <p className="text-green-500 text-sm font-semibold uppercase tracking-wider flex items-center gap-2 mb-4">
                      <Check className="h-4 w-4" /> Do
                    </p>
                    <ul className="space-y-3 text-foreground/70 text-sm">
                      <li className="flex items-start gap-2">
                        <span className="text-green-500 mt-0.5">•</span>
                        Use adequate clear space around the logo
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-green-500 mt-0.5">•</span>
                        Maintain original proportions when scaling
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-green-500 mt-0.5">•</span>
                        Use on high-contrast backgrounds
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-green-500 mt-0.5">•</span>
                        Use official brand colors only
                      </li>
                    </ul>
                  </div>

                  <div className="p-6 rounded-lg border border-foreground/10">
                    <p className="text-red-500 text-sm font-semibold uppercase tracking-wider flex items-center gap-2 mb-4">
                      <X className="h-4 w-4" /> Don&apos;t
                    </p>
                    <ul className="space-y-3 text-foreground/70 text-sm">
                      <li className="flex items-start gap-2">
                        <span className="text-red-500 mt-0.5">•</span>
                        Stretch, skew, or distort the logo
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-red-500 mt-0.5">•</span>
                        Change or modify logo colors
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-red-500 mt-0.5">•</span>
                        Add shadows, gradients, or effects
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-red-500 mt-0.5">•</span>
                        Place on busy or low-contrast backgrounds
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Download CTA */}
              <div className="py-12 flex flex-col sm:flex-row items-center justify-between gap-6">
                <div>
                  <h3 className="text-foreground text-lg font-semibold">Need the full brand kit?</h3>
                  <p className="text-foreground/50 text-sm mt-1">Get logos in all formats, color specifications, and more.</p>
                </div>
                <a
                  href="mailto:hello@kloudlite.io?subject=Brand%20Assets%20Request"
                  className="inline-flex items-center gap-2 px-5 py-2.5 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors"
                >
                  <Download className="h-4 w-4" />
                  Request Brand Kit
                </a>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
