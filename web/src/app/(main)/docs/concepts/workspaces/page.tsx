import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Code2,
  Terminal,
  Package,
  Network,
  Layers,
  HardDrive,
  Lock,
  Zap,
  CheckCircle2,
  Info,
  Settings,
  Globe,
  FolderOpen,
} from 'lucide-react'

const tocItems = [
  { id: 'what-is-workspace', title: 'What is a Workspace?' },
  { id: 'key-features', title: 'Key Features' },
  { id: 'architecture', title: 'Workspace Architecture' },
  { id: 'access-methods', title: 'Access Methods' },
  { id: 'storage', title: 'Storage & Persistence' },
  { id: 'networking', title: 'Networking' },
]

export default function WorkspaceOverviewPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        Workspace Overview
      </h1>

      {/* What is a Workspace? */}
      <section id="what-is-workspace" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Code2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            What is a Workspace?
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          A workspace is your personal Dev Container - an isolated, fully-configured development
          environment that runs in the cloud. Each workspace comes with everything you need to
          code, including a complete Linux environment, your choice of packages and tools, and
          multiple ways to access and work with your code.
        </p>

        <div className="bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-950 dark:to-cyan-950 rounded-lg border-2 border-blue-300 dark:border-blue-700 p-4 sm:p-6 mb-6">
          <h4 className="text-blue-900 dark:text-blue-100 font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5" />
            Dev Container
          </h4>
          <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed mb-3">
            Unlike traditional development environments that run locally on your machine,
            workspaces run as containerized environments on remote infrastructure. This provides:
          </p>
          <ul className="text-blue-800 dark:text-blue-200 text-sm space-y-1 m-0 list-disc list-inside">
            <li>Consistent environment across team members</li>
            <li>Access from any device with a browser</li>
            <li>Isolation from your local machine</li>
            <li>Reproducible setup with package management</li>
          </ul>
        </div>

        <div className="grid gap-4 sm:gap-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Zap className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Instant Development
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Start coding immediately without setting up local development environments,
                  installing dependencies, or configuring tools.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Lock className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Isolated & Safe
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Each workspace runs in isolation with its own network namespace, preventing
                  conflicts and keeping your development secure.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Key Features */}
      <section id="key-features" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <CheckCircle2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Key Features</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Globe className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Multiple Access Methods
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0 mb-3">
                  Access your workspace through various methods depending on your workflow:
                </p>
                <ul className="text-muted-foreground text-sm space-y-1.5">
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                    Web-based IDE in your browser
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                    SSH access for any desktop IDE (VS Code, Cursor, IntelliJ, etc.)
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                    Web terminal for quick access
                  </li>
                  <li className="flex items-center gap-2">
                    <CheckCircle2 className="h-3 w-3 text-green-500 flex-shrink-0" />
                    AI-powered coding assistants
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Package className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Package Management
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Powered by Nix package manager, packages are installed and persisted at the
                  workmachine host level. Each workspace gets access to packages based on its
                  configuration, making binaries instantly available in the workspace PATH without
                  reinstallation.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Network className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Environment Connection
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Connect your workspace to environments to access services like databases, caches,
                  and APIs using simple service names. The workspace&apos;s network namespace
                  switches to enable seamless service access.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <Settings className="text-primary h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Service Interception
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Intercept services running in connected environments to route their traffic to
                  your workspace. Debug and test with real traffic without affecting the actual
                  service.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Architecture */}
      <section id="architecture" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Layers className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Workspace Architecture
          </h2>
        </div>

        {/* Workspace Architecture Diagram */}
        <div className="bg-gradient-to-br from-slate-50 to-emerald-50/30 dark:from-slate-900/50 dark:to-slate-800/50 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 sm:p-12 mb-8">
          <div className="max-w-4xl mx-auto">
            {/* Main Workspace Container */}
            <div className="relative">
              <div className="absolute -inset-1 bg-gradient-to-r from-emerald-600 to-green-600 rounded-2xl blur opacity-25"></div>
              <div className="relative bg-white dark:bg-slate-950 rounded-2xl border-2 border-emerald-500 shadow-xl p-8">
                <div className="absolute -top-4 left-6 bg-emerald-600 text-white rounded-lg px-4 py-1.5 text-xs font-bold uppercase tracking-wide shadow-lg">
                  Workspace Container
                </div>

                <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-2">
                  {/* Access */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <Code2 className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">Access</div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Web IDE, SSH, Terminal
                    </div>
                  </div>

                  {/* Packages */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <Package className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">
                      <a href="/docs/workspace-internals/packages" className="hover:text-purple-600 dark:hover:text-purple-400">
                        Packages
                      </a>
                    </div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Nix packages
                    </div>
                  </div>

                  {/* Code */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <FolderOpen className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">Code</div>
                    <div className="text-xs text-slate-600 dark:text-slate-400 font-mono">
                      ~/workspaces/[name]
                    </div>
                  </div>

                  {/* Home */}
                  <div className="text-center">
                    <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-xl p-3 shadow-lg mb-3 mx-auto w-fit">
                      <HardDrive className="h-8 w-8 text-white" />
                    </div>
                    <div className="font-semibold text-sm mb-1 text-slate-900 dark:text-slate-100">Shared Home</div>
                    <div className="text-xs text-slate-600 dark:text-slate-400">
                      Configs, dotfiles
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Learn More Links */}
            <div className="mt-6">
              <div className="flex flex-wrap gap-6 justify-center items-center text-sm">
                <a href="/docs/workspace-internals/cli" className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 font-medium">
                  <Terminal className="h-4 w-4" />
                  CLI Reference
                </a>
                <a href="/docs/workspace-internals/environment-connection" className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 font-medium">
                  <Network className="h-4 w-4" />
                  Environment Connection
                </a>
                <a href="/docs/workspace-internals/intercepts" className="flex items-center gap-2 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 font-medium">
                  <Settings className="h-4 w-4" />
                  Service Intercepts
                </a>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Workmachine Host
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Workspaces run on workmachines - virtual machines that host multiple workspaces. Each
            workmachine provides:
          </p>
          <ul className="text-muted-foreground text-sm space-y-2">
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Shared Home Directory (~):</strong> All workspaces on the same workmachine
                share the home directory, persisting tool configurations, SSH keys, and git settings
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Package Storage:</strong> Nix packages installed at the host level and shared
                efficiently across workspaces
              </span>
            </li>
          </ul>
        </div>

        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Workspace Container
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Each workspace is a containerized environment with:
          </p>
          <ul className="text-muted-foreground text-sm space-y-2">
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Dedicated Code Directory:</strong> Workspace code stored at{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  ~/workspaces/[workspace-name]
                </code>
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Package Access:</strong> Configured packages available in PATH based on
                workspace spec
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>Resource Limits:</strong> CPU, memory, and storage quotas per workspace
              </span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span>
                <strong>IDE Server:</strong> Running IDE server for browser-based editing and SSH access for desktop IDEs
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Access Methods */}
      <section id="access-methods" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Access Methods</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Code2 className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Web-Based IDE
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Full IDE experience in your browser with extensions, debugging, and terminal
                  access. No installation required.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Terminal className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  SSH Access for Desktop IDEs
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Connect via SSH with your preferred desktop IDE (VS Code, Cursor, IntelliJ, etc.)
                  for native development experience.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Globe className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Web Terminal
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Quick browser-based terminal access for running commands and checking workspace
                  status without full IDE.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Zap className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  AI-Powered Coding Assistants
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Integrate AI assistants that work directly with your workspace for enhanced
                  development productivity.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Storage & Persistence */}
      <section id="storage" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <HardDrive className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Storage & Persistence
          </h2>
        </div>

        <div className="space-y-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <FolderOpen className="text-primary h-5 w-5" />
              Shared Home Directory
            </h4>
            <p className="text-muted-foreground text-sm mb-3 leading-relaxed">
              All workspaces on the same workmachine share the home directory (
              <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">~</code>). This means:
            </p>
            <ul className="text-muted-foreground text-sm space-y-2">
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>
                  Tool configurations (
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">.bashrc</code>,{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">.vimrc</code>,{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">.gitconfig</code>)
                  are shared
                </span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>SSH keys and credentials persist across workspaces</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>Global tool installations remain available</span>
              </li>
            </ul>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <FolderOpen className="text-primary h-5 w-5" />
              Workspace Code Storage
            </h4>
            <p className="text-muted-foreground text-sm mb-3 leading-relaxed">
              Each workspace has its own dedicated directory for code:
            </p>
            <div className="bg-muted rounded p-3 font-mono text-xs mb-3">
              <pre className="m-0 leading-relaxed">~/workspaces/[workspace-name]/</pre>
            </div>
            <p className="text-muted-foreground text-sm m-0 leading-relaxed">
              This keeps each workspace&apos;s code isolated while sharing common configurations
              in the home directory.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Package className="text-primary h-5 w-5" />
              Package Persistence
            </h4>
            <p className="text-muted-foreground text-sm m-0 leading-relaxed">
              Packages installed via Nix are stored at the workmachine host level and persist
              across workspace restarts. When a workspace starts, packages specified in its
              configuration are immediately available in the PATH without reinstallation.
            </p>
          </div>
        </div>

        <div className="mt-6 bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Data Persistence
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                All data in the home directory and workspace directories persists across workspace
                suspensions and restarts. Only when a workspace is deleted is its data removed.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Networking */}
      <section id="networking" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Networking</h2>
        </div>

        <div className="space-y-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
              Network Isolation
            </h4>
            <p className="text-muted-foreground text-sm m-0 leading-relaxed">
              Each workspace has its own network namespace, providing complete network isolation.
              By default, workspaces can access the external internet but cannot directly access
              environment services.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
              Environment Connection
            </h4>
            <p className="text-muted-foreground text-sm mb-3 leading-relaxed">
              When you connect a workspace to an environment using{' '}
              <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl env connect</code>,
              the workspace&apos;s network namespace switches to the environment&apos;s namespace.
              This enables:
            </p>
            <ul className="text-muted-foreground text-sm space-y-2">
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>
                  DNS resolution of service names (e.g.,{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">postgres</code>,{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">redis</code>)
                </span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>Direct access to environment services by name and port</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
                <span>Ability to intercept services for debugging</span>
              </li>
            </ul>
          </div>
        </div>
      </section>

    </DocsContentLayout>
  )
}
