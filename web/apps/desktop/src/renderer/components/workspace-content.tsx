import { cn } from '@/lib/utils'
import { useState } from 'react'
import { Monitor, Globe, Terminal, Laptop, ExternalLink, Copy, Check, Package, GitBranch, Trash2, Plus } from 'lucide-react'
import { SnapshotTree, generateSnapshots } from './snapshot-tree'
import { EmptyState } from './empty-state'

interface WorkspaceContentProps {
  wsName: string
  wsId: string
  activeTab: string
}

// Dummy workspace data keyed by id
const WS_DATA: Record<string, {
  status: string
  env: string
  git: string
  branch: string
  ports: number[]
  packages: { name: string; version: string; manager: string }[]
}> = {
  'ws-1': {
    status: 'running', env: 'Staging',
    git: 'github.com/kloudlite/api', branch: 'main',
    ports: [8080, 5432, 6379],
    packages: [
      { name: 'go', version: '1.24.0', manager: 'nix' },
      { name: 'nodejs', version: '22.5.0', manager: 'nix' },
      { name: 'kubectl', version: '1.31.0', manager: 'nix' },
      { name: 'docker', version: '27.0.0', manager: 'nix' },
    ],
  },
  'ws-2': {
    status: 'running', env: 'Development',
    git: 'github.com/kloudlite/web', branch: 'feat/ui',
    ports: [5173, 3000],
    packages: [
      { name: 'bun', version: '1.3.8', manager: 'nix' },
      { name: 'nodejs', version: '22.5.0', manager: 'nix' },
    ],
  },
  'ws-3': {
    status: 'stopped', env: 'Staging',
    git: '', branch: '',
    ports: [],
    packages: [
      { name: 'go', version: '1.24.0', manager: 'nix' },
      { name: 'delve', version: '1.23.0', manager: 'nix' },
    ],
  },
  'ws-4': {
    status: 'running', env: 'QA Testing',
    git: 'github.com/kloudlite/api', branch: 'fix/migration',
    ports: [8080],
    packages: [
      { name: 'go', version: '1.24.0', manager: 'nix' },
      { name: 'postgresql', version: '16.0', manager: 'nix' },
    ],
  },
  'ws-5': {
    status: 'failed', env: 'Production',
    git: 'github.com/kloudlite/api', branch: 'hotfix/auth',
    ports: [],
    packages: [
      { name: 'go', version: '1.24.0', manager: 'nix' },
    ],
  },
}

const IDE_OPTIONS = [
  { category: 'Desktop IDEs', items: [
    { name: 'VS Code', icon: Monitor, accent: '#007ACC' },
    { name: 'Cursor', icon: Monitor, accent: '#6B5CE7' },
    { name: 'Zed', icon: Monitor, accent: '#084CCF' },
    { name: 'JetBrains', icon: Laptop, accent: '#FF318C' },
  ]},
  { category: 'Web-Based', items: [
    { name: 'VS Code Web', icon: Globe, accent: '#007ACC' },
    { name: 'Terminal', icon: Terminal, accent: '#22C55E' },
  ]},
  { category: 'AI Assistants', items: [
    { name: 'Claude Code', icon: Terminal, accent: '#D97706' },
  ]},
]

