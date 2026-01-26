import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { Step } from '@/components/docs/step'
import { NextLinkCard } from '@/components/docs/next-link-card'
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
      <PageTitle>Getting Started</PageTitle>

      <section id="steps" className="mb-16">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Set up your first environment and workspace.
        </p>

        <div className="space-y-10">
          <Step
            number={1}
            title="Create a user"
            description="Login with super admin access and create a new user from the admin panel."
          >
            <CreateUserPreview />
          </Step>

          <Step
            number={2}
            title="Login as the user"
            description="Sign out and login with the newly created user credentials."
          >
            <LoginPreview />
          </Step>

          <Step
            number={3}
            title="Setup a Workmachine"
            description="Go to Workmachines → Create Workmachine. Select a machine type for your workspaces."
          >
            <WorkmachinePreview />
          </Step>

          <Step
            number={4}
            title="Create an Environment"
            description="Go to Environments → Create Environment. Add your services using Docker Compose."
          >
            <EnvironmentPreview />
          </Step>

          <Step
            number={5}
            title="Create a Workspace"
            description="Go to Workspaces → Create Workspace. Select your packages and workmachine."
          >
            <WorkspacePreview />
          </Step>

          <Step
            number={6}
            title="Connect to Environment"
            description="From your workspace, run kl env connect to access environment services."
          >
            <ConnectEnvironmentPreview />
          </Step>

          <Step
            number={7}
            title="Start coding"
            description="Access via VS Code Web, SSH, or terminal. Your services are available by name."
          >
            <StartCodingPreview />
          </Step>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/concepts/environments"
          title="Environments"
          description="Learn more about environments and services"
        />
        <NextLinkCard
          href="/docs/concepts/workspaces"
          title="Workspaces"
          description="Learn more about workspaces and packages"
        />
      </div>
    </DocsContentLayout>
  )
}
