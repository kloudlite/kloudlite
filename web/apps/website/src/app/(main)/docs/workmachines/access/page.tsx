import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight, Key, Terminal, Users } from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'your-keys', title: 'Your SSH Keys' },
  { id: 'authorized-keys', title: 'Authorized Keys' },
]

export default function AccessPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Access
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12">
        <p className="text-muted-foreground mb-8 leading-relaxed">
          SSH keys provide <strong className="text-foreground">secure access</strong> to your
          workspaces. Add your own keys to connect, or authorize others to access your workspaces.
        </p>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Key className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Key types supported</h3>
          </div>
          <div className="space-y-2 text-muted-foreground text-sm">
            <p className="m-0"><strong className="text-foreground">Ed25519</strong> — Modern, secure, and fast (recommended)</p>
            <p className="m-0"><strong className="text-foreground">RSA</strong> — 2048-bit or higher</p>
            <p className="m-0"><strong className="text-foreground">ECDSA</strong> — Elliptic curve keys</p>
          </div>
        </div>
      </section>

      {/* Your SSH Keys */}
      <section id="your-keys" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Your SSH Keys</h2>

        <div className="space-y-6">
          <div className="bg-card border p-6">
            <div className="flex items-center gap-3 mb-3">
              <Terminal className="h-5 w-5 text-primary" />
              <h3 className="text-card-foreground text-lg font-semibold m-0">Generate a key</h3>
            </div>
            <div className="bg-muted p-3 font-mono text-sm overflow-x-auto">
              <pre className="m-0">ssh-keygen -t ed25519 -C "your-email@example.com"</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 m-0">
              Creates a new Ed25519 key pair in <code className="bg-muted px-1.5 py-0.5 font-mono">~/.ssh/</code>
            </p>
          </div>

          <div className="bg-card border p-6">
            <h3 className="text-card-foreground text-lg font-semibold mb-3">Add to Kloudlite</h3>
            <div className="bg-muted/50 p-4 space-y-3">
              <div>
                <p className="text-foreground text-sm font-medium mb-1">1. Copy your public key</p>
                <div className="bg-muted p-2 font-mono text-xs overflow-x-auto">
                  <pre className="m-0">cat ~/.ssh/id_ed25519.pub</pre>
                </div>
              </div>
              <div>
                <p className="text-foreground text-sm font-medium mb-1">2. Add in settings</p>
                <p className="text-muted-foreground text-xs m-0">
                  Go to your account settings and paste the public key.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Authorized Keys */}
      <section id="authorized-keys" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-6">Authorized Keys</h2>

        <div className="bg-card border p-6">
          <div className="flex items-center gap-3 mb-3">
            <Users className="h-5 w-5 text-primary" />
            <h3 className="text-card-foreground text-lg font-semibold m-0">Grant access to others</h3>
          </div>
          <p className="text-muted-foreground text-sm leading-relaxed mb-4">
            Add SSH keys from teammates or CI systems to let them access your workspaces.
          </p>

          <div className="bg-muted/50 p-4 space-y-3">
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Team access</p>
              <p className="text-muted-foreground text-xs m-0">
                Add teammate public keys to grant them SSH access to your workspaces.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">CI/CD access</p>
              <p className="text-muted-foreground text-xs m-0">
                Add keys for automated systems that need workspace access.
              </p>
            </div>
            <div>
              <p className="text-foreground text-sm font-medium mb-1">Revoke access</p>
              <p className="text-muted-foreground text-xs m-0">
                Remove a key to immediately revoke access.
              </p>
            </div>
          </div>
        </div>

        <div className="bg-muted/50 border p-4 mt-6">
          <p className="text-foreground text-sm font-medium mb-1">Security note</p>
          <p className="text-muted-foreground text-xs m-0">
            Only add keys from trusted sources. Anyone with an authorized key can access your workspace via SSH.
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
