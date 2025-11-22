import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Network,
  Link2,
  Database,
  Code2,
  CheckCircle2,
  Info,
  Terminal,
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'how-it-works', title: 'How It Works' },
  { id: 'connecting', title: 'Connecting to Environment' },
  { id: 'accessing-services', title: 'Accessing Services by Name' },
  { id: 'use-cases', title: 'Use Cases' },
]

export default function EnvironmentConnectionPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        Environment Connection
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Link2 className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Environment Connection allows workspaces to connect to environments and access their
          services directly by name. When a workspace connects to an environment, its network
          namespace switches to enable seamless access to all services running in that environment.
        </p>

        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-1">
                Network Namespace Switching
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed">
                When you connect a workspace to an environment, the workspace&apos;s network
                namespace switches to access services from the environment by their names (e.g.,
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs mx-1">
                  postgres
                </code>
                ,
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs mx-1">redis</code>
                ).
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Network className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">How It Works</h2>
        </div>

        <div className="space-y-6">
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                1
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Workspace Without Connection
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  By default, a workspace runs in its own isolated network namespace. It has access
                  to external internet but cannot directly access environment services.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                2
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Connect to Environment
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  When you connect the workspace to an environment, the workspace&apos;s network
                  configuration changes. The workspace can now resolve environment service names
                  through DNS.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <div className="flex items-start gap-4">
              <div className="bg-primary text-primary-foreground mt-1 flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold">
                3
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-2 m-0">
                  Service Access by Name
                </h4>
                <p className="text-muted-foreground text-sm m-0 leading-relaxed">
                  Services are now accessible using their service names as hostnames. For example,
                  connect to PostgreSQL at{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                    postgres:5432
                  </code>{' '}
                  or Redis at{' '}
                  <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">redis:6379</code>
                  .
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Connecting to Environment */}
      <section id="connecting" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Connecting to Environment
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Use the <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-sm">kl env connect</code>{' '}
          command from within your workspace to connect to an environment. This switches your
          workspace&apos;s network namespace to access environment services.
        </p>

        {/* Interactive Mode */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <CheckCircle2 className="text-green-500 h-5 w-5" />
            Interactive Mode (Recommended)
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            Run the command without arguments to see a list of available environments and select
            one interactively using fuzzy-find.
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">kl env connect</pre>
          </div>
          <div className="mt-3 bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded border p-3">
            <p className="text-blue-800 dark:text-blue-200 text-xs m-0 leading-relaxed">
              This will display all available environments and let you select using arrow keys and
              search.
            </p>
          </div>
        </div>

        {/* Direct Connection */}
        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <CheckCircle2 className="text-green-500 h-5 w-5" />
            Connect to Specific Environment
          </h4>
          <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
            If you know the environment name, you can connect directly:
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-2">
            <pre className="m-0 leading-relaxed">kl env connect my-env</pre>
            <pre className="m-0 leading-relaxed">kl env connect team-env</pre>
            <pre className="m-0 leading-relaxed">kl e c test-env          # Using aliases</pre>
          </div>
        </div>

        {/* What Happens After Connection */}
        <div className="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-950 dark:to-emerald-950 rounded-lg border-2 border-green-300 dark:border-green-700 p-4 sm:p-6">
          <h4 className="text-green-900 dark:text-green-100 font-semibold mb-3 m-0 text-base flex items-center gap-2">
            <Info className="text-green-600 dark:text-green-400 h-5 w-5" />
            What Happens After Connection
          </h4>
          <ul className="space-y-2 m-0">
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-600 dark:text-green-400 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-green-800 dark:text-green-200">
                <strong>Network Namespace Switch:</strong> Your workspace&apos;s network configuration
                changes to the environment&apos;s namespace
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-600 dark:text-green-400 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-green-800 dark:text-green-200">
                <strong>DNS Resolution:</strong> Service names in the environment become resolvable
                as hostnames
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-600 dark:text-green-400 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-green-800 dark:text-green-200">
                <strong>Immediate Access:</strong> You can now access services using their short names
                (e.g., <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">postgres</code>,{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">redis</code>)
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Accessing Services */}
      <section id="accessing-services" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Database className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Accessing Services by Name
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          After connecting to an environment, all services running in that environment become
          accessible using their service names as hostnames. You no longer need fully qualified
          domain names or IP addresses.
        </p>

        {/* How Service Names Work */}
        <div className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 rounded-lg border p-3 sm:p-4 mb-6">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-blue-600 dark:text-blue-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-blue-900 dark:text-blue-100 text-sm font-medium m-0 mb-2">
                How It Works
              </p>
              <p className="text-blue-800 dark:text-blue-200 text-sm m-0 leading-relaxed mb-3">
                When you connect to an environment, your workspace&apos;s DNS configuration is updated
                to resolve service names to their cluster IPs. This means:
              </p>
              <ul className="text-blue-800 dark:text-blue-200 text-sm space-y-1 m-0 list-disc list-inside">
                <li>Service <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">postgres</code> resolves to its cluster IP</li>
                <li>Service <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">redis</code> resolves to its cluster IP</li>
                <li>Service <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">api-server</code> resolves to its cluster IP</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Examples Section */}
        <div className="space-y-6">
          {/* From Terminal */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Terminal className="text-primary h-5 w-5" />
              From Terminal
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Access services directly from your workspace terminal using their service names:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-4">
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Connect to PostgreSQL</p>
                <pre className="m-0 leading-relaxed">psql -h postgres -p 5432 -U myuser -d myapp</pre>
              </div>
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Connect to Redis</p>
                <pre className="m-0 leading-relaxed">redis-cli -h redis -p 6379</pre>
              </div>
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Ping a service</p>
                <pre className="m-0 leading-relaxed">ping postgres</pre>
              </div>
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Check service connectivity</p>
                <pre className="m-0 leading-relaxed">curl http://api-server:8080/health</pre>
              </div>
            </div>
          </div>

          {/* From Application Code */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Code2 className="text-primary h-5 w-5" />
              In Application Code
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Use service names directly in your application configuration and code:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto space-y-4">
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Environment Variables (.env)</p>
                <pre className="m-0 leading-relaxed">{`DATABASE_URL=postgresql://user:pass@postgres:5432/myapp
REDIS_URL=redis://redis:6379
API_ENDPOINT=http://api-server:8080`}</pre>
              </div>
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Node.js Example</p>
                <pre className="m-0 leading-relaxed">{`const redis = require('redis');
const client = redis.createClient({
  host: 'redis',
  port: 6379
});`}</pre>
              </div>
              <div>
                <p className="text-muted-foreground mb-2 text-xs"># Python Example</p>
                <pre className="m-0 leading-relaxed">{`import psycopg2

conn = psycopg2.connect(
    host="postgres",
    port=5432,
    database="myapp"
)`}</pre>
              </div>
            </div>
          </div>

          {/* Service Discovery */}
          <div className="bg-card rounded-lg border p-4 sm:p-6">
            <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base flex items-center gap-2">
              <Network className="text-primary h-5 w-5" />
              Find Available Services
            </h4>
            <p className="text-muted-foreground text-sm mb-4 leading-relaxed">
              Use <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">kl env status</code>{' '}
              to see all services available in the connected environment:
            </p>
            <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
              <pre className="m-0 leading-relaxed">kl env status</pre>
            </div>
            <p className="text-muted-foreground text-xs mt-3 leading-relaxed">
              This command shows the environment name, namespace, and lists all available services
              with their ports.
            </p>
          </div>
        </div>

        {/* Important Note */}
        <div className="mt-6 bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-amber-600 dark:text-amber-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-amber-900 dark:text-amber-100 text-sm font-medium m-0 mb-1">
                Port Numbers
              </p>
              <p className="text-amber-800 dark:text-amber-200 text-sm m-0 leading-relaxed">
                Always specify the port number when connecting to services (e.g.,{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">postgres:5432</code>).
                The service name alone won&apos;t include the port.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Use Cases */}
      <section id="use-cases" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Terminal className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Use Cases</h2>
        </div>

        <div className="grid gap-4 sm:gap-6">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Code2 className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Development Against Real Services
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Develop your application code in the workspace while connecting to real databases,
                  caches, and APIs running in the environment
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Database className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Database Migrations
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Run database migrations from your workspace directly against the environment
                  database
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Terminal className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Testing & Debugging
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Test your code against production-like services and debug issues in a realistic
                  environment setup
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Network className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Switch Between Environments
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Connect to different environments as needed to test against various service
                  configurations and data sets
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-6 bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <CheckCircle2 className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Seamless Development Experience
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                Environment connection eliminates the need to run services locally or manage complex
                port forwarding. Simply connect and start coding against real services.
              </p>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
