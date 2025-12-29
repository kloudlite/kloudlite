import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight } from 'lucide-react'
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
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Installation - BYOC
      </h1>

      <section id="steps" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Deploy Kloudlite on your own cloud infrastructure with Bring Your Own Compute (BYOC).
        </p>

        <div className="space-y-10">
          {/* Step 1 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
              1
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Register your account</p>
              <p className="text-muted-foreground text-sm mb-4">
                Go to{' '}
                <a
                  href="https://console.kloudlite.io"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  console.kloudlite.io
                </a>
                {' '}and sign up using GitHub, Google, or Microsoft.
              </p>
              <SignUpPreview />
            </div>
          </div>

          {/* Step 2 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
              2
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Create your installation</p>
              <p className="text-muted-foreground text-sm mb-4">
                Click "New Installation" and provide a name, description, and subdomain for your installation.
              </p>
              <CreateInstallationPreview />
            </div>
          </div>

          {/* Step 3 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
              3
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Select your cloud provider</p>
              <p className="text-muted-foreground text-sm mb-4">
                Choose AWS, GCP, or Azure and select the region where you want to deploy.
              </p>
              <CloudProviderPreview />
            </div>
          </div>

          {/* Step 4 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
              4
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Run the installation command</p>
              <p className="text-muted-foreground text-sm mb-4">
                Copy the generated command and run it in your terminal. Make sure you have the required CLI tools installed.
              </p>
              <InstallCommandPreview />
            </div>
          </div>

          {/* Step 5 */}
          <div className="flex items-start gap-4">
            <div className="bg-primary text-primary-foreground flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
              5
            </div>
            <div className="flex-1">
              <p className="text-foreground font-medium mb-1">Access your installation</p>
              <p className="text-muted-foreground text-sm mb-4">
                Once the installation is verified, you can access your Kloudlite dashboard and start setting up your development environment.
              </p>
              <InstallationCompletePreview />
            </div>
          </div>
        </div>
      </section>

      <div className="border-t pt-8">
        <Link
          href="/docs/introduction/getting-started"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Getting Started</p>
            <p className="text-muted-foreground text-sm m-0">Create your first environment and workspace</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
