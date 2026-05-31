import { cn } from '@/lib/utils'
import { Copy, Check, Pencil, Trash2, Eye, EyeOff, Key, FileText as FileIcon, Loader2, RefreshCw } from 'lucide-react'
import { useState, useEffect } from 'react'
import { CodeEditor } from './code-editor'
import { SnapshotTree, generateSnapshots } from './snapshot-tree'
import { ServicesGraph } from './services-graph'
import { LogsViewer } from './services-graph/logs-viewer'
import { parseComposeServices } from '../lib/compose-services'
import { buildConfigMap, buildEnvSecret, buildFileConfigMap, ENV_CONFIG_NAME, ENV_SECRET_NAME, filenamePattern } from '../lib/config-resource-helpers'

const API_NAMESPACE = 'wm-karthik-dev'

interface EnvironmentContentProps {
  envName: string
  envHash: string
  activeTab: string
  onDeleted?: () => void
}

// Dummy services data — port-level intercepts + volumes
interface ServicePort {
  port: number
  targetPort: number
  protocol: string
  interceptedBy?: string  // workspace id if this port is intercepted
}
interface ServiceVolume {
  name: string
  mountPath: string
  type: 'persistent' | 'config' | 'secret' | 'host'
}
interface ServiceData {
  id: string
  name: string
  type: 'ClusterIP' | 'LoadBalancer' | 'NodePort'
  clusterIP: string
  ports: ServicePort[]
  volumes: ServiceVolume[]
  dns: string
}
const SERVICES: Record<string, ServiceData[]> = {
  'a1b2c3': [
    { id: 'frontend', name: 'frontend', type: 'ClusterIP', clusterIP: '10.96.45.12', dns: 'frontend-a1b2c3.staging.local',
      ports: [{ port: 3000, targetPort: 3000, protocol: 'TCP' }],
      volumes: [{ name: 'static-assets', mountPath: '/usr/share/nginx/html', type: 'config' }] },
    { id: 'api-server', name: 'api-server', type: 'ClusterIP', clusterIP: '10.96.45.13', dns: 'api-server-a1b2c3.staging.local',
      ports: [
        { port: 8080, targetPort: 8080, protocol: 'TCP', interceptedBy: 'ws-1' },
        { port: 9090, targetPort: 9090, protocol: 'TCP' },
        { port: 50051, targetPort: 50051, protocol: 'TCP' },
      ],
      volumes: [
        { name: 'app-config', mountPath: '/etc/config', type: 'config' },
        { name: 'tls-certs', mountPath: '/etc/tls', type: 'secret' },
      ] },
    { id: 'redis', name: 'redis', type: 'ClusterIP', clusterIP: '10.96.45.14', dns: 'redis-a1b2c3.staging.local',
      ports: [{ port: 6379, targetPort: 6379, protocol: 'TCP' }],
      volumes: [{ name: 'redis-data', mountPath: '/data', type: 'persistent' }] },
    { id: 'postgres', name: 'postgres', type: 'ClusterIP', clusterIP: '10.96.45.15', dns: 'postgres-a1b2c3.staging.local',
      ports: [{ port: 5432, targetPort: 5432, protocol: 'TCP', interceptedBy: 'ws-3' }],
      volumes: [
        { name: 'pg-data', mountPath: '/var/lib/postgresql/data', type: 'persistent' },
        { name: 'pg-credentials', mountPath: '/etc/secrets', type: 'secret' },
      ] },
  ],
  'd4e5f6': [
    { id: 'web-app', name: 'web-app', type: 'ClusterIP', clusterIP: '10.96.50.10', dns: 'web-app-d4e5f6.dev.local',
      ports: [
        { port: 5173, targetPort: 5173, protocol: 'TCP', interceptedBy: 'ws-2' },
        { port: 24678, targetPort: 24678, protocol: 'TCP' },
      ],
      volumes: [{ name: 'src', mountPath: '/app/src', type: 'host' }] },
    { id: 'auth-service', name: 'auth-service', type: 'ClusterIP', clusterIP: '10.96.50.11', dns: 'auth-d4e5f6.dev.local',
      ports: [{ port: 9090, targetPort: 9090, protocol: 'TCP' }],
      volumes: [] },
  ],
  'g7h8i9': [
    { id: 'gateway', name: 'gateway', type: 'LoadBalancer', clusterIP: '10.96.60.10', dns: 'gateway-g7h8i9.prod.local',
      ports: [
        { port: 443, targetPort: 8443, protocol: 'TCP' },
        { port: 80, targetPort: 8080, protocol: 'TCP' },
      ],
      volumes: [
        { name: 'tls-cert', mountPath: '/etc/ssl/certs', type: 'secret' },
        { name: 'gateway-config', mountPath: '/etc/gateway', type: 'config' },
      ] },
    { id: 'dashboard', name: 'dashboard', type: 'ClusterIP', clusterIP: '10.96.60.11', dns: 'dashboard-g7h8i9.prod.local',
      ports: [{ port: 3000, targetPort: 3000, protocol: 'TCP' }],
      volumes: [] },
  ],
}

