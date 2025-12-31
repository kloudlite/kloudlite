'use client'

import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { cn } from '@kloudlite/lib'
import { Download, Check, X } from 'lucide-react'

function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

function ColorSwatch({ name, value, hex, className }: { name: string; value: string; hex: string; className?: string }) {
  return (
    <div className="flex flex-col">
      <div className={cn('h-24 w-full rounded-lg border border-foreground/10', className)} style={{ backgroundColor: value }} />
      <p className="text-foreground mt-3 text-sm font-medium">{name}</p>
      <p className="text-foreground/50 text-xs font-mono">{hex}</p>
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
              <div className="py-20 lg:py-24">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground/40">B</span><span className="text-foreground">rand</span>{' '}
                    <span className="text-foreground/40">A</span><span className="text-foreground">ssets</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-md text-lg leading-relaxed">
                    Resources and guidelines for using
                    <br />
                    the Kloudlite brand.
                  </p>
                </div>
              </div>

              {/* Logo Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-t border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em]">Logo</h2>
                  <p className="text-foreground/50 mt-3 text-sm leading-relaxed">
                    Our logo represents the essence of Kloudlite - cloud development simplified.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10 border-b border-foreground/10">
                  <div className="grid sm:grid-cols-2 gap-8">
                    {/* Full Logo - Light */}
                    <div className="flex flex-col items-center justify-center p-8 bg-white rounded-lg border border-foreground/10">
                      <KloudliteLogo showText={true} linkToHome={false} className="scale-150" />
                      <p className="text-gray-500 mt-6 text-xs">Full Logo - Light Background</p>
                    </div>
                    {/* Full Logo - Dark */}
                    <div className="flex flex-col items-center justify-center p-8 bg-gray-900 rounded-lg">
                      <KloudliteLogo showText={true} linkToHome={false} variant="white" className="scale-150" />
                      <p className="text-gray-400 mt-6 text-xs">Full Logo - Dark Background</p>
                    </div>
                    {/* Icon Only - Light */}
                    <div className="flex flex-col items-center justify-center p-8 bg-white rounded-lg border border-foreground/10">
                      <KloudliteLogo showText={false} linkToHome={false} className="scale-[2]" />
                      <p className="text-gray-500 mt-6 text-xs">Icon Only - Light Background</p>
                    </div>
                    {/* Icon Only - Dark */}
                    <div className="flex flex-col items-center justify-center p-8 bg-gray-900 rounded-lg">
                      <KloudliteLogo showText={false} linkToHome={false} variant="white" className="scale-[2]" />
                      <p className="text-gray-400 mt-6 text-xs">Icon Only - Dark Background</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Colors Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em]">Colors</h2>
                  <p className="text-foreground/50 mt-3 text-sm leading-relaxed">
                    Our color palette is minimal and purposeful.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
                    <ColorSwatch name="Primary" value="#6366f1" hex="#6366F1" className="bg-primary" />
                    <ColorSwatch name="Black" value="#09090b" hex="#09090B" />
                    <ColorSwatch name="White" value="#ffffff" hex="#FFFFFF" />
                    <ColorSwatch name="Gray" value="#71717a" hex="#71717A" />
                  </div>
                </div>
              </div>

              {/* Typography Section */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em]">Typography</h2>
                  <p className="text-foreground/50 mt-3 text-sm leading-relaxed">
                    We use Inter for its clarity and modern feel.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <div className="space-y-6">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-2">Headings</p>
                      <p className="text-foreground text-4xl font-bold tracking-tight">Inter Bold</p>
                    </div>
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-2">Body</p>
                      <p className="text-foreground text-lg">Inter Regular - The quick brown fox jumps over the lazy dog.</p>
                    </div>
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-2">Code</p>
                      <p className="text-foreground font-mono text-sm bg-foreground/5 px-3 py-2 rounded">JetBrains Mono - const workspace = new Kloudlite()</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Usage Guidelines */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em]">Usage</h2>
                  <p className="text-foreground/50 mt-3 text-sm leading-relaxed">
                    Guidelines for using our brand correctly.
                  </p>
                </div>
                <div className="lg:col-span-2 p-8 lg:p-10">
                  <div className="grid sm:grid-cols-2 gap-6">
                    <div className="space-y-4">
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider flex items-center gap-2">
                        <Check className="h-4 w-4 text-green-500" /> Do
                      </p>
                      <ul className="space-y-2 text-foreground/70 text-sm">
                        <li>Use adequate spacing around the logo</li>
                        <li>Maintain original proportions</li>
                        <li>Use on contrasting backgrounds</li>
                        <li>Use official color palette</li>
                      </ul>
                    </div>
                    <div className="space-y-4">
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider flex items-center gap-2">
                        <X className="h-4 w-4 text-red-500" /> Don&apos;t
                      </p>
                      <ul className="space-y-2 text-foreground/70 text-sm">
                        <li>Stretch or distort the logo</li>
                        <li>Change the logo colors</li>
                        <li>Add effects or shadows</li>
                        <li>Place on busy backgrounds</li>
                      </ul>
                    </div>
                  </div>
                </div>
              </div>

              {/* Download Section */}
              <div className="p-8 lg:p-10 -mx-6 lg:-mx-12 flex flex-col sm:flex-row items-center justify-between gap-4">
                <div>
                  <p className="text-foreground font-medium">Need the assets?</p>
                  <p className="text-foreground/50 text-sm">Contact us for the full brand kit.</p>
                </div>
                <a
                  href="mailto:hello@kloudlite.io?subject=Brand%20Assets%20Request"
                  className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded-md hover:bg-primary/90 transition-colors"
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
