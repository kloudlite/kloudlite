import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { CommandBlock, CodeExample, CodeLine } from '@/components/docs/command-block'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { Key, Github } from 'lucide-react'
import { VPNConnectionPreview, AuthorizedKeysPreview, SSHPublicKeyPreview } from './_components/access-previews'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'vpn', title: 'VPN Connection' },
  { id: 'authorized-keys', title: 'Authorized Keys' },
  { id: 'ssh-public-key', title: 'SSH Public Key' },
]

export default function AccessPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Access</PageTitle>

      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-6 leading-relaxed">
          Access your workmachine through a <strong className="text-foreground">VPN connection</strong>.
          Once connected, use SSH to access workspaces directly from your IDE or terminal.
        </p>

        <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
          <p className="text-foreground font-semibold text-sm mb-3">How access works</p>
          <ol className="text-muted-foreground text-sm space-y-2 list-decimal list-inside m-0">
            <li>Connect to VPN to join the workmachine network</li>
            <li>Add your SSH public key to authorized keys</li>
            <li>SSH into workspaces from your IDE or terminal</li>
          </ol>
        </div>
      </section>

      <section id="vpn" className="mb-12">
        <SectionTitle id="vpn">VPN Connection</SectionTitle>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          The VPN connects your local machine to the workmachine network. This enables direct access
          to workspaces via SSH and to environment services by their service names.
        </p>

        <VPNConnectionPreview />
      </section>

      <section id="authorized-keys" className="mb-12">
        <SectionTitle id="authorized-keys">Authorized Keys</SectionTitle>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Add your SSH public keys to enable SSH access from your IDE or terminal.
          These are <strong className="text-foreground">your keys</strong> that authorize access to the workmachine.
        </p>

        <div className="mb-6">
          <AuthorizedKeysPreview />
        </div>

        <div className="space-y-6">
          <CommandBlock
            icon={Key}
            title="Add your public key"
            description="Enable SSH access by adding your public key to the workmachine."
          >
            <div className="space-y-4">
              <div>
                <p className="text-foreground text-sm font-semibold mb-2">1. Copy your public key</p>
                <CodeExample>
                  <CodeLine>cat ~/.ssh/id_ed25519.pub</CodeLine>
                </CodeExample>
              </div>
              <div>
                <p className="text-foreground text-sm font-semibold mb-2">2. Add in workmachine settings</p>
                <p className="text-muted-foreground text-sm m-0">
                  Open the workmachine settings panel and paste your public key in the Authorized Keys section.
                </p>
              </div>
            </div>
          </CommandBlock>

          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <h3 className="text-foreground text-base font-semibold mb-3">IDE Integration</h3>
            <p className="text-muted-foreground text-sm leading-relaxed mb-4">
              Once your key is authorized, connect to workspaces using SSH:
            </p>
            <CodeExample>
              <CodeLine>ssh user@workspace-name.workmachine</CodeLine>
            </CodeExample>
            <p className="text-muted-foreground text-sm mt-4 m-0">
              Works with VS Code Remote SSH, JetBrains Gateway, Cursor, and any SSH-capable editor.
            </p>
          </div>
        </div>
      </section>

      <section id="ssh-public-key" className="mb-16">
        <SectionTitle id="ssh-public-key">SSH Public Key</SectionTitle>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Each workmachine has a pre-generated SSH key pair. Use the public key to integrate
          with external services like GitHub, GitLab, or Bitbucket.
        </p>

        <div className="mb-6">
          <SSHPublicKeyPreview />
        </div>

        <CommandBlock
          icon={Github}
          title="Add to GitHub"
          description="Integrate your workmachine with GitHub to clone private repositories."
        >
          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-semibold mb-2">1. Copy the workmachine public key</p>
              <p className="text-muted-foreground text-sm m-0">
                Find the SSH public key in your workmachine settings and click the copy button.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-semibold mb-2">2. Add to GitHub</p>
              <p className="text-muted-foreground text-sm m-0">
                Go to GitHub → Settings → SSH and GPG Keys → New SSH Key, and paste the public key.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-semibold mb-2">3. Clone repositories</p>
              <p className="text-muted-foreground text-sm m-0">
                Your workspaces can now clone private repositories using SSH URLs.
              </p>
            </div>
          </div>
        </CommandBlock>

        <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-4 mt-6">
          <p className="text-foreground text-sm font-semibold mb-2">Tip</p>
          <p className="text-muted-foreground text-sm m-0">
            The same public key works for GitLab, Bitbucket, and any SSH-based service.
          </p>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/workmachines"
          title="Workmachines Overview"
          description="Learn about workmachine features"
        />
      </div>
    </DocsContentLayout>
  )
}
