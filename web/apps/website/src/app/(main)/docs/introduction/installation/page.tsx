import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { Step } from '@/components/docs/step'
import { NextLinkCard } from '@/components/docs/next-link-card'
import {
  SignUpPreview,
  CreateInstallationPreview,
  CloudProviderPreview,
  InstallCommandPreview,
  InstallationCompletePreview,
} from './_components/step-previews'

const tocItems = [
  { id: 'steps', title: 'Installation Steps' },
]

export default function InstallationPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Installation - BYOC</PageTitle>

      <section id="steps" className="mb-16">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Deploy Kloudlite on your own cloud infrastructure with Bring Your Own Compute (BYOC).
        </p>

        <div className="space-y-10">
          <Step
            number={1}
            title="Register your account"
            description="Go to console.kloudlite.io and sign up using GitHub, Google, or Microsoft."
          >
            <SignUpPreview />
          </Step>

          <Step
            number={2}
            title="Create your installation"
            description='Click "New Installation" and provide a name, description, and subdomain for your installation.'
          >
            <CreateInstallationPreview />
          </Step>

          <Step
            number={3}
            title="Select your cloud provider"
            description="Choose AWS, GCP, or Azure and select the region where you want to deploy."
          >
            <CloudProviderPreview />
          </Step>

          <Step
            number={4}
            title="Run the installation command"
            description="Copy the generated command and run it in your terminal. Make sure you have the required CLI tools installed."
          >
            <InstallCommandPreview />
          </Step>

          <Step
            number={5}
            title="Access your installation"
            description="Once the installation is verified, you can access your Kloudlite dashboard and start setting up your development environment."
          >
            <InstallationCompletePreview />
          </Step>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/introduction/getting-started"
          title="Getting Started"
          description="Create your first environment and workspace"
        />
      </div>
    </DocsContentLayout>
  )
}