function CopyBtn({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <button
      className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground/40 transition-colors hover:bg-accent hover:text-muted-foreground"
      onClick={() => { navigator.clipboard.writeText(text); setCopied(true); setTimeout(() => setCopied(false), 1500) }}
    >
      {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  )
}

function ConnectView({ wsId, wsName }: { wsId: string; wsName: string }) {
  const data = WS_DATA[wsId]
  const isRunning = data?.status === 'running'

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className={cn(
          'flex h-10 w-10 items-center justify-center rounded-xl',
          isRunning ? 'bg-emerald-500/10' : data?.status === 'failed' ? 'bg-red-500/10' : 'bg-muted'
        )}>
          <Terminal className={cn(
            'h-5 w-5',
            isRunning ? 'text-emerald-500' : data?.status === 'failed' ? 'text-red-500' : 'text-muted-foreground'
          )} />
        </div>
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">{wsName}</h2>
          <p className="text-[12px] text-muted-foreground">{data?.env}</p>
        </div>
        {isRunning && (
          <span className="ml-auto rounded-full bg-emerald-500/10 px-2.5 py-1 text-[11px] font-medium text-emerald-600">Running</span>
        )}
      </div>

      {!isRunning && (
        <div className="mt-4 flex items-center gap-3 rounded-xl border border-amber-500/20 bg-amber-50 px-4 py-3 dark:bg-amber-500/5">
          <div className="h-2 w-2 rounded-full bg-amber-500" />
          <p className="flex-1 text-[13px] text-amber-700 dark:text-amber-400">Workspace is not running</p>
          <button className="rounded-lg bg-amber-500 px-3 py-1.5 text-[12px] font-medium text-white transition-colors hover:bg-amber-600">
            Start
          </button>
        </div>
      )}

      {isRunning && (
        <>
          {/* SSH */}
          <div className="mt-5 rounded-xl border border-border/50 bg-[#1e1e2e] px-4 py-3" style={{ fontFamily: 'SF Mono, Fira Code, JetBrains Mono, Menlo, Consolas, monospace' }}>
            <div className="flex items-center gap-3">
              <span className="text-[10px] text-[#6c7086]">$</span>
              <span className="flex-1 text-[13px] text-[#cdd6f4]">ssh {wsName}.workspace.local</span>
              <CopyBtn text={`ssh ${wsName}.workspace.local`} />
            </div>
          </div>

          {/* Open in — all in one card */}
          <div className="mt-5 rounded-xl border border-border/50 overflow-hidden">
            {IDE_OPTIONS.map((cat, catIdx) => (
              <div key={cat.category}>
                {catIdx > 0 && <div className="border-t border-border/30" />}
                <div className="px-4 pt-3 pb-1">
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/40">{cat.category}</p>
                </div>
                <div className="flex flex-wrap gap-0 px-2 pb-2">
                  {cat.items.map((ide) => (
                    <button
                      key={ide.name}
                      className="flex items-center gap-2.5 rounded-lg px-3 py-2 transition-colors hover:bg-accent/50"
                    >
                      <div
                        className="flex h-8 w-8 items-center justify-center rounded-lg"
                        style={{ backgroundColor: `${ide.accent}10` }}
                      >
                        <ide.icon className="h-4 w-4" style={{ color: ide.accent }} />
                      </div>
                      <span className="text-[12px] font-medium text-foreground">{ide.name}</span>
                      <ExternalLink className="h-3 w-3 text-muted-foreground/30" />
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>

          {/* Ports */}
          {data.ports.length > 0 && (
            <div className="mt-5 overflow-hidden rounded-xl border border-border/50">
              <div className="border-b border-border/30 bg-accent/20 px-4 py-2">
                <p className="text-[11px] font-medium text-muted-foreground/60">Exposed Ports</p>
              </div>
              {data.ports.map((port, i) => (
                <div key={port} className={cn(
                  'flex items-center gap-3 px-4 py-2.5 transition-colors hover:bg-accent/20',
                  i < data.ports.length - 1 && 'border-b border-border/20'
                )}>
                  <div className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                  <span className="font-mono text-[13px] font-medium text-foreground">{port}</span>
                  <div className="flex-1" />
                  <code className="font-mono text-[10px] text-muted-foreground/50">{wsName}.workspace.local:{port}</code>
                  <CopyBtn text={`${wsName}.workspace.local:${port}`} />
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}

function PackagesView({ wsId }: { wsId: string }) {
  const data = WS_DATA[wsId]
  const packages = data?.packages || []

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">Packages</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">{packages.length} packages installed via Nix</p>
        </div>
        <button className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90">
          <Plus className="h-3.5 w-3.5" />
          Add Package
        </button>
      </div>

      <div className="mt-4 overflow-hidden rounded-xl border border-border/50">
        <table className="w-full text-left text-[13px]">
          <thead>
            <tr className="border-b border-border/50 bg-accent/30">
              <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Package</th>
              <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Version</th>
              <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Manager</th>
              <th className="w-16 px-4 py-2.5"></th>
            </tr>
          </thead>
          <tbody>
            {packages.map((pkg) => (
              <tr key={pkg.name} className="h-12 border-b border-border/30 last:border-0 hover:bg-accent/20">
                <td className="px-4">
                  <span className="font-mono font-medium text-foreground">{pkg.name}</span>
                </td>
                <td className="px-4 font-mono text-[12px] text-muted-foreground">{pkg.version}</td>
                <td className="px-4">
                  <span className="rounded-full bg-accent px-2 py-0.5 text-[10px] font-medium text-muted-foreground">{pkg.manager}</span>
                </td>
                <td className="px-4 text-right">
                  <button className="rounded p-1 text-muted-foreground/40 transition-colors hover:bg-accent hover:text-red-500">
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function GitView({ wsId }: { wsId: string }) {
  const data = WS_DATA[wsId]

  if (!data?.git) {
    return (
      <EmptyState
        title="No git repository configured"
        description="Connect a repository to enable git integration"
        action={{ label: 'Connect Repository', onClick: () => {} }}
      />
    )
  }

  return (
    <div className="p-6">
      <h2 className="text-[16px] font-semibold text-foreground">Git Repository</h2>

      <div className="mt-4 rounded-xl border border-border/50 bg-card p-5">
        <div className="flex flex-col gap-3 text-[13px]">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Repository</span>
            <div className="flex items-center gap-1.5">
              <span className="font-mono text-[12px] font-medium text-foreground">{data.git}</span>
              <CopyBtn text={data.git} />
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Branch</span>
            <span className="rounded-full bg-accent px-2.5 py-0.5 text-[12px] font-mono font-medium text-foreground">{data.branch}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

function WsSettingsView({ wsName, wsId }: { wsName: string; wsId: string }) {
  const data = WS_DATA[wsId]

  return (
    <div className="p-6">
      <h2 className="text-[16px] font-semibold text-foreground">Settings</h2>

      <div className="mt-5 flex flex-col gap-5">
        <div className="rounded-xl border border-border/50 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-foreground">General</h3>
          <div className="mt-3 flex flex-col gap-3 text-[13px]">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Name</span>
              <span className="font-medium text-foreground">{wsName}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Environment</span>
              <span className="font-medium text-foreground">{data?.env}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Idle Timeout</span>
              <span className="font-medium text-foreground">30 minutes</span>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-red-500/20 bg-card p-5">
          <h3 className="text-[13px] font-semibold text-red-500">Danger Zone</h3>
          <div className="mt-3 flex gap-2">
            <button className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20">
              Archive Workspace
            </button>
            <button className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[12px] font-medium text-red-500 transition-colors hover:bg-red-500/20">
              Delete Workspace
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export function WorkspaceContent({ wsName, wsId, activeTab }: WorkspaceContentProps) {
  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto max-w-4xl">
        {activeTab === 'connect' && <ConnectView wsId={wsId} wsName={wsName} />}
        {activeTab === 'packages' && <PackagesView wsId={wsId} />}
        {activeTab === 'git' && <GitView wsId={wsId} />}
        {activeTab === 'snapshots' && (
          <SnapshotTree
            snapshots={generateSnapshots(`ws-${wsId}`)}
            title="Snapshots"
            subtitle={`Snapshots for ${wsName}`}
          />
        )}
        {activeTab === 'settings' && <WsSettingsView wsName={wsName} wsId={wsId} />}
      </div>
    </div>
  )
}

export function NewWorkspaceDialog({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [visibility, setVisibility] = useState<'private' | 'shared'>('private')
  const [gitUrl, setGitUrl] = useState('')
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
          <h2 className="text-[16px] font-semibold text-foreground">Create Workspace</h2>
          <p className="mt-0.5 text-[12px] text-muted-foreground">Set up a new development workspace</p>
        </div>

        <div className="flex flex-col gap-4 px-6 py-5">
          <div>
            <label className="mb-1.5 block text-[12px] font-medium text-foreground">Name</label>
            <input
              type="text"
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-[13px] text-foreground outline-none transition-colors focus:border-primary"
              placeholder="e.g. api-dev, frontend-dev"
              value={name}
              onChange={(e) => setName(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-'))}
              autoFocus
            />
          </div>

          <div>
            <label className="mb-1.5 block text-[12px] font-medium text-foreground">Visibility</label>
            <div className="flex gap-2">
              {(['private', 'shared'] as const).map((v) => (
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

          <div>
            <label className="mb-1.5 block text-[12px] font-medium text-foreground">Git Repository <span className="text-muted-foreground">(optional)</span></label>
            <input
              type="text"
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-[13px] text-foreground outline-none transition-colors focus:border-primary"
              placeholder="github.com/org/repo"
              value={gitUrl}
              onChange={(e) => setGitUrl(e.target.value)}
            />
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-border/30 px-6 py-4">
          <button className="rounded-lg px-4 py-2 text-[12px] font-medium text-muted-foreground transition-colors hover:bg-accent" onClick={close}>
            Cancel
          </button>
          <button
            className="rounded-lg bg-primary px-4 py-2 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            disabled={!name.trim()}
            onClick={close}
          >
            Create Workspace
          </button>
        </div>
      </div>
    </div>
  )
}
