import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Network, Globe, Terminal, Server, Database, Code2, Info, CheckCircle2, AlertCircle, Play, Square } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'how-it-works', title: 'How It Works' },
  { id: 'using-intercepts', title: 'Using Service Intercepts' },
  { id: 'managing-intercepts', title: 'Managing Intercepts' },
  { id: 'use-cases', title: 'Use Cases' },
]

export default function ServiceInterceptsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">Service Intercepts</h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">What are Service Intercepts?</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Service Intercepts allow developers to route traffic from a running service in an environment
          directly to their workspace for debugging and testing. When a service is intercepted, all
          incoming traffic is redirected to the workspace instead of the original service.
        </p>

        {/* Intercept Flow Diagram */}
        <div className="bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-xl border-2 border-slate-200 dark:border-slate-700 p-6 sm:p-8 mb-6">
          <div className="grid md:grid-cols-2 gap-8">
            {/* Before Intercept */}
            <div>
              <h3 className="text-foreground text-lg font-semibold mb-4 flex items-center gap-2">
                <div className="bg-slate-300 dark:bg-slate-700 rounded px-2 py-1 text-xs font-bold">
                  BEFORE
                </div>
                Normal Flow
              </h3>
              <div className="bg-white dark:bg-slate-950 rounded-lg border-2 border-slate-300 dark:border-slate-600 p-6">
                <div className="flex flex-col items-center space-y-4">
                  {/* Client Request at top */}
                  <div className="inline-flex items-center gap-2 bg-slate-100 dark:bg-slate-800 rounded-lg px-4 py-3 border-2 border-slate-400 dark:border-slate-600">
                    <Globe className="h-5 w-5 text-slate-600 dark:text-slate-400" />
                    <span className="text-slate-700 dark:text-slate-300 text-sm font-semibold">
                      Client Request
                    </span>
                  </div>

                  {/* Arrow down */}
                  <div className="flex flex-col items-center">
                    <div className="w-0.5 h-12 bg-slate-400 dark:bg-slate-600"></div>
                    <div className="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[10px] border-t-slate-400 dark:border-t-slate-600"></div>
                  </div>

                  {/* Environment with Service */}
                  <div className="w-full bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950 dark:to-orange-950 rounded-lg border-2 border-amber-400 dark:border-amber-600 p-4">
                    <div className="flex items-center gap-2 mb-3">
                      <Database className="h-5 w-5 text-amber-600 dark:text-amber-400" />
                      <h4 className="text-amber-800 dark:text-amber-200 font-semibold text-sm m-0">
                        Environment
                      </h4>
                    </div>
                    <div className="bg-white dark:bg-amber-900/20 rounded p-3 border border-amber-300 dark:border-amber-700">
                      <div className="flex items-center gap-2 justify-center">
                        <Server className="h-4 w-4 text-amber-700 dark:text-amber-300" />
                        <span className="text-amber-900 dark:text-amber-100 text-sm font-mono font-semibold">
                          api-service:8080
                        </span>
                      </div>
                    </div>
                  </div>

                  <p className="text-center text-xs text-slate-500 dark:text-slate-400 mt-2">
                    Traffic flows to service in environment
                  </p>
                </div>
              </div>
            </div>

            {/* After Intercept */}
            <div>
              <h3 className="text-foreground text-lg font-semibold mb-4 flex items-center gap-2">
                <div className="bg-emerald-500 text-white rounded px-2 py-1 text-xs font-bold">
                  AFTER
                </div>
                Intercepted Flow
              </h3>
              <div className="bg-white dark:bg-slate-950 rounded-lg border-2 border-emerald-400 dark:border-emerald-600 p-6">
                <div className="flex flex-col items-center space-y-4">
                  {/* Client Request at top */}
                  <div className="inline-flex items-center gap-2 bg-slate-100 dark:bg-slate-800 rounded-lg px-4 py-3 border-2 border-slate-400 dark:border-slate-600">
                    <Globe className="h-5 w-5 text-slate-600 dark:text-slate-400" />
                    <span className="text-slate-700 dark:text-slate-300 text-sm font-semibold">
                      Client Request
                    </span>
                  </div>

                  {/* Gray arrow down to environment (showing it would go there) */}
                  <div className="flex flex-col items-center">
                    <div className="w-0.5 h-8 bg-slate-300 dark:bg-slate-600"></div>
                    <div className="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[10px] border-t-slate-300 dark:border-t-slate-600"></div>
                  </div>

                  {/* Environment with Service - DISABLED/BYPASSED */}
                  <div className="relative w-full">
                    <div className="w-full bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950 dark:to-orange-950 rounded-lg border-2 border-amber-300 dark:border-amber-700 p-4 opacity-40 relative">
                      <div className="flex items-center gap-2 mb-3">
                        <Database className="h-5 w-5 text-amber-600 dark:text-amber-400" />
                        <h4 className="text-amber-800 dark:text-amber-200 font-semibold text-sm m-0">
                          Environment
                        </h4>
                      </div>
                      <div className="bg-white dark:bg-amber-900/20 rounded p-3 border border-amber-300 dark:border-amber-700 relative">
                        <div className="flex items-center gap-2 justify-center">
                          <Server className="h-4 w-4 text-amber-700 dark:text-amber-300" />
                          <span className="text-amber-900 dark:text-amber-100 text-sm font-mono line-through">
                            api-service:8080
                          </span>
                        </div>
                      </div>
                      {/* BYPASSED badge on top right */}
                      <div className="absolute -top-2 -right-2 bg-red-500 text-white text-xs font-bold px-2 py-1 rounded shadow-lg">
                        BYPASSED
                      </div>
                    </div>
                  </div>

                  {/* Curved redirect to workspace */}
                  <div className="flex flex-col items-center -mt-2">
                    <div className="text-emerald-600 dark:text-emerald-400 text-xs font-bold uppercase tracking-wide mb-1">
                      ↓ Redirected ↓
                    </div>
                    <div className="w-1 h-8 bg-gradient-to-b from-emerald-500 to-emerald-600 rounded"></div>
                    <div className="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[10px] border-t-emerald-600"></div>
                  </div>

                  {/* Workspace receiving traffic */}
                  <div className="w-full bg-gradient-to-br from-emerald-50 to-green-50 dark:from-emerald-950 dark:to-green-950 rounded-lg border-2 border-emerald-500 dark:border-emerald-600 p-4 ring-2 ring-emerald-400">
                    <div className="flex items-center gap-2 mb-3">
                      <Code2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
                      <h4 className="text-emerald-800 dark:text-emerald-200 font-semibold text-sm m-0">
                        Workspace
                      </h4>
                    </div>
                    <div className="bg-white dark:bg-emerald-900/20 rounded p-3 border border-emerald-300 dark:border-emerald-700">
                      <div className="flex items-center gap-2 justify-center">
                        <Terminal className="h-4 w-4 text-emerald-700 dark:text-emerald-300" />
                        <span className="text-emerald-900 dark:text-emerald-100 text-sm font-mono font-semibold">
                          localhost:8080
                        </span>
                      </div>
                    </div>
                  </div>

                  <p className="text-center text-xs text-emerald-600 dark:text-emerald-400 mt-2 font-medium">
                    Traffic flows to workspace instead
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Legend */}
          <div className="mt-6 pt-6 border-t border-slate-200 dark:border-slate-700">
            <div className="flex flex-wrap gap-4 justify-center text-xs text-slate-600 dark:text-slate-400">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-slate-400"></div>
                <span>Normal traffic flow</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-emerald-500"></div>
                <span>Intercepted traffic</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                <span>Service bypassed</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 text-2xl sm:text-3xl font-bold">How Service Intercepts Work</h2>
        <div className="grid gap-4 sm:gap-6 mb-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <div className="text-primary font-bold text-lg">1</div>
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-base font-semibold mb-2 m-0">
                  Enable Intercept
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  From your workspace, you activate an intercept for a specific service running in an
                  environment (e.g., <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">api-service</code>)
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <div className="text-primary font-bold text-lg">2</div>
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-base font-semibold mb-2 m-0">
                  Traffic Redirection
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  All incoming requests to the service are automatically redirected to your workspace.
                  The original service is bypassed completely.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <div className="text-primary font-bold text-lg">3</div>
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-base font-semibold mb-2 m-0">
                  Debug & Test
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  Handle requests directly in your workspace with your local code changes. Set
                  breakpoints, inspect variables, and test fixes in real-time with live traffic.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 rounded-lg p-2 flex-shrink-0">
                <div className="text-primary font-bold text-lg">4</div>
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground text-base font-semibold mb-2 m-0">
                  Disable Intercept
                </h3>
                <p className="text-muted-foreground text-sm leading-relaxed m-0">
                  When you&apos;re done debugging, disable the intercept and traffic automatically
                  flows back to the original service in the environment.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Using Service Intercepts */}
      <section id="using-intercepts" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Using Service Intercepts
          </h2>
        </div>

        {/* Prerequisite */}
        <div className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded-lg border p-3 sm:p-4 mb-6">
          <div className="flex gap-2 sm:gap-3">
            <AlertCircle className="text-amber-600 dark:text-amber-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-amber-900 dark:text-amber-100 text-sm font-medium m-0 mb-1">
                Prerequisite
              </p>
              <p className="text-amber-800 dark:text-amber-200 text-sm m-0 leading-relaxed">
                You must first connect to an environment using{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl env connect</code>{' '}
                before you can intercept services. Service intercepts only work within a connected
                environment.
              </p>
            </div>
          </div>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Use the <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl intercept start</code>{' '}
          command from within your workspace to intercept a service running in the connected
          environment.
        </p>

        {/* Interactive Mode */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Play className="text-green-500 h-5 w-5" />
            Interactive Mode (Recommended)
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Run the command without arguments to see a list of available services in the connected
            environment and select one interactively using fuzzy-find.
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">kl intercept start</pre>
          </div>
          <div className="mt-3 bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded border p-3">
            <p className="text-blue-800 dark:text-blue-200 text-xs m-0 leading-relaxed">
              This will display all services in the environment and let you select using arrow keys
              and search. You&apos;ll then be prompted to configure port mappings.
            </p>
          </div>
        </div>

        {/* Direct Intercept */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Play className="text-green-500 h-5 w-5" />
            Intercept Specific Service
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            If you know the service name, you can intercept it directly:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl intercept start api-server</pre>
            <pre className="m-0 leading-relaxed">kl intercept start backend-api</pre>
            <pre className="m-0 leading-relaxed">kl i s api-server          # Using aliases</pre>
          </div>
        </div>

        {/* Port Mapping */}
        <div className="bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-950 dark:to-cyan-950 rounded-lg border-2 border-blue-300 dark:border-blue-700 p-4 sm:p-6">
          <h4 className="text-blue-900 dark:text-blue-100 font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5" />
            Port Mapping Configuration
          </h4>
          <p className="text-blue-800 dark:text-blue-200 text-sm mb-3 leading-relaxed">
            When you start an intercept, you need to map service ports to your workspace ports.
            For example, if the service runs on port 8080, you might map it to port 3000 in your
            workspace where your local development server is running.
          </p>
          <div className="bg-white dark:bg-blue-900/20 rounded p-3 border border-blue-300 dark:border-blue-700">
            <p className="text-blue-900 dark:text-blue-100 text-xs font-medium mb-2">Example Mapping:</p>
            <ul className="text-blue-800 dark:text-blue-200 text-xs space-y-1 m-0 list-disc list-inside">
              <li>Service port <code className="bg-muted rounded px-1 py-0.5 font-mono">8080</code> → Workspace port <code className="bg-muted rounded px-1 py-0.5 font-mono">3000</code></li>
              <li>Service port <code className="bg-muted rounded px-1 py-0.5 font-mono">9090</code> → Workspace port <code className="bg-muted rounded px-1 py-0.5 font-mono">9090</code> (same port)</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Managing Intercepts */}
      <section id="managing-intercepts" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Managing Intercepts
          </h2>
        </div>

        <div className="space-y-6">
          {/* List Active Intercepts */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Server className="text-primary h-5 w-5" />
              List Active Intercepts
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              View all active service intercepts in the connected environment:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
              <pre className="m-0 leading-relaxed">kl intercept list</pre>
              <pre className="m-0 leading-relaxed">kl i ls              # Using aliases</pre>
            </div>
            <div className="mt-4">
              <p className="text-muted-foreground text-xs font-medium mb-2">Output shows:</p>
              <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
                <li>Service name and phase (Pending, Active, Failed)</li>
                <li>Port mappings (service port → workspace port)</li>
                <li>Error messages for failed intercepts</li>
              </ul>
            </div>
          </div>

          {/* Check Intercept Status */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Info className="text-primary h-5 w-5" />
              Check Intercept Status
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Get detailed status information about a specific service intercept:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
              <pre className="m-0 leading-relaxed">kl intercept status api-server</pre>
              <pre className="m-0 leading-relaxed">kl i st api-server              # Using aliases</pre>
            </div>
            <div className="mt-4">
              <p className="text-muted-foreground text-xs font-medium mb-2">Detailed information includes:</p>
              <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
                <li>Service and workspace details</li>
                <li>Phase and status messages</li>
                <li>Port mappings</li>
                <li>Workspace pod details (name, IP address)</li>
                <li>List of affected pods</li>
                <li>Intercept start time</li>
              </ul>
            </div>
          </div>

          {/* Stop Intercept */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Square className="text-red-500 h-5 w-5" />
              Stop Service Intercept
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Stop intercepting a service and restore normal traffic routing. You can use
              interactive mode or specify the service name directly:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
              <pre className="m-0 leading-relaxed">kl intercept stop              # Interactive selection</pre>
              <pre className="m-0 leading-relaxed">kl intercept stop api-server   # Stop specific service</pre>
              <pre className="m-0 leading-relaxed">kl i sp api-server             # Using aliases</pre>
            </div>
            <div className="mt-4 bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded border p-3">
              <div className="flex gap-2">
                <CheckCircle2 className="text-green-600 dark:text-green-400 h-4 w-4 flex-shrink-0 mt-0.5" />
                <p className="text-green-800 dark:text-green-200 text-xs m-0 leading-relaxed">
                  After stopping, traffic will automatically route back to the original service in
                  the environment within ~30 seconds.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Use Cases */}
      <section id="use-cases" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Code2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Use Cases</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Terminal className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Debug with Real Traffic
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Intercept a service to debug issues using real incoming requests from other
                  services or external clients. Set breakpoints and inspect variables with actual
                  production-like data.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Code2 className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Test API Changes
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Test new API endpoints or modifications by intercepting the service and handling
                  requests from dependent services without deploying changes to the environment.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Server className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Develop with Service Integration
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Develop new features locally while staying integrated with other running services
                  in the environment. Your workspace receives real requests from other services.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Database className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Performance Testing
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Profile and optimize your service by intercepting it and analyzing performance
                  with real traffic patterns and load from the environment.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Network className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Reproduce and Fix Bugs
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Reproduce bugs that only occur with specific service interactions or data by
                  intercepting traffic and iterating on fixes in real-time.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-6 bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <CheckCircle2 className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Safe Testing Environment
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                Service intercepts let you safely test and debug with live traffic without affecting
                the actual service running in the environment. Stop the intercept anytime to restore
                normal operation.
              </p>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
