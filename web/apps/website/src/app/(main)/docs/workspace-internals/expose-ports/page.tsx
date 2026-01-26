import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { Globe, Terminal } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'exposing', title: 'Exposing Ports' },
  { id: 'managing', title: 'Managing Exposed Ports' },
]

export default function ExposePortsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Expose Ports</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          Expose workspace ports to get <strong className="text-foreground">public URLs</strong> for
          your HTTP services. Share your local development server with teammates or test webhooks
          without deploying.
        </p>

        <CommandBlock
          icon={Globe}
          title="URL format"
          description="Exposed ports get unique URLs with your workspace identifier."
        >
          <CodeExample>
            <CodeLine>https://p3000-abc123.subdomain.khost.dev</CodeLine>
          </CodeExample>
          <p className="text-muted-foreground text-sm mt-4 m-0">
            The URL includes the port number and a unique hash for your workspace.
          </p>
        </CommandBlock>
      </section>

      <section id="exposing" className="mb-12">
        <SectionTitle id="exposing">Exposing Ports</SectionTitle>

        <CommandBlock
          icon={Terminal}
          title="kl expose"
          description="Expose a port to get a public URL. Works with any HTTP service running in your workspace."
        >
          <CodeExample>
            <CodeLine>kl expose 3000               # Expose port 3000</CodeLine>
            <CodeLine>kl expose 8080               # Expose port 8080</CodeLine>
          </CodeExample>
        </CommandBlock>

        <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4 mt-6">
          <p className="text-foreground text-sm font-semibold mb-2">HTTP only</p>
          <p className="text-muted-foreground text-sm m-0">
            Exposed ports are designed for HTTP/HTTPS services. Other protocols are not supported.
          </p>
        </div>
      </section>

      <section id="managing" className="mb-16">
        <SectionTitle id="managing">Managing Exposed Ports</SectionTitle>

        <div className="grid gap-6 md:grid-cols-2">
          <CommandBlock
            icon={Terminal}
            title="List exposed ports"
            description="Shows all exposed ports with their public URLs."
          >
            <CodeExample>
              <CodeLine>kl expose list</CodeLine>
            </CodeExample>
          </CommandBlock>

          <CommandBlock
            icon={Terminal}
            title="Remove exposed port"
            description="Removes the public URL for the specified port."
          >
            <CodeExample>
              <CodeLine>kl expose remove 3000</CodeLine>
            </CodeExample>
          </CommandBlock>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/forking-sharing"
          title="Forking, Cloning & Sharing"
          description="Share workspaces with your team"
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
