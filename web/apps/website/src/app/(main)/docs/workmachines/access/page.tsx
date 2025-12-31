import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Key, Github } from 'lucide-react'
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
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Access
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-6 leading-relaxed">
          Access your workmachine through a <strong className="text-foreground">VPN connection</strong>.
          Once connected, use SSH to access workspaces directly from your IDE or terminal.
        </p>

        <div className="bg-muted/50 border p-4 text-sm">
          <p className="text-foreground font-medium mb-2">How access works</p>
          <ol className="text-muted-foreground space-y-1 list-decimal list-inside m-0">
            <li>Connect to VPN to join the workmachine network</li>
            <li>Add your SSH public key to authorized keys</li>
            <li>SSH into workspaces from your IDE or terminal</li>
          </ol>
        </div>
      </section>

      {/* VPN Connection */}
      <section id="vpn" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">VPN Connection</h2>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          The VPN connects your local machine to the workmachine network. This enables direct access
          to workspaces via SSH and to environment services by their service names.
        </p>

        <VPNConnectionPreview />
      </section>

      {/* Authorized Keys */}
      <section id="authorized-keys" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Authorized Keys</h2>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Add your SSH public keys to enable SSH access from your IDE or terminal.
          These are <strong className="text-foreground">your keys</strong> that authorize access to the workmachine.
        </p>

        <div className="mb-6">
          <AuthorizedKeysPreview />
        </div>

        <div className="space-y-4">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Key className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Add your public key</h3>
            </div>
            <div className="bg-muted/50 p-4 space-y-3">
              <div>
                <p className="text-foreground text-sm font-medium mb-1">1. Copy your public key</p>
                <div className="bg-muted p-2 font-mono text-xs overflow-x-auto">
                  <pre className="m-0">cat ~/.ssh/id_ed25519.pub</pre>
                </div>
              </div>
              <div>
                <p className="text-foreground text-sm font-medium mb-1">2. Add in workmachine settings</p>
                <p className="text-muted-foreground text-xs m-0">
                  Open the workmachine settings panel and paste your public key in the Authorized Keys section.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">IDE Integration</h3>
            <p className="text-muted-foreground text-sm leading-relaxed mb-4">
              Once your key is authorized, connect to workspaces using SSH:
            </p>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">ssh user@workspace-name.workmachine</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Works with VS Code Remote SSH, JetBrains Gateway, Cursor, and any SSH-capable editor.
            </p>
          </div>
        </div>
      </section>

      {/* SSH Public Key */}
      <section id="ssh-public-key" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">SSH Public Key</h2>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Each workmachine has a pre-generated SSH key pair. Use the public key to integrate
          with external services like GitHub, GitLab, or Bitbucket.
        </p>

        <div className="mb-6">
          <SSHPublicKeyPreview />
        </div>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Github className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Add to GitHub</h3>
          </div>
          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">1. Copy the workmachine public key</p>
              <p className="text-muted-foreground text-xs m-0">
                Find the SSH public key in your workmachine settings and click the copy button.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">2. Add to GitHub</p>
              <p className="text-muted-foreground text-xs m-0">
                Go to GitHub → Settings → SSH and GPG Keys → New SSH Key, and paste the public key.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">3. Clone repositories</p>
              <p className="text-muted-foreground text-xs m-0">
                Your workspaces can now clone private repositories using SSH URLs.
              </p>
            </div>
          </div>
        </div>

        <div className="bg-muted/50 border p-4 mt-6">
          <p className="text-foreground text-sm font-medium mb-1">Tip</p>
          <p className="text-muted-foreground text-xs m-0">
            The same public key works for GitLab, Bitbucket, and any SSH-based service.
          </p>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8">
        <Link
          href="/docs/workmachines"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Workmachines Overview</p>
            <p className="text-muted-foreground text-sm m-0">Learn about workmachine features</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
