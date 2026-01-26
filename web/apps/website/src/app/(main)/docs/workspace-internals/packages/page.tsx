import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Package, Terminal } from 'lucide-react'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'commands', title: 'Commands' },
]

export default function PackageManagementPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Package Management</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          Install development tools using <strong className="text-foreground">Nix packages</strong>.
          Packages are installed at the host level and shared across workspaces.
        </p>

        <CommandBlock
          icon={Package}
          title="Quick start"
          description="Get started with package management."
        >
          <CodeExample>
            <CodeLine>kl pkg add                   # Interactive mode</CodeLine>
            <CodeLine>kl pkg add nodejs python     # Add multiple packages</CodeLine>
            <CodeLine>kl pkg list                  # List installed</CodeLine>
          </CodeExample>
        </CommandBlock>
      </section>

      <section id="commands" className="mb-16">
        <SectionTitle id="commands">Commands</SectionTitle>

        <div className="space-y-6">
          <CommandBlock
            icon={Terminal}
            title="Search packages"
            description="Search the Nix package registry for available packages."
          >
            <CodeExample>
              <CodeLine>kl pkg search nodejs</CodeLine>
            </CodeExample>
          </CommandBlock>

          <CommandBlock
            icon={Package}
            title="Add packages"
            description="Add packages to your workspace."
          >
            <CodeExample>
              <CodeLine>kl pkg add                   # Interactive with fuzzy search</CodeLine>
              <CodeLine>kl pkg add git vim curl      # Add by name</CodeLine>
              <CodeLine>kl pkg add nodejs --exact    # Pin to exact version</CodeLine>
            </CodeExample>
          </CommandBlock>

          <CommandBlock
            icon={Package}
            title="Install with version control"
            description="Specify exact versions, channels, or commits for reproducibility."
          >
            <CodeExample>
              <CodeLine>kl pkg install nodejs --version 20.0.0</CodeLine>
              <CodeLine>kl pkg install python --channel nixos-24.05</CodeLine>
              <CodeLine>kl pkg install curl --commit abc123def</CodeLine>
            </CodeExample>
          </CommandBlock>

          <div className="grid gap-6 md:grid-cols-2">
            <CommandBlock
              icon={Terminal}
              title="List installed"
              description="Shows packages with their source and status."
            >
              <CodeExample>
                <CodeLine>kl pkg list</CodeLine>
              </CodeExample>
            </CommandBlock>

            <CommandBlock
              icon={Terminal}
              title="Remove package"
              description="Removes from your workspace config only."
            >
              <CodeExample>
                <CodeLine>kl pkg uninstall vim</CodeLine>
              </CodeExample>
            </CommandBlock>
          </div>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/cli"
          title="CLI Reference"
          description="Full kl command reference"
        />
        <NextLinkCard
          href="/docs/workspace-internals/environment-connection"
          title="Environment Connection"
          description="Connect to environments"
        />
      </div>
    </DocsContentLayout>
  )
}
