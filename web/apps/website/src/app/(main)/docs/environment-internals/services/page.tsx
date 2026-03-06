import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { PageTitle } from '@/components/docs/page-title'
import { SectionTitle } from '@/components/docs/section-title'
import { NextLinkCard } from '@/components/docs/next-link-card'

const tocItems = [
  { id: 'composition', title: 'Composition' },
  { id: 'images-ports', title: 'Images & Ports' },
  { id: 'volumes', title: 'Volumes' },
  { id: 'networking', title: 'Networking' },
]

export default function ServicesPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <PageTitle>Services</PageTitle>

      <p className="text-muted-foreground mb-8 leading-relaxed">
        Define services using Docker Compose syntax. Databases, caches, APIs—anything that runs
        in a container can be deployed to your environment.
      </p>

      <section id="composition" className="mb-12">
        <SectionTitle id="composition">Composition</SectionTitle>

        <div className="bg-zinc-950 border border-foreground/10 rounded-sm overflow-hidden">
          <div className="bg-zinc-900 px-4 py-2 border-b border-foreground/10">
            <span className="text-zinc-400 text-xs font-mono">docker-compose.yml</span>
          </div>
          <pre className="p-4 overflow-x-auto">
            <code className="text-zinc-300 font-mono text-sm leading-normal">{`services:
  postgres:
    image: postgres:16
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: \${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:`}</code>
          </pre>
        </div>
      </section>

      <section id="images-ports" className="mb-12">
        <SectionTitle id="images-ports">Images & Ports</SectionTitle>

        <div className="grid sm:grid-cols-2 gap-6">
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-3 m-0">Images</p>
            <div className="bg-muted/50 p-3 font-mono text-xs text-muted-foreground space-y-1 rounded-sm">
              <div>image: postgres:16</div>
              <div>image: ghcr.io/org/api:latest</div>
            </div>
          </div>
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-3 m-0">Ports</p>
            <div className="bg-muted/50 p-3 font-mono text-xs text-muted-foreground space-y-1 rounded-sm">
              <div>- &quot;5432:5432&quot; # host:container</div>
              <div>- &quot;3000:8080&quot; # map 8080 to 3000</div>
            </div>
          </div>
        </div>
      </section>

      <section id="volumes" className="mb-12">
        <SectionTitle id="volumes">Volumes</SectionTitle>

        <div className="grid sm:grid-cols-2 gap-6">
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-3 m-0">Named Volumes</p>
            <div className="bg-muted/50 p-3 font-mono text-xs text-muted-foreground rounded-sm">
              - postgres_data:/var/lib/postgresql/data
            </div>
            <p className="text-muted-foreground text-sm mt-3">Persist data across restarts</p>
          </div>
          <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
            <p className="font-semibold text-foreground text-sm mb-3 m-0">Config Files</p>
            <div className="bg-muted/50 p-3 font-mono text-xs text-muted-foreground rounded-sm">
              - /files/nginx.conf:/etc/nginx/nginx.conf
            </div>
            <p className="text-muted-foreground text-sm mt-3">
              Mount files using{' '}
              <Link href="/docs/environment-internals/configs-secrets#config-files" className="text-primary hover:underline">
                /files/ prefix
              </Link>
            </p>
          </div>
        </div>
      </section>

      <section id="networking" className="mb-16">
        <SectionTitle id="networking">Networking</SectionTitle>

        <p className="text-muted-foreground mb-6 text-sm leading-relaxed">
          Services communicate using service names as hostnames. DNS resolution is automatic.
        </p>

        <div className="bg-foreground/[0.02] border border-foreground/10 rounded-sm p-5">
          <div className="bg-muted/50 p-3 font-mono text-xs text-muted-foreground space-y-1 rounded-sm">
            <div>postgres://user:pass@<span className="text-primary">postgres</span>:5432/db</div>
            <div>redis://<span className="text-primary">redis</span>:6379</div>
            <div>http://<span className="text-primary">api</span>:8080/health</div>
          </div>
        </div>
      </section>

      <div className="border-t border-foreground/10 pt-8">
        <NextLinkCard
          href="/docs/environment-internals/configs-secrets"
          title="Configs & Secrets"
          description="Manage environment variables and config files"
        />
      </div>
    </DocsContentLayout>
  )
}
