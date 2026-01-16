'use client'

import { useEffect, useState } from 'react'
import { Shield, ShieldCheck, Copy, Check, Plus, Trash2, ChevronDown, Terminal, Key } from 'lucide-react'

// Preview frame wrapper with browser chrome
function PreviewFrame({ children, url = 'console.kloudlite.io' }: { children: React.ReactNode; url?: string }) {
  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50">
      {/* Browser chrome */}
      <div className="bg-zinc-800 px-4 py-2.5 flex items-center gap-3">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-none bg-[#ff5f57]" />
          <div className="w-3 h-3 rounded-none bg-[#febc2e]" />
          <div className="w-3 h-3 rounded-none bg-[#28c840]" />
        </div>
        <div className="flex-1 flex justify-center">
          <div className="bg-zinc-700/50 rounded px-3 py-1 text-zinc-400 text-[10px] flex items-center gap-2">
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            {url}
          </div>
        </div>
        <div className="text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>
      {/* Content */}
      <div className="bg-card">
        {children}
      </div>
    </div>
  )
}

// Hook for step-based animation with looping
function useAnimationSteps(totalSteps: number, stepDuration: number = 1200, pauseAtEnd: number = 2000) {
  const [step, setStep] = useState(0)

  useEffect(() => {
    let currentStep = 0

    const stepInterval = setInterval(() => {
      currentStep = (currentStep + 1) % (totalSteps + Math.ceil(pauseAtEnd / stepDuration))
      setStep(currentStep < totalSteps ? currentStep : totalSteps - 1)
    }, stepDuration)

    return () => clearInterval(stepInterval)
  }, [totalSteps, stepDuration, pauseAtEnd])

  return step
}

// Dashboard header component
function DashboardHeader({ vpnStatus = 'disconnected', showDropdown = false }: { vpnStatus?: 'connected' | 'disconnected'; showDropdown?: boolean }) {
  return (
    <div className="bg-background border-b px-4 h-12 flex items-center justify-between text-xs">
      <div className="flex items-center gap-6">
        <span className="font-bold text-sm">Kloudlite</span>
        <div className="flex items-center gap-1 text-muted-foreground">
          <span className="px-2 py-1 hover:text-foreground cursor-pointer">Environments</span>
          <span className="px-2 py-1 bg-accent text-foreground">Workspaces</span>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <div className="relative">
          <button className={`flex items-center gap-2 px-3 py-1.5 border transition-colors ${showDropdown ? 'border-primary bg-primary/5' : ''}`}>
            {vpnStatus === 'connected' ? (
              <ShieldCheck className="h-4 w-4 text-green-500" />
            ) : (
              <Shield className="h-4 w-4 text-muted-foreground" />
            )}
            <span className={vpnStatus === 'connected' ? 'text-green-600' : 'text-muted-foreground'}>
              {vpnStatus === 'connected' ? 'Connected' : 'VPN'}
            </span>
            <ChevronDown className={`h-3 w-3 text-muted-foreground transition-transform ${showDropdown ? 'rotate-180' : ''}`} />
          </button>
        </div>
        <div className="w-7 h-7 rounded-none bg-muted" />
      </div>
    </div>
  )
}

