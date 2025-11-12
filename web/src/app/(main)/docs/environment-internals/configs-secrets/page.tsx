import { DocsContentLayout } from '@/components/docs/docs-content-layout'
import {
  Settings,
  Lock,
  FileText,
  Key,
  CheckCircle2,
  Info,
  AlertTriangle,
} from 'lucide-react'

const tocItems = [
  { id: 'overview', title: 'Overview' },
  { id: 'environment-variables', title: 'Environment Variables' },
  { id: 'config-files', title: 'Config Files' },
  { id: 'secrets', title: 'Secrets Management' },
  { id: 'usage', title: 'Using in Services' },
]

export default function ConfigsSecretsPage() {
  return (
    <DocsContentLayout tocItems={tocItems}>
      <h1 className="text-foreground mb-8 text-3xl sm:text-4xl font-bold">
        Configs & Secrets
      </h1>

      {/* Overview */}
      <section id="overview" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Settings className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">Overview</h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Kloudlite provides two ways to manage configuration and sensitive data for your services:
          Environment Variables for key-value configuration and Config Files for complex
          configuration files. Secrets are stored securely and encrypted at rest.
        </p>

        <div className="grid gap-4 sm:gap-6 mb-6 md:grid-cols-2">
          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <Key className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Environment Variables
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Key-value pairs for simple configuration (API keys, database URLs, feature flags)
                </p>
              </div>
            </div>
          </div>

          <div className="bg-card rounded-lg border p-4">
            <div className="flex items-start gap-3">
              <div className="bg-primary/10 rounded-lg p-2">
                <FileText className="text-primary h-5 w-5" />
              </div>
              <div className="flex-1">
                <h4 className="text-card-foreground font-semibold mb-1 m-0 text-sm">
                  Config Files
                </h4>
                <p className="text-muted-foreground text-xs m-0 leading-relaxed">
                  Full configuration files that services can mount and use (JSON, YAML, etc.)
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Environment Variables */}
      <section id="environment-variables" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Key className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Environment Variables
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Environment variables are key-value pairs that are injected into your services at runtime.
          They&apos;re perfect for configuration that changes between environments.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Common Use Cases:
          </h4>
          <ul className="space-y-2 m-0">
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Database Connection Strings:</strong>{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  DATABASE_URL=postgres://user:pass@postgres:5432/myapp
                </code>
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>API Keys & Tokens:</strong>{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  STRIPE_API_KEY=sk_test_...
                </code>
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Feature Flags:</strong>{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  ENABLE_BETA_FEATURES=true
                </code>
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Environment Names:</strong>{' '}
                <code className="bg-muted rounded px-1 py-0.5 font-mono text-xs">
                  NODE_ENV=development
                </code>
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Config Files */}
      <section id="config-files" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <FileText className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Configuration Files
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Config files allow you to upload complete configuration files that can be mounted into
          your services. This is useful for complex configurations that don&apos;t fit well as
          environment variables.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6 mb-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Supported File Types:
          </h4>
          <ul className="space-y-2 m-0">
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>JSON/YAML Configs:</strong> Application configuration files
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Nginx/Apache Configs:</strong> Web server configuration
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>SSL Certificates:</strong> TLS certificates and keys
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Any Text File:</strong> Custom configuration formats
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Secrets Management */}
      <section id="secrets" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Lock className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Secrets Management
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Secrets are sensitive data like passwords, API keys, and tokens. They are stored
          encrypted at rest and only decrypted when injected into services.
        </p>

        <div className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800 rounded-lg border p-3 sm:p-4 mb-6">
          <div className="flex gap-2 sm:gap-3">
            <AlertTriangle className="text-amber-600 dark:text-amber-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-amber-900 dark:text-amber-100 text-sm font-medium m-0 mb-1">
                Security Best Practices
              </p>
              <ul className="text-amber-800 dark:text-amber-200 text-sm m-0 space-y-1 list-disc list-inside">
                <li>Never commit secrets to version control</li>
                <li>Use environment-specific secrets for different stages</li>
                <li>Rotate secrets regularly</li>
                <li>Limit secret access to only services that need them</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <h4 className="text-card-foreground font-semibold mb-3 m-0 text-base">
            Secret Features:
          </h4>
          <ul className="space-y-2 m-0">
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Encrypted Storage:</strong> All secrets encrypted at rest
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Secure Injection:</strong> Secrets only visible to authorized services
              </span>
            </li>
            <li className="flex items-start gap-2 text-sm">
              <CheckCircle2 className="text-green-500 h-4 w-4 flex-shrink-0 mt-0.5" />
              <span className="text-muted-foreground">
                <strong>Version History:</strong> Track changes to secret values
              </span>
            </li>
          </ul>
        </div>
      </section>

      {/* Usage in Services */}
      <section id="usage" className="mb-12 sm:mb-16">
        <div className="mb-6 flex items-center gap-3">
          <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-full">
            <Settings className="text-primary-foreground h-6 w-6" />
          </div>
          <h2 className="text-foreground m-0 text-2xl sm:text-3xl font-bold">
            Using in Services
          </h2>
        </div>

        <p className="text-muted-foreground mb-6 leading-relaxed">
          Reference configs and secrets in your Docker Compose service definitions using environment
          variable substitution.
        </p>

        <div className="bg-card rounded-lg border p-4 sm:p-6">
          <p className="text-card-foreground text-sm mb-3 m-0 font-medium">
            Example: Using Environment Variables in Services
          </p>
          <div className="bg-muted rounded p-4 font-mono text-xs overflow-x-auto">
            <pre className="m-0 leading-relaxed">
{`services:
  api:
    image: myapp/api:latest
    environment:
      # Reference environment-level variables
      DATABASE_URL: \${DATABASE_URL}
      REDIS_URL: \${REDIS_URL}
      API_SECRET: \${API_SECRET}
      NODE_ENV: production
    ports:
      - "3000:3000"`}
            </pre>
          </div>
        </div>

        <div className="mt-6 bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800 rounded-lg border p-3 sm:p-4">
          <div className="flex gap-2 sm:gap-3">
            <Info className="text-green-600 dark:text-green-400 h-5 w-5 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-900 dark:text-green-100 text-sm font-medium m-0 mb-1">
                Environment-Level Configuration
              </p>
              <p className="text-green-800 dark:text-green-200 text-sm m-0 leading-relaxed">
                All configs and secrets are defined at the environment level in the Environment
                Settings. They are then available to all services in that environment via variable
                substitution.
              </p>
            </div>
          </div>
        </div>
      </section>
    </DocsContentLayout>
  )
}
