'use client'

import { useEffect, useState } from 'react'
import { Check, Loader2, Copy, ExternalLink, ChevronDown, Github } from 'lucide-react'

// Preview frame wrapper with browser chrome
function PreviewFrame({ children, url = 'console.kloudlite.io' }: { children: React.ReactNode; url?: string }) {
  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50">
      {/* Browser chrome */}
      <div className="bg-zinc-800 px-3 sm:px-4 py-2 sm:py-2.5 flex items-center gap-2 sm:gap-3">
        <div className="flex gap-1.5">
          <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#ff5f57]" />
          <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#febc2e]" />
          <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#28c840]" />
        </div>
        <div className="flex-1 flex justify-center">
          <div className="bg-zinc-700/50 rounded px-2 sm:px-3 py-1 text-zinc-400 text-[9px] sm:text-[10px] flex items-center gap-1 sm:gap-2">
            <svg className="w-2.5 sm:w-3 h-2.5 sm:h-3 hidden sm:block" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <span className="truncate max-w-[120px] sm:max-w-none">{url}</span>
          </div>
        </div>
        <div className="text-[8px] sm:text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-1.5 sm:px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>
      {/* Content */}
      <div className="bg-card overflow-x-auto">
        {children}
      </div>
    </div>
  )
}

// Hook for step-based animation with looping
function useAnimationSteps(totalSteps: number, stepDuration: number = 1000, pauseAtEnd: number = 2000) {
  const [step, setStep] = useState(0)

  useEffect(() => {
    let currentStep = 0

    const stepInterval = setInterval(() => {
      currentStep = (currentStep + 1) % (totalSteps + Math.ceil(pauseAtEnd / stepDuration))
      setStep(currentStep < totalSteps ? currentStep : 0)
    }, stepDuration)

    return () => clearInterval(stepInterval)
  }, [totalSteps, stepDuration, pauseAtEnd])

  return step
}

// Progress indicator showing 3 steps
function InstallationProgress({ currentStep }: { currentStep: number }) {
  const steps = ['Create', 'Install', 'Complete']

  return (
    <div className="flex items-center justify-center gap-2 mb-6">
      {steps.map((label, index) => (
        <div key={label} className="flex items-center gap-2">
          <div className={`w-6 h-6 rounded-none flex items-center justify-center text-[10px] font-medium transition-colors ${
            index < currentStep
              ? 'bg-primary text-primary-foreground'
              : index === currentStep
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground'
          }`}>
            {index < currentStep ? <Check className="h-3 w-3" /> : index + 1}
          </div>
          <span className={`text-[10px] ${index === currentStep ? 'font-medium' : 'text-muted-foreground'}`}>{label}</span>
          {index < steps.length - 1 && <div className={`w-8 h-px ${index < currentStep ? 'bg-primary' : 'bg-muted'}`} />}
        </div>
      ))}
    </div>
  )
}

// Google icon component
function GoogleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
      <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
      <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
      <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
      <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
    </svg>
  )
}

// Microsoft icon component
function MicrosoftIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
      <path fill="#F25022" d="M1 1h10v10H1z"/>
      <path fill="#00A4EF" d="M1 13h10v10H1z"/>
      <path fill="#7FBA00" d="M13 1h10v10H13z"/>
      <path fill="#FFB900" d="M13 13h10v10H13z"/>
    </svg>
  )
}

