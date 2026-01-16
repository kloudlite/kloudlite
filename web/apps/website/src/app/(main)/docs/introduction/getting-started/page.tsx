import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight } from 'lucide-react'
import {
  CreateUserPreview,
  LoginPreview,
  WorkmachinePreview,
  EnvironmentPreview,
  WorkspacePreview,
  ConnectEnvironmentPreview,
  StartCodingPreview,
} from './_components/step-previews'

const tocItems = [
  { id: 'steps', title: 'Steps' },
]

export default function GettingStartedPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Getting Started
      </h1>

      <section id="steps" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Set up your first environment and workspace.
        </p>

        <div className="space-y-10">
          {/* Step 1 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              1
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Create a user</p>
              <p className="text-muted-foreground text-sm mb-4">
                Login with super admin access and create a new user from the admin panel.
              </p>
              <CreateUserPreview />
            </div>
          </div>

          {/* Step 2 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              2
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Login as the user</p>
              <p className="text-muted-foreground text-sm mb-4">
                Sign out and login with the newly created user credentials.
              </p>
              <LoginPreview />
            </div>
          </div>

          {/* Step 3 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              3
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Setup a Workmachine</p>
              <p className="text-muted-foreground text-sm mb-4">
                Go to Workmachines → Create Workmachine. Select a machine type for your workspaces.
              </p>
              <WorkmachinePreview />
            </div>
          </div>

          {/* Step 4 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              4
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Create an Environment</p>
              <p className="text-muted-foreground text-sm mb-4">
                Go to Environments → Create Environment. Add your services using Docker Compose.
              </p>
              <EnvironmentPreview />
            </div>
          </div>

          {/* Step 5 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              5
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Create a Workspace</p>
              <p className="text-muted-foreground text-sm mb-4">
                Go to Workspaces → Create Workspace. Select your packages and workmachine.
              </p>
              <WorkspacePreview />
            </div>
          </div>

          {/* Step 6 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              6
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Connect to Environment</p>
              <p className="text-muted-foreground text-sm mb-4">
                From your workspace, run{' '}
                <code className="bg-muted px-1.5 py-0.5 font-mono text-xs">kl env connect</code>{' '}
                to access environment services.
              </p>
              <ConnectEnvironmentPreview />
            </div>
          </div>

          {/* Step 7 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-none text-sm font-bold">
              7
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Start coding</p>
              <p className="text-muted-foreground text-sm mb-4">
                Access via VS Code Web, SSH, or terminal. Your services are available by name.
              </p>
              <StartCodingPreview />
            </div>
          </div>
        </div>
      </section>

      <div className="border-t pt-8 space-y-4">
        <Link
          href="/docs/concepts/environments"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Environments</p>
            <p className="text-muted-foreground text-sm m-0">Learn more about environments and services</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>

        <Link
          href="/docs/concepts/workspaces"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Workspaces</p>
            <p className="text-muted-foreground text-sm m-0">Learn more about workspaces and packages</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