// Workspaces connected to each environment
interface ConnectedWorkspace {
  id: string
  name: string
  owner: string
  status: 'running' | 'stopped' | 'failed'
}
const ENV_WORKSPACES: Record<string, ConnectedWorkspace[]> = {
  'a1b2c3': [
    { id: 'ws-1', name: 'api-dev', owner: 'karthik', status: 'running' },
    { id: 'ws-3', name: 'debug-session', owner: 'sohail', status: 'stopped' },
  ],
  'd4e5f6': [
    { id: 'ws-2', name: 'frontend-dev', owner: 'karthik', status: 'running' },
  ],
  'g7h8i9': [],
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
      - POSTGRES_PASSWORD=<set-in-secret>`,
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

function ServicesView({ envHash, envName }: { envHash: string; envName: string }) {
  const [services, setServices] = useState<ServiceData[]>([])
  const [workspaces, setWorkspaces] = useState<ConnectedWorkspace[]>([])
  const [compose, setCompose] = useState('')
  const [loading, setLoading] = useState(true)
  const [refreshKey, setRefreshKey] = useState(0)
  const [composeOpen, setComposeOpen] = useState(false)
  const [composeExiting, setComposeExiting] = useState(false)
  const [saved, setSaved] = useState(false)
  const [saving, setSaving] = useState(false)
  const [logsService, setLogsService] = useState<string | null>(null)

  // Fetch environment data from API
  useEffect(() => {
    setLoading(true)

    window.electronAPI.listEnvironments(API_NAMESPACE).then((result) => {
      if (result.error) return
      const env = (result.items || []).find((e: any) =>
        e.metadata?.name === envName || e.metadata?.name === envHash || e.metadata?.labels?.['kloudlite.io/environment-name'] === envName
      )
      if (!env) {
        setServices(SERVICES[envHash] || [])
        setWorkspaces(ENV_WORKSPACES[envHash] || [])
        setLoading(false)
        return
      }
      // Parse services from spec.compose.composeContent (source of truth)
      const cc = env.spec?.compose?.composeContent
      if (cc) {
        const parsed = parseComposeServices(cc, env.spec?.targetNamespace || '')
        if (parsed.length > 0) setServices(parsed)
      }
      if (cc && !composeOpen) setCompose(cc)
      setWorkspaces(ENV_WORKSPACES[envHash] || [])
      setLoading(false)
    }).catch(() => {
      setServices(SERVICES[envHash] || [])
      setWorkspaces(ENV_WORKSPACES[envHash] || [])
      setCompose(COMPOSITIONS[envHash] || '')
      setLoading(false)
    })
  }, [envHash, envName, refreshKey])

  useEffect(() => {
    function handler(e: Event) {
      const detail = (e as CustomEvent).detail
      setLogsService(detail.name)
    }
    window.addEventListener('open-service-logs', handler)
    return () => window.removeEventListener('open-service-logs', handler)
  }, [])

  function closeCompose(immediate?: boolean) {
    if (immediate) {
      setComposeOpen(false)
      setComposeExiting(false)
      return
    }
    setComposeExiting(true)
    setTimeout(() => {
      setComposeOpen(false)
      setComposeExiting(false)
    }, 150)
  }

  if (loading && services.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground/50" />
      </div>
    )
  }

  const graphServices = services.map((s) => ({
    id: s.id,
    name: s.name,
    dns: s.dns,
    type: s.type,
    ports: s.ports.map((p) => ({ port: p.port, targetPort: p.targetPort, protocol: p.protocol, interceptedBy: p.interceptedBy })),
    volumes: s.volumes,
  }))

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex shrink-0 items-center justify-between px-6 pt-6 pb-4">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Services</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">
            {services.length} services · {workspaces.length} connected workspace{workspaces.length !== 1 ? 's' : ''}
          </p>
        </div>
        <button
          className={cn(
            'rounded-lg border px-3 py-1.5 text-[12px] font-medium transition-colors',
            composeOpen
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-accent'
          )}
          onClick={async () => {
            if (composeOpen) { closeCompose(); return }
            // Always fetch latest compose from API when opening
            try {
              const result = await window.electronAPI.listEnvironments(API_NAMESPACE)
              const env = (result?.items || []).find((e: any) =>
                e.metadata?.name === envName || e.metadata?.labels?.['kloudlite.io/environment-name'] === envName
              )
              const cc = env?.spec?.compose?.composeContent
              if (cc) setCompose(cc)
            } catch {}
            setComposeOpen(true)
          }}
        >
          Composition
        </button>
      </div>

      {/* Composition editor */}
      {composeOpen && (
        <div className="mx-6 mb-4 overflow-hidden rounded-xl border border-border/50" style={{ animation: composeExiting ? 'popover-out 150ms ease-in forwards' : 'popover-in 150ms ease-out' }}>
          <div className="flex items-center justify-between border-b border-border/30 bg-card px-4 py-2">
            <span className="text-[11px] font-medium text-muted-foreground">docker-compose.yml</span>
            <div className="flex items-center gap-2">
              <button
                className="rounded-md px-3 py-1 text-[11px] font-medium text-muted-foreground transition-colors hover:bg-accent"
                onClick={() => closeCompose()}
              >
                Cancel
              </button>
              <button
                className="rounded-md bg-primary px-3 py-1 text-[11px] font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                disabled={saving}
                onClick={async () => {
                  if (saving || !compose.trim()) return
                  setSaving(true)
                  try {
                    await window.electronAPI.patchResource(
                      'wm-karthik-dev',
                      'environments',
                      envName,
                      {
                        apiVersion: 'environments.kloudlite.io/v1',
                        kind: 'Environment',
                        metadata: { name: envName, namespace: 'wm-karthik-dev' },
                        spec: {
                          compose: {
                            displayName: 'Compose App',
                            composeContent: compose,
                            composeFormat: 'v3.8'
                          }
                        }
                      }
                    )
                    closeCompose(true)
                    setRefreshKey((k) => k + 1)
                  } catch {
                  } finally {
                    setSaving(false)
                  }
                }}
              >
                {saving ? 'Saving...' : saved ? 'Saved!' : 'Apply'}
              </button>
            </div>
          </div>
          <CodeEditor value={compose} onChange={setCompose} height="300px" />
        </div>
      )}

      {/* Graph fills remaining space */}
      <div className="min-h-0 flex-1 relative" key={`svc-${services.length}-${services.map(s => s.id).join('-')}`}>
        {services.length > 0 && (
          <div className="absolute top-2 left-2 z-10 rounded-lg border border-border bg-background/90 px-3 py-1.5 text-[11px] text-foreground backdrop-blur-sm">
            {services.length} service{services.length !== 1 ? 's' : ''}: {services.map(s => s.name).join(', ')}
          </div>
        )}
        <ServicesGraph
          key={'g-' + services.map(s => s.id).join('-') + '-p' + services.map(s => s.ports.map(p => `${p.port}-${p.interceptedBy || ''}`).join(',')).join('-')}
          services={graphServices}
          workspaces={workspaces}
        />
      </div>

      {/* Logs viewer */}
      {logsService && <LogsViewer serviceName={logsService} onClose={() => setLogsService(null)} />}
    </div>
  )
}

interface EnvVarRow {
  id: string
  key: string
  value: string
  type: 'config' | 'secret'
  source: string
}

interface ConfigFileRow {
  id: string
  name: string
  configMapName: string
  source: string
  content: string
}

type ConfigDialogState =
  | { kind: 'env'; mode: 'create' | 'edit'; item?: EnvVarRow }
  | { kind: 'file'; mode: 'create' | 'edit'; item?: ConfigFileRow }
  | null

function decodeSecretValue(value: unknown): string {
  if (typeof value !== 'string') return ''
  try {
    return atob(value)
  } catch {
    return value
  }
}

function readStringMap(value: unknown): Record<string, string> {
  if (!value || typeof value !== 'object') return {}
  return Object.fromEntries(
    Object.entries(value as Record<string, unknown>)
      .filter(([, v]) => typeof v === 'string')
      .map(([k, v]) => [k, v as string])
  )
}

function getLabels(item: Record<string, unknown>): Record<string, string> {
  return readStringMap((item as any).metadata?.labels)
}

function isEnvConfigMap(item: Record<string, unknown>): boolean {
  const labels = getLabels(item)
  const name = (item as any).metadata?.name
  return name === 'env-config'
    || labels['kloudlite.io/config-type'] === 'envvars'
    || labels['kloudlite.io/resource-type'] === 'config'
}

function isEnvSecret(item: Record<string, unknown>): boolean {
  const labels = getLabels(item)
  const name = (item as any).metadata?.name
  return (name === 'env-secret'
    || labels['kloudlite.io/config-type'] === 'envvars'
    || labels['kloudlite.io/resource-type'] === 'secret')
    && (item as any).type !== 'kubernetes.io/service-account-token'
}

function isConfigFileMap(item: Record<string, unknown>): boolean {
  return getLabels(item)['kloudlite.io/resource-type'] === 'file'
}

function mapConfigObjects(configmaps: Record<string, unknown>[], secrets: Record<string, unknown>[]) {
  const envvars: EnvVarRow[] = []
  const files: ConfigFileRow[] = []

  for (const item of configmaps.filter(isEnvConfigMap)) {
    const name = (item as any).metadata?.name || 'configmap'
    const data = readStringMap((item as any).data)
    for (const [key, value] of Object.entries(data)) {
      envvars.push({ id: `config-${name}-${key}`, key, value, type: 'config', source: name })
    }
  }

  for (const item of secrets.filter(isEnvSecret)) {
    const name = (item as any).metadata?.name || 'secret'
    const data = readStringMap((item as any).data)
    const stringData = readStringMap((item as any).stringData)
    const keys = new Set([...Object.keys(data), ...Object.keys(stringData)])
    for (const key of keys) {
      const value = key in stringData ? stringData[key] : decodeSecretValue(data[key])
      envvars.push({ id: `secret-${name}-${key}`, key, value, type: 'secret', source: name })
    }
  }

  for (const item of configmaps.filter(isConfigFileMap)) {
    const labels = getLabels(item)
    const configMapName = (item as any).metadata?.name || 'config-file'
    const filename = labels['kloudlite.io/filename'] || configMapName
    const data = readStringMap((item as any).data)
    const content = data.content || Object.values(data)[0] || ''
    files.push({
      id: `file-${configMapName}`,
      name: filename,
      configMapName,
      source: configMapName,
      content,
    })
  }

  return { envvars, files }
}

function ConfigEditorDialog({
  dialog,
  saving,
  error,
  onClose,
  onSaveEnv,
  onSaveFile,
}: {
  dialog: NonNullable<ConfigDialogState>
  saving: boolean
  error: string | null
  onClose: () => void
  onSaveEnv: (input: { key: string; value: string; type: 'config' | 'secret'; original?: EnvVarRow }) => void
  onSaveFile: (input: { filename: string; content: string; original?: ConfigFileRow }) => void
}) {
  const envItem = dialog.kind === 'env' ? dialog.item : undefined
  const fileItem = dialog.kind === 'file' ? dialog.item : undefined
  const [key, setKey] = useState(envItem?.key || '')
  const [value, setValue] = useState(envItem?.value || '')
  const [type, setType] = useState<'config' | 'secret'>(envItem?.type || 'config')
  const [filename, setFilename] = useState(fileItem?.name || '')
  const [content, setContent] = useState(fileItem?.content || '')

  const isEnv = dialog.kind === 'env'
  const filenameValid = filenamePattern.test(filename.trim())
  const canSubmit = isEnv ? Boolean(key.trim() && value.length > 0) : Boolean(filename.trim() && filenameValid)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={onClose}>
      <div className="w-full max-w-lg overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="border-b border-border/30 px-5 py-4">
          <h3 className="text-[15px] font-semibold text-foreground">
            {dialog.mode === 'create' ? 'Add' : 'Edit'} {isEnv ? 'environment variable' : 'config file'}
          </h3>
          <p className="mt-0.5 text-[12px] text-muted-foreground">
            {isEnv ? 'Used by compose during interpolation and service env injection.' : 'Mounted into compose services with /files/<filename> references.'}
          </p>
        </div>

        <div className="flex flex-col gap-4 px-5 py-4">
          {error && <div className="rounded-lg border border-red-500/20 bg-red-500/[0.04] px-3 py-2 text-[12px] text-red-500">{error}</div>}

          {isEnv ? (
            <>
              <div>
                <label className="mb-1.5 block text-[12px] font-medium text-foreground">Key</label>
                <input
                  className="w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-[13px] outline-none focus:border-primary"
                  value={key}
                  onChange={(e) => setKey(e.target.value)}
                  placeholder="DATABASE_URL"
                  autoFocus
                />
              </div>
              <div>
                <label className="mb-1.5 block text-[12px] font-medium text-foreground">Value</label>
                <textarea
                  className="h-24 w-full resize-none rounded-lg border border-border bg-background px-3 py-2 font-mono text-[13px] outline-none focus:border-primary"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="value"
                />
              </div>
              <div>
                <label className="mb-1.5 block text-[12px] font-medium text-foreground">Type</label>
                <div className="flex gap-2">
                  {(['config', 'secret'] as const).map((option) => (
                    <button
                      key={option}
                      className={cn(
                        'flex-1 rounded-lg border px-3 py-2 text-[12px] font-medium capitalize transition-colors',
                        type === option ? 'border-primary bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:bg-accent'
                      )}
                      onClick={() => setType(option)}
                    >
                      {option}
                    </button>
                  ))}
                </div>
              </div>
            </>
          ) : (
            <>
              <div>
                <label className="mb-1.5 block text-[12px] font-medium text-foreground">Filename</label>
                <input
                  className="w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-[13px] outline-none focus:border-primary"
                  value={filename}
                  onChange={(e) => setFilename(e.target.value)}
                  placeholder="app-config.yaml"
                  autoFocus
                />
                {!filenameValid && filename.trim() && (
                  <p className="mt-1 text-[11px] text-red-500">Use 1-63 label-safe characters. Start/end with a letter or number; use letters, numbers, dots, underscores, or dashes.</p>
                )}
              </div>
              <div>
                <label className="mb-1.5 block text-[12px] font-medium text-foreground">Content</label>
                <textarea
                  className="h-52 w-full resize-none rounded-lg border border-border bg-background px-3 py-2 font-mono text-[12px] leading-relaxed outline-none focus:border-primary"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="file contents"
                />
              </div>
            </>
          )}
        </div>

        <div className="flex justify-end gap-2 border-t border-border/30 px-5 py-4">
          <button className="rounded-lg px-4 py-2 text-[12px] font-medium text-muted-foreground transition-colors hover:bg-accent" onClick={onClose} disabled={saving}>
            Cancel
          </button>
          <button
            className="rounded-lg bg-primary px-4 py-2 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            disabled={!canSubmit || saving}
            onClick={() => {
              if (isEnv) onSaveEnv({ key: key.trim(), value, type, original: envItem })
              else onSaveFile({ filename: filename.trim(), content, original: fileItem })
            }}
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}

function ConfigsView({ envHash, envName }: { envHash: string; envName: string }) {
  const [activeSection, setActiveSection] = useState<'envvars' | 'files'>('envvars')
  const [revealedSecrets, setRevealedSecrets] = useState<Set<string>>(new Set())
  const [envvars, setEnvvars] = useState<EnvVarRow[]>([])
  const [files, setFiles] = useState<ConfigFileRow[]>([])
  const [configmaps, setConfigmaps] = useState<Record<string, unknown>[]>([])
  const [secrets, setSecrets] = useState<Record<string, unknown>[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [mutationError, setMutationError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [dialog, setDialog] = useState<ConfigDialogState>(null)
  const [refreshKey, setRefreshKey] = useState(0)
  const [targetNamespace, setTargetNamespace] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    window.electronAPI.listEnvironments(API_NAMESPACE).then((envResult) => {
      if (cancelled) return
      if (envResult.error) throw new Error(envResult.error)
      const env = (envResult.items || []).find((e: any) =>
        e.metadata?.name === envName || e.metadata?.name === envHash || e.metadata?.labels?.['kloudlite.io/environment-name'] === envName
      )
      const namespace = (env as any)?.spec?.targetNamespace
      if (!namespace) throw new Error('Environment target namespace not found')
      setTargetNamespace(namespace)
      return Promise.all([
        window.electronAPI.listResources(namespace, 'configmaps'),
        window.electronAPI.listResources(namespace, 'secrets'),
      ])
    }).then((result) => {
      if (cancelled) return
      if (!result) return
      const [configmapsResult, secretsResult] = result
      if (configmapsResult.error || secretsResult.error) {
        setError(configmapsResult.error || secretsResult.error || 'Failed to load configs and secrets')
        setEnvvars([])
        setFiles([])
        setConfigmaps([])
        setSecrets([])
        return
      }

      const mapped = mapConfigObjects(configmapsResult.items || [], secretsResult.items || [])
      setConfigmaps(configmapsResult.items || [])
      setSecrets(secretsResult.items || [])
      setEnvvars(mapped.envvars)
      setFiles(mapped.files)
    }).catch((err) => {
      if (cancelled) return
      setError((err as Error).message)
      setEnvvars([])
      setFiles([])
      setConfigmaps([])
      setSecrets([])
    }).finally(() => {
      if (!cancelled) setLoading(false)
    })

    return () => { cancelled = true }
  }, [envHash, envName, refreshKey])

  function toggleReveal(key: string) {
    setRevealedSecrets((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  function openCreateDialog() {
    setMutationError(null)
    setDialog(activeSection === 'envvars'
      ? { kind: 'env', mode: 'create' }
      : { kind: 'file', mode: 'create' })
  }

  function dataForConfigMap(name: string): Record<string, string> {
    const item = configmaps.find((cm: any) => cm.metadata?.name === name)
    return readStringMap((item as any)?.data)
  }

  function dataForSecret(name: string): Record<string, string> {
    const item = secrets.find((secret: any) => secret.metadata?.name === name)
    const decoded: Record<string, string> = {}
    for (const [key, value] of Object.entries(readStringMap((item as any)?.data))) {
      decoded[key] = decodeSecretValue(value)
    }
    return decoded
  }

  async function upsertResource(resource: 'configmaps' | 'secrets', name: string, object: Record<string, unknown>) {
    if (!targetNamespace) throw new Error('Target namespace not loaded')
    const exists = resource === 'configmaps'
      ? configmaps.some((cm: any) => cm.metadata?.name === name)
      : secrets.some((secret: any) => secret.metadata?.name === name)

    const result = exists
      ? await window.electronAPI.patchResource(targetNamespace, resource, name, object)
      : await window.electronAPI.createResource(targetNamespace, resource, object)
    if ((result as any).error) throw new Error((result as any).error)
  }

  async function saveEnvVar(input: { key: string; value: string; type: 'config' | 'secret'; original?: EnvVarRow }) {
    if (!targetNamespace) throw new Error('Target namespace not loaded')
    const old = input.original

    if (old && old.type === input.type) {
      const sourceName = old.source || (input.type === 'config' ? ENV_CONFIG_NAME : ENV_SECRET_NAME)
      const exists = input.type === 'config'
        ? configmaps.some((cm: any) => cm.metadata?.name === sourceName)
        : secrets.some((secret: any) => secret.metadata?.name === sourceName)

      if (!exists) {
        if (input.type === 'config') await upsertResource('configmaps', sourceName, buildConfigMap(targetNamespace, { [input.key]: input.value }))
        else await upsertResource('secrets', sourceName, buildEnvSecret(targetNamespace, { [input.key]: input.value }))
        return
      }

      const dataPatch: Record<string, string | null> = { [input.key]: input.value }
      if (old.key !== input.key) dataPatch[old.key] = null

      const result = input.type === 'config'
        ? await window.electronAPI.patchResource(targetNamespace, 'configmaps', sourceName, {
          apiVersion: 'v1',
          kind: 'ConfigMap',
          metadata: { name: sourceName, namespace: targetNamespace },
          data: dataPatch,
        })
        : await window.electronAPI.patchResource(targetNamespace, 'secrets', sourceName, {
          apiVersion: 'v1',
          kind: 'Secret',
          type: 'Opaque',
          metadata: { name: sourceName, namespace: targetNamespace },
          data: old.key !== input.key ? { [old.key]: null } : undefined,
          stringData: { [input.key]: input.value },
        })

      if ((result as any).error) throw new Error((result as any).error)
      return
    }

    if (old && old.type !== input.type) {
      await deleteEnvVar(old)
    }

    if (input.type === 'config') {
      const data = { ...dataForConfigMap(ENV_CONFIG_NAME), [input.key]: input.value }
      await upsertResource('configmaps', ENV_CONFIG_NAME, buildConfigMap(targetNamespace, data))
      return
    }

    const data = { ...dataForSecret(ENV_SECRET_NAME), [input.key]: input.value }
    await upsertResource('secrets', ENV_SECRET_NAME, buildEnvSecret(targetNamespace, data))
  }

  async function deleteEnvVar(item: EnvVarRow) {
    if (!targetNamespace) throw new Error('Target namespace not loaded')
    if (item.type === 'config') {
      const sourceName = item.source || ENV_CONFIG_NAME
      const data = { ...dataForConfigMap(sourceName) }
      delete data[item.key]
      if (Object.keys(data).length === 0) {
        const result = await window.electronAPI.deleteResource(targetNamespace, 'configmaps', sourceName)
        if (result.error) throw new Error(result.error)
        return
      }

      const result = await window.electronAPI.patchResource(targetNamespace, 'configmaps', sourceName, {
        apiVersion: 'v1',
        kind: 'ConfigMap',
        metadata: { name: sourceName, namespace: targetNamespace },
        data: { [item.key]: null },
      })
      if ((result as any).error) throw new Error((result as any).error)
      return
    }

    const sourceName = item.source || ENV_SECRET_NAME
    const data = { ...dataForSecret(sourceName) }
    delete data[item.key]
    if (Object.keys(data).length === 0) {
      const result = await window.electronAPI.deleteResource(targetNamespace, 'secrets', sourceName)
      if (result.error) throw new Error(result.error)
      return
    }

    const result = await window.electronAPI.patchResource(targetNamespace, 'secrets', sourceName, {
      apiVersion: 'v1',
      kind: 'Secret',
      type: 'Opaque',
      metadata: { name: sourceName, namespace: targetNamespace },
      data: { [item.key]: null },
    })
    if ((result as any).error) throw new Error((result as any).error)
  }

  async function saveFile(input: { filename: string; content: string; original?: ConfigFileRow }) {
    if (!targetNamespace) throw new Error('Target namespace not loaded')
    const object = buildFileConfigMap(targetNamespace, input.filename, input.content)
    const name = (object.metadata as { name: string }).name
    const originalName = input.original?.configMapName

    if (originalName && originalName !== name) {
      const deleted = await window.electronAPI.deleteResource(targetNamespace, 'configmaps', originalName)
      if (deleted.error) throw new Error(deleted.error)
    }

    await upsertResource('configmaps', name, object)
  }

  async function deleteFile(item: ConfigFileRow) {
    if (!targetNamespace) throw new Error('Target namespace not loaded')
    const result = await window.electronAPI.deleteResource(targetNamespace, 'configmaps', item.configMapName)
    if (result.error) throw new Error(result.error)
  }

  async function runMutation(fn: () => Promise<void>) {
    setSaving(true)
    setMutationError(null)
    try {
      await fn()
      setDialog(null)
      setRefreshKey((k) => k + 1)
    } catch (err) {
      setMutationError((err as Error).message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Configs & Secrets</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">
            Compose environment variables and files{targetNamespace ? ` from ${targetNamespace}` : ''}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="rounded-lg bg-primary px-3 py-1.5 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            onClick={openCreateDialog}
            disabled={loading || !targetNamespace}
          >
            Add {activeSection === 'envvars' ? 'Variable' : 'File'}
          </button>
          <button
            className="flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-[12px] font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            onClick={() => setRefreshKey((k) => k + 1)}
            disabled={loading}
          >
            <RefreshCw className={cn('h-3.5 w-3.5', loading && 'animate-spin')} />
            Refresh
          </button>
        </div>
      </div>

      {mutationError && (
        <div className="mt-4 rounded-xl border border-red-500/20 bg-red-500/[0.04] px-4 py-3 text-[12px] text-red-500">
          {mutationError}
        </div>
      )}

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

      {error && (
        <div className="mt-4 flex items-center justify-between rounded-xl border border-red-500/20 bg-red-500/[0.04] px-4 py-3">
          <p className="text-[12px] text-red-500">{error}</p>
          <button className="text-[12px] font-medium text-red-500 hover:underline" onClick={() => setRefreshKey((k) => k + 1)}>Retry</button>
        </div>
      )}

      {loading && (
        <div className="mt-4 flex h-32 items-center justify-center rounded-xl border border-border/50">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground/50" />
        </div>
      )}

      {/* Envvars table */}
      {!loading && !error && activeSection === 'envvars' && (
        <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
          {envvars.length === 0 ? (
            <div className="px-4 py-10 text-center text-[13px] text-muted-foreground">No ConfigMap or Secret keys found.</div>
          ) : (
            <table className="w-full text-left text-[13px]">
              <thead>
                <tr className="border-b border-border/50 bg-accent/30">
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Key</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Value</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Source</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Type</th>
                  <th className="w-20 px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody>
                {envvars.map((env) => (
                  <tr key={env.id} className="h-12 border-b border-border/30 transition-colors hover:bg-accent/20">
                    <td className="px-4">
                      <span className="font-mono text-[12px] font-medium text-foreground">{env.key}</span>
                    </td>
                    <td className="max-w-[320px] px-4">
                      <span className="block truncate font-mono text-[12px] text-muted-foreground">
                        {env.type === 'secret' && !revealedSecrets.has(env.id)
                          ? '••••••••••••'
                          : env.value}
                      </span>
                    </td>
                    <td className="px-4">
                      <span className="font-mono text-[11px] text-muted-foreground/70">{env.source}</span>
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
                            onClick={() => toggleReveal(env.id)}
                          >
                            {revealedSecrets.has(env.id)
                              ? <EyeOff className="h-3.5 w-3.5" />
                              : <Eye className="h-3.5 w-3.5" />}
                          </button>
                        )}
                        <button
                          className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground"
                          onClick={() => {
                            setMutationError(null)
                            setDialog({ kind: 'env', mode: 'edit', item: env })
                          }}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button
                          className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-red-500"
                          onClick={() => runMutation(() => deleteEnvVar(env))}
                          disabled={saving}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Config files table */}
      {!loading && !error && activeSection === 'files' && (
        <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
          {files.length === 0 ? (
            <div className="px-4 py-10 text-center text-[13px] text-muted-foreground">No file-like config keys found.</div>
          ) : (
            <table className="w-full text-left text-[13px]">
              <thead>
                <tr className="border-b border-border/50 bg-accent/30">
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">File Name</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">ConfigMap</th>
                  <th className="w-20 px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody>
                {files.map((file) => (
                  <tr key={file.id} className="h-12 border-b border-border/30 transition-colors hover:bg-accent/20">
                    <td className="px-4">
                      <div className="flex items-center gap-2">
                        <FileIcon className="h-4 w-4 text-muted-foreground/50" />
                        <span className="font-mono text-[12px] font-medium text-foreground">{file.name}</span>
                      </div>
                    </td>
                    <td className="px-4 font-mono text-[11px] text-muted-foreground/70">{file.configMapName}</td>
                    <td className="px-4">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground"
                          onClick={() => {
                            setMutationError(null)
                            setDialog({ kind: 'file', mode: 'edit', item: file })
                          }}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button
                          className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-red-500"
                          onClick={() => runMutation(() => deleteFile(file))}
                          disabled={saving}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {dialog && (
        <ConfigEditorDialog
          dialog={dialog}
          saving={saving}
          error={mutationError}
          onClose={() => {
            if (!saving) {
              setDialog(null)
              setMutationError(null)
            }
          }}
          onSaveEnv={(input) => runMutation(() => saveEnvVar(input))}
          onSaveFile={(input) => runMutation(() => saveFile(input))}
        />
      )}
    </div>
  )
}

// Snapshots — uses shared SnapshotTree component
function SnapshotsView({ envHash, envName }: { envHash: string; envName: string }) {
  const snapshots = generateSnapshots(envHash)
  return <SnapshotTree snapshots={snapshots} title="Snapshots" subtitle={`${snapshots.length} snapshots for ${envName}`} />
}

function SettingsView({ envName, envHash, onDeleted }: { envName: string; envHash: string; onDeleted?: () => void }) {
  const [deactivating, setDeactivating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [settingsError, setSettingsError] = useState<string | null>(null)

  async function deactivate() {
    setDeactivating(true)
    setSettingsError(null)
    try {
      const result = await window.electronAPI.patchResource(API_NAMESPACE, 'environments', envName, {
        apiVersion: 'environments.kloudlite.io/v1',
        kind: 'Environment',
        metadata: { name: envName, namespace: API_NAMESPACE },
        spec: { activated: false },
      })
      if ((result as any).error) throw new Error((result as any).error)
    } catch (err) {
      setSettingsError((err as Error).message)
    } finally {
      setDeactivating(false)
    }
  }

  async function deleteEnvironment() {
    setDeleting(true)
    setSettingsError(null)
    try {
      const result = await window.electronAPI.deleteEnvironment(API_NAMESPACE, envName)
      if (result.error) throw new Error(result.error)
      onDeleted?.()
    } catch (err) {
      setSettingsError((err as Error).message)
      setDeleting(false)
    }
  }

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

        {settingsError && (
          <div className="rounded-xl border border-red-500/20 bg-red-500/[0.04] px-4 py-3 text-[12px] text-red-500">
            {settingsError}
          </div>
        )}

        {/* Danger Zone */}
        <div className="rounded-xl border border-red-500/20 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-red-500">Danger Zone</h3>
          <p className="mt-1 text-[12px] text-muted-foreground">These actions are destructive and cannot be undone.</p>
          <div className="mt-3 flex gap-2">
            <button
              className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20 disabled:opacity-50"
              onClick={deactivate}
              disabled={deactivating || deleting}
            >
              {deactivating ? 'Deactivating...' : 'Deactivate Environment'}
            </button>
            <button
              className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20 disabled:opacity-50"
              onClick={deleteEnvironment}
              disabled={deactivating || deleting}
            >
              {deleting ? 'Deleting...' : 'Delete Environment'}
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

export function EnvironmentContent({ envName, envHash, activeTab, onDeleted }: EnvironmentContentProps) {
  // Services view needs full height (graph), others get scrollable max-width
  if (activeTab === 'services') {
    return (
      <div className="h-full bg-background">
        <ServicesView envHash={envHash} envName={envName} />
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto max-w-4xl">
        {activeTab === 'configs' && <ConfigsView envHash={envHash} envName={envName} />}
        {activeTab === 'snapshots' && <SnapshotsView envHash={envHash} envName={envName} />}
        {activeTab === 'settings' && <SettingsView envName={envName} envHash={envHash} onDeleted={onDeleted} />}
      </div>
    </div>
  )
}
