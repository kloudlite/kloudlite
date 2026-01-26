import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { NextLinkCard } from '@/components/docs/next-link-card'
import { EnvVarsPreview, ConfigFilesPreview } from './_components/step-previews'

const tocItems = [
  { id: 'environment-variables', title: 'Environment Variables' },
  { id: 'config-files', title: 'Config Files' },
  { id: 'usage', title: 'Usage' },
]

export default function ConfigsSecretsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Configs & Secrets</PageTitle>

      <p className="text-muted-foreground mb-8 leading-relaxed">
        Store configuration separately from your service definitions. Update values without
        changing your composition.
      </p>

      <section id="environment-variables" className="mb-12">
        <SectionTitle id="environment-variables">Environment Variables</SectionTitle>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Add variables as <strong className="text-foreground">Config</strong> (visible) or{' '}
          <strong className="text-foreground">Secret</strong> (encrypted). Reference them in your
          composition using <code className="bg-muted px-1.5 py-0.5 text-xs font-mono border border-foreground/10 rounded-sm">{'${VAR_NAME}'}</code>.
        </p>

        <EnvVarsPreview />
      </section>

      <section id="config-files" className="mb-12">
        <SectionTitle id="config-files">Config Files</SectionTitle>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Upload configuration files to mount into containers. Reference them using the{' '}
          <code className="bg-muted px-1.5 py-0.5 text-xs font-mono border border-foreground/10 rounded-sm">/files/</code> prefix.
        </p>

        <ConfigFilesPreview />
      </section>

      <section id="usage" className="mb-16">
        <SectionTitle id="usage">Usage</SectionTitle>

        <div className="bg-zinc-950 border border-foreground/10 rounded-sm overflow-hidden mb-6">
          <div className="bg-zinc-900 px-4 py-2 border-b border-foreground/10">
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

        <div className="grid sm:grid-cols-2 gap-6">
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-2">Variables</p>
            <code className="bg-muted px-2 py-1 text-xs font-mono border border-foreground/10 rounded-sm">{'${VAR_NAME}'}</code>
          </div>
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-2">Files</p>
            <code className="bg-muted px-2 py-1 text-xs font-mono border border-foreground/10 rounded-sm">/files/name:/path</code>
          </div>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/concepts/workspaces"
          title="Workspaces"
          description="Connect to environments from your workspace"
        />
      </div>
    </DocsContentLayout>
  )
}
