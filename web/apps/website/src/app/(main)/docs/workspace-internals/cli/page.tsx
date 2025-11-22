import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Terminal,
  Package,
  Network,
  Shield,
  Info,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'workspace', title: 'Workspace Commands' },
  { id: 'packages', title: 'Package Management' },
  { id: 'environment', title: 'Environment Connection' },
  { id: 'intercept', title: 'Service Interception' },
]

export default function CLIReferencePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        CLI Reference
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          The <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl</code> CLI is
          a command-line tool for managing Kloudlite workspaces from within your Dev Container. It
          provides commands to manage workspace configuration, packages, environment connections,
          and service interception.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <p className="text-card-foreground text-sm mb-3 m-0 font-medium">
            Example: Common Commands
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-3">
            <div>
              <p className="text-muted-foreground mb-1 text-xs"># Show workspace status</p>
              <pre className="m-0 leading-relaxed">kl status</pre>
            </div>
            <div>
              <p className="text-muted-foreground mb-1 text-xs"># Add packages</p>
              <pre className="m-0 leading-relaxed">kl pkg add git vim curl</pre>
            </div>
            <div>
              <p className="text-muted-foreground mb-1 text-xs"># Connect to environment</p>
              <pre className="m-0 leading-relaxed">kl env connect my-env</pre>
            </div>
            <div>
              <p className="text-muted-foreground mb-1 text-xs"># Intercept a service</p>
              <pre className="m-0 leading-relaxed">kl intercept start api-server</pre>
            </div>
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Interactive Mode
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                Many commands support interactive mode using fuzzy-find (fzf) when run without
                arguments, making it easy to select packages, environments, or services.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Workspace Commands */}
      <section id="workspace" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Workspace Commands
          </h2>
        </div>

        {/* status */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl status <span className="text-muted-foreground text-sm font-normal">(aliases: st, s)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Display comprehensive workspace information including metadata, resource usage, access
            URLs, and timing information.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">kl status</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Displays:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Workspace name, namespace, and owner</li>
              <li>Phase and status message</li>
              <li>Resource usage (CPU, Memory, Storage)</li>
              <li>Access URLs for services</li>
              <li>Runtime and activity timestamps</li>
            </ul>
          </div>
        </div>

        {/* version */}
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl version <span className="text-muted-foreground text-sm font-normal">(aliases: v)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Show version information for the kl CLI tool.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">kl version</pre>
          </div>
        </div>
      </section>

      {/* Package Management */}
      <section id="packages" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Package className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Package Management
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Manage Nix packages in your workspace using the{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl pkg</code>{' '}
          command (aliases: <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">package</code>,{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">p</code>).
        </p>

        {/* pkg search */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl pkg search <span className="text-muted-foreground text-sm font-normal">(aliases: find, s)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Search for available Nix packages using the Devbox package registry.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg search nodejs</pre>
            <pre className="m-0 leading-relaxed">kl p s python</pre>
          </div>
        </div>

        {/* pkg add */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl pkg add <span className="text-muted-foreground text-sm font-normal">(aliases: a)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Add packages to your workspace. When run without arguments, enters interactive mode
            with fuzzy search and version selection. When run with package names, automatically
            installs the latest version.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg add              # Interactive mode</pre>
            <pre className="m-0 leading-relaxed">kl pkg add git vim curl # Add multiple packages</pre>
            <pre className="m-0 leading-relaxed">kl p a nodejs python    # Using aliases</pre>
          </div>
          <div className="mt-4 bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded border p-3">
            <div className="flex gap-2">
              <AlertCircle className="text-amber-600 dark:text-amber-400 h-4 w-4 flex-shrink-0 mt-0.5" />
              <p className="text-amber-800 dark:text-amber-200 text-xs m-0 leading-relaxed">
                <strong>Timeout:</strong> Package installation waits up to 5 minutes for completion.
              </p>
            </div>
          </div>
        </div>

        {/* pkg install */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl pkg install <span className="text-muted-foreground text-sm font-normal">(aliases: i, in)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Install a specific Nix package with optional version, channel, or commit specification.
          </p>
          <div className="mb-3">
            <p className="text-muted-foreground text-xs font-medium mb-2">Flags:</p>
            <ul className="text-muted-foreground text-xs space-y-1">
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--version</code> - Semantic version (e.g., 24.0.0)
              </li>
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--channel</code> - Nixpkgs channel (e.g., nixos-24.05)
              </li>
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--commit</code> - Exact nixpkgs commit hash
              </li>
            </ul>
          </div>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg install nodejs</pre>
            <pre className="m-0 leading-relaxed">kl pkg install nodejs --version 20.0.0</pre>
            <pre className="m-0 leading-relaxed">kl pkg install vim --channel nixos-24.05</pre>
            <pre className="m-0 leading-relaxed">kl p i git --channel unstable</pre>
          </div>
        </div>

        {/* pkg uninstall */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl pkg uninstall <span className="text-muted-foreground text-sm font-normal">(aliases: remove, rm, un)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Remove a package from your workspace.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg uninstall git</pre>
            <pre className="m-0 leading-relaxed">kl p rm vim</pre>
          </div>
        </div>

        {/* pkg list */}
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl pkg list <span className="text-muted-foreground text-sm font-normal">(aliases: ls, l)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Display all packages in the workspace spec and their installation status.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg list</pre>
            <pre className="m-0 leading-relaxed">kl p ls</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Output includes:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Workspace spec packages with channel/commit info</li>
              <li>Installed packages with version and binary path</li>
              <li>Failed packages (if any)</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Environment Connection */}
      <section id="environment" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Environment Connection
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Manage workspace environment connections using the{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl env</code>{' '}
          command (aliases: <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">e</code>,{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">environment</code>).
          When connected to an environment, services can be accessed using short DNS names.
        </p>

        {/* env connect */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl env connect <span className="text-muted-foreground text-sm font-normal">(aliases: c, conn)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Connect workspace to an environment for simplified service access. If no environment
            name is provided, an interactive list will be shown.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl env connect           # Interactive selection</pre>
            <pre className="m-0 leading-relaxed">kl env connect my-env    # Connect to specific env</pre>
            <pre className="m-0 leading-relaxed">kl e c my-env            # Using aliases</pre>
          </div>
          <div className="mt-4 bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded border p-3">
            <div className="flex gap-2">
              <CheckCircle2 className="text-green-600 dark:text-green-400 h-4 w-4 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-green-900 dark:text-green-100 text-xs font-medium m-0 mb-1">
                  After Connecting
                </p>
                <p className="text-green-800 dark:text-green-200 text-xs m-0 leading-relaxed">
                  Access services using short names like{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono">postgres:5432</code> instead of
                  full qualified names.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* env disconnect */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl env disconnect <span className="text-muted-foreground text-sm font-normal">(aliases: d, disc)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Disconnect workspace from the connected environment. This removes all active intercepts
            for that environment.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl env disconnect</pre>
            <pre className="m-0 leading-relaxed">kl e d</pre>
          </div>
        </div>

        {/* env status */}
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl env status <span className="text-muted-foreground text-sm font-normal">(aliases: st, stat)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Show the current environment connection status and available services.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl env status</pre>
            <pre className="m-0 leading-relaxed">kl e st</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Displays:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Connection status and environment name</li>
              <li>Target namespace</li>
              <li>Service access examples</li>
              <li>Active service intercepts</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Service Interception */}
      <section id="intercept" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Shield className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Service Interception
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Manage service interception using the{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl intercept</code>{' '}
          command (aliases: <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">i</code>,{' '}
          <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">int</code>).
          Service interception redirects traffic from a service to your workspace.
        </p>

        <div className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded-lg border p-3 sm:p-4 mb-6">
          <div className="flex gap-2 sm:gap-3">
            <AlertCircle className="text-amber-600 dark:text-amber-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-amber-900 dark:text-amber-100 text-sm font-medium m-0 mb-1">
                Prerequisite
              </p>
              <p className="text-amber-800 dark:text-amber-200 text-sm m-0 leading-relaxed">
                You must connect to an environment using{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl env connect</code>{' '}
                before you can intercept services.
              </p>
            </div>
          </div>
        </div>

        {/* intercept start */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl intercept start <span className="text-muted-foreground text-sm font-normal">(aliases: s, begin)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Start intercepting a service to redirect its traffic to your workspace. If no service
            name is provided, an interactive list will be shown with port mapping options.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl intercept start            # Interactive selection</pre>
            <pre className="m-0 leading-relaxed">kl intercept start api-server # Intercept specific service</pre>
            <pre className="m-0 leading-relaxed">kl i s api-server             # Using aliases</pre>
          </div>
        </div>

        {/* intercept stop */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl intercept stop <span className="text-muted-foreground text-sm font-normal">(aliases: sp, end)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Stop intercepting a service and restore normal traffic routing. If no service name is
            provided, an interactive list of active intercepts will be shown.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl intercept stop            # Interactive selection</pre>
            <pre className="m-0 leading-relaxed">kl intercept stop api-server # Stop specific intercept</pre>
            <pre className="m-0 leading-relaxed">kl i sp api-server           # Using aliases</pre>
          </div>
        </div>

        {/* intercept list */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl intercept list <span className="text-muted-foreground text-sm font-normal">(aliases: ls)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            List all active service intercepts in the connected environment.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl intercept list</pre>
            <pre className="m-0 leading-relaxed">kl i ls</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Output includes:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Service name and phase (Pending, Active, Failed)</li>
              <li>Port mappings (service → workspace)</li>
              <li>Error messages (if failed)</li>
            </ul>
          </div>
        </div>

        {/* intercept status */}
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h3 className="text-card-foreground text-lg font-semibold mb-3 m-0">
            kl intercept status <span className="text-muted-foreground text-sm font-normal">(aliases: st)</span>
          </h3>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Show detailed status of service intercept(s). Can show all intercepts or a specific
            service intercept.
          </p>
          <div className="bg-muted rounded p-3 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl intercept status            # Show all intercepts</pre>
            <pre className="m-0 leading-relaxed">kl intercept status api-server # Show specific service</pre>
            <pre className="m-0 leading-relaxed">kl i st api-server             # Using aliases</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Displays:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Service and workspace details</li>
              <li>Phase and status messages</li>
              <li>Port mappings</li>
              <li>Workspace pod details (name, IP)</li>
              <li>Affected pods and start time</li>
            </ul>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
