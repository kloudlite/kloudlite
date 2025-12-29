import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight } from 'lucide-react'
import { EnvVarsPreview, ConfigFilesPreview } from './_components/step-previews'

const tocItems = [
  { id: 'environment-variables', title: 'Environment Variables' },
  { id: 'config-files', title: 'Config Files' },
  { id: 'usage', title: 'Usage' },
]

export default function ConfigsSecretsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">
        Configs & Secrets
      </h1>

      <p className="text-muted-foreground mb-8 leading-relaxed">
        Store configuration separately from your service definitions. Update values without
        changing your composition.
      </p>

      {/* Environment Variables */}
      <section id="environment-variables" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-4">Environment Variables</h2>

        <p className="text-muted-foreground mb-6 text-sm">
          Add variables as <strong className="text-foreground">Config</strong> (visible) or{' '}
          <strong className="text-foreground">Secret</strong> (encrypted). Reference them in your
          composition using <code className="bg-muted px-1.5 py-0.5 text-xs font-mono">{'${VAR_NAME}'}</code>.
        </p>

        <EnvVarsPreview />
      </section>

      {/* Config Files */}
      <section id="config-files" className="mb-12">
        <h2 className="text-foreground text-2xl font-bold mb-4">Config Files</h2>

        <p className="text-muted-foreground mb-6 text-sm">
          Upload configuration files to mount into containers. Reference them using the{' '}
          <code className="bg-muted px-1.5 py-0.5 text-xs font-mono">/files/</code> prefix.
        </p>

        <ConfigFilesPreview />
      </section>

      {/* Usage */}
      <section id="usage" className="mb-10">
        <h2 className="text-foreground text-2xl font-bold mb-4">Usage</h2>

        <div className="bg-zinc-950 border border-zinc-800 overflow-hidden mb-6">
          <div className="bg-zinc-900 px-4 py-2 border-b border-zinc-800">
            <span className="text-zinc-400 text-xs font-mono">docker-compose.yml</span>
          </div>
          <pre className="p-4 overflow-x-auto">
            <code className="text-zinc-300 font-mono text-sm leading-normal">{`services:
  api:
    image: myapp/api
    environment:
      DATABASE_URL: \${DATABASE_URL}
      JWT_SECRET: \${JWT_SECRET}

  nginx:
    image: nginx:alpine
    volumes:
      - /files/nginx.conf:/etc/nginx/nginx.conf`}</code>
          </pre>
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-1">Variables</p>
            <code className="bg-muted px-2 py-1 text-xs font-mono">{'${VAR_NAME}'}</code>
          </div>
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-1">Files</p>
            <code className="bg-muted px-2 py-1 text-xs font-mono">/files/name:/path</code>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8">
        <Link
          href="/docs/concepts/workspaces"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Workspaces</p>
            <p className="text-muted-foreground text-sm m-0">Connect to environments from your workspace</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