export function VPNConnectionPreview() {
  const step = useAnimationSteps(6, 1400, 2500)

  const showDropdown = step >= 1
  const showTabs = step >= 2
  const showCommand = step >= 3
  const showTerminal = step >= 3
  const showInstalling = step >= 4
  const showConnected = step >= 5

  return (
    <PreviewFrame>
      <div className="text-xs h-[280px]">
        <DashboardHeader vpnStatus={showConnected ? 'connected' : 'disconnected'} showDropdown={showDropdown && !showConnected} />

        <div className="relative h-[232px]">
          {/* Dropdown overlay */}
          <div className={`absolute top-2 right-4 w-72 border bg-background shadow-xl z-10 transition-opacity duration-200 ${showDropdown && !showConnected ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}>
            <div className="p-3 border-b">
              <p className="font-medium text-foreground">VPN Connection</p>
              <p className="text-muted-foreground text-[10px] mt-1">Install kltun to connect</p>
            </div>

            <div className={`p-3 transition-opacity duration-200 ${showTabs ? 'opacity-100' : 'opacity-0'}`}>
              <div className="flex gap-1 mb-3">
                <span className="px-2 py-1 bg-primary text-primary-foreground text-[10px]">macOS/Linux</span>
                <span className="px-2 py-1 text-muted-foreground text-[10px]">Windows</span>
              </div>

              <div className={`transition-opacity duration-200 ${showCommand ? 'opacity-100' : 'opacity-0'}`}>
                <div className="bg-zinc-900 p-2 font-mono text-[9px] text-zinc-300 relative overflow-hidden">
                  <pre className="m-0 whitespace-nowrap">curl -fsSL .../install | bash</pre>
                </div>
              </div>
            </div>
          </div>

          {/* Connected state */}
          <div className={`absolute top-2 right-4 w-64 border bg-background shadow-xl z-10 transition-opacity duration-200 ${showConnected ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}>
            <div className="p-3 border-b flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-green-500" />
              <div>
                <p className="font-medium text-green-600">Connected</p>
                <p className="text-muted-foreground text-[10px]">Uptime: 2h 34m</p>
              </div>
            </div>
            <div className="p-3 space-y-2 text-[10px]">
              <div className="flex justify-between">
                <span className="text-muted-foreground">VPN IP</span>
                <span className="font-mono">10.13.0.5</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Tunnel</span>
                <span className="font-mono text-green-600">Active</span>
              </div>
            </div>
          </div>

          {/* Terminal section */}
          <div className="m-4 rounded overflow-hidden border border-zinc-700">
            <div className="bg-zinc-800 px-3 py-1.5 flex items-center gap-2">
              <div className="flex gap-1">
                <div className="w-2 h-2 rounded-none bg-zinc-600" />
                <div className="w-2 h-2 rounded-none bg-zinc-600" />
                <div className="w-2 h-2 rounded-none bg-zinc-600" />
              </div>
              <span className="text-zinc-500 text-[10px]">Terminal</span>
            </div>

            <div className="bg-zinc-900 text-zinc-300 p-3 font-mono text-[11px] leading-relaxed h-[140px]">
              <div className={`flex items-center gap-2 transition-opacity duration-200 ${showTerminal ? 'opacity-100' : 'opacity-0'}`}>
                <span className="text-green-400">~</span>
                <span className="text-zinc-500">$</span>
                <span>curl -fsSL https://kl.kloudlite.io/install | bash<span className={`${step === 3 ? 'animate-pulse' : 'opacity-0'}`}>_</span></span>
              </div>

              <div className={`mt-2 text-zinc-500 transition-opacity duration-200 ${showInstalling ? 'opacity-100' : 'opacity-0'}`}>
                <div>Installing kltun...</div>
                <div className="text-green-400 mt-1">✓ kltun installed</div>
                <div className="text-zinc-500 mt-1">Connecting to VPN...</div>
              </div>

              <div className={`transition-opacity duration-200 ${showConnected ? 'opacity-100' : 'opacity-0'}`}>
                <div className="text-green-400">✓ VPN connected (10.13.0.5)</div>
                <div className="mt-3 flex items-center gap-2">
                  <span className="text-green-400">~</span>
                  <span className="text-zinc-500">$</span>
                  <span className="animate-pulse">_</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function AuthorizedKeysPreview() {
  const step = useAnimationSteps(6, 1200, 2000)

  const showInput = step >= 1
  const typingKey = step >= 2
  const keyTyped = step >= 3
  const addClicked = step >= 4
  const keyAdded = step >= 5

  const keyText = keyTyped ? 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@laptop' : ''

  return (
    <PreviewFrame>
      <div className="text-xs h-[320px]">
        <DashboardHeader vpnStatus="connected" />

        <div className="p-4 h-[272px]">
          {/* Settings panel simulation */}
          <div className="border-l-4 border-l-primary bg-background p-4 shadow-lg h-full">
            <div className="flex items-center gap-2 mb-4">
              <Key className="h-4 w-4 text-primary" />
              <p className="font-medium">Authorized Keys</p>
            </div>

            {/* Existing keys list - fixed height */}
            <div className="space-y-2 mb-4 h-[72px]">
              <div className="flex items-center justify-between p-2 bg-muted/50 group">
                <div className="flex items-center gap-2">
                  <Key className="h-3 w-3 text-muted-foreground" />
                  <span className="font-mono text-[10px]">ssh-ed25519</span>
                  <span className="text-muted-foreground text-[10px]">john@workstation</span>
                </div>
                <Trash2 className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100" />
              </div>

              <div className={`flex items-center justify-between p-2 bg-green-500/10 border border-green-500/20 transition-opacity duration-200 ${keyAdded ? 'opacity-100' : 'opacity-0'}`}>
                <div className="flex items-center gap-2">
                  <Key className="h-3 w-3 text-green-600" />
                  <span className="font-mono text-[10px]">ssh-ed25519</span>
                  <span className="text-muted-foreground text-[10px]">user@laptop</span>
                </div>
                <Check className="h-3 w-3 text-green-600" />
              </div>
            </div>

            {/* Add new key */}
            <div className="space-y-2">
              <label className="text-muted-foreground text-[10px]">Add SSH Public Key</label>
              <div className="flex gap-2">
                <div className={`flex-1 border h-8 flex items-center px-2 font-mono text-[10px] transition-colors overflow-hidden ${showInput && !keyAdded ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                  <span className={`transition-opacity duration-200 ${typingKey && !keyAdded ? 'opacity-100' : 'opacity-0'}`}>
                    {keyText}
                    <span className={`${step === 2 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
                  </span>
                  <span className={`text-muted-foreground absolute transition-opacity duration-200 ${!typingKey && !keyAdded ? 'opacity-100' : 'opacity-0'}`}>
                    ssh-rsa AAAAB3... user@example.com
                  </span>
                </div>
                <button className={`px-3 h-8 bg-primary text-primary-foreground flex items-center gap-1 transition-all ${addClicked && !keyAdded ? 'scale-95 brightness-90' : ''}`}>
                  <Plus className="h-3 w-3" />
                  Add
                </button>
              </div>
            </div>

            <p className="text-muted-foreground text-[10px] mt-3">
              Add your public key to enable SSH access from your IDE or terminal.
            </p>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function SSHPublicKeyPreview() {
  const step = useAnimationSteps(4, 1500, 2500)

  const showKey = step >= 1
  const hoverCopy = step >= 2
  const copied = step >= 3

  const sshKey = 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGx7...kF2m workmachine-abc123'

  return (
    <PreviewFrame>
      <div className="text-xs h-[340px]">
        <DashboardHeader vpnStatus="connected" />

        <div className="p-4 h-[292px]">
          {/* Settings panel simulation */}
          <div className="border-l-4 border-l-primary bg-background p-4 shadow-lg h-full">
            <div className="flex items-center gap-2 mb-4">
              <Terminal className="h-4 w-4 text-primary" />
              <p className="font-medium">SSH Public Key</p>
            </div>

            <p className="text-muted-foreground text-[10px] mb-3">
              Add this key to GitHub, GitLab, or other services to enable SSH access from your workspaces.
            </p>

            {/* SSH key display - fixed height */}
            <div className={`relative bg-muted p-3 font-mono text-[10px] h-[44px] transition-all ${hoverCopy && !copied ? 'ring-1 ring-primary/30' : ''}`}>
              <pre className={`m-0 overflow-hidden text-foreground transition-opacity duration-300 ${showKey ? 'opacity-100' : 'opacity-0'}`}>
                {sshKey}
              </pre>

              <button className={`absolute top-2 right-2 p-1.5 transition-all ${hoverCopy ? 'bg-primary text-primary-foreground' : 'bg-background border'} ${copied ? 'bg-green-600' : ''}`}>
                {copied ? (
                  <Check className="h-3 w-3" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </button>
            </div>

            {/* Fixed height for copied message */}
            <div className="h-5 mt-2">
              <p className={`text-green-600 text-[10px] transition-opacity duration-200 ${copied ? 'opacity-100' : 'opacity-0'}`}>
                Copied to clipboard!
              </p>
            </div>

            <div className="mt-2 p-3 bg-muted/50 border">
              <p className="text-foreground text-[10px] font-medium mb-1">How to use</p>
              <ol className="text-muted-foreground text-[10px] space-y-1 list-decimal list-inside m-0">
                <li>Copy the public key above</li>
                <li>Go to GitHub → Settings → SSH Keys</li>
                <li>Click "New SSH Key" and paste</li>
              </ol>
            </div>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}
