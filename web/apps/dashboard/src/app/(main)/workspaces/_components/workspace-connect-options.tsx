'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Terminal, Copy, Check, Sparkles } from 'lucide-react'
import { SiAnthropic, SiZedindustries } from 'react-icons/si'
import { VscVscode } from 'react-icons/vsc'
import { CursorIcon } from '@/components/icons/cursor-icon'
import { OpenCodeIcon } from '@/components/icons/opencode-icon'
import { AntigravityIcon } from '@/components/icons/antigravity-icon'
import type { Workspace } from '@kloudlite/types'

interface WorkspaceConnectOptionsProps {
  workspaceId: string
  workspace: Workspace
}

interface AccessMethod {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  command?: string
  url?: string
  available: boolean
  comingSoon?: boolean
  category: string
}

export function WorkspaceConnectOptions({
  workspaceId: _workspaceId,
  workspace,
}: WorkspaceConnectOptionsProps) {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)
  const [sshDialogOpen, setSshDialogOpen] = useState(false)

  const workspaceName = workspace.metadata?.name || 'workspace'

  // Get hash and subdomain from workspace status (computed by controller)
  const wsHash = workspace.status?.hash || ''
  const subdomain = workspace.status?.subdomain || ''

  // Generate VPN-accessible URLs using hash pattern: {prefix}-{hash}.{subdomain}
  // Example: vscode-a1b2c3d4.beanbag.khost.dev
  const generateAccessUrl = (prefix: string): string => {
    if (!wsHash || !subdomain) {
      return ''
    }
    return `https://${prefix}-${wsHash}.${subdomain}`
  }

  // Use URLs from status (already computed by controller), or generate from hash if available
  const codeServerUrl = workspace.status?.accessUrls?.['code-server'] || generateAccessUrl('vscode') || workspace.status?.accessUrl
  const ttydUrl = workspace.status?.accessUrls?.['ttyd'] || generateAccessUrl('tty')
  const claudeTtydUrl = workspace.status?.accessUrls?.['claude-ttyd'] || generateAccessUrl('claude')
  const codexTtydUrl = workspace.status?.accessUrls?.['codex-ttyd'] || generateAccessUrl('codex')
  const opencodeTtydUrl = workspace.status?.accessUrls?.['opencode-ttyd'] || generateAccessUrl('opencode')

  // Generate SSH host using pattern: {workspaceName}-{hash}.{subdomain}
  // Example: platform-base-58890cd7.beanbag.khost.dev
  const workspaceDir = `/home/kl/workspaces/${workspaceName}`
  const sshHost = wsHash && subdomain ? `${workspaceName}-${wsHash}.${subdomain}` : ''

  // SSH URLs for IDEs: vscode://vscode-remote/ssh-remote+kl@{host}{path}
  const vscodeUrl = sshHost ? `vscode://vscode-remote/ssh-remote+kl@${sshHost}${workspaceDir}` : ''
  const cursorUrl = sshHost ? `cursor://vscode-remote/ssh-remote+kl@${sshHost}${workspaceDir}` : ''
  const zedUrl = sshHost ? `zed://ssh/${sshHost}${workspaceDir}` : ''
  const antigravityUrl = sshHost ? `antigravity://ssh-remote+kl@${sshHost}${workspaceDir}` : ''
  const sshCommand = sshHost ? `ssh kl@${sshHost}` : ''

  const accessMethods: AccessMethod[] = [
    {
      id: 'vscode',
      name: 'VS Code',
      description: 'Remote development via SSH',
      icon: <VscVscode className="h-4 w-4 flex-shrink-0" />,
      available: !!vscodeUrl,
      url: vscodeUrl,
      category: 'Desktop IDEs',
    },
    {
      id: 'cursor',
      name: 'Cursor',
      description: 'AI-powered editor via SSH',
      icon: <CursorIcon className="h-4 w-4 flex-shrink-0" />,
      available: !!cursorUrl,
      url: cursorUrl,
      category: 'Desktop IDEs',
    },
    {
      id: 'zed',
      name: 'Zed',
      description: 'Fast collaborative editor',
      icon: <SiZedindustries className="h-4 w-4 flex-shrink-0" />,
      available: !!zedUrl,
      url: zedUrl,
      category: 'Desktop IDEs',
    },
    {
      id: 'antigravity',
      name: 'Antigravity',
      description: 'AI-powered IDE',
      icon: <AntigravityIcon className="h-4 w-4 flex-shrink-0" />,
      available: !!antigravityUrl,
      url: antigravityUrl,
      category: 'Desktop IDEs',
    },
    {
      id: 'code-server',
      name: 'VS Code Web',
      description: 'Full IDE in your browser',
      icon: <VscVscode className="h-4 w-4 flex-shrink-0" />,
      available: !!codeServerUrl,
      url: codeServerUrl,
      category: 'Web-Based IDEs',
    },
    {
      id: 'ttyd-terminal',
      name: 'Web Terminal',
      description: 'Browser-based terminal',
      icon: <Terminal className="h-4 w-4 flex-shrink-0" />,
      available: !!ttydUrl,
      url: ttydUrl,
      category: 'Web Terminal & AI Assistants',
    },
    {
      id: 'claude-code-web',
      name: 'Claude Code',
      description: 'AI coding assistant',
      icon: <SiAnthropic className="h-4 w-4 flex-shrink-0" />,
      available: !!claudeTtydUrl,
      url: claudeTtydUrl,
      category: 'Web Terminal & AI Assistants',
    },
    {
      id: 'opencode',
      name: 'OpenCode',
      description: 'AI coding assistant',
      icon: <OpenCodeIcon className="h-4 w-4 flex-shrink-0" />,
      available: !!opencodeTtydUrl,
      url: opencodeTtydUrl,
      category: 'Web Terminal & AI Assistants',
    },
    {
      id: 'codex',
      name: 'Codex',
      description: 'AI coding assistant',
      icon: <Sparkles className="h-4 w-4 flex-shrink-0" />,
      available: !!codexTtydUrl,
      url: codexTtydUrl,
      category: 'Web Terminal & AI Assistants',
    },
    {
      id: 'ssh',
      name: 'SSH Terminal',
      description: 'Direct terminal access',
      icon: <Terminal className="h-4 w-4 flex-shrink-0" />,
      available: !!sshCommand,
      command: sshCommand,
      category: 'Direct Access',
    },
  ]

  // Group methods by category
  const groupedMethods = accessMethods.reduce(
    (acc, method) => {
      if (!acc[method.category]) {
        acc[method.category] = []
      }
      acc[method.category].push(method)
      return acc
    },
    {} as Record<string, AccessMethod[]>,
  )

  const handleCopyCommand = (command: string, methodId: string) => {
    navigator.clipboard.writeText(command)
    setCopiedCommand(methodId)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const handleConnect = (method: AccessMethod) => {
    if (method.url) {
      window.open(method.url, '_blank', 'noopener,noreferrer')
    } else if (method.id === 'ssh') {
      setSshDialogOpen(true)
    } else if (method.command) {
      handleCopyCommand(method.command, method.id)
    }
  }

  return (
    <>
      <div className="bg-card rounded-lg border p-6">
        <h3 className="mb-4 text-sm font-medium">Connect to Workspace</h3>

        <div className="space-y-6">
          {Object.entries(groupedMethods).map(([category, methods]) => (
            <div key={category}>
              <h4 className="text-muted-foreground mb-3 text-xs font-medium">{category}</h4>
              <div className="flex flex-wrap gap-2">
                {methods.map((method) => (
                  <button
                    key={method.id}
                    onClick={() => handleConnect(method)}
                    disabled={!method.available || method.comingSoon}
                    className={`inline-flex h-8 items-center gap-2 rounded-full border px-3 transition-all ${
                      !method.available || method.comingSoon
                        ? 'bg-muted/30 cursor-not-allowed opacity-50'
                        : 'hover:bg-muted/50 hover:border-primary/50'
                    }`}
                  >
                    {method.icon}
                    <span className="text-sm leading-none font-medium whitespace-nowrap">
                      {method.name}
                    </span>
                    {method.comingSoon && (
                      <span className="text-muted-foreground text-[10px] leading-none">Soon</span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* SSH Connection Dialog */}
      <Dialog open={sshDialogOpen} onOpenChange={setSshDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>SSH Connection</DialogTitle>
            <DialogDescription>
              Use this command to connect to your workspace via SSH
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="bg-muted rounded-md p-4">
              <code className="font-mono text-sm break-all">{sshCommand}</code>
            </div>

            <Button
              onClick={() => {
                handleCopyCommand(sshCommand, 'ssh')
                setTimeout(() => setSshDialogOpen(false), 1000)
              }}
              className="w-full gap-2"
            >
              {copiedCommand === 'ssh' ? (
                <>
                  <Check className="h-4 w-4" />
                  Copied!
                </>
              ) : (
                <>
                  <Copy className="h-4 w-4" />
                  Copy Command
                </>
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
