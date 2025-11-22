import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Zap,
  RefreshCw,
  Cloud,
  Code2,
  Users,
  Sparkles,
  CheckCircle2,
  Layers,
  Clock,
  ShieldCheck,
  GitBranch,
  BarChart3,
  Cpu
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'why-kloudlite', title: 'Why Kloudlite?' },
  { id: 'key-features', title: 'Key Features' },
  { id: 'architectural-principles', title: 'Architectural Principles' },
]

export default function WhatIsKloudlitePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">What is Kloudlite?</h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Sparkles className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 text-lg leading-relaxed">
          Kloudlite is a <strong>Development Environment As a Service (DEaaS)</strong> platform that connects
          cloud-based development workspaces directly to your application environments and services.
        </p>

        <div className="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 rounded-xl border-2 border-blue-200 dark:border-blue-800 p-6 sm:p-8">
          <p className="text-blue-900 dark:text-blue-100 text-base leading-relaxed m-0">
            Kloudlite provides seamless integration between cloud-based workspaces and environments, enabling
            developers to work with production-like setups without the complexity of managing infrastructure.
          </p>
        </div>
      </section>

      {/* Why Kloudlite */}
      <section id="why-kloudlite" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Zap className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Why Use Kloudlite?</h2>
        </div>

        <p className="text-muted-foreground mb-8 leading-relaxed">
          Modern applications have become increasingly distributed and complex, leading to longer development
          cycles. Kloudlite aims to reduce build and deployment time, allowing developers to concentrate on
          coding and innovation to boost productivity.
        </p>

        <div className="grid gap-6 sm:gap-8">
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-destructive/10 rounded-lg p-3 flex-shrink-0">
                <span className="text-destructive text-2xl font-bold">❌</span>
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
                  Traditional Development Challenges
                </h3>
                <ul className="text-muted-foreground text-sm space-y-2 m-0 list-disc list-inside">
                  <li>Long build and deployment cycles slow down iteration</li>
                  <li>Complex environment configuration and maintenance</li>
                  <li>Difficulty replicating production environments locally</li>
                  <li>Limited collaboration on shared development environments</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-emerald-500/10 rounded-lg p-3 flex-shrink-0">
                <CheckCircle2 className="h-8 w-8 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
                  Kloudlite Solution
                </h3>
                <ul className="text-muted-foreground text-sm space-y-2 m-0 list-disc list-inside">
                  <li>Instant testing without build or deploy steps</li>
                  <li>Automated environment setup and management</li>
                  <li>Production-like environments on demand</li>
                  <li>Real-time collaboration and debugging capabilities</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div className="bg-card rounded-lg border p-5">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Clock className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Faster Feedback Loops
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Teams dramatically shrink time-to-first-test by intercepting services instead of
                  waiting for container builds or cluster rollouts.
                </p>
              </div>
            </div>
          </div>
          <div className="bg-card rounded-lg border p-5">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <ShieldCheck className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Production Parity
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Every workspace uses the same Kubernetes namespaces, Secrets, and DNS that your
                  service will use in production—no more “works on my machine” regressions.
                </p>
              </div>
            </div>
          </div>
          <div className="bg-card rounded-lg border p-5">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <GitBranch className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Less Ops Overhead
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Platform teams define guard-railed templates once; developers self-serve new
                  environments and workspaces without needing cluster access.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-8 bg-gradient-to-r from-emerald-500/10 via-primary/5 to-blue-500/10 dark:from-emerald-500/20 dark:via-primary/10 dark:to-blue-500/20 border border-primary/20 rounded-2xl p-6 sm:p-8">
          <h3 className="text-foreground text-xl font-semibold mb-4 m-0 flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-primary" />
            Measurable Impact
          </h3>
          <div className="grid gap-4 sm:grid-cols-3 text-sm text-muted-foreground">
            <div>
              <p className="text-foreground text-2xl font-bold mb-1">⏱️ Faster loops</p>
              <p className="m-0 leading-relaxed">
                Inner-loop testing happens in seconds because code changes flow straight from your workspace into the environment.
              </p>
            </div>
            <div>
              <p className="text-foreground text-2xl font-bold mb-1">🔐 0 drift</p>
              <p className="m-0 leading-relaxed">
                Workspaces inherit the same Compose specs, Secrets, and routing rules that run in production.
              </p>
            </div>
            <div>
              <p className="text-foreground text-2xl font-bold mb-1">🧑‍🤝‍🧑 Lower ops load</p>
              <p className="m-0 leading-relaxed">
                Platform teams spend their time on strategic improvements rather than repetitive sandbox provisioning.
              </p>
            </div>
          </div>
        </div>

        <div className="mt-10 grid gap-6 md:grid-cols-3">
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Code2 className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  For Developers
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Ship features without fighting fleet configs. Spin up a workspace, intercept a service, and
                  debug against real dependencies in seconds.
                </p>
              </div>
            </div>
          </div>
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Cpu className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  For Platform Teams
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Define reproducible blueprints for environments and machine types. Enforce policies and observability
                  without blocking day-to-day delivery.
                </p>
              </div>
            </div>
          </div>
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Users className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  For Engineering Leaders
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Reduce cycle time, improve release quality, and give teams on-demand, cost-aware environments while
                  keeping compliance requirements intact.
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
            <Sparkles className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Key Features</h2>
        </div>

        <div className="grid gap-6 sm:gap-8">
          {/* Feature 1 */}
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Zap className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Instant Testing Without Build or Deploy
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Skip standard build/deploy steps and test changes immediately in live environments through
                  connected workspaces. See your code changes reflected instantly without waiting for lengthy
                  deployment pipelines.
                </p>
              </div>
            </div>
          </div>

          {/* Feature 2 */}
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <RefreshCw className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Seamless Environment Switching
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Move between environments without reconfiguring settings or connections, keeping focus on code
                  development. Switch from development to staging to production-like environments with a single command.
                </p>
              </div>
            </div>
          </div>

          {/* Feature 3 */}
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Cloud className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Ephemeral and Stateless Environments
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Use lightweight, easily cloned environments where each developer can work independently on parallel
                  development tasks. Create and destroy environments on demand without affecting others.
                </p>
              </div>
            </div>
          </div>

          {/* Feature 4 */}
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Code2 className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Dev Containers with IDE Support
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Develop in SSH-enabled containers with IDE attachment capabilities, using Nix package manager
                  for dependency management. Connect your favorite IDE (VS Code, JetBrains, etc.) directly to
                  cloud-based workspaces.
                </p>
              </div>
            </div>
          </div>

          {/* Feature 6 */}
          <div className="bg-card rounded-lg border p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
                <Users className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-lg font-semibold mb-2 m-0">
                  Collaborative Coding and Debugging
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Enable real-time collaboration across multiple services within the same environment, with app
                  interception and debugging capabilities. Work together with your team on the same codebase in
                  shared development environments.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Key Architectural Principles */}
      <section id="architectural-principles" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Layers className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Key Architectural Principles</h2>
        </div>

        <p className="text-muted-foreground mb-8 leading-relaxed">
          Kloudlite is built on a core architectural principle that fundamentally transforms the development experience.
        </p>

        <div className="bg-card rounded-lg border p-6 mb-6">
          <div className="flex items-start gap-4 mb-6">
            <div className="bg-primary/10 rounded-lg p-3 flex-shrink-0">
              <Zap className="h-6 w-6 text-primary" />
            </div>
            <div className="flex-1">
              <h3 className="text-card-foreground text-xl font-semibold mb-3 m-0">
                Reduced Development Loop
              </h3>
              <p className="text-muted-foreground text-sm leading-relaxed m-0">
                The development loop—write code, test, debug—determines how fast you can iterate. Traditional
                workflows require building and deploying before you can test, taking minutes to hours. Kloudlite
                eliminates these steps, reducing iterations from minutes to seconds.
              </p>
            </div>
          </div>

          <div className="space-y-6">
            {/* Traditional Development Loop */}
            <div className="border-l-4 border-destructive pl-4">
              <h4 className="text-card-foreground font-semibold mb-2 text-base">
                Traditional Development Loop (Minutes to Hours)
              </h4>
              <ol className="text-muted-foreground text-sm space-y-2 list-decimal list-inside">
                <li><strong>Write Code:</strong> Make changes to your application</li>
                <li><strong>Build:</strong> Compile code, bundle assets, create container images (3-15 minutes)</li>
                <li><strong>Deploy:</strong> Push images, update services, wait for rollout (2-10 minutes)</li>
                <li><strong>Test:</strong> Manually test or run automated tests (1-5 minutes)</li>
                <li><strong>Debug:</strong> If issues found, repeat entire cycle</li>
              </ol>
              <p className="text-destructive text-xs mt-3 font-medium">
                ⏱️ Total: 6-30 minutes per iteration
              </p>
            </div>

            {/* Kloudlite Development Loop */}
            <div className="border-l-4 border-emerald-500 pl-4">
              <h4 className="text-card-foreground font-semibold mb-2 text-base">
                Kloudlite Development Loop (Seconds)
              </h4>
              <ol className="text-muted-foreground text-sm space-y-2 list-decimal list-inside">
                <li><strong>Write Code:</strong> Make changes in your workspace</li>
                <li><strong>Test Instantly:</strong> Changes reflected immediately via service intercepts</li>
                <li><strong>Debug in Real-Time:</strong> Use actual environment services and data</li>
              </ol>
              <p className="text-emerald-600 dark:text-emerald-400 text-xs mt-3 font-medium">
                ⚡ Total: Seconds per iteration
              </p>
            </div>

            {/* How Kloudlite Achieves This */}
            <div className="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 rounded-lg border border-blue-200 dark:border-blue-800 p-4">
              <h4 className="text-blue-900 dark:text-blue-100 font-semibold mb-3 text-base">
                How Kloudlite Achieves This
              </h4>
              <ul className="space-y-3">
                <li className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <strong className="text-blue-900 dark:text-blue-100 text-sm">Service Intercepts:</strong>
                    <span className="text-blue-800 dark:text-blue-200 text-sm"> Route traffic from environment services directly to your workspace. Your code runs in the workspace but serves real requests from the environment.</span>
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <strong className="text-blue-900 dark:text-blue-100 text-sm">Environment Connection:</strong>
                    <span className="text-blue-800 dark:text-blue-200 text-sm"> Access all environment services (databases, caches, APIs) from your workspace by name. No need to mock services or maintain local copies.</span>
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <strong className="text-blue-900 dark:text-blue-100 text-sm">No Build/Deploy Steps:</strong>
                    <span className="text-blue-800 dark:text-blue-200 text-sm"> Skip container builds and deployments entirely. Code changes in your workspace are immediately effective.</span>
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <strong className="text-blue-900 dark:text-blue-100 text-sm">Live Debugging:</strong>
                    <span className="text-blue-800 dark:text-blue-200 text-sm"> Set breakpoints, inspect variables, and debug with your IDE using real environment data and services.</span>
                  </div>
                </li>
              </ul>
            </div>

            {/* Impact */}
            <div className="bg-gradient-to-br from-emerald-50 to-green-50 dark:from-emerald-950/30 dark:to-green-950/30 rounded-lg border border-emerald-200 dark:border-emerald-800 p-4">
              <h4 className="text-emerald-900 dark:text-emerald-100 font-semibold mb-2 text-base">
                The Impact
              </h4>
              <p className="text-emerald-800 dark:text-emerald-200 text-sm leading-relaxed m-0 mb-3">
                By eliminating build and deploy steps, Kloudlite reduces the development loop from minutes to seconds.
                This means:
              </p>
              <ul className="text-emerald-800 dark:text-emerald-200 text-sm space-y-1.5">
                <li className="flex items-center gap-2">
                  <Zap className="h-4 w-4 flex-shrink-0" />
                  <strong>Orders-of-magnitude faster iterations</strong> - Test changes in seconds instead of minutes
                </li>
                <li className="flex items-center gap-2">
                  <Zap className="h-4 w-4 flex-shrink-0" />
                  <strong>More experiments</strong> - Try different approaches without the time penalty
                </li>
                <li className="flex items-center gap-2">
                  <Zap className="h-4 w-4 flex-shrink-0" />
                  <strong>Better code quality</strong> - Test thoroughly because testing is fast
                </li>
                <li className="flex items-center gap-2">
                  <Zap className="h-4 w-4 flex-shrink-0" />
                  <strong>Maintained focus</strong> - Stay in flow state without context switching
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
