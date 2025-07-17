import { DocsPage } from '@/components/docs/docs-page'
import { 
  NumberedSection, 
  InfoBox, 
  BulletList, 
  FeatureCard, 
  ResourceList 
} from '@/components/docs/ui'
import { BookOpen, MessageSquare, Bug } from 'lucide-react'

export default function GettingStartedPage() {
  return (
    <DocsPage
      title="Getting Started with Kloudlite"
      description="Learn how to set up and start using Kloudlite platform for your development workflow."
      lastUpdated="January 15, 2025"
      editUrl="https://github.com/kloudlite/kloudlite/edit/main/web/src/app/docs/getting-started/page.tsx"
      difficulty="beginner"
      estimatedTime={10}
      tags={['setup', 'quickstart', 'beginner']}
    >
      <div className="space-y-12">
        <section>
          <h2>Welcome to Kloudlite</h2>
          <p>
            Kloudlite is a comprehensive cloud-native development platform that reduces development loop time by providing fast, reliable, and consistent environments for developers.
          </p>
          
          <InfoBox variant="primary" title="What you'll learn">
            <BulletList 
              variant="compact"
              bulletColor="primary"
              items={[
                "How to create your first Kloudlite project",
                "Setting up development environments",
                "Using service intercepts for local development",
                "Deploying your applications to the cloud"
              ]}
            />
          </InfoBox>
        </section>

        <section>
          <h2>Prerequisites</h2>
          <p>Before getting started, make sure you have:</p>
          <BulletList 
            items={[
              "A modern web browser (Chrome, Firefox, Safari, or Edge)",
              "Git installed on your machine",
              "Node.js 18+ or Docker installed",
              "A GitHub account (for authentication)"
            ]}
          />
        </section>

        <section>
          <h2>Quick Start</h2>
          <p>
            Get up and running with Kloudlite in just a few minutes:
          </p>
          
          <div className="space-y-8">
            <NumberedSection number={1} title="Sign up for Kloudlite">
              <p>
                Visit our platform and create your account using GitHub OAuth.
              </p>
              <pre>
                <code>https://kloudlite.io/signup</code>
              </pre>
            </NumberedSection>

            <NumberedSection number={2} title="Create your first project">
              <p>
                Use the dashboard to create a new project and configure your development environment.
              </p>
              <pre>
                <code>kloudlite project create my-first-project</code>
              </pre>
            </NumberedSection>

            <NumberedSection number={3} title="Start developing">
              <p>
                Clone your project and start coding with instant cloud environments.
              </p>
              <pre>
                <code>git clone https://github.com/yourusername/my-first-project.git</code>
              </pre>
            </NumberedSection>
          </div>
        </section>

        <section>
          <h2>Next Steps</h2>
          <p>
            Now that you have Kloudlite set up, explore these key features:
          </p>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <FeatureCard 
              title="Development Environments"
              description="Learn how to create and manage cloud-based development environments."
              href="/docs/concepts/environments"
            />

            <FeatureCard 
              title="Service Intercepts"
              description="Discover how to intercept traffic from cloud services to your local machine."
              href="/docs/concepts/intercepts"
            />

            <FeatureCard 
              title="API Reference"
              description="Explore the complete API documentation for programmatic access."
              href="/docs/api"
            />

            <FeatureCard 
              title="Deployment"
              description="Learn how to deploy your applications to production environments."
              href="/docs/deployment"
            />
          </div>
        </section>

        <section>
          <h2>Need Help?</h2>
          <p>
            If you encounter any issues or have questions, here are some resources:
          </p>
          
          <InfoBox variant="default">
            <ResourceList 
              items={[
                {
                  icon: BookOpen,
                  label: "Browse the full documentation",
                  href: "/docs"
                },
                {
                  icon: MessageSquare,
                  label: "Join our Discord community",
                  href: "https://discord.gg/kloudlite"
                },
                {
                  icon: Bug,
                  label: "Report issues on GitHub",
                  href: "https://github.com/kloudlite/kloudlite/issues"
                }
              ]}
            />
          </InfoBox>
        </section>
      </div>
    </DocsPage>
  )
}