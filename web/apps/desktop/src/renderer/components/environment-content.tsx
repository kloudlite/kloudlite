import { cn } from '@/lib/utils'
import { Copy, Check, Pencil, Trash2, Eye, EyeOff, Plus, Key, FileText as FileIcon } from 'lucide-react'
import { useState } from 'react'
import { CodeEditor } from './code-editor'
import { SnapshotTree, generateSnapshots } from './snapshot-tree'

interface EnvironmentContentProps {
  envName: string
  envHash: string
  activeTab: string
}

// Dummy services data
const SERVICES: Record<string, { name: string; type: string; clusterIP: string; ports: { port: number; targetPort: string; protocol: string }[]; dns: string }[]> = {
  'a1b2c3': [
    { name: 'frontend', type: 'ClusterIP', clusterIP: '10.96.45.12', ports: [{ port: 3000, targetPort: '3000', protocol: 'TCP' }], dns: 'frontend-a1b2c3.staging.local' },
    { name: 'api-server', type: 'ClusterIP', clusterIP: '10.96.45.13', ports: [{ port: 8080, targetPort: '8080', protocol: 'TCP' }], dns: 'api-server-a1b2c3.staging.local' },
    { name: 'redis', type: 'ClusterIP', clusterIP: '10.96.45.14', ports: [{ port: 6379, targetPort: '6379', protocol: 'TCP' }], dns: 'redis-a1b2c3.staging.local' },
    { name: 'postgres', type: 'ClusterIP', clusterIP: '10.96.45.15', ports: [{ port: 5432, targetPort: '5432', protocol: 'TCP' }], dns: 'postgres-a1b2c3.staging.local' },
  ],
  'd4e5f6': [
    { name: 'web-app', type: 'ClusterIP', clusterIP: '10.96.50.10', ports: [{ port: 5173, targetPort: '5173', protocol: 'TCP' }], dns: 'web-app-d4e5f6.dev.local' },
    { name: 'auth-service', type: 'ClusterIP', clusterIP: '10.96.50.11', ports: [{ port: 9090, targetPort: '9090', protocol: 'TCP' }], dns: 'auth-d4e5f6.dev.local' },
  ],
  'g7h8i9': [
    { name: 'gateway', type: 'LoadBalancer', clusterIP: '10.96.60.10', ports: [{ port: 443, targetPort: '8443', protocol: 'TCP' }, { port: 80, targetPort: '8080', protocol: 'TCP' }], dns: 'gateway-g7h8i9.prod.local' },
    { name: 'dashboard', type: 'ClusterIP', clusterIP: '10.96.60.11', ports: [{ port: 3000, targetPort: '3000', protocol: 'TCP' }], dns: 'dashboard-g7h8i9.prod.local' },
  ],
}

// Dummy configs
const CONFIGS: Record<string, { name: string; type: 'configmap' | 'secret'; keys: string[]; updated: string }[]> = {
  'a1b2c3': [
    { name: 'app-config', type: 'configmap', keys: ['DATABASE_URL', 'REDIS_URL', 'API_KEY'], updated: '2 hours ago' },
    { name: 'tls-certs', type: 'secret', keys: ['tls.crt', 'tls.key', 'ca.crt'], updated: '5 days ago' },
    { name: 'feature-flags', type: 'configmap', keys: ['ENABLE_DARK_MODE', 'ENABLE_BETA'], updated: '1 day ago' },
  ],
  'd4e5f6': [
    { name: 'dev-config', type: 'configmap', keys: ['DEBUG', 'LOG_LEVEL', 'PORT'], updated: '30 min ago' },
    { name: 'db-credentials', type: 'secret', keys: ['username', 'password'], updated: '3 days ago' },
  ],
  'g7h8i9': [
    { name: 'prod-config', type: 'configmap', keys: ['NODE_ENV', 'CDN_URL', 'SENTRY_DSN'], updated: '1 week ago' },
    { name: 'api-keys', type: 'secret', keys: ['STRIPE_KEY', 'SENDGRID_KEY'], updated: '2 weeks ago' },
  ],
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <button
      className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground"
      onClick={() => {
        navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      }}
    >
      {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  )
}

