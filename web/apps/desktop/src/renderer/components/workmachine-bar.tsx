import { useState } from 'react'
import { ChevronLeft, ChevronRight, Check, Copy, Cpu, HardDrive, KeyRound, Power, RotateCw, Settings, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { DotGrid } from './ui/dot-grid'
import { Button, Dialog } from './ui'

// Dummy workmachine state
const WORKMACHINE = {
  name: "Karthik's WorkMachine",
  status: 'running' as 'running' | 'stopped' | 'starting' | 'error',
  type: '4 vCPU · 8 GB',
  cpu: 42,
  memory: 67,
  uptime: '3d 14h',
}

const statusConfig = {
  running: { color: 'bg-emerald-400', label: 'Running', textColor: 'text-emerald-400' },
  stopped: { color: 'bg-sidebar-foreground/25', label: 'Stopped', textColor: 'text-sidebar-foreground/60' },
  starting: { color: 'bg-amber-400', label: 'Starting', textColor: 'text-amber-400' },
  error: { color: 'bg-red-400', label: 'Error', textColor: 'text-red-400' },
}

const MACHINE_TYPES = [
  { id: 'small', name: 'Small', cpu: '2 vCPU', memory: '4 GB', note: 'Light dev work' },
  { id: 'standard', name: 'Standard', cpu: '4 vCPU', memory: '8 GB', note: 'Recommended' },
  { id: 'large', name: 'Large', cpu: '8 vCPU', memory: '16 GB', note: 'Builds and services' },
  { id: 'xlarge', name: 'XLarge', cpu: '16 vCPU', memory: '32 GB', note: 'Heavy workloads' },
]

const SSH_PUBLIC_KEY = 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKloudliteMachineKey kloudlite-workmachine'

export function WorkMachineBar() {
  const [showActions, setShowActions] = useState(false)
  const [manageView, setManageView] = useState<'main' | 'keys' | 'machine'>('main')
  const [drillDirection, setDrillDirection] = useState<'forward' | 'back'>('forward')
  const [selectedMachineType, setSelectedMachineType] = useState('standard')
  const [authorizedKeys, setAuthorizedKeys] = useState<string[]>([
    'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIUserLaptopKey karthik@macbook',
  ])
  const [authorizedKey, setAuthorizedKey] = useState('')
  const [copied, setCopied] = useState(false)
  const navigateManage = (view: 'main' | 'keys' | 'machine') => {
    setDrillDirection(view === 'main' ? 'back' : 'forward')
    setManageView(view)
  }

  const addAuthorizedKey = () => {
    const key = authorizedKey.trim()
    if (!key || authorizedKeys.includes(key)) return
    setAuthorizedKeys([...authorizedKeys, key])
    setAuthorizedKey('')
  }
  const wm = WORKMACHINE
  const config = statusConfig[wm.status]

  return (
    <div className="shrink-0 p-3">
      <div className="rounded-xl border border-sidebar-foreground/[0.08] bg-sidebar-foreground/[0.03] p-3 shadow-sm">
        <div className="flex items-start gap-3">
          <div className="relative mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sidebar-foreground/[0.06]">
            <HardDrive className="h-4 w-4 text-sidebar-foreground/75" />
            <div className={cn('absolute bottom-1 right-1 h-2 w-2 rounded-full ring-2 ring-sidebar', config.color)} />
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="truncate text-[13px] font-semibold text-sidebar-foreground/90">{wm.name}</h3>
              <span className={cn('shrink-0 rounded-full bg-sidebar-foreground/[0.06] px-2 py-0.5 text-[10px] font-medium', config.textColor)}>{config.label}</span>
            </div>
            <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-sidebar-foreground/50">
              <span>{wm.type}</span>
              <span>·</span>
              <span>Up {wm.uptime}</span>
            </div>
          </div>

          <button
            className="no-drag inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-sidebar-foreground/[0.08] text-sidebar-foreground/60 transition-colors hover:bg-sidebar-foreground/[0.08] hover:text-sidebar-foreground/90"
            aria-label="Manage WorkMachine"
            onClick={() => { setManageView('main'); setDrillDirection('forward'); setShowActions(true) }}
          >
            <Settings className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-2">
          <div className="rounded-lg bg-sidebar-foreground/[0.035] px-2.5 py-2">
            <div className="mb-1.5 flex items-center justify-between text-[10px] text-sidebar-foreground/50">
              <span>CPU</span>
              <span className="font-mono text-sidebar-foreground/75">{wm.cpu}%</span>
            </div>
            <DotGrid value={wm.cpu} total={10} size={6} />
          </div>
          <div className="rounded-lg bg-sidebar-foreground/[0.035] px-2.5 py-2">
            <div className="mb-1.5 flex items-center justify-between text-[10px] text-sidebar-foreground/50">
              <span>Memory</span>
              <span className="font-mono text-sidebar-foreground/75">{wm.memory}%</span>
            </div>
            <DotGrid value={wm.memory} total={10} size={6} />
          </div>
        </div>
      </div>

      {showActions && (
        <Dialog
          title={manageView === 'main' ? 'Manage WorkMachine' : manageView === 'keys' ? 'Authorized keys' : 'Machine type'}
          description={manageView === 'main' ? 'Power, access, and compute settings' : manageView === 'keys' ? 'Manage who can SSH into this machine' : 'Choose the compute size for this machine'}
          onClose={() => setShowActions(false)}
          maxWidth="40rem"
          footer={(close) => (
            <div className="flex w-full items-center justify-between gap-3">
              <span className="text-[11px] text-muted-foreground">{wm.name}</span>
              <div className="flex gap-2">
                {manageView !== 'main' && <Button variant="secondary" onClick={() => navigateManage('main')}>Back</Button>}
                <Button variant={manageView === 'main' ? 'secondary' : 'primary'} onClick={close}>{manageView === 'main' ? 'Close' : 'Save'}</Button>
              </div>
            </div>
          )}
        >
          <div
            key={manageView}
            className="will-change-transform"
            style={{
              animation: drillDirection === 'forward' ? 'drill-in-right 180ms ease-out' : 'drill-in-left 180ms ease-out',
            }}
          >
          {manageView === 'main' && (
            <div className="space-y-4">
              <div className="rounded-xl border border-border/50 bg-card/60 p-4">
                <div className="flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={cn('h-2 w-2 rounded-full', config.color)} />
                      <h3 className="truncate text-[14px] font-semibold text-foreground">{wm.name}</h3>
                    </div>
                    <p className="mt-1 text-[12px] text-muted-foreground">{wm.type} · Up {wm.uptime}</p>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="secondary" size="sm"><RotateCw className="h-3.5 w-3.5" />Restart</Button>
                    <Button variant="danger" size="sm"><Power className="h-3.5 w-3.5" />Shutdown</Button>
                  </div>
                </div>
              </div>

              <div className="overflow-hidden rounded-xl border border-border/50 bg-card/60">
                <button className="flex w-full items-center gap-3 px-4 py-4 text-left transition-colors hover:bg-accent/70" onClick={() => navigateManage('keys')}>
                  <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted text-muted-foreground"><KeyRound className="h-4 w-4" /></span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-[13px] font-semibold text-foreground">Authorized keys</span>
                    <span className="mt-0.5 block text-[12px] text-muted-foreground">{authorizedKeys.length} SSH key{authorizedKeys.length === 1 ? '' : 's'} allowed</span>
                  </span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
                <button className="flex w-full items-center gap-3 border-t border-border/50 px-4 py-4 text-left transition-colors hover:bg-accent/70" onClick={() => navigateManage('machine')}>
                  <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted text-muted-foreground"><Cpu className="h-4 w-4" /></span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-[13px] font-semibold text-foreground">Machine type</span>
                    <span className="mt-0.5 block text-[12px] text-muted-foreground">{MACHINE_TYPES.find((type) => type.id === selectedMachineType)?.name} · {wm.type}</span>
                  </span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
              </div>
            </div>
          )}

          {manageView === 'keys' && (
            <div className="space-y-4">
              <button className="inline-flex items-center gap-1.5 text-[12px] text-muted-foreground hover:text-foreground" onClick={() => navigateManage('main')}>
                <ChevronLeft className="h-3.5 w-3.5" /> WorkMachine settings
              </button>

              <div className="rounded-xl border border-border/50 bg-card/60 p-3">
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-[12px] font-medium text-muted-foreground">Machine public key</span>
                  <button
                    className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    onClick={async () => {
                      await navigator.clipboard?.writeText(SSH_PUBLIC_KEY)
                      setCopied(true)
                      setTimeout(() => setCopied(false), 1200)
                    }}
                  >
                    {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                    {copied ? 'Copied' : 'Copy'}
                  </button>
                </div>
                <code className="block truncate rounded-lg bg-muted px-3 py-2 font-mono text-[11px] text-foreground/80">{SSH_PUBLIC_KEY}</code>
              </div>

              <div className="overflow-hidden rounded-xl border border-border/50 bg-card/60">
                <div className="flex items-center justify-between border-b border-border/50 px-4 py-3">
                  <span className="text-[12px] font-semibold text-foreground">Authorized keys</span>
                  <span className="text-[11px] text-muted-foreground">{authorizedKeys.length} added</span>
                </div>
                {authorizedKeys.map((key) => (
                  <div key={key} className="flex items-center gap-2 border-b border-border/50 px-4 py-3 last:border-b-0">
                    <code className="min-w-0 flex-1 truncate font-mono text-[11px] text-foreground/80">{key}</code>
                    <button className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-red-500/10 hover:text-red-400" onClick={() => setAuthorizedKeys(authorizedKeys.filter((item) => item !== key))}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                ))}
              </div>

              <div className="rounded-xl border border-border/50 bg-card/60 p-3">
                <label className="block text-[12px] font-medium text-muted-foreground">Add authorized key</label>
                <textarea
                  className="mt-2 h-20 w-full resize-none rounded-lg border border-border bg-background px-3 py-2 font-mono text-[12px] text-foreground outline-none transition-colors placeholder:text-muted-foreground/50 focus:border-primary/60"
                  placeholder="Paste an ssh-ed25519 or ssh-rsa public key"
                  value={authorizedKey}
                  onChange={(e) => setAuthorizedKey(e.target.value)}
                />
                <div className="mt-2 flex justify-end">
                  <Button size="sm" disabled={!authorizedKey.trim()} onClick={addAuthorizedKey}>Add key</Button>
                </div>
              </div>
            </div>
          )}

          {manageView === 'machine' && (
            <div className="space-y-4">
              <button className="inline-flex items-center gap-1.5 text-[12px] text-muted-foreground hover:text-foreground" onClick={() => navigateManage('main')}>
                <ChevronLeft className="h-3.5 w-3.5" /> WorkMachine settings
              </button>

              <div className="overflow-hidden rounded-xl border border-border/50 bg-card/60">
                {MACHINE_TYPES.map((type, index) => {
                  const selected = selectedMachineType === type.id
                  return (
                    <button
                      key={type.id}
                      className={cn(
                        'grid w-full grid-cols-[1fr_auto_auto_auto] items-center gap-4 px-4 py-3 text-left transition-colors',
                        index > 0 && 'border-t border-border/50',
                        selected ? 'bg-primary/[0.08]' : 'hover:bg-accent/70'
                      )}
                      onClick={() => setSelectedMachineType(type.id)}
                    >
                      <span className="flex items-center gap-3">
                        <span className={cn('flex h-5 w-5 items-center justify-center rounded-full border', selected ? 'border-primary bg-primary text-primary-foreground' : 'border-border')}>
                          {selected && <Check className="h-3 w-3" />}
                        </span>
                        <span>
                          <span className="block text-[13px] font-semibold text-foreground">{type.name}</span>
                          <span className="mt-0.5 block text-[11px] text-muted-foreground">{type.note}</span>
                        </span>
                      </span>
                      <span className="font-mono text-[12px] text-muted-foreground">{type.cpu}</span>
                      <span className="font-mono text-[12px] text-muted-foreground">{type.memory}</span>
                      <span className={cn('text-[11px] font-medium', selected ? 'text-primary' : 'text-muted-foreground/60')}>{selected ? 'Selected' : 'Select'}</span>
                    </button>
                  )
                })}
              </div>
            </div>
          )}
          </div>
        </Dialog>
      )}
    </div>
  )
}
