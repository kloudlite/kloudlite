'use client'

import { useEffect, useState } from 'react'
import { Plus, Lock, Globe, Server, Cpu, HardDrive, FileCode, GitBranch, Monitor, ChevronDown, Check, X } from 'lucide-react'

// Preview frame wrapper with browser chrome
function PreviewFrame({ children, url = 'console.kloudlite.io' }: { children: React.ReactNode; url?: string }) {
  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50">
      {/* Browser chrome */}
      <div className="bg-zinc-800 px-4 py-2.5 flex items-center gap-3">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
          <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
          <div className="w-3 h-3 rounded-full bg-[#28c840]" />
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
function useAnimationSteps(totalSteps: number, stepDuration: number = 1000, pauseAtEnd: number = 2000) {
  const [step, setStep] = useState(0)

  useEffect(() => {
    let currentStep = 0

    // Progress through steps
    const stepInterval = setInterval(() => {
      currentStep = (currentStep + 1) % (totalSteps + Math.ceil(pauseAtEnd / stepDuration))
      setStep(currentStep < totalSteps ? currentStep : 0)
    }, stepDuration)

    return () => clearInterval(stepInterval)
  }, [totalSteps, stepDuration, pauseAtEnd])

  return step
}

// Shared navigation header for dashboard previews
function DashboardHeader({ activeItem }: { activeItem?: string }) {
  return (
    <div className="bg-background border-b px-4 h-10 flex items-center justify-between text-xs">
      <div className="flex items-center gap-4">
        <span className="font-bold">Kloudlite</span>
        <div className="flex items-center gap-1">
          {['Environments', 'Workspaces'].map((item) => (
            <span
              key={item}
              className={`px-2 py-1 ${activeItem === item ? 'bg-accent font-medium' : 'text-muted-foreground'}`}
            >
              {item}
            </span>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-6 h-6 rounded-full bg-muted" />
      </div>
    </div>
  )
}

// Admin header for admin panel previews
function AdminHeader() {
  return (
    <div className="bg-zinc-900 text-white px-4 h-10 flex items-center justify-between text-xs">
      <div className="flex items-center gap-4">
        <span className="font-bold">Kloudlite Admin</span>
        <div className="flex items-center gap-1">
          {['Users', 'Machine Configs', 'Settings'].map((item, i) => (
            <span
              key={item}
              className={`px-2 py-1 ${i === 0 ? 'bg-white/10' : 'text-zinc-400'}`}
            >
              {item}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}

export function CreateUserPreview() {
  const step = useAnimationSteps(6, 1200, 1500)

  const emailText = step >= 1 ? 'jane@example.com' : ''
  const usernameText = step >= 2 ? 'janedoe' : ''
  const showCheck = step >= 3
  const roleSelected = step >= 4
  const buttonActive = step >= 5

  return (
    <PreviewFrame url="admin.kloudlite.io">
      <div className="text-xs">
        <AdminHeader />
      <div className="p-4">
        {/* Header with filters and add button */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <span className="bg-muted px-2 py-1">All</span>
            <span className="text-muted-foreground px-2 py-1">Admin</span>
            <span className="text-muted-foreground px-2 py-1">User</span>
          </div>
          <button className="bg-primary text-primary-foreground px-3 py-1.5 flex items-center gap-1 transition-transform active:scale-95">
            <Plus className="h-3 w-3" />
            Add User
          </button>
        </div>

        {/* Users table */}
        <div className="border">
          <div className="bg-muted/50 border-b px-4 py-2 grid grid-cols-5 gap-4 text-muted-foreground">
            <span>User</span>
            <span>Role</span>
            <span>Status</span>
            <span>Last Login</span>
            <span>Actions</span>
          </div>
          <div className="divide-y">
            <div className="px-4 py-3 grid grid-cols-5 gap-4 items-center hover:bg-muted/50">
              <div>
                <div className="font-medium">John Doe</div>
                <div className="text-muted-foreground">john@example.com</div>
              </div>
              <span className="bg-info/10 text-info px-2 py-0.5 w-fit">Admin</span>
              <span className="bg-success/10 text-success px-2 py-0.5 w-fit">Active</span>
              <span className="text-muted-foreground">2 hours ago</span>
              <span className="text-muted-foreground">•••</span>
            </div>
          </div>
        </div>

        {/* Add User Dialog Overlay */}
        <div className="mt-4 border p-4 bg-background shadow-lg max-w-sm animate-in fade-in slide-in-from-bottom-2 duration-300">
          <p className="font-medium mb-4">Add New User</p>
          <div className="space-y-3">
            <div>
              <label className="text-muted-foreground mb-1 block">Email</label>
              <div className={`border h-8 flex items-center px-3 bg-background transition-colors ${step >= 1 && step < 2 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                {emailText}<span className={`${step >= 1 && step < 2 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
              </div>
            </div>
            <div>
              <label className="text-muted-foreground mb-1 block">Username</label>
              <div className={`border h-8 flex items-center px-3 bg-background justify-between transition-colors ${step >= 2 && step < 3 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                <span>{usernameText}<span className={`${step >= 2 && step < 3 ? 'animate-pulse' : 'opacity-0'}`}>|</span></span>
                <Check className={`h-3 w-3 text-success transition-all duration-300 ${showCheck ? 'opacity-100 scale-100' : 'opacity-0 scale-50'}`} />
              </div>
              <span className={`text-muted-foreground text-[10px] transition-opacity duration-300 ${showCheck ? 'opacity-100' : 'opacity-0'}`}>Username is available</span>
            </div>
            <div>
              <label className="text-muted-foreground mb-1 block">Roles</label>
              <div className="flex gap-1">
                <span className={`px-2 py-1 transition-all duration-300 ${roleSelected ? 'bg-primary text-primary-foreground' : 'border text-muted-foreground'}`}>User</span>
                <span className="border px-2 py-1 text-muted-foreground">Admin</span>
              </div>
            </div>
            <button className={`bg-primary text-primary-foreground w-full py-2 mt-2 transition-all duration-200 ${buttonActive ? 'scale-95 brightness-90' : ''}`}>
              Create User
            </button>
          </div>
        </div>
      </div>
      </div>
    </PreviewFrame>
  )
}

export function LoginPreview() {
  const step = useAnimationSteps(4, 1200, 1500)

  const emailText = step >= 1 ? 'jane@example.com' : ''
  const passwordDots = step >= 2 ? '••••••••' : ''
  const buttonActive = step >= 3

  return (
    <PreviewFrame url="auth.kloudlite.io">
      <div className="text-xs flex min-h-[280px]">
        {/* Left branding section */}
        <div className="w-1/3 bg-zinc-900 text-white p-6 flex flex-col justify-center">
          <div className="font-bold text-lg mb-2">Kloudlite</div>
          <p className="text-zinc-400 text-[10px]">Cloud Development Environment</p>
        </div>

        {/* Right form section */}
        <div className="flex-1 p-6 flex items-center justify-center">
          <div className="w-full max-w-xs space-y-4">
            <div className="text-center mb-6">
              <p className="font-medium">Sign in to your account</p>
            </div>

            <div className="space-y-3">
              <div>
                <label className="text-muted-foreground mb-1 block">Email</label>
                <div className={`border h-8 flex items-center px-3 transition-colors ${step === 1 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                  {emailText}<span className={`${step === 1 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
                </div>
              </div>
              <div>
                <label className="text-muted-foreground mb-1 block">Password</label>
                <div className={`border h-8 flex items-center px-3 text-muted-foreground transition-colors ${step === 2 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                  {passwordDots}<span className={`${step === 2 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
                </div>
              </div>
            </div>

            <button className={`bg-primary text-primary-foreground w-full py-2 transition-all duration-200 ${buttonActive ? 'scale-95 brightness-90' : ''}`}>
              Sign In
            </button>

            <div className="relative my-4">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t" />
              </div>
              <div className="relative flex justify-center">
                <span className="bg-card px-2 text-muted-foreground text-[10px]">Or continue with</span>
              </div>
            </div>

            <div className="grid grid-cols-3 gap-2">
              <button className="border py-2 text-center hover:bg-muted transition-colors">Google</button>
              <button className="border py-2 text-center hover:bg-muted transition-colors">GitHub</button>
              <button className="border py-2 text-center hover:bg-muted transition-colors">Microsoft</button>
            </div>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

export function WorkmachinePreview() {
  const step = useAnimationSteps(5, 1200, 1500)

  const dropdownOpen = step === 1
  const machineSelected = step >= 2
  const detailsShown = step >= 3
  const buttonActive = step >= 4

  return (
    <PreviewFrame>
      <div className="text-xs">
        <DashboardHeader />
      <div className="p-6 flex items-center justify-center min-h-[300px]">
        <div className="text-center max-w-md">
          {/* Welcome icon */}
          <div className="bg-primary/10 w-12 h-12 rounded-full flex items-center justify-center mx-auto mb-4">
            <Server className="h-6 w-6 text-primary" />
          </div>

          <h3 className="font-medium text-base mb-2">Welcome to Kloudlite!</h3>
          <p className="text-muted-foreground mb-6">Let's set up your development environment</p>

          {/* Machine type selector */}
          <div className="border p-4 text-left space-y-4">
            <div className="relative">
              <label className="text-muted-foreground mb-2 block">Select Machine Type</label>
              <div className={`border h-10 flex items-center justify-between px-3 transition-colors ${dropdownOpen ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                <span className={machineSelected ? '' : 'text-muted-foreground'}>{machineSelected ? 'Medium (4 CPU, 8GB RAM)' : 'Select a machine type...'}</span>
                <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform ${dropdownOpen ? 'rotate-180' : ''}`} />
              </div>

              {/* Dropdown menu */}
              {dropdownOpen && (
                <div className="absolute top-full left-0 right-0 mt-1 border bg-background shadow-lg z-10 animate-in fade-in slide-in-from-top-1 duration-150">
                  <div className="p-1 border-b text-muted-foreground hover:bg-muted">Small (2 CPU, 4GB RAM)</div>
                  <div className="p-1 border-b bg-primary/10 text-primary font-medium">Medium (4 CPU, 8GB RAM)</div>
                  <div className="p-1 text-muted-foreground hover:bg-muted">Large (8 CPU, 16GB RAM)</div>
                </div>
              )}
            </div>

            {/* Selected machine details */}
            <div className={`bg-muted/50 p-3 space-y-2 transition-all duration-300 ${detailsShown ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-2'}`}>
              <div className="font-medium">Medium Instance</div>
              <div className="text-muted-foreground text-[10px]">Balanced performance for most workloads</div>
              <div className="grid grid-cols-2 gap-2 mt-2">
                <div className="flex items-center gap-2">
                  <div className="bg-primary/10 p-1">
                    <Cpu className="h-3 w-3 text-primary" />
                  </div>
                  <span>4 CPU Cores</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="bg-primary/10 p-1">
                    <HardDrive className="h-3 w-3 text-primary" />
                  </div>
                  <span>8GB Memory</span>
                </div>
              </div>
              <div className="mt-2">
                <span className="bg-info/10 text-info border border-info/20 px-2 py-0.5 text-[10px]">
                  General Purpose
                </span>
              </div>
            </div>

            <button className={`bg-primary text-primary-foreground w-full py-2 transition-all duration-200 ${buttonActive ? 'scale-95 brightness-90' : ''}`}>
              Create Workmachine
            </button>
            <p className={`text-muted-foreground text-[10px] text-center transition-opacity duration-300 ${buttonActive ? 'opacity-100' : 'opacity-0'}`}>
              This will take a few moments to provision...
            </p>
          </div>
        </div>
      </div>
      </div>
    </PreviewFrame>
  )
}

export function EnvironmentPreview() {
  const step = useAnimationSteps(7, 1000, 1500)

  const nameText = step >= 1 ? 'staging' : ''
  const showCompose = step >= 2
  const showLine1 = step >= 3
  const showLine2 = step >= 4
  const privateSelected = step >= 5
  const buttonActive = step >= 6

  return (
    <PreviewFrame>
      <div className="text-xs">
        <DashboardHeader activeItem="Environments" />
      <div className="p-4">
        {/* Header with filters */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <span className="bg-muted px-2 py-1">All</span>
            <span className="text-muted-foreground px-2 py-1">Mine</span>
            <span className="text-muted-foreground ml-4">2 environments</span>
          </div>
          <button className="bg-primary text-primary-foreground px-3 py-1.5 flex items-center gap-1 transition-transform">
            <Plus className="h-3 w-3" />
            Create Environment
          </button>
        </div>

        {/* Create Dialog with Docker Compose */}
        <div className="border p-4 bg-background shadow-lg animate-in fade-in slide-in-from-bottom-2 duration-300">
          <p className="font-medium mb-4">Create Environment</p>
          <div className="grid grid-cols-2 gap-4">
            {/* Left column - Form */}
            <div className="space-y-3">
              <div>
                <label className="text-muted-foreground mb-1 block">Name</label>
                <div className={`border h-8 flex items-center px-3 transition-colors ${step === 1 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                  {nameText}<span className={`${step === 1 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
                </div>
              </div>
              <div>
                <label className="text-muted-foreground mb-1 block">Visibility</label>
                <div className="space-y-1.5">
                  <div className={`border p-2 flex items-center gap-2 transition-all duration-300 ${privateSelected ? 'bg-primary/5 border-primary' : ''}`}>
                    <Lock className={`h-3 w-3 transition-colors ${privateSelected ? 'text-foreground' : 'text-muted-foreground'}`} />
                    <div className={privateSelected ? 'font-medium' : ''}>Private</div>
                  </div>
                  <div className="border p-2 flex items-center gap-2">
                    <Globe className="h-3 w-3 text-muted-foreground" />
                    <div>Open</div>
                  </div>
                </div>
              </div>
              <button className={`bg-primary text-primary-foreground w-full py-2 transition-all duration-200 ${buttonActive ? 'scale-95 brightness-90' : ''}`}>Create</button>
            </div>

            {/* Right column - Docker Compose Editor */}
            <div>
              <label className="text-muted-foreground mb-1 block flex items-center gap-1">
                <FileCode className="h-3 w-3" />
                Docker Compose
              </label>
              <div className={`border bg-[#1e1e1e] p-3 font-mono text-[10px] leading-relaxed h-[180px] overflow-hidden transition-colors ${step >= 2 && step <= 4 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                <div className="text-[#9cdcfe]">services:</div>
                {showCompose && (
                  <div className="animate-in fade-in duration-200">
                    <div className="pl-2"><span className="text-[#9cdcfe]">postgres</span><span className="text-[#d4d4d4]">:</span></div>
                    <div className="pl-4"><span className="text-[#9cdcfe]">image</span><span className="text-[#d4d4d4]">: </span><span className="text-[#ce9178]">postgres:15</span></div>
                    <div className="pl-4"><span className="text-[#9cdcfe]">ports</span><span className="text-[#d4d4d4]">:</span></div>
                    <div className="pl-6 text-[#ce9178]">- "5432:5432"</div>
                  </div>
                )}
                {showLine1 && (
                  <div className="animate-in fade-in duration-200">
                    <div className="pl-2"><span className="text-[#9cdcfe]">redis</span><span className="text-[#d4d4d4]">:</span></div>
                    <div className="pl-4"><span className="text-[#9cdcfe]">image</span><span className="text-[#d4d4d4]">: </span><span className="text-[#ce9178]">redis:7</span></div>
                    <div className="pl-4"><span className="text-[#9cdcfe]">ports</span><span className="text-[#d4d4d4]">:</span></div>
                    <div className="pl-6 text-[#ce9178]">- "6379:6379"</div>
                  </div>
                )}
                {showLine2 && (
                  <div className="animate-in fade-in duration-200">
                    <div className="pl-2"><span className="text-[#9cdcfe]">api</span><span className="text-[#d4d4d4]">:</span></div>
                    <div className="pl-4"><span className="text-[#9cdcfe]">image</span><span className="text-[#d4d4d4]">: </span><span className="text-[#ce9178]">myapp/api:latest</span></div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
      </div>
    </PreviewFrame>
  )
}

export function WorkspacePreview() {
  const step = useAnimationSteps(6, 1200, 1500)

  const nameText = step >= 1 ? 'backend-api' : ''
  const visibilityShown = step >= 2
  const repoShown = step >= 3
  const repoText = step >= 4 ? 'https://github.com/user/repo.git' : ''
  const buttonActive = step >= 5

  return (
    <PreviewFrame>
      <div className="text-xs">
        <DashboardHeader activeItem="Workspaces" />
      <div className="p-4">
        {/* Header with filters */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <span className="bg-muted px-2 py-1">All</span>
            <span className="text-muted-foreground px-2 py-1">Mine</span>
            <span className="text-muted-foreground ml-2">|</span>
            <span className="text-muted-foreground px-2 py-1">Active</span>
            <span className="text-muted-foreground ml-4">1 workspace</span>
          </div>
          <button className="bg-primary text-primary-foreground px-3 py-1.5 flex items-center gap-1 transition-transform">
            <Plus className="h-3 w-3" />
            New Workspace
          </button>
        </div>

        {/* Workspaces table */}
        <div className="border">
          <div className="bg-muted/50 border-b px-4 py-2 grid grid-cols-5 gap-4 text-muted-foreground">
            <span>Name</span>
            <span>Status</span>
            <span>Environment</span>
            <span>Created</span>
            <span>Actions</span>
          </div>
          <div className="px-4 py-3 grid grid-cols-5 gap-4 items-center hover:bg-muted/50">
            <div className="flex items-center gap-2">
              <Monitor className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium hover:text-primary">my-workspace</span>
            </div>
            <span className="bg-success/10 text-success px-2 py-0.5 w-fit">Running</span>
            <span className="text-muted-foreground">development</span>
            <span className="text-muted-foreground">2 hours ago</span>
            <span className="text-muted-foreground">•••</span>
          </div>
        </div>

        {/* Create Sheet (right side panel simulation) */}
        <div className="mt-4 border-l-4 border-l-primary p-4 bg-background shadow-lg animate-in fade-in slide-in-from-right-2 duration-300">
          <p className="font-medium mb-4">Create Workspace</p>
          <div className="space-y-4">
            <div>
              <label className="text-muted-foreground mb-1 block">Workspace Name</label>
              <div className={`border h-8 flex items-center px-3 transition-colors ${step === 1 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                {nameText}<span className={`${step === 1 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
              </div>
              <span className="text-muted-foreground text-[10px]">Lowercase letters, numbers, and hyphens only</span>
            </div>

            <div className={`transition-all duration-300 ${visibilityShown ? 'opacity-100' : 'opacity-50'}`}>
              <label className="text-muted-foreground mb-1 block">Visibility</label>
              <div className={`border h-8 flex items-center justify-between px-3 transition-colors ${step === 2 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                <div className="flex items-center gap-2">
                  <Lock className="h-3 w-3" />
                  <span>Private</span>
                </div>
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </div>
            </div>

            <div className={`transition-all duration-300 ${repoShown ? 'opacity-100' : 'opacity-50'}`}>
              <label className="text-muted-foreground mb-1 block flex items-center gap-2">
                <GitBranch className="h-3 w-3" />
                Git Repository (Optional)
              </label>
              <div className={`border h-8 flex items-center px-3 font-mono text-[10px] transition-colors ${step === 4 ? 'border-primary ring-1 ring-primary/20' : ''}`}>
                {repoText}<span className={`${step === 4 ? 'animate-pulse' : 'opacity-0'}`}>|</span>
              </div>
            </div>

            <div className="flex gap-2 pt-2">
              <button className="border px-3 py-2 flex-1 transition-colors">Cancel</button>
              <button className={`bg-primary text-primary-foreground px-3 py-2 flex-1 transition-all duration-200 ${buttonActive ? 'scale-95 brightness-90' : ''}`}>Create Workspace</button>
            </div>
          </div>
        </div>
      </div>
      </div>
    </PreviewFrame>
  )
}

export function ConnectEnvironmentPreview() {
  const step = useAnimationSteps(7, 1000, 1500)

  const cmd1Text = step >= 1 ? 'kl env list' : ''
  const showList = step >= 2
  const cmd2Text = step >= 3 ? 'kl env connect development' : ''
  const showConnected = step >= 4
  const showServices = step >= 5
  const showFinalPrompt = step >= 6

  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50 text-xs">
      {/* Terminal header */}
      <div className="bg-zinc-800 px-4 py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
            <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
            <div className="w-3 h-3 rounded-full bg-[#28c840]" />
          </div>
          <span className="text-zinc-400 ml-2">Terminal — my-workspace</span>
        </div>
        <div className="text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>

      {/* Terminal content */}
      <div className="bg-zinc-900 text-zinc-300 p-4 font-mono text-xs leading-relaxed min-h-[200px]">
        <div className="flex items-center gap-2">
          <span className="text-green-400">kl@my-workspace</span>
          <span className="text-zinc-500">~</span>
          <span className="text-zinc-400">$</span>
          <span>{cmd1Text}<span className={`${step === 1 ? 'animate-pulse' : 'opacity-0'}`}>_</span></span>
        </div>

        {showList && (
          <div className="animate-in fade-in duration-200">
            <div className="mt-2 text-zinc-400">
              NAME          STATUS    SERVICES
            </div>
            <div className="text-zinc-300">
              development   active    3
            </div>
            <div className="text-zinc-300">
              staging       active    3
            </div>
          </div>
        )}

        {step >= 3 && (
          <div className="mt-4 flex items-center gap-2 animate-in fade-in duration-200">
            <span className="text-green-400">kl@my-workspace</span>
            <span className="text-zinc-500">~</span>
            <span className="text-zinc-400">$</span>
            <span>{cmd2Text}<span className={`${step === 3 ? 'animate-pulse' : 'opacity-0'}`}>_</span></span>
          </div>
        )}

        {showConnected && (
          <div className="mt-2 animate-in fade-in duration-200">
            <span className="text-green-400">✓</span> Connected to <span className="text-cyan-400">development</span>
          </div>
        )}

        {showServices && (
          <div className="animate-in fade-in slide-in-from-bottom-1 duration-300">
            <div className="text-zinc-400 mt-1">Available services:</div>
            <div className="text-zinc-300 ml-2">• postgres:5432</div>
            <div className="text-zinc-300 ml-2">• redis:6379</div>
            <div className="text-zinc-300 ml-2">• api:8080</div>
          </div>
        )}

        {showFinalPrompt && (
          <div className="mt-4 flex items-center gap-2 animate-in fade-in duration-200">
            <span className="text-green-400">kl@my-workspace</span>
            <span className="text-cyan-400">[development]</span>
            <span className="text-zinc-500">~</span>
            <span className="text-zinc-400">$</span>
            <span className="animate-pulse">_</span>
          </div>
        )}
      </div>
    </div>
  )
}

export function StartCodingPreview() {
  const step = useAnimationSteps(5, 1200, 1500)
  const [cursorVisible, setCursorVisible] = useState(true)

  // Cursor blink effect
  useEffect(() => {
    const interval = setInterval(() => {
      setCursorVisible(v => !v)
    }, 530)
    return () => clearInterval(interval)
  }, [])

  const showLine1 = step >= 1
  const showLine2 = step >= 2
  const showLine3 = step >= 3
  const showComment = step >= 4

  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50 text-xs">
      {/* VS Code title bar */}
      <div className="bg-[#323233] text-zinc-300 px-3 py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
            <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
            <div className="w-3 h-3 rounded-full bg-[#28c840]" />
          </div>
          <span className="text-zinc-400 ml-2">my-workspace — VS Code</span>
        </div>
        <div className="text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>

      {/* VS Code content */}
      <div className="flex h-[220px]">
        {/* Activity bar */}
        <div className="w-10 bg-[#333333] flex flex-col items-center py-2 gap-3">
          <div className="w-6 h-6 flex items-center justify-center text-white bg-white/10">
            <svg viewBox="0 0 24 24" className="w-4 h-4 fill-current">
              <path d="M17.5 0h-9L7 1.5V6H2.5L1 7.5v15.07L2.5 24h12.07L16 22.57V18h4.7l1.3-1.43V4.5L17.5 0zm0 2.12l2.38 2.38H17.5V2.12zm-3 20.38H3v-15h4.5v8.07L9 17h5.5v5.5zm1-6.5h-5.5v-9H16v9z"/>
            </svg>
          </div>
          <div className="w-6 h-6 flex items-center justify-center text-zinc-500">
            <svg viewBox="0 0 24 24" className="w-4 h-4 fill-current">
              <path d="M15.25 0a8.25 8.25 0 0 0-6.18 13.72L1 22.88l1.12 1.12 8.05-8.05A8.25 8.25 0 1 0 15.25 0zm0 15a6.75 6.75 0 1 1 0-13.5 6.75 6.75 0 0 1 0 13.5z"/>
            </svg>
          </div>
        </div>

        {/* Sidebar */}
        <div className="w-44 bg-[#252526] border-r border-[#3c3c3c] overflow-hidden">
          <div className="text-[10px] text-zinc-400 px-3 py-2 uppercase tracking-wide">Explorer</div>
          <div className="px-2 text-zinc-300">
            <div className="flex items-center gap-1 py-0.5 bg-[#37373d] px-1">
              <span className="text-yellow-400">▼</span>
              <span>backend-api</span>
            </div>
            <div className="pl-4 py-0.5 flex items-center gap-1">
              <span className="text-yellow-400">▼</span>
              <span>src</span>
            </div>
            <div className="pl-6 py-0.5 text-zinc-400 flex items-center gap-1">
              <span className="text-blue-400">◆</span>
              index.ts
            </div>
            <div className={`pl-6 py-0.5 flex items-center gap-1 ${step >= 1 ? 'bg-[#37373d] text-zinc-200' : 'text-zinc-400'}`}>
              <span className="text-blue-400">◆</span>
              db.ts
            </div>
            <div className="pl-4 py-0.5 flex items-center gap-1 text-zinc-400">
              <span className="text-green-400">◆</span>
              package.json
            </div>
            <div className="pl-4 py-0.5 flex items-center gap-1 text-zinc-400">
              <span className="text-zinc-500">◆</span>
              .env
            </div>
          </div>
        </div>

        {/* Editor */}
        <div className="flex-1 bg-[#1e1e1e] flex flex-col">
          {/* Tabs */}
          <div className="bg-[#252526] flex text-zinc-400 border-b border-[#3c3c3c]">
            <div className="px-3 py-1 bg-[#1e1e1e] text-zinc-200 border-t-2 border-t-blue-500 flex items-center gap-2">
              <span className="text-blue-400">◆</span>
              db.ts
              <X className="h-3 w-3 hover:bg-white/10" />
            </div>
          </div>

          {/* Editor content - VS Code Dark+ theme colors */}
          <div className="flex-1 p-3 font-mono text-[11px] leading-relaxed overflow-hidden text-[#d4d4d4]">
            <div className="flex">
              <span className="text-[#858585] w-6 text-right mr-4">1</span>
              <span><span className="text-[#c586c0]">import</span> {'{ '}<span className="text-[#9cdcfe]">Client</span>{' }'} <span className="text-[#c586c0]">from</span> <span className="text-[#ce9178]">'pg'</span></span>
            </div>
            <div className="flex">
              <span className="text-[#858585] w-6 text-right mr-4">2</span>
              <span></span>
            </div>
            <div className="flex">
              <span className="text-[#858585] w-6 text-right mr-4">3</span>
              <span><span className="text-[#569cd6]">const</span> <span className="text-[#9cdcfe]">client</span> = <span className="text-[#569cd6]">new</span> <span className="text-[#4ec9b0]">Client</span>{'({'}</span>
            </div>
            <div className={`flex transition-colors duration-200 ${step === 1 ? 'bg-[#2a2d2e]' : ''}`}>
              <span className="text-[#858585] w-6 text-right mr-4">4</span>
              <span className="pl-4">
                {showLine1 ? (
                  <span className="animate-in fade-in duration-200"><span className="text-[#9cdcfe]">host</span>: <span className="text-[#ce9178]">'postgres'</span>,</span>
                ) : (
                  <span className={cursorVisible ? 'border-l border-[#aeafad]' : ''}>&nbsp;</span>
                )}
              </span>
            </div>
            <div className={`flex transition-colors duration-200 ${step === 2 ? 'bg-[#2a2d2e]' : ''}`}>
              <span className="text-[#858585] w-6 text-right mr-4">5</span>
              <span className="pl-4">
                {showLine2 ? (
                  <span className="animate-in fade-in duration-200"><span className="text-[#9cdcfe]">port</span>: <span className="text-[#b5cea8]">5432</span>,</span>
                ) : (
                  <span>&nbsp;</span>
                )}
              </span>
            </div>
            <div className={`flex transition-colors duration-200 ${step === 3 ? 'bg-[#2a2d2e]' : ''}`}>
              <span className="text-[#858585] w-6 text-right mr-4">6</span>
              <span className="pl-4">
                {showLine3 ? (
                  <span className="animate-in fade-in duration-200"><span className="text-[#9cdcfe]">database</span>: <span className="text-[#ce9178]">'myapp'</span></span>
                ) : (
                  <span>&nbsp;</span>
                )}
              </span>
            </div>
            <div className="flex">
              <span className="text-[#858585] w-6 text-right mr-4">7</span>
              <span>{'}'});</span>
            </div>
            <div className="flex">
              <span className="text-[#858585] w-6 text-right mr-4">8</span>
              <span></span>
            </div>
            <div className={`flex transition-colors duration-200 ${step === 4 ? 'bg-[#2a2d2e]' : ''}`}>
              <span className="text-[#858585] w-6 text-right mr-4">9</span>
              {showComment ? (
                <span className="text-[#6a9955] animate-in fade-in duration-200">// Services available via kl env connect<span className={cursorVisible ? 'border-l border-[#aeafad] ml-0.5' : 'ml-0.5'}>&nbsp;</span></span>
              ) : (
                <span></span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
