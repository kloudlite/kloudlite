import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { Route, Terminal } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'starting', title: 'Starting Intercepts' },
  { id: 'managing', title: 'Managing Intercepts' },
]

export default function ServiceInterceptsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Service Intercepts</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Intercept a service to <strong className="text-foreground">route its traffic to your workspace</strong>.
          Requests that would go to the service in the environment are redirected to your local code instead.
        </p>

        <CommandBlock
          icon={Route}
          title="How it works"
          description="Redirect service traffic to your workspace for real-time debugging with production requests."
        >
          <div className="space-y-3 text-muted-foreground text-sm">
            <p className="m-0">
              <strong className="text-foreground">1.</strong> Start an intercept for a service (e.g., <code className="bg-muted px-1.5 py-0.5 font-mono text-xs border border-foreground/10 rounded-sm">api</code>)
            </p>
            <p className="m-0">
              <strong className="text-foreground">2.</strong> All traffic to that service is redirected to your workspace
            </p>
            <p className="m-0">
              <strong className="text-foreground">3.</strong> Debug with real requests, then stop the intercept to restore normal flow
            </p>
          </div>
        </CommandBlock>

        <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4 mt-6">
          <p className="text-foreground text-sm font-semibold mb-2">Prerequisite</p>
          <p className="text-muted-foreground text-sm m-0">
            You must first connect to an environment with{' '}
            <code className="bg-muted px-1.5 py-0.5 font-mono text-xs border border-foreground/10 rounded-sm">kl env connect</code> before intercepting services.
          </p>
        </div>
      </section>

      <section id="starting" className="mb-12">
        <SectionTitle id="starting">Starting Intercepts</SectionTitle>

        <div className="space-y-6">
          <CommandBlock
            icon={Terminal}
            title="kl intercept start"
            description="Start intercepting a service. You'll be prompted to configure port mapping."
          >
            <CodeExample>
              <CodeLine>kl intercept start           # Interactive service selection</CodeLine>
              <CodeLine>kl intercept start api       # Intercept specific service</CodeLine>
            </CodeExample>
          </CommandBlock>

          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <h3 className="text-foreground text-base font-semibold mb-3">Port mapping</h3>
            <p className="text-muted-foreground text-sm leading-relaxed mb-4">
              Map the service port to your workspace port where your local server runs.
            </p>
            <CodeExample>
              <CodeLine># Service port 8080 → Workspace port 3000</CodeLine>
              <CodeLine># Traffic to api:8080 now reaches localhost:3000</CodeLine>
            </CodeExample>
          </div>
        </div>
      </section>

      <section id="managing" className="mb-16">
        <SectionTitle id="managing">Managing Intercepts</SectionTitle>

        <div className="space-y-6">
          <CommandBlock
            icon={Terminal}
            title="List active intercepts"
            description="Shows service name, phase (Active/Pending/Failed), and port mappings."
          >
            <CodeExample>
              <CodeLine>kl intercept list</CodeLine>
            </CodeExample>
          </CommandBlock>

          <CommandBlock
            icon={Terminal}
            title="Check status"
            description="Detailed info including workspace pod, port mappings, and start time."
          >
            <CodeExample>
              <CodeLine>kl intercept status api</CodeLine>
            </CodeExample>
          </CommandBlock>

          <CommandBlock
            icon={Terminal}
            title="Stop intercept"
            description="Traffic routes back to the original service within seconds."
          >
            <CodeExample>
              <CodeLine>kl intercept stop            # Interactive selection</CodeLine>
              <CodeLine>kl intercept stop api        # Stop specific service</CodeLine>
            </CodeExample>
          </CommandBlock>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/environment-connection"
          title="Environment Connection"
          description="Connect to environments first"
        />
        <NextLinkCard
          href="/docs/workspace-internals/cli"
          title="CLI Reference"
          description="Full kl command reference"
        />
      </div>
    </DocsContentLayout>
  )
}
