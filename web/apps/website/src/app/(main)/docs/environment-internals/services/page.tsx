import Link from 'next/link'
import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import { ArrowRight } from 'lucide-react'

const tocItems = [
  { id: 'composition', title: 'Composition' },
  { id: 'images-ports', title: 'Images & Ports' },
  { id: 'volumes', title: 'Volumes' },
  { id: 'networking', title: 'Networking' },
]

export default function ServicesPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-6 text-3xl sm:text-4xl font-bold tracking-tight">Services</h1>

      <p className="text-muted-foreground mb-8 leading-relaxed">
        Define services using Docker Compose syntax. Databases, caches, APIs—anything that runs
        in a container can be deployed to your environment.
      </p>

      {/* Composition */}
      <section id="composition" className="mb-10">
        <h2 className="text-foreground text-xl font-bold mb-4">Composition</h2>

        <div className="bg-zinc-950 border border-zinc-800 overflow-hidden">
          <div className="bg-zinc-900 px-4 py-2 border-b border-zinc-800">
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

      {/* Images & Ports */}
      <section id="images-ports" className="mb-10">
        <h2 className="text-foreground text-xl font-bold mb-4">Images & Ports</h2>

        <div className="grid sm:grid-cols-2 gap-4">
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-2 m-0">Images</p>
            <div className="bg-muted/50 p-2 font-mono text-xs text-muted-foreground space-y-1">
              <div>image: postgres:16</div>
              <div>image: ghcr.io/org/api:latest</div>
            </div>
          </div>
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-2 m-0">Ports</p>
            <div className="bg-muted/50 p-2 font-mono text-xs text-muted-foreground space-y-1">
              <div>- "5432:5432" # host:container</div>
              <div>- "3000:8080" # map 8080 to 3000</div>
            </div>
          </div>
        </div>
      </section>

      {/* Volumes */}
      <section id="volumes" className="mb-10">
        <h2 className="text-foreground text-xl font-bold mb-4">Volumes</h2>

        <div className="grid sm:grid-cols-2 gap-4">
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-2 m-0">Named Volumes</p>
            <div className="bg-muted/50 p-2 font-mono text-xs text-muted-foreground">
              - postgres_data:/var/lib/postgresql/data
            </div>
            <p className="text-muted-foreground text-xs mt-2">Persist data across restarts</p>
          </div>
          <div className="bg-card border p-4">
            <p className="font-medium text-foreground text-sm mb-2 m-0">Config Files</p>
            <div className="bg-muted/50 p-2 font-mono text-xs text-muted-foreground">
              - /files/nginx.conf:/etc/nginx/nginx.conf
            </div>
            <p className="text-muted-foreground text-xs mt-2">
              Mount files using{' '}
              <Link href="/docs/environment-internals/configs-secrets#config-files" className="text-primary hover:underline">
                /files/ prefix
              </Link>
            </p>
          </div>
        </div>
      </section>

      {/* Networking */}
      <section id="networking" className="mb-10">
        <h2 className="text-foreground text-xl font-bold mb-4">Networking</h2>

        <p className="text-muted-foreground mb-4 text-sm">
          Services communicate using service names as hostnames. DNS resolution is automatic.
        </p>

        <div className="bg-card border p-4">
          <div className="bg-muted/50 p-2 font-mono text-xs text-muted-foreground space-y-1">
            <div>postgres://user:pass@<span className="text-primary">postgres</span>:5432/db</div>
            <div>redis://<span className="text-primary">redis</span>:6379</div>
            <div>http://<span className="text-primary">api</span>:8080/health</div>
          </div>
        </div>
      </section>

      {/* Next Links */}
      <div className="border-t pt-8">
        <Link
          href="/docs/environment-internals/configs-secrets"
          className="bg-card border p-4 hover:border-primary transition-colors flex items-center justify-between gap-4 no-underline"
        >
          <div>
            <p className="text-foreground font-medium m-0">Configs & Secrets</p>
            <p className="text-muted-foreground text-sm m-0">Manage environment variables and config files</p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
        </Link>
      </div>
    </DocsContentLayout>
  )
}
