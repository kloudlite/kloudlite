import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { Cloud, Server, Settings, Activity, AlertTriangle, Lock, UsersRound, Network, Boxes, FolderTree, UserCog, MapPin } from 'lucide-react'

export default function InstallationPage() {
  return (
    <div className="prose prose-slate dark:prose-invert mx-auto max-w-3xl px-4 pt-8 pb-16 sm:px-6 lg:px-8 xl:pr-16">
      {/* Header */}
      <div className="mb-12 sm:mb-16">
        <h1 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl break-words leading-tight sm:leading-tight">
          Installation
        </h1>
        <p className="text-muted-foreground mt-4 text-base sm:text-lg lg:text-xl leading-relaxed">
          Learn what a Kloudlite installation is and how to create and manage your installation
          on console.kloudlite.io
        </p>
      </div>

      {/* What is an Installation */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">What is an Installation?</h2>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-4 sm:mb-6">
          <p className="text-card-foreground mb-4 text-lg leading-relaxed">
            An <strong>Installation</strong> is your team&apos;s complete Kloudlite deployment -
            a dedicated platform instance running on its own domain at{' '}
            <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">{'{subdomain}'}.khost.dev</code>
          </p>
        </div>

        <h3 className="text-foreground mb-4 text-xl font-semibold leading-snug">Key Features</h3>
        <div className="mb-4 sm:mb-6 grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Lock className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Isolated Resources
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  No data is shared with Kloudlite. Everything is installed and runs from your installation - either in your cloud (BYOC) or on isolated dedicated infrastructure (Cloud mode)
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Settings className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Administration
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  Team management, OAuth configurations, and machine type settings are managed within your installation at <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code>
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <UsersRound className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Team Collaboration
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  Members discover and interact with environments and workspaces within the scope of the installation
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <MapPin className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Proximity to Team
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  Your installation runs in the cloud region closest to your team. By deploying in your preferred geographic location (AWS, GCP, or Azure regions), you get low latency access to workspaces and environments—ensuring fast, responsive development experiences
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-muted/50 border-muted-foreground/20 rounded-lg border p-3 sm:p-4">
          <p className="text-card-foreground mb-0 text-sm leading-relaxed">
            <strong>Note:</strong> Installation is registered in <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">console.kloudlite.io</code> for subdomain allocation and root user access management only.
          </p>
        </div>

        <h3 className="text-foreground mt-6 sm:mt-8 mb-4 text-xl font-semibold leading-snug">How it Works</h3>
        <div className="mb-4 sm:mb-6 grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Network className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Management Node
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  One dedicated node manages your entire installation - handling team access, environments, workspaces, orchestrating all workmachines, and responsible for backups and recovery
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Boxes className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Workmachines for Users
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  Additional VM instances are provisioned as workmachines where team members run their workspaces and development environments
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <FolderTree className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Centralized Management
                </h4>
                <p className="text-muted-foreground leading-relaxed">
                  All team environments, workspaces, backups, and regional resources are managed centrally through the management node
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Creating an Installation */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">Creating an Installation</h2>
        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h3 className="text-card-foreground mb-4 text-xl font-semibold leading-snug">
            Follow these steps to create your installation
          </h3>
          <ol className="space-y-4">
            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                1
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium leading-snug">Access the Console</p>
                <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                  Navigate to{' '}
                  <a
                    href="https://console.kloudlite.io"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    console.kloudlite.io
                  </a>{' '}
                  and sign in with your account. If you don&apos;t have an account yet, click &quot;Sign Up&quot;
                  to create one.
                </p>
              </div>
            </li>

            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                2
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium leading-snug">Create New Installation</p>
                <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                  Once logged in, you&apos;ll see the installations dashboard. Click on the{' '}
                  <strong>&quot;Create Installation&quot;</strong> button to begin the setup process.
                </p>
              </div>
            </li>

            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                3
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium leading-snug">Choose Installation Type</p>
                <div className="mt-3 space-y-4">
                  <div className="bg-muted rounded-lg border p-3 sm:p-4">
                    <div className="mb-2 flex items-center gap-2">
                      <Cloud className="text-primary h-6 w-6" />
                      <h4 className="text-card-foreground text-sm font-semibold leading-snug">
                        Kloudlite Cloud
                      </h4>
                      <span className="bg-warning/20 text-warning rounded px-2 py-0.5 text-xs font-medium">
                        Coming Soon
                      </span>
                    </div>
                    <p className="text-muted-foreground text-xs leading-relaxed">
                      Quick setup with infrastructure managed by Kloudlite. Perfect for getting
                      started quickly.
                    </p>
                  </div>
                  <div className="bg-muted rounded-lg border p-3 sm:p-4">
                    <div className="mb-2 flex items-center gap-2">
                      <Server className="text-primary h-6 w-6" />
                      <h4 className="text-card-foreground text-sm font-semibold leading-snug">BYOC (Bring Your Own Cloud)</h4>
                    </div>
                    <p className="text-muted-foreground text-xs leading-relaxed">
                      Kloudlite creates and manages the cluster on your cloud provider (AWS, Azure, GCP) for complete control over infrastructure and data locality.
                    </p>
                  </div>
                </div>
              </div>
            </li>

            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                4
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium leading-snug">Run Installation (BYOC Only)</p>
                <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                  For BYOC mode, follow the detailed instructions provided for your specific cloud provider (AWS, Azure, or GCP) to run the installation script and complete the setup.
                </p>
              </div>
            </li>

            <li className="flex items-start gap-3">
              <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                5
              </span>
              <div className="flex-1">
                <p className="text-card-foreground font-medium leading-snug">Reserve Domain Name</p>
                <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                  After the installation completes, reserve your domain name to make your installation accessible at <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code>
                </p>
              </div>
            </li>
          </ol>
        </div>
      </section>

      {/* Managing Installations */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">Managing Installations</h2>
        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Activity className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Installation Details
                </h3>
                <p className="text-muted-foreground mb-3 leading-relaxed">
                  From <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">console.kloudlite.io</code>, you can view details about your installations:
                </p>
                <ul className="text-muted-foreground space-y-1 text-sm leading-relaxed">
                  <li>• Installation URL (<code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code>)</li>
                  <li>• Current status and health</li>
                  <li>• Cloud provider and region</li>
                  <li>• Installation owner information</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <Settings className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Super User Access
                </h3>
                <p className="text-muted-foreground mb-3 leading-relaxed">
                  The installation owner (root user) has full administrative access to the installation:
                </p>
                <ul className="text-muted-foreground space-y-1 text-sm leading-relaxed">
                  <li>• Access installation admin panel at <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code></li>
                  <li>• Manage all team members, environments, and workspaces</li>
                  <li>• Configure OAuth providers and authentication settings</li>
                  <li>• Manage machine types and resource allocations</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-primary flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
                <UserCog className="text-primary-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Transfer Ownership
                </h3>
                <p className="text-muted-foreground mb-3 leading-relaxed">
                  The current installation owner can transfer ownership to another user through <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">console.kloudlite.io</code>:
                </p>
                <ul className="text-muted-foreground space-y-1 text-sm leading-relaxed">
                  <li>• Navigate to the installation card in the console</li>
                  <li>• Click on &quot;Transfer Ownership&quot;</li>
                  <li>• Enter the email address of the new owner</li>
                  <li>• Confirm the transfer</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="mb-4 flex items-start gap-4">
              <div className="bg-destructive flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg">
              <AlertTriangle className="text-destructive-foreground h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">
                  Uninstallation
                </h3>
                <div className="bg-destructive/10 border-destructive/50 mb-3 rounded-lg border p-3">
                  <p className="text-card-foreground mb-1 text-sm font-semibold leading-snug">
                    Warning: This action is irreversible
                  </p>
                  <p className="text-muted-foreground mb-0 text-xs leading-relaxed">
                    Uninstalling will permanently remove all associated workspaces, environments, and data.
                  </p>
                </div>
                <p className="text-muted-foreground text-sm leading-relaxed">
                  Follow the uninstallation process specific to your cloud provider to clean up all resources.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-muted/50 border-muted-foreground/20 mt-6 rounded-lg border p-3 sm:p-4">
          <p className="text-card-foreground mb-0 text-sm leading-relaxed">
            <strong>Note:</strong> The new owner will receive root user access and full administrative privileges. Ownership transfers cannot be undone unless the new owner transfers it back.
          </p>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">Next Steps</h2>
        <div className="space-y-3 sm:space-y-4">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <Link href="/docs">
              <h3 className="text-card-foreground mb-3 text-lg font-semibold hover:text-primary cursor-pointer transition-colors leading-snug">
                Access Installation Dashboard
              </h3>
            </Link>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Navigate to your installation at <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">{'{subdomain}'}.khost.dev</code> and sign in with your root user credentials to access the admin dashboard.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <Link href="/docs">
              <h3 className="text-card-foreground mb-3 text-lg font-semibold hover:text-primary cursor-pointer transition-colors leading-snug">
                Create Your First Environment
              </h3>
            </Link>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Set up isolated environments for development, staging, and production to organize your team&apos;s work.
            </p>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <Link href="/docs">
              <h3 className="text-card-foreground mb-3 text-lg font-semibold hover:text-primary cursor-pointer transition-colors leading-snug">
                Create Your First Workspace
              </h3>
            </Link>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Launch a development workspace with your preferred tools and start coding immediately.
            </p>
          </div>
        </div>
      </section>

      {/* Common Issues */}
      <section className="mb-12 sm:mb-16">
        <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">Common Issues</h2>
        <div className="bg-card space-y-3 sm:space-y-4 rounded-lg border p-4 sm:p-6">
          <div>
            <h3 className="text-card-foreground mb-2 text-base font-semibold leading-snug">
              Installation Creation Failed
            </h3>
            <p className="text-muted-foreground mb-2 text-sm leading-relaxed">If your installation fails to create, check:</p>
            <ul className="text-muted-foreground space-y-1 text-sm leading-relaxed">
              <li>• Your account has sufficient credits (for Cloud installations)</li>
              <li>• The installation name is unique and follows naming conventions</li>
              <li>• For BYOC: Your cloud provider credentials are valid and have necessary permissions</li>
              <li>• For BYOC: Your cloud account has sufficient quota for cluster resources</li>
            </ul>
          </div>

          <div className="border-t pt-4">
            <h3 className="text-card-foreground mb-2 text-base font-semibold leading-snug">
              Cannot Access Installation
            </h3>
            <p className="text-muted-foreground mb-2 text-sm leading-relaxed">
              If you&apos;re having trouble accessing your installation:
            </p>
            <ul className="text-muted-foreground space-y-1 text-sm leading-relaxed">
              <li>• Verify your account has the correct permissions</li>
              <li>• Check that the installation status is &quot;Active&quot; in the console</li>
              <li>• Clear your browser cache and cookies</li>
              <li>• Try accessing from a different browser or incognito mode</li>
            </ul>
          </div>

          <div className="border-t pt-4">
            <h3 className="text-card-foreground mb-2 text-base font-semibold">Need More Help?</h3>
            <div className="text-muted-foreground space-y-1 text-sm">
              <p>• Check our <Link href="/docs/faq" className="text-primary hover:underline">FAQ page</Link> for common questions</p>
              <p>• <Link href="/contact" className="text-primary hover:underline">Contact our support team</Link> for personalized assistance</p>
              <p>• Visit our <a href="https://github.com/kloudlite/kloudlite/issues" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">GitHub repository</a> to report bugs</p>
            </div>
          </div>
        </div>
      </section>

      {/* Ready to get started */}
      <section className="mb-12 sm:mb-16">
        <div className="bg-card rounded-lg border p-6 sm:p-8 text-center">
          <h2 className="text-card-foreground mb-4 text-2xl font-bold leading-tight">Ready to get started?</h2>
          <p className="text-muted-foreground mb-6 leading-relaxed">
            Head over to the console and create your first installation today.
          </p>
          <Button asChild size="lg">
            <a href="https://console.kloudlite.io" target="_blank" rel="noopener noreferrer">
              Open Console
            </a>
          </Button>
        </div>
      </section>
    </div>
  )
}