// Dummy compositions
const COMPOSITIONS: Record<string, string> = {
  'a1b2c3': `version: "3.8"
services:
  frontend:
    image: kloudlite/frontend:latest
    ports:
      - "3000:3000"
    environment:
      - API_URL=http://api-server:8080
  api-server:
    image: kloudlite/api:latest
    ports:
      - "8080:8080"
    depends_on:
      - redis
      - postgres
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=app
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=secret`,
  'd4e5f6': `version: "3.8"
services:
  web-app:
    image: kloudlite/web:dev
    ports:
      - "5173:5173"
    volumes:
      - ./src:/app/src
  auth-service:
    image: kloudlite/auth:dev
    ports:
      - "9090:9090"`,
  'g7h8i9': `version: "3.8"
services:
  gateway:
    image: kloudlite/gateway:stable
    ports:
      - "443:8443"
      - "80:8080"
  dashboard:
    image: kloudlite/dashboard:stable
    ports:
      - "3000:3000"`,
}

function ServicesView({ envHash }: { envHash: string }) {
  const services = SERVICES[envHash] || []
  const [compose, setCompose] = useState(COMPOSITIONS[envHash] || '')
  const [composeOpen, setComposeOpen] = useState(false)
  const [composeExiting, setComposeExiting] = useState(false)
  const [saved, setSaved] = useState(false)

  function closeCompose() {
    setComposeExiting(true)
    setTimeout(() => {
      setComposeOpen(false)
      setComposeExiting(false)
    }, 150)
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Services</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">{services.length} services deployed</p>
        </div>
        <button
          className={cn(
            'rounded-lg border px-3 py-1.5 text-[12px] font-medium transition-colors',
            composeOpen
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-accent'
          )}
          onClick={() => composeOpen ? closeCompose() : setComposeOpen(true)}
        >
          Composition
        </button>
      </div>

      {/* Composition editor */}
      {composeOpen && (
        <div className="mt-4 overflow-hidden rounded-xl border border-border/50" style={{ animation: composeExiting ? 'popover-out 150ms ease-in forwards' : 'popover-in 150ms ease-out' }}>
          <div className="flex items-center justify-between border-b border-border/30 bg-card px-4 py-2">
            <span className="text-[11px] font-medium text-muted-foreground">docker-compose.yml</span>
            <div className="flex items-center gap-2">
              <button
                className="rounded-md px-3 py-1 text-[11px] font-medium text-muted-foreground transition-colors hover:bg-accent"
                onClick={() => {
                  setCompose(COMPOSITIONS[envHash] || '')
                  closeCompose()
                }}
              >
                Cancel
              </button>
              <button
                className="rounded-md bg-primary px-3 py-1 text-[11px] font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                onClick={() => {
                  setSaved(true)
                  setTimeout(() => setSaved(false), 2000)
                }}
              >
                {saved ? 'Saved!' : 'Apply'}
              </button>
            </div>
          </div>
          <CodeEditor
            value={compose}
            onChange={setCompose}
            height="300px"
          />
        </div>
      )}

      {/* Services table */}
      <div className="mt-5">
        <table className="w-full text-left text-[13px]">
          <thead>
            <tr className="border-b border-border/50 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">
              <th className="pb-2.5 pr-4">Name</th>
              <th className="pb-2.5 pr-4">DNS Hostname</th>
              <th className="pb-2.5 pr-4">Cluster IP</th>
              <th className="pb-2.5 pr-4">Ports</th>
              <th className="pb-2.5">Type</th>
            </tr>
          </thead>
          <tbody>
            {services.map((svc) => (
              <tr key={svc.name} className="h-12 border-b border-border/30 transition-colors hover:bg-accent/30">
                <td className="py-3 pr-4">
                  <span className="font-medium text-foreground">{svc.name}</span>
                </td>
                <td className="py-3 pr-4">
                  <div className="flex items-center gap-1.5">
                    <span className="text-muted-foreground">{svc.dns}</span>
                    <CopyButton text={svc.dns} />
                  </div>
                </td>
                <td className="py-3 pr-4 text-muted-foreground">{svc.clusterIP}</td>
                <td className="py-3 pr-4">
                  <div className="flex flex-wrap gap-1">
                    {svc.ports.map((p) => (
                      <span key={p.port} className="rounded bg-accent px-1.5 py-0.5 text-[11px] text-muted-foreground">
                        {p.port}→{p.targetPort}/{p.protocol}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="py-3">
                  <span className={cn(
                    'rounded-full px-2 py-0.5 text-[10px] font-medium',
                    svc.type === 'LoadBalancer' ? 'bg-blue-500/10 text-blue-600' : 'bg-accent text-muted-foreground'
                  )}>
                    {svc.type}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// Dummy envvars
const ENVVARS: Record<string, { key: string; value: string; type: 'config' | 'secret' }[]> = {
  'a1b2c3': [
    { key: 'DATABASE_URL', value: 'postgresql://admin:secret@postgres:5432/app', type: 'secret' },
    { key: 'REDIS_URL', value: 'redis://redis:6379', type: 'config' },
    { key: 'API_KEY', value: 'sk-kloudlite-a1b2c3d4e5f6', type: 'secret' },
    { key: 'NODE_ENV', value: 'staging', type: 'config' },
    { key: 'LOG_LEVEL', value: 'debug', type: 'config' },
  ],
  'd4e5f6': [
    { key: 'DEBUG', value: 'true', type: 'config' },
    { key: 'LOG_LEVEL', value: 'verbose', type: 'config' },
    { key: 'PORT', value: '5173', type: 'config' },
    { key: 'DB_PASSWORD', value: 'dev-secret-123', type: 'secret' },
  ],
  'g7h8i9': [
    { key: 'NODE_ENV', value: 'production', type: 'config' },
    { key: 'CDN_URL', value: 'https://cdn.kloudlite.io', type: 'config' },
    { key: 'SENTRY_DSN', value: 'https://abc@sentry.io/123', type: 'secret' },
    { key: 'STRIPE_KEY', value: 'sk_live_kloudlite_prod', type: 'secret' },
  ],
}

const CONFIG_FILES: Record<string, { name: string; size: string }[]> = {
  'a1b2c3': [
    { name: 'nginx.conf', size: '2.4 KB' },
    { name: 'app-config.json', size: '1.1 KB' },
  ],
  'd4e5f6': [
    { name: 'vite.config.ts', size: '0.8 KB' },
  ],
  'g7h8i9': [
    { name: 'gateway.yaml', size: '3.2 KB' },
    { name: 'tls.crt', size: '4.1 KB' },
    { name: 'tls.key', size: '1.6 KB' },
  ],
}

function ConfigsView({ envHash }: { envHash: string }) {
  const [activeSection, setActiveSection] = useState<'envvars' | 'files'>('envvars')
  const [revealedSecrets, setRevealedSecrets] = useState<Set<string>>(new Set())
  const envvars = ENVVARS[envHash] || []
  const files = CONFIG_FILES[envHash] || []

  function toggleReveal(key: string) {
    setRevealedSecrets((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Configs & Secrets</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">Environment variables and configuration files</p>
        </div>
        <button className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90">
          <Plus className="h-3.5 w-3.5" />
          Add
        </button>
      </div>

      {/* Section tabs */}
      <div className="mt-4 flex gap-1 rounded-lg bg-accent/50 p-0.5">
        <button
          className={cn(
            'flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-[12px] font-medium transition-colors',
            activeSection === 'envvars'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onClick={() => setActiveSection('envvars')}
        >
          <Key className="h-3.5 w-3.5" />
          Env Variables
        </button>
        <button
          className={cn(
            'flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-[12px] font-medium transition-colors',
            activeSection === 'files'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onClick={() => setActiveSection('files')}
        >
          <FileIcon className="h-3.5 w-3.5" />
          Config Files
        </button>
      </div>

      {/* Envvars table */}
      {activeSection === 'envvars' && (
        <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
          <table className="w-full text-left text-[13px]">
            <thead>
              <tr className="border-b border-border/50 bg-accent/30">
                <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Key</th>
                <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Value</th>
                <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Type</th>
                <th className="w-20 px-4 py-2.5"></th>
              </tr>
            </thead>
            <tbody>
              {envvars.map((env) => (
                <tr key={env.key} className="h-12 border-b border-border/30 transition-colors hover:bg-accent/20">
                  <td className="px-4">
                    <span className="font-mono text-[12px] font-medium text-foreground">{env.key}</span>
                  </td>
                  <td className="px-4">
                    <span className="font-mono text-[12px] text-muted-foreground">
                      {env.type === 'secret' && !revealedSecrets.has(env.key)
                        ? '••••••••••••'
                        : env.value}
                    </span>
                  </td>
                  <td className="px-4">
                    <span className={cn(
                      'rounded-full px-2 py-0.5 text-[10px] font-medium',
                      env.type === 'secret'
                        ? 'bg-purple-500/10 text-purple-600'
                        : 'bg-blue-500/10 text-blue-600'
                    )}>
                      {env.type === 'secret' ? 'Secret' : 'Config'}
                    </span>
                  </td>
                  <td className="px-4">
                    <div className="flex items-center justify-end gap-1">
                      {env.type === 'secret' && (
                        <button
                          className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground"
                          onClick={() => toggleReveal(env.key)}
                        >
                          {revealedSecrets.has(env.key)
                            ? <EyeOff className="h-3.5 w-3.5" />
                            : <Eye className="h-3.5 w-3.5" />}
                        </button>
                      )}
                      <button className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground">
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-red-500">
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Config files table */}
      {activeSection === 'files' && (
        <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
          <table className="w-full text-left text-[13px]">
            <thead>
              <tr className="border-b border-border/50 bg-accent/30">
                <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">File Name</th>
                <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Size</th>
                <th className="w-20 px-4 py-2.5"></th>
              </tr>
            </thead>
            <tbody>
              {files.map((file) => (
                <tr key={file.name} className="h-12 border-b border-border/30 transition-colors hover:bg-accent/20">
                  <td className="px-4">
                    <div className="flex items-center gap-2">
                      <FileIcon className="h-4 w-4 text-muted-foreground/50" />
                      <span className="font-mono text-[12px] font-medium text-foreground">{file.name}</span>
                    </div>
                  </td>
                  <td className="px-4 text-[12px] text-muted-foreground">{file.size}</td>
                  <td className="px-4">
                    <div className="flex items-center justify-end gap-1">
                      <button className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground">
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-red-500">
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

// Snapshots — uses shared SnapshotTree component
function SnapshotsView({ envHash, envName }: { envHash: string; envName: string }) {
  const snapshots = generateSnapshots(envHash)
  return <SnapshotTree snapshots={snapshots} title="Snapshots" subtitle={`${snapshots.length} snapshots for ${envName}`} />
}

function SettingsView({ envName, envHash }: { envName: string; envHash: string }) {
  return (
    <div className="p-6">
      <h2 className="text-[16px] font-semibold text-foreground">Settings</h2>
      <p className="mt-1 text-[13px] text-muted-foreground">Environment configuration</p>

      <div className="mt-5 flex flex-col gap-5">
        {/* General */}
        <div className="rounded-xl border border-border/50 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-foreground">General</h3>
          <div className="mt-3 flex flex-col gap-3 text-[13px]">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Name</span>
              <span className="font-medium text-foreground">{envName}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Hash</span>
              <span className="font-mono text-[12px] text-muted-foreground">{envHash}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Namespace</span>
              <span className="font-mono text-[12px] text-muted-foreground">env-{envHash}</span>
            </div>
          </div>
        </div>

        {/* Resource Quotas */}
        <div className="rounded-xl border border-border/50 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-foreground">Resource Quotas</h3>
          <div className="mt-3 flex flex-col gap-3 text-[13px]">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">CPU Limit</span>
              <span className="font-medium text-foreground">4 cores</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Memory Limit</span>
              <span className="font-medium text-foreground">8 GiB</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Storage</span>
              <span className="font-medium text-foreground">20 GiB</span>
            </div>
          </div>
        </div>

        {/* Danger Zone */}
        <div className="rounded-xl border border-red-500/20 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-red-500">Danger Zone</h3>
          <p className="mt-1 text-[12px] text-muted-foreground">These actions are destructive and cannot be undone.</p>
          <div className="mt-3 flex gap-2">
            <button className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20">
              Deactivate Environment
            </button>
            <button className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20">
              Delete Environment
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}


function CompositionView({ envHash }: { envHash: string }) {
  const [compose, setCompose] = useState(COMPOSITIONS[envHash] || '')
  const [saved, setSaved] = useState(false)

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Composition</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">Docker Compose definition for this environment</p>
        </div>
        <button
          className="rounded-lg bg-primary px-4 py-2 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90"
          onClick={() => {
            setSaved(true)
            setTimeout(() => setSaved(false), 2000)
          }}
        >
          {saved ? 'Saved!' : 'Apply Changes'}
        </button>
      </div>

      <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
        <textarea
          className="h-[400px] w-full resize-none bg-card p-4 font-mono text-[12px] leading-relaxed text-foreground outline-none"
          value={compose}
          onChange={(e) => setCompose(e.target.value)}
          spellCheck={false}
        />
      </div>
    </div>
  )
}

// ---------- New Environment Dialog ----------

export function NewEnvironmentDialog({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [visibility, setVisibility] = useState<'private' | 'shared' | 'open'>('private')
  const [exiting, setExiting] = useState(false)

  function close() {
    setExiting(true)
    setTimeout(onClose, 150)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={close}>
      <div
        className="w-full max-w-md overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl"
        style={{ animation: exiting ? 'popover-out 150ms ease-in forwards' : 'popover-in 150ms ease-out' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="border-b border-border/30 px-6 py-4">
          <h2 className="text-[16px] font-semibold text-foreground">Create Environment</h2>
          <p className="mt-0.5 text-[12px] text-muted-foreground">Set up a new isolated environment</p>
        </div>

        <div className="flex flex-col gap-4 px-6 py-5">
          <div>
            <label className="mb-1.5 block text-[12px] font-medium text-foreground">Name</label>
            <input
              type="text"
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-[13px] text-foreground outline-none transition-colors focus:border-primary"
              placeholder="e.g. staging, development"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
            />
          </div>

          <div>
            <label className="mb-1.5 block text-[12px] font-medium text-foreground">Visibility</label>
            <div className="flex gap-2">
              {(['private', 'shared', 'open'] as const).map((v) => (
                <button
                  key={v}
                  className={cn(
                    'flex-1 rounded-lg border px-3 py-2 text-[12px] font-medium capitalize transition-colors',
                    visibility === v
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border text-muted-foreground hover:bg-accent'
                  )}
                  onClick={() => setVisibility(v)}
                >
                  {v}
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-border/30 px-6 py-4">
          <button
            className="rounded-lg px-4 py-2 text-[12px] font-medium text-muted-foreground transition-colors hover:bg-accent"
            onClick={close}
          >
            Cancel
          </button>
          <button
            className="rounded-lg bg-primary px-4 py-2 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            disabled={!name.trim()}
            onClick={close}
          >
            Create Environment
          </button>
        </div>
      </div>
    </div>
  )
}

export function EnvironmentContent({ envName, envHash, activeTab }: EnvironmentContentProps) {
  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto max-w-4xl">
        {activeTab === 'services' && <ServicesView envHash={envHash} />}
        {activeTab === 'configs' && <ConfigsView envHash={envHash} />}
        {activeTab === 'snapshots' && <SnapshotsView envHash={envHash} envName={envName} />}
        {activeTab === 'settings' && <SettingsView envName={envName} envHash={envHash} />}
      </div>
    </div>
  )
}
