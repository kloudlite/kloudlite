import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { Terminal, Package, Network, Route, Settings, Globe, Container, Sparkles } from 'lucide-react'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { FeatureCard } from '@/components/docs/feature-card'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'pkg', title: 'kl pkg' },
  { id: 'env', title: 'kl env' },
  { id: 'intercept', title: 'kl intercept' },
  { id: 'status', title: 'kl status' },
  { id: 'config', title: 'kl config' },
  { id: 'expose', title: 'kl expose' },
  { id: 'image', title: 'kl image' },
  { id: 'other', title: 'Other Commands' },
]

export default function CLIReferencePage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>CLI Reference</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground leading-relaxed mb-8">
          The <code className="bg-muted px-1.5 py-0.5 font-mono text-sm border border-foreground/10 rounded-sm">kl</code> CLI manages
          your workspace from within the dev container. Most commands support{' '}
          <strong className="text-foreground">interactive mode</strong> when run without arguments.
        </p>

        <FeatureCard
          icon={Terminal}
          title="Quick Start"
          description={
            <CodeExample>
              <CodeLine>kl status                    # Show workspace info</CodeLine>
              <CodeLine>kl pkg add nodejs python     # Install packages</CodeLine>
              <CodeLine>kl env connect               # Connect to environment (interactive)</CodeLine>
              <CodeLine>kl intercept start           # Intercept a service (interactive)</CodeLine>
              <CodeLine>kl expose 3000               # Expose port with public URL</CodeLine>
            </CodeExample>
          }
        />
      </section>

      <section id="pkg" className="mb-12">
        <CommandBlock
          icon={Package}
          title="kl pkg"
          alias="aliases: package, p"
          description="Manage Nix packages in your workspace."
        >
          <div className="space-y-5">
            <CodeExample title="Search packages">
              <CodeLine>kl pkg search &lt;query&gt;        # Search available packages</CodeLine>
              <CodeLine>kl p s nodejs                # Short form</CodeLine>
            </CodeExample>

            <CodeExample title="Add packages">
              <CodeLine>kl pkg add                   # Interactive mode with fuzzy search</CodeLine>
              <CodeLine>kl pkg add nodejs python     # Add multiple packages</CodeLine>
              <CodeLine>kl pkg add nodejs --exact    # Pin to exact version (slower)</CodeLine>
            </CodeExample>

            <CodeExample title="Install with version control">
              <CodeLine>kl pkg install nodejs --version 20.0.0</CodeLine>
              <CodeLine>kl pkg install python --channel nixos-24.05</CodeLine>
              <CodeLine>kl pkg install curl --commit abc123def</CodeLine>
            </CodeExample>

            <CodeExample title="Manage installed">
              <CodeLine>kl pkg list                  # List installed packages</CodeLine>
              <CodeLine>kl pkg uninstall vim         # Remove package</CodeLine>
            </CodeExample>
          </div>
        </CommandBlock>
      </section>

      <section id="env" className="mb-12">
        <CommandBlock
          icon={Network}
          title="kl env"
          alias="aliases: environment, e"
          description="Connect to environments and access services by name."
        >
          <CodeExample>
            <CodeLine>kl env connect               # Interactive environment selection</CodeLine>
            <CodeLine>kl env connect my-env        # Connect to specific environment</CodeLine>
            <CodeLine>kl env disconnect            # Disconnect (removes intercepts)</CodeLine>
            <CodeLine>kl env status                # Show connection status</CodeLine>
          </CodeExample>
        </CommandBlock>
      </section>

      <section id="intercept" className="mb-12">
        <CommandBlock
          icon={Route}
          title="kl intercept"
          alias="aliases: int, i"
          description="Route service traffic to your workspace for debugging."
        >
          <CodeExample>
            <CodeLine>kl intercept start           # Interactive service & port selection</CodeLine>
            <CodeLine>kl intercept start api       # Start intercepting specific service</CodeLine>
            <CodeLine>kl intercept stop api        # Stop intercepting</CodeLine>
            <CodeLine>kl intercept list            # List active intercepts</CodeLine>
            <CodeLine>kl intercept status          # Show intercept status</CodeLine>
          </CodeExample>
        </CommandBlock>
      </section>

      <section id="status" className="mb-12">
        <CommandBlock
          icon={Terminal}
          title="kl status"
          alias="aliases: st, s"
          description="Display workspace information including phase, resources, access URLs, and timing."
        >
          <CodeExample>
            <CodeLine>kl status                    # Show workspace info</CodeLine>
          </CodeExample>

          <div className="mt-5 grid sm:grid-cols-2 gap-4">
            <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4">
              <p className="text-foreground font-semibold text-sm mb-2">Shows:</p>
              <ul className="text-muted-foreground text-sm space-y-1.5 m-0 pl-4">
                <li>• Workspace name & owner</li>
                <li>• Current phase</li>
                <li>• Resource usage (CPU, Memory, Storage)</li>
              </ul>
            </div>
            <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4">
              <p className="text-foreground font-semibold text-sm mb-2">Also includes:</p>
              <ul className="text-muted-foreground text-sm space-y-1.5 m-0 pl-4">
                <li>• Access URLs (VS Code, SSH)</li>
                <li>• Start time & last activity</li>
                <li>• Active connections count</li>
              </ul>
            </div>
          </div>
        </CommandBlock>
      </section>

      <section id="config" className="mb-12">
        <CommandBlock
          icon={Settings}
          title="kl config"
          alias="aliases: cfg, c"
          description="View and update workspace configuration."
        >
          <CodeExample>
            <CodeLine>kl config get                # Show all configuration</CodeLine>
            <CodeLine>kl config get display-name   # Get specific value</CodeLine>
            <CodeLine>kl config set display-name &quot;My Workspace&quot;</CodeLine>
            <CodeLine>kl config set git.user-email &quot;dev@example.com&quot;</CodeLine>
          </CodeExample>

          <div className="mt-5 bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4">
            <p className="text-foreground font-semibold text-sm mb-3">Available keys:</p>
            <div className="grid sm:grid-cols-2 gap-x-6 gap-y-2 text-sm font-mono text-muted-foreground">
              <span>display-name</span>
              <span>git.user-name</span>
              <span>description</span>
              <span>git.user-email</span>
              <span>auto-stop</span>
              <span>git.default-branch</span>
              <span>idle-timeout</span>
              <span>env.&lt;VAR_NAME&gt;</span>
              <span>max-runtime</span>
              <span>vscode-version</span>
            </div>
          </div>
        </CommandBlock>
      </section>

      <section id="expose" className="mb-12">
        <CommandBlock
          icon={Globe}
          title="kl expose"
          description="Expose workspace ports with public URLs for HTTP services."
        >
          <CodeExample>
            <CodeLine>kl expose 3000               # Expose port 3000</CodeLine>
            <CodeLine>kl expose list               # List exposed ports</CodeLine>
            <CodeLine>kl expose remove 3000        # Remove exposed port</CodeLine>
          </CodeExample>

          <p className="text-muted-foreground text-sm mt-4">
            Generates URLs like{' '}
            <code className="bg-muted px-1.5 py-0.5 font-mono text-xs border border-foreground/10 rounded-sm">https://p3000-abc.subdomain.khost.dev</code>
          </p>
        </CommandBlock>
      </section>

      <section id="image" className="mb-12">
        <CommandBlock
          icon={Container}
          title="kl image"
          description="Build and push container images to Kloudlite registry."
        >
          <CodeExample>
            <CodeLine>kl image push myapp          # Build and push image</CodeLine>
            <CodeLine>kl image push -f Dockerfile.prod myapp</CodeLine>
            <CodeLine>kl image push --build-arg VERSION=1.0 myapp</CodeLine>
          </CodeExample>
        </CommandBlock>
      </section>

      <section id="other" className="mb-16">
        <SectionTitle id="other">Other Commands</SectionTitle>

        <div className="space-y-5">
          <div className="grid gap-5 md:grid-cols-2">
            <FeatureCard
              icon={Sparkles}
              title="kl mcp"
              description="Start MCP server for AI assistants. Exposes 29 tools for Claude and other AI tools."
            />
            <FeatureCard
              icon={Terminal}
              title="kl completion"
              description="Generate shell completions for bash, zsh, fish, or powershell."
            />
          </div>

          <FeatureCard
            icon={Terminal}
            title={
              <>
                kl version <span className="text-muted-foreground text-xs font-mono ml-2">alias: v</span>
              </>
            }
            description="Display CLI version information."
          />
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8 space-y-4">
        <NextLinkCard
          href="/docs/workspace-internals/packages"
          title="Packages"
          description="Learn more about package management"
        />
        <NextLinkCard
          href="/docs/workspace-internals/environment-connection"
          title="Environment Connection"
          description="Learn more about connecting to environments"
        />
        <NextLinkCard
          href="/docs/workspace-internals/intercepts"
          title="Service Intercepts"
          description="Learn more about service interception"
        />
      </div>
    </DocsContentLayout>
  )
}
