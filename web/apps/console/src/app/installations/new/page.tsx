'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button, RadioGroup, RadioGroupItem, Label } from '@kloudlite/ui'
import { Cloud, Server, ArrowRight } from 'lucide-react'

export default function NewInstallationPage() {
  const router = useRouter()
  const [hostingType, setHostingType] = useState<'kloudlite' | 'byoc'>('kloudlite')

  const handleContinue = () => {
    if (hostingType === 'kloudlite') {
      router.push('/installations/new-kl-cloud')
    } else {
      router.push('/installations/new-byoc')
    }
  }

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          <div className="absolute -top-4 -right-4 -z-10 opacity-5">
            <svg width="300" height="300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
          </div>

          {/* What's Next Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">What happens next?</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold">1</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Choose hosting type</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Select how you want to deploy</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Create installation</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Set up your installation details</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">3</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy & verify</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Deploy to cloud and confirm it&apos;s ready</p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Tips Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Quick Tips</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span><strong>Kloudlite Cloud</strong> is the fastest way to get started &mdash; no CLI or credentials needed</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span><strong>Bring your own Cloud</strong> gives you full control over your infrastructure</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* Right Column */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        {/* Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">
            New Installation
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Choose how you want to deploy Kloudlite
          </p>
        </div>

        {/* Hosting Type Card */}
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="p-8">
            <RadioGroup
              value={hostingType}
              onValueChange={(v) => setHostingType(v as 'kloudlite' | 'byoc')}
              className="grid gap-4"
            >
              <Label
                htmlFor="hosting-kloudlite"
                className={`flex items-start gap-4 rounded-lg border p-5 cursor-pointer transition-colors ${
                  hostingType === 'kloudlite'
                    ? 'border-primary bg-primary/5'
                    : 'border-foreground/10 hover:border-foreground/20'
                }`}
              >
                <RadioGroupItem value="kloudlite" id="hosting-kloudlite" className="mt-1" />
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <Cloud className="size-4 text-primary" />
                    <p className="text-sm font-semibold text-foreground">Kloudlite Cloud</p>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    We manage the infrastructure for you. No CLI or cloud credentials required.
                    Just fill in your installation details and we&apos;ll handle the rest.
                  </p>
                </div>
              </Label>

              <Label
                htmlFor="hosting-byoc"
                className={`flex items-start gap-4 rounded-lg border p-5 cursor-pointer transition-colors ${
                  hostingType === 'byoc'
                    ? 'border-primary bg-primary/5'
                    : 'border-foreground/10 hover:border-foreground/20'
                }`}
              >
                <RadioGroupItem value="byoc" id="hosting-byoc" className="mt-1" />
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <Server className="size-4 text-muted-foreground" />
                    <p className="text-sm font-semibold text-foreground">Bring your own Cloud</p>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Deploy to your own AWS, GCP, or Azure account using CLI commands.
                    You&apos;ll need cloud provider credentials and a configured CLI.
                  </p>
                </div>
              </Label>
            </RadioGroup>

            <div className="flex justify-end pt-6">
              <Button size="default" onClick={handleContinue}>
                Continue
                <ArrowRight className="ml-2 size-4" />
              </Button>
            </div>
          </div>
        </div>

        {/* Help text */}
        <div className="flex items-start gap-3 text-sm text-muted-foreground">
          <div className="flex-shrink-0 mt-0.5">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p>
              Need help deciding?{' '}
              <a
                href="https://docs.kloudlite.io"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary relative inline-block"
              >
                <span className="relative">
                  View our documentation
                  <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 hover:scale-x-100 transition-transform duration-300 origin-left" />
                </span>
              </a>
              {' '}for a comparison of hosting options.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
