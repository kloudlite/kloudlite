import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Network, Terminal, Database } from 'lucide-react'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Environment Connection - Kloudlite Documentation',
  description: 'Connect your workspace to an environment to access services by name',
}

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'connecting', title: 'Connecting' },
  { id: 'accessing-services', title: 'Accessing Services' },
]

export default function EnvironmentConnectionPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Environment Connection</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          Connect your workspace to an environment to{' '}
          <strong className="text-foreground">access services by name</strong>. Once connected,
          services like <code className="bg-muted px-1.5 py-0.5 font-mono text-sm border border-foreground/10 rounded-sm">postgres</code> or{' '}
          <code className="bg-muted px-1.5 py-0.5 font-mono text-sm border border-foreground/10 rounded-sm">redis</code> become directly
          reachable from your workspace.
        </p>

        <CommandBlock
          icon={Network}
          title="How it works"
          description="When you connect, your workspace's DNS is configured to resolve environment service names. No port forwarding or IP addresses needed."
        >
          <div></div>
        </CommandBlock>
      </section>

      <section id="connecting" className="mb-12">
        <SectionTitle id="connecting">Connecting</SectionTitle>

        <div className="space-y-6">
          <CommandBlock
            icon={Terminal}
            title="kl env connect"
            description="Connect to an environment interactively or by name."
          >
            <CodeExample>
              <CodeLine>kl env connect               # Interactive selection</CodeLine>
              <CodeLine>kl env connect my-env        # Connect to specific environment</CodeLine>
            </CodeExample>
          </CommandBlock>

          <div className="grid gap-6 md:grid-cols-2">
            <CommandBlock
              icon={Terminal}
              title="Disconnect"
              description="Removes active intercepts and clears DNS config."
            >
              <CodeExample>
                <CodeLine>kl env disconnect</CodeLine>
              </CodeExample>
            </CommandBlock>

            <CommandBlock
              icon={Terminal}
              title="Check status"
              description="Shows connected environment and available services."
            >
              <CodeExample>
                <CodeLine>kl env status</CodeLine>
              </CodeExample>
            </CommandBlock>
          </div>
        </div>
      </section>

      <section id="accessing-services" className="mb-16">
        <SectionTitle id="accessing-services">Accessing Services</SectionTitle>

        <p className="text-muted-foreground mb-6">
          After connecting, use <strong className="text-foreground">service names as hostnames</strong>.
          Always include the port number.
        </p>

        <CommandBlock
          icon={Database}
          title="Examples"
          description="Use service names in your application configuration."
        >
          <CodeExample title="From terminal">
            <CodeLine>psql -h postgres -p 5432 -U myuser</CodeLine>
            <CodeLine>redis-cli -h redis -p 6379</CodeLine>
            <CodeLine>curl http://api:8080/health</CodeLine>
          </CodeExample>

          <CodeExample title="In application config">
            <CodeLine>{`DATABASE_URL=postgresql://user:pass@postgres:5432/myapp`}</CodeLine>
            <CodeLine>REDIS_URL=redis://redis:6379</CodeLine>
            <CodeLine>API_ENDPOINT=http://api:8080</CodeLine>
          </CodeExample>

          <div className="mt-5 bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4">
            <p className="text-foreground font-semibold text-sm mb-2">Available service names</p>
            <p className="text-muted-foreground text-sm m-0">
              Run <code className="bg-muted px-1.5 py-0.5 font-mono text-xs border border-foreground/10 rounded-sm">kl env status</code> to see
              all services and their ports in the connected environment.
            </p>
          </div>
        </CommandBlock>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/intercepts"
          title="Service Intercepts"
          description="Route service traffic to your workspace"
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