export function SignUpPreview() {
  const step = useAnimationSteps(4, 1500, 2000)

  const hoveredButton = step === 1 ? 'github' : step === 2 ? 'google' : step === 3 ? 'microsoft' : null

  return (
    <PreviewFrame url="auth.kloudlite.io">
      <div className="text-xs flex flex-col sm:flex-row min-h-[280px] sm:min-h-[320px]">
        {/* Left branding section - hidden on mobile */}
        <div className="hidden sm:flex w-2/5 bg-zinc-900 text-white p-6 flex-col justify-between">
          <div>
            <div className="font-bold text-lg mb-4">Kloudlite</div>
            <p className="text-zinc-400 text-[11px] leading-relaxed">
              Remote local environments with service mesh and VPN for seamless cloud development.
            </p>
          </div>
          <div className="space-y-3">
            <div className="flex items-start gap-2">
              <Check className="h-3 w-3 text-green-400 mt-0.5 flex-shrink-0" />
              <span className="text-zinc-300 text-[10px]">Isolated dev environments</span>
            </div>
            <div className="flex items-start gap-2">
              <Check className="h-3 w-3 text-green-400 mt-0.5 flex-shrink-0" />
              <span className="text-zinc-300 text-[10px]">Cloud-native workspaces</span>
            </div>
            <div className="flex items-start gap-2">
              <Check className="h-3 w-3 text-green-400 mt-0.5 flex-shrink-0" />
              <span className="text-zinc-300 text-[10px]">Bring your own compute</span>
            </div>
          </div>
        </div>

        {/* Right form section */}
        <div className="flex-1 p-4 sm:p-6 flex items-center justify-center bg-background">
          <div className="w-full max-w-xs space-y-4">
            <div className="text-center mb-4 sm:mb-6">
              <p className="font-medium text-base">Get started for free</p>
              <p className="text-muted-foreground text-[11px] mt-1">Sign in with your preferred provider</p>
            </div>

            <div className="space-y-3">
              <button className={`w-full border py-2.5 px-4 flex items-center justify-center gap-2 transition-all ${hoveredButton === 'github' ? 'border-primary bg-primary/5 scale-[1.02]' : 'hover:bg-muted'}`}>
                <Github className="h-4 w-4" />
                <span>Continue with GitHub</span>
              </button>
              <button className={`w-full border py-2.5 px-4 flex items-center justify-center gap-2 transition-all ${hoveredButton === 'google' ? 'border-primary bg-primary/5 scale-[1.02]' : 'hover:bg-muted'}`}>
                <GoogleIcon className="h-4 w-4" />
                <span>Continue with Google</span>
              </button>
              <button className={`w-full border py-2.5 px-4 flex items-center justify-center gap-2 transition-all ${hoveredButton === 'microsoft' ? 'border-primary bg-primary/5 scale-[1.02]' : 'hover:bg-muted'}`}>
                <MicrosoftIcon className="h-4 w-4" />
                <span>Continue with Microsoft</span>
              </button>
            </div>

            <p className="text-center text-muted-foreground text-[10px] mt-4">
              By signing in, you agree to our Terms of Service
            </p>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function CreateInstallationPreview() {
  const step = useAnimationSteps(7, 1200, 2000)

  const nameText = step >= 1 ? (step >= 2 ? 'My Production Cluster' : 'My Producti') : ''
  const descText = step >= 3 ? 'Production environment for the engineering team' : ''
  const subdomainText = step >= 4 ? 'my-prod-cluster' : ''
  const checking = step === 5
  const available = step >= 6

  return (
    <PreviewFrame>
      <div className="text-xs">
        {/* Header */}
        <div className="bg-background border-b px-3 sm:px-4 py-2 sm:py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-muted-foreground hidden sm:inline">Back</span>
            <span className="font-bold">Kloudlite</span>
          </div>
        </div>

        <div className="p-4 sm:p-6 max-w-md mx-auto">
          <div className="hidden sm:block"><InstallationProgress currentStep={0} /></div>

        <h2 className="font-semibold text-base mb-1">Create Installation</h2>
        <p className="text-muted-foreground text-[11px] mb-6">Set up your Kloudlite installation details</p>

        <div className="space-y-4">
          {/* Installation Name */}
          <div>
            <label className="text-[11px] font-medium mb-1.5 block">Installation Name</label>
            <div className={`border h-9 flex items-center px-3 transition-colors ${step >= 1 && step < 3 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
              {nameText}<span className={`${step >= 1 && step < 3 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
            </div>
            <p className="text-muted-foreground text-[10px] mt-1">3-50 characters, alphanumeric with spaces or hyphens</p>
          </div>

          {/* Description */}
          <div>
            <label className="text-[11px] font-medium mb-1.5 block">Description <span className="text-muted-foreground">(optional)</span></label>
            <div className={`border h-16 flex items-start p-3 transition-colors ${step === 3 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
              <span className="text-[11px]">{descText}<span className={`${step === 3 ? 'animate-pulse' : 'opacity-0'}`}>|</span></span>
            </div>
          </div>

          {/* Subdomain */}
          <div>
            <label className="text-[11px] font-medium mb-1.5 block">Installation Subdomain</label>
            <div className="flex items-center gap-2">
              <div className={`border h-9 flex-1 flex items-center px-3 transition-colors ${step >= 4 && step < 6 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                {subdomainText}<span className={`${step >= 4 && step < 6 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
              </div>
              <span className="text-muted-foreground">.kloudlite.io</span>
            </div>
            <div className="flex items-center gap-1.5 mt-1.5">
              {checking && (
                <div className="flex items-center gap-1 text-muted-foreground animate-in fade-in">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  <span className="text-[10px]">Checking availability...</span>
                </div>
              )}
              {available && (
                <div className="flex items-center gap-1 text-green-600 animate-in fade-in">
                  <Check className="h-3 w-3" />
                  <span className="text-[10px]">Subdomain is available</span>
                </div>
              )}
            </div>
          </div>

          <button className={`w-full bg-primary text-primary-foreground py-2.5 mt-2 transition-all ${step >= 6 ? 'opacity-100' : 'opacity-50'}`}>
            Continue
          </button>
        </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function CloudProviderPreview() {
  const step = useAnimationSteps(6, 1500, 2000)

  const activeTab = step < 2 ? 'aws' : step < 4 ? 'gcp' : 'azure'
  const regionSelected = step >= 1

  const selectedRegion = {
    aws: 'us-east-1',
    gcp: 'us-central1',
    azure: 'eastus'
  }

  return (
    <PreviewFrame>
      <div className="text-xs">
        {/* Header */}
        <div className="bg-background border-b px-3 sm:px-4 py-2 sm:py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-muted-foreground hidden sm:inline">Back</span>
            <span className="font-bold">Kloudlite</span>
          </div>
        </div>

        <div className="p-4 sm:p-6 max-w-lg mx-auto">
        <div className="hidden sm:block"><InstallationProgress currentStep={1} /></div>

        <h2 className="font-semibold text-base mb-1">Install Kloudlite</h2>
        <p className="text-muted-foreground text-[11px] mb-6">Deploy Kloudlite on your cloud infrastructure</p>

        {/* Cloud Provider Tabs */}
        <div className="flex border-b mb-4">
          {['aws', 'gcp', 'azure'].map((provider) => (
            <button
              key={provider}
              className={`px-4 py-2 text-[11px] font-medium transition-colors border-b-2 -mb-px ${
                activeTab === provider
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {provider.toUpperCase()}
            </button>
          ))}
        </div>

        {/* Region Selection */}
        <div className="space-y-4">
          <div>
            <label className="text-[11px] font-medium mb-1.5 block">Select Region</label>
            <div className={`border h-9 flex items-center justify-between px-3 transition-colors ${step === 1 || step === 3 || step === 5 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
              <span className={regionSelected ? '' : 'text-muted-foreground'}>
                {regionSelected ? selectedRegion[activeTab as keyof typeof selectedRegion] : 'Select a region...'}
              </span>
              <ChevronDown className="h-3 w-3 text-muted-foreground" />
            </div>
          </div>

          {/* Simple World Map Representation */}
          <div className="border bg-muted/30 p-4 rounded">
            <div className="flex items-center justify-between mb-2">
              <span className="text-[10px] text-muted-foreground">Selected Region</span>
              <span className="text-[10px] font-medium">{selectedRegion[activeTab as keyof typeof selectedRegion]}</span>
            </div>
            <div className="h-24 bg-background border rounded flex items-center justify-center relative overflow-hidden">
              {/* Simplified map dots */}
              <div className="absolute inset-0 flex items-center justify-center opacity-20">
                <div className="grid grid-cols-12 gap-1">
                  {Array.from({ length: 48 }).map((_, i) => (
                    <div key={i} className="w-1.5 h-1.5 rounded-none bg-foreground" />
                  ))}
                </div>
              </div>
              {/* Active region indicator */}
              <div className="absolute animate-pulse">
                <div className="w-3 h-3 rounded-none bg-primary" />
                <div className="absolute inset-0 w-3 h-3 rounded-none bg-primary animate-ping opacity-50" />
              </div>
            </div>
          </div>

          {/* Prerequisites */}
          <div className="bg-muted/50 p-3 rounded space-y-2">
            <p className="text-[10px] font-medium">Prerequisites for {activeTab.toUpperCase()}</p>
            <ul className="text-[10px] text-muted-foreground space-y-1">
              <li className="flex items-center gap-1.5">
                <Check className="h-2.5 w-2.5 text-green-500" />
                {activeTab === 'aws' && 'AWS CLI configured with credentials'}
                {activeTab === 'gcp' && 'gcloud CLI authenticated'}
                {activeTab === 'azure' && 'Azure CLI logged in'}
              </li>
              <li className="flex items-center gap-1.5">
                <Check className="h-2.5 w-2.5 text-green-500" />
                Terraform installed
              </li>
              <li className="flex items-center gap-1.5">
                <Check className="h-2.5 w-2.5 text-green-500" />
                kubectl installed
              </li>
            </ul>
          </div>

          <button className="w-full bg-primary text-primary-foreground py-2.5">
            Generate Install Command
          </button>
        </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function InstallCommandPreview() {
  const step = useAnimationSteps(5, 1500, 2500)

  const showCommand = step >= 1
  const copied = step >= 2 && step < 4
  const running = step >= 3

  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50 text-xs">
      {/* Terminal header */}
      <div className="bg-zinc-800 px-3 sm:px-4 py-2 sm:py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#ff5f57]" />
            <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#febc2e]" />
            <div className="w-2.5 sm:w-3 h-2.5 sm:h-3 rounded-none bg-[#28c840]" />
          </div>
          <span className="text-zinc-400 ml-2 hidden sm:inline">Terminal</span>
        </div>
        <div className="text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>

      <div className="bg-card p-3 sm:p-4 space-y-4">
        {/* Command Box */}
        <div className="bg-zinc-900 border border-zinc-700 rounded">
          <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-700">
            <span className="text-zinc-400 text-[10px]">Installation Command</span>
            <button className={`flex items-center gap-1 px-2 py-1 rounded text-[10px] transition-colors ${copied ? 'bg-green-500/20 text-green-400' : 'hover:bg-zinc-800 text-zinc-400'}`}>
              {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
          <div className="p-3 font-mono text-[10px] text-zinc-300 overflow-x-auto">
            {showCommand && (
              <div className="animate-in fade-in">
                <span className="text-green-400">$</span> curl -sSL https://install.kloudlite.io | \<br />
                <span className="pl-4">INSTALL_KEY=<span className="text-yellow-400">kl_inst_abc123xyz</span> \</span><br />
                <span className="pl-4">REGION=<span className="text-cyan-400">us-east-1</span> \</span><br />
                <span className="pl-4">bash</span>
              </div>
            )}
          </div>
        </div>

        {/* Running output */}
        {running && (
          <div className="bg-zinc-900 p-3 font-mono text-[10px] text-zinc-300 space-y-1 animate-in fade-in slide-in-from-bottom-2">
            <div className="text-green-400">Initializing Kloudlite installation...</div>
            <div className="text-zinc-400">Checking prerequisites...</div>
            <div className="flex items-center gap-2">
              <Check className="h-3 w-3 text-green-400" />
              <span>AWS credentials verified</span>
            </div>
            <div className="flex items-center gap-2">
              <Check className="h-3 w-3 text-green-400" />
              <span>Terraform v1.5.0 found</span>
            </div>
            <div className="flex items-center gap-2">
              <Check className="h-3 w-3 text-green-400" />
              <span>kubectl v1.28.0 found</span>
            </div>
            <div className="flex items-center gap-2 text-cyan-400">
              <Loader2 className="h-3 w-3 animate-spin" />
              <span>Provisioning infrastructure...</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export function InstallationCompletePreview() {
  const step = useAnimationSteps(4, 2000, 3000)

  const verifying = step === 0 || step === 1
  const complete = step >= 2

  return (
    <PreviewFrame>
      <div className="text-xs">
        {/* Header */}
        <div className="bg-background border-b px-3 sm:px-4 py-2 sm:py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="font-bold">Kloudlite</span>
          </div>
        </div>

        <div className="p-4 sm:p-6 max-w-md mx-auto text-center">
        <div className="hidden sm:block"><InstallationProgress currentStep={complete ? 3 : 2} /></div>

        {verifying ? (
          <div className="animate-in fade-in">
            <div className="w-16 h-16 rounded-none bg-primary/10 flex items-center justify-center mx-auto mb-4">
              <Loader2 className="h-8 w-8 text-primary animate-spin" />
            </div>
            <h2 className="font-semibold text-base mb-2">Verifying Installation</h2>
            <p className="text-muted-foreground text-[11px] mb-4">
              Waiting for your cluster to connect...
            </p>
            <div className="bg-muted/50 p-3 rounded text-left">
              <p className="text-[10px] text-muted-foreground mb-2">Installation URL</p>
              <div className="flex items-center gap-2">
                <code className="text-[10px] bg-background px-2 py-1 rounded flex-1">
                  https://my-prod-cluster.kloudlite.io
                </code>
                <button className="p-1 hover:bg-muted rounded">
                  <Copy className="h-3 w-3 text-muted-foreground" />
                </button>
              </div>
            </div>
            <p className="text-muted-foreground text-[10px] mt-4">
              Check #{step + 1}... This may take a few minutes
            </p>
          </div>
        ) : (
          <div className="animate-in fade-in slide-in-from-bottom-2">
            <div className="w-16 h-16 rounded-none bg-green-500/10 flex items-center justify-center mx-auto mb-4">
              <Check className="h-8 w-8 text-green-500" />
            </div>
            <h2 className="font-semibold text-base mb-2">Installation Complete!</h2>
            <p className="text-muted-foreground text-[11px] mb-6">
              Your Kloudlite installation is ready to use
            </p>

            <button className="w-full bg-primary text-primary-foreground py-2.5 flex items-center justify-center gap-2 mb-4">
              <ExternalLink className="h-3.5 w-3.5" />
              Open Installation Dashboard
            </button>

            <div className="bg-muted/50 p-4 rounded text-left space-y-3">
              <p className="text-[11px] font-medium">What's next?</p>
              <div className="space-y-2 text-[10px]">
                <div className="flex items-start gap-2">
                  <div className="w-4 h-4 rounded-none bg-primary/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-[8px] font-medium text-primary">1</span>
                  </div>
                  <span className="text-muted-foreground">Create users in the admin panel</span>
                </div>
                <div className="flex items-start gap-2">
                  <div className="w-4 h-4 rounded-none bg-primary/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-[8px] font-medium text-primary">2</span>
                  </div>
                  <span className="text-muted-foreground">Set up environments with your services</span>
                </div>
                <div className="flex items-start gap-2">
                  <div className="w-4 h-4 rounded-none bg-primary/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-[8px] font-medium text-primary">3</span>
                  </div>
                  <span className="text-muted-foreground">Create workspaces and start coding</span>
                </div>
              </div>
            </div>
          </div>
        )}
        </div>
      </div>
    </PreviewFrame>
  )
}
