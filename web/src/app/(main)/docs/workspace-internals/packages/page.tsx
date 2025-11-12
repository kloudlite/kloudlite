import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Package,
  Search,
  Download,
  Trash2,
  List,
  Info,
  CheckCircle2,
  AlertCircle,
  Terminal,
  Layers,
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'how-it-works', title: 'How It Works' },
  { id: 'searching', title: 'Searching for Packages' },
  { id: 'adding', title: 'Adding Packages' },
  { id: 'managing', title: 'Managing Packages' },
  { id: 'best-practices', title: 'Best Practices' },
]

export default function PackageManagementPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        Package Management
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Package className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Kloudlite uses Nix package manager to provide reproducible, declarative package management
          for your workspaces. Packages are installed at the workmachine (host) level and made
          available in each workspace&apos;s PATH based on its configuration.
        </p>

        <div className="grid gap-4 sm:gap-6 mb-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Layers className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Host-Level Installation
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Packages are installed once at the workmachine host and shared across workspaces,
                  saving disk space and installation time.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <CheckCircle2 className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Per-Workspace Configuration
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Each workspace has its own package configuration. Only the packages specified
                  for a workspace are available in its PATH.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Powered by Nix
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                All packages come from the Nix package repository via the Devbox registry,
                providing access to thousands of packages with reproducible builds and version pinning.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Layers className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">How It Works</h2>
        </div>

        <div className="space-y-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                1
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Add Package to Workspace
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  When you add a package using{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl pkg add</code>,
                  it&apos;s added to your workspace&apos;s package configuration.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                2
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Host-Level Installation
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Nix installs the package at the workmachine host level. If another workspace
                  uses the same package, it references the same installation.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                3
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Available in Workspace PATH
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  The package binary becomes available in your workspace&apos;s PATH. You can
                  immediately use it from the terminal or in your code.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                4
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Persistent Across Restarts
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Packages persist at the host level. When you restart a workspace, all configured
                  packages are immediately available without reinstallation.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Searching for Packages */}
      <section id="searching" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Search className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Searching for Packages
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Use the <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl pkg search</code>{' '}
          command to search the Devbox package registry for available packages.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Search className="text-primary h-5 w-5" />
            Search Command
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Search for packages by name or keyword:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg search nodejs</pre>
            <pre className="m-0 leading-relaxed">kl pkg search python</pre>
            <pre className="m-0 leading-relaxed">kl p s postgresql      # Using aliases</pre>
          </div>
          <div className="mt-4">
            <p className="text-muted-foreground text-xs font-medium mb-2">Output shows:</p>
            <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
              <li>Package name</li>
              <li>Available versions</li>
            </ul>
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Devbox Package Registry
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                Search queries the Devbox package registry, which provides curated access to Nix
                packages with simplified versioning.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Adding Packages */}
      <section id="adding" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Download className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Adding Packages
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          There are multiple ways to add packages to your workspace, depending on your needs for
          version control and interactive selection.
        </p>

        {/* Interactive Mode */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <CheckCircle2 className="text-green-500 h-5 w-5" />
            Interactive Mode (Recommended)
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Run without arguments to search, select packages, and choose specific versions
            interactively using fuzzy-find:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">kl pkg add</pre>
          </div>
          <div className="mt-3 bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded border p-3">
            <p className="text-blue-800 dark:text-blue-200 text-xs m-0 leading-relaxed">
              This lets you search for packages, view available versions, and select the exact
              version you want to install.
            </p>
          </div>
        </div>

        {/* Quick Add */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Download className="text-primary h-5 w-5" />
            Quick Add (Latest Version)
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Provide package names directly to install the latest version of each:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg add git vim curl</pre>
            <pre className="m-0 leading-relaxed">kl pkg add nodejs python</pre>
            <pre className="m-0 leading-relaxed">kl p a jq           # Using aliases</pre>
          </div>
        </div>

        {/* Specific Version Installation */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Package className="text-primary h-5 w-5" />
            Install Specific Version
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Use <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl pkg install</code>{' '}
            to specify exact versions, channels, or commits:
          </p>
          <div className="mb-3">
            <p className="text-muted-foreground text-xs font-medium mb-2">Available Options:</p>
            <ul className="text-muted-foreground text-xs space-y-1">
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--version</code> - Semantic version (e.g., 20.0.0)
              </li>
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--channel</code> - Nixpkgs channel (e.g., nixos-24.05, unstable)
              </li>
              <li>
                <code className="bg-muted rounded px-1 py-0.5 font-mono">--commit</code> - Exact nixpkgs commit hash
              </li>
            </ul>
          </div>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl pkg install nodejs --version 20.0.0</pre>
            <pre className="m-0 leading-relaxed">kl pkg install python --version 3.11.0</pre>
            <pre className="m-0 leading-relaxed">kl pkg install vim --channel nixos-24.05</pre>
            <pre className="m-0 leading-relaxed">kl p i go --channel unstable</pre>
          </div>
        </div>

        {/* Installation Process */}
        <div className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <AlertCircle className="text-amber-600 dark:text-amber-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-amber-900 dark:text-amber-100 text-sm font-medium m-0 mb-1">
                Installation Timeout
              </p>
              <p className="text-amber-800 dark:text-amber-200 text-sm m-0 leading-relaxed">
                Package installation may take a few minutes depending on the package size. The CLI
                waits up to 5 minutes for installation to complete.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Managing Packages */}
      <section id="managing" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Managing Packages
          </h2>
        </div>

        <div className="space-y-6">
          {/* List Packages */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <List className="text-primary h-5 w-5" />
              List Installed Packages
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              View all packages configured for your workspace and their installation status:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
              <pre className="m-0 leading-relaxed">kl pkg list</pre>
              <pre className="m-0 leading-relaxed">kl p ls        # Using aliases</pre>
            </div>
            <div className="mt-4">
              <p className="text-muted-foreground text-xs font-medium mb-2">Output includes:</p>
              <ul className="text-muted-foreground text-xs space-y-1 list-disc list-inside">
                <li>Workspace spec packages with channel/commit information</li>
                <li>Installed packages with version and binary path</li>
                <li>Failed packages with error messages (if any)</li>
              </ul>
            </div>
          </div>

          {/* Uninstall Packages */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Trash2 className="text-red-500 h-5 w-5" />
              Uninstall Packages
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Remove packages from your workspace configuration:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
              <pre className="m-0 leading-relaxed">kl pkg uninstall nodejs</pre>
              <pre className="m-0 leading-relaxed">kl pkg remove python</pre>
              <pre className="m-0 leading-relaxed">kl p rm vim              # Using aliases</pre>
            </div>
            <div className="mt-4 bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded border p-3">
              <div className="flex gap-2">
                <Info className="text-blue-600 dark:text-blue-400 h-4 w-4 flex-shrink-0 mt-0.5" />
                <p className="text-blue-800 dark:text-blue-200 text-xs m-0 leading-relaxed">
                  Removing a package only removes it from your workspace&apos;s configuration. The
                  package remains at the host level for other workspaces that may use it.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Best Practices */}
      <section id="best-practices" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <CheckCircle2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Best Practices
          </h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <CheckCircle2 className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Pin Versions for Reproducibility
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Use <code className="bg-muted rounded px-1 py-0.5 font-mono">kl pkg install</code> with{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono">--version</code> or{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono">--commit</code> to ensure
                  consistent package versions across team members and environments.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Package className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Keep Package List Minimal
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Only install packages you actively need. This keeps your workspace configuration
                  clean and reduces startup time.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <List className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Use Interactive Mode for Discovery
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Use <code className="bg-muted rounded px-1 py-0.5 font-mono">kl pkg add</code> without
                  arguments to explore available versions and make informed decisions about which
                  version to install.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Layers className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Leverage Host-Level Caching
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Common packages are shared across workspaces at the host level. Using the same
                  package versions across workspaces maximizes this efficiency.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Search className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Search Before Installing
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Use <code className="bg-muted rounded px-1 py-0.5 font-mono">kl pkg search</code> to
                  verify package names and check available versions before installation.
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
                Reproducible Development Environment
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                By using Nix package manager, your workspace environment is fully reproducible.
                Package installations are deterministic and can be shared with your team through
                workspace configuration.
              </p>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
