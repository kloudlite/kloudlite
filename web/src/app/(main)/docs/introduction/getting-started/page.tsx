import { Button } from '@/components/ui/button'
import Link from 'next/link'
import {
  Cpu,
  Database,
  FolderTree,
  Code2,
  Terminal,
  Laptop,
  Globe,
  Sparkles,
  CheckCircle2,
  Info,
  Zap,
  Lock,
  Network,
  Settings
} from 'lucide-react'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'

const tocItems = [
  { id: 'choose-machine', title: 'Choose Your Work Machine' },
  { id: 'create-environment', title: 'Create an Environment' },
  { id: 'create-workspace', title: 'Create Your Workspace' },
  { id: 'access-workspace', title: 'Access Your Workspace' },
  { id: 'next-steps', title: 'Next Steps' },
  { id: 'tips', title: 'Tips & Best Practices' },
]

export default function GettingStartedPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      {/* Header */}
      <div className="mb-12 sm:mb-16">
        <h1 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl break-words leading-tight sm:leading-tight">
          Getting Started
        </h1>
        <p className="text-muted-foreground mt-4 text-base sm:text-lg lg:text-xl leading-relaxed">
          Get up and running with Kloudlite in four simple steps
        </p>
      </div>

      {/* Progress Overview */}
      <section id="next-steps" className="mb-12 sm:mb-16">
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h2 className="text-card-foreground mb-6 text-xl font-semibold">Your Journey</h2>
          <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4">
            <div className="flex flex-col items-center text-center">
              <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-full">
                <span className="text-primary-foreground font-bold">1</span>
              </div>
              <p className="text-sm font-medium">Choose Machine</p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-full">
                <span className="text-primary-foreground font-bold">2</span>
              </div>
              <p className="text-sm font-medium">Create Environment</p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-full">
                <span className="text-primary-foreground font-bold">3</span>
              </div>
              <p className="text-sm font-medium">Create Workspace</p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="bg-primary mb-3 flex h-12 w-12 items-center justify-center rounded-full">
                <span className="text-primary-foreground font-bold">4</span>
              </div>
              <p className="text-sm font-medium">Start Coding</p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 1: Choose Work Machine */}
      <section id="choose-machine" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <span className="text-primary-foreground text-lg font-bold">1</span>
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Choose Your Work Machine</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          A work machine is your dedicated development environment that runs your workspaces. Select a machine type from the available options in your installation based on your resource requirements.
        </p>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">Machine Types</p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                Your installation administrator configures the available machine types. Choose one that matches your development workload - you can always create additional work machines with different specifications later.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 2: Create Environment */}
      <section id="create-environment" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <span className="text-primary-foreground text-lg font-bold">2</span>
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Create an Environment</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Environments are isolated spaces where your applications and services run. Think of them as different stages like development, staging, or production - each with its own services and configurations.
        </p>

        <h3 className="text-foreground mb-4 text-xl font-semibold">Creating Your First Environment</h3>
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <ol className="space-y-4 m-0">
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                1
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Navigate to Environments</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Click on &quot;Environments&quot; in your dashboard
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                2
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Click &quot;Create Environment&quot;</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  This opens a dialog to create your new environment
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                3
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Enter Environment Name</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Choose a name like &quot;development&quot; or &quot;staging&quot; (lowercase letters, numbers, and hyphens only, max 63 characters)
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                4
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Create Environment</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Click the create button and your environment will be ready in seconds
                </p>
              </div>
            </li>
          </ol>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">What Happens Next?</p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                After creation, you can add services to your environment using docker-compose files or manually. Environments are initially inactive to save resources - activate them when you&apos;re ready to use them.
              </p>
            </div>
          </div>
        </div>

        <h3 className="text-foreground mt-8 mb-4 text-xl font-semibold">What&apos;s Inside an Environment?</h3>
        <p className="text-muted-foreground mb-4 leading-relaxed">
          Once created, your environment has several sections to manage different aspects:
        </p>

        <div className="grid gap-4 mb-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Database className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0">Services</h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Add and manage your application services using docker-compose. View service details like DNS names, IPs, and ports. Set up service intercepts to route traffic to your workspace for debugging.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Settings className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0">Configs &amp; Secrets</h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  <strong>Environment Variables:</strong> Store key-value pairs for configuration (API keys, database URLs, etc.).<br />
                  <strong>Config Files:</strong> Upload configuration files that services can mount and use.
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
                <h4 className="text-card-foreground font-semibold mb-1 m-0">Settings</h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Configure environment properties, access control, network policies, and security settings. Manage who can access the environment and how services communicate.
                </p>
              </div>
            </div>
          </div>
        </div>

        <h3 className="text-foreground mt-8 mb-4 text-xl font-semibold">Adding Services with Docker Compose</h3>
        <p className="text-muted-foreground mb-4 leading-relaxed">
          The easiest way to add services to your environment is using docker-compose. Click on your environment, go to the Services tab, and click the edit button to open the composition editor.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <p className="text-card-foreground text-sm mb-3 m-0 font-medium">Example: Adding a PostgreSQL database and Redis cache</p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">
{`services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: devpassword
      POSTGRES_DB: myapp
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:`}
            </pre>
          </div>
          <p className="text-muted-foreground text-xs mt-3 m-0">
            Save the composition and Kloudlite will create the services in your environment. They&apos;ll be accessible via their service names (e.g., <code className="bg-muted rounded px-1 py-0.5 font-mono">postgres</code>, <code className="bg-muted rounded px-1 py-0.5 font-mono">redis</code>). Data will persist across Environment restarts.
          </p>
        </div>
      </section>

      {/* Step 3: Create Workspace */}
      <section id="create-workspace" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <span className="text-primary-foreground text-lg font-bold">3</span>
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Create Your Workspace</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          A workspace is your Dev Container with different ways to access (VS Code Web, desktop IDEs via SSH, web terminal, AI assistants like Claude Code) and package management powered by Nix. All your code, tools, and configurations live here.
        </p>

        <h3 className="text-foreground mb-4 text-xl font-semibold">Setting Up Your Workspace</h3>
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <ol className="space-y-4 m-0">
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                1
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Navigate to Workspaces</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Access the &quot;Workspaces&quot; section from your dashboard
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                2
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Click &quot;Create Workspace&quot;</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Start the workspace creation process
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                3
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Enter Basic Information</p>
                <ul className="text-muted-foreground text-sm mt-2 space-y-1 pl-4">
                  <li><strong>Display Name:</strong> User-friendly name (e.g., &quot;My Dev Workspace&quot;)</li>
                  <li><strong>Description:</strong> Optional description of the workspace purpose</li>
                </ul>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                4
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Select Development Packages</p>
                <p className="text-muted-foreground text-sm mt-1 mb-2 m-0">
                  Choose the tools and runtimes you need. Examples:
                </p>
                <div className="bg-muted rounded p-3">
                  <div className="grid gap-2 text-xs font-mono">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-3 w-3 text-green-500" />
                      <span>nodejs (Node.js runtime)</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-3 w-3 text-green-500" />
                      <span>python3 (Python interpreter)</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-3 w-3 text-green-500" />
                      <span>go (Go compiler)</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-3 w-3 text-green-500" />
                      <span>git (Version control)</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-3 w-3 text-green-500" />
                      <span>vim (Text editor)</span>
                    </div>
                  </div>
                </div>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                5
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Choose Machine Type</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Select the machine type you created earlier based on your resource needs
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                6
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Connect to Environment (Optional)</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Link your workspace to an environment for service access and integrations
                </p>
              </div>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                7
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium m-0 leading-snug">Create Workspace</p>
                <p className="text-muted-foreground text-sm mt-1 m-0">
                  Click create and wait for your workspace to provision (usually takes 30-60 seconds)
                </p>
              </div>
            </li>
          </ol>
        </div>

        <div className="bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <CheckCircle2 className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">Reproducible Environments</p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                Kloudlite uses Nix package manager for declarative, reproducible package installations. Your exact package versions are preserved and can be shared across your team.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 4: Access Workspace */}
      <section id="tips" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <span className="text-primary-foreground text-lg font-bold">4</span>
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Access Your Workspace</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Kloudlite offers multiple ways to access your workspace. Choose the method that fits your workflow best.
        </p>

        {/* SSH Setup */}
        <h3 className="text-foreground mb-4 text-xl font-semibold">Setup: SSH Configuration (Required First)</h3>
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <p className="text-muted-foreground text-sm mb-3 m-0">
            Add this configuration to your <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">~/.ssh/config</code> file:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">
{`Host your-workspace-name
  HostName workspace-your-workspace-name
  User kl
  ProxyJump kloudlite@localhost:2222
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null`}
            </pre>
          </div>
          <p className="text-muted-foreground text-xs mt-3 m-0">
            Replace <code className="bg-muted rounded px-1 py-0.5 font-mono">your-workspace-name</code> with your actual workspace name.
          </p>
        </div>

        {/* Desktop IDEs */}
        <h3 className="text-foreground mb-4 text-xl font-semibold">Desktop IDEs</h3>
        <div className="grid gap-4 mb-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-3">
              <Code2 className="text-primary h-6 w-6" />
              <h4 className="text-card-foreground text-base font-semibold m-0">VS Code Extension</h4>
            </div>
            <p className="text-muted-foreground text-sm mb-2 m-0">
              Install the Kloudlite VS Code extension and click the connection link from your workspace dashboard.
            </p>
            <div className="bg-muted rounded p-2 font-mono text-xs break-all">
              vscode://kloudlite.kloudlite-workspace/connect
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-3">
              <Code2 className="text-primary h-6 w-6" />
              <h4 className="text-card-foreground text-base font-semibold m-0">VS Code (SSH Remote)</h4>
            </div>
            <p className="text-muted-foreground text-sm mb-2 m-0">
              Connect using VS Code&apos;s built-in SSH Remote extension:
            </p>
            <div className="bg-muted rounded p-2 font-mono text-xs overflow-x-auto">
              <pre className="m-0 leading-relaxed">code --remote ssh-remote+your-workspace-name</pre>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-3">
              <Terminal className="text-primary h-6 w-6" />
              <h4 className="text-card-foreground text-base font-semibold m-0">Cursor, IntelliJ, Zed</h4>
            </div>
            <p className="text-muted-foreground text-sm m-0">
              These IDEs also support SSH remote connections. Use your workspace SSH config for seamless access.
            </p>
          </div>
        </div>

        {/* Web-based Access */}
        <h3 className="text-foreground mb-4 text-xl font-semibold">Web-Based Access</h3>
        <div className="grid gap-4 mb-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-3">
              <Globe className="text-primary h-6 w-6" />
              <h4 className="text-card-foreground text-base font-semibold m-0">VS Code Web</h4>
            </div>
            <p className="text-muted-foreground text-sm m-0">
              Full VS Code IDE in your browser. No installation needed - just click and code!
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-3">
              <Terminal className="text-primary h-6 w-6" />
              <h4 className="text-card-foreground text-base font-semibold m-0">Web Terminal</h4>
            </div>
            <p className="text-muted-foreground text-sm m-0">
              Browser-based terminal with full shell access to your workspace.
            </p>
          </div>
        </div>

        {/* AI Assistants */}
        <h3 className="text-foreground mb-4 text-xl font-semibold">AI-Powered Development</h3>
        <div className="grid gap-4 mb-6 sm:grid-cols-2 lg:grid-cols-3">
          <div className="bg-card rounded-lg border p-4">
            <div className="mb-2 flex items-center gap-2">
              <Sparkles className="text-primary h-5 w-5" />
              <h4 className="text-card-foreground text-sm font-semibold m-0">Claude Code</h4>
            </div>
            <p className="text-muted-foreground text-xs m-0">
              Anthropic&apos;s AI coding assistant directly in your terminal
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="mb-2 flex items-center gap-2">
              <Sparkles className="text-primary h-5 w-5" />
              <h4 className="text-card-foreground text-sm font-semibold m-0">OpenCode</h4>
            </div>
            <p className="text-muted-foreground text-xs m-0">
              AI assistant for code generation and debugging
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="mb-2 flex items-center gap-2">
              <Sparkles className="text-primary h-5 w-5" />
              <h4 className="text-card-foreground text-sm font-semibold m-0">Codex</h4>
            </div>
            <p className="text-muted-foreground text-xs m-0">
              AI-powered code completion and assistance
            </p>
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">Recommendation for Beginners</p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                Start with <strong>VS Code Web</strong> for the easiest setup. No configuration needed - just click the link from your workspace dashboard and start coding immediately.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Next Steps */}
      <section id="next-steps" className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-6 text-2xl sm:text-3xl font-bold">Next Steps</h2>
        <div className="grid gap-4 sm:gap-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold leading-snug m-0">
              <Link href="/docs/introduction/installation" className="hover:text-primary transition-colors">
                Installation Guide
              </Link>
            </h3>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Learn how to set up your own Kloudlite installation for your team
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold leading-snug m-0">
              Advanced Configuration
            </h3>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Explore workspace settings, dotfiles, environment variables, and custom configurations
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold leading-snug m-0">
              Package Management
            </h3>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Deep dive into Nix package management, channels, and version pinning
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h3 className="text-card-foreground mb-3 text-lg font-semibold leading-snug m-0">
              CLI Reference
            </h3>
            <p className="text-muted-foreground text-sm leading-relaxed m-0">
              Complete guide to the <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl</code> CLI tool commands
            </p>
          </div>
        </div>
      </section>

      {/* Tips & Best Practices */}
      <section id="next-steps" className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-6 text-2xl sm:text-3xl font-bold">Tips & Best Practices</h2>
        <div className="bg-card space-y-4 rounded-lg border p-4 sm:p-6">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="text-green-500 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-card-foreground text-sm font-medium m-0 mb-1">Start Small</p>
              <p className="text-muted-foreground text-sm m-0">
                Begin with a Development or General Purpose machine type. You can always scale up later.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <CheckCircle2 className="text-green-500 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-card-foreground text-sm font-medium m-0 mb-1">Save Resources</p>
              <p className="text-muted-foreground text-sm m-0">
                Keep environments inactive when not in use and suspend workspaces to save resources.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <CheckCircle2 className="text-green-500 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-card-foreground text-sm font-medium m-0 mb-1">Use Semantic Versioning</p>
              <p className="text-muted-foreground text-sm m-0">
                When installing packages, specify versions for reproducibility (e.g., nodejs@20.10.0).
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <CheckCircle2 className="text-green-500 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-card-foreground text-sm font-medium m-0 mb-1">Configure SSH Early</p>
              <p className="text-muted-foreground text-sm m-0">
                Set up your SSH config file early for seamless access across all desktop IDEs.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <CheckCircle2 className="text-green-500 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-card-foreground text-sm font-medium m-0 mb-1">Leverage AI Assistants</p>
              <p className="text-muted-foreground text-sm m-0">
                Try Claude Code or other AI assistants for faster development and learning.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Support */}
      <section id="tips" className="mb-12 sm:mb-16">
        <div className="bg-card rounded-lg border p-4 sm:p-6 lg:p-8 text-center">
          <h2 className="text-card-foreground mb-4 text-xl sm:text-2xl font-bold leading-tight">Ready to Start Building?</h2>
          <p className="text-muted-foreground mb-6 leading-relaxed">
            You&apos;re all set! Head to your dashboard and create your first workspace.
          </p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button asChild variant="outline" size="lg">
              <Link href="https://github.com/kloudlite/kloudlite">View on GitHub</Link>
            </Button>
            <Button asChild size="lg">
              <Link href="/contact">Get Support</Link>
            </Button>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
