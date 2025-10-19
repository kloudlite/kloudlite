'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Terminal,
  ExternalLink,
  Copy,
  Check,
  Sparkles,
  Globe,
  Zap
} from 'lucide-react'
import {
  SiIntellijidea,
  SiAnthropic,
  SiZedindustries
} from 'react-icons/si'
import { VscVscode } from 'react-icons/vsc'
import { CursorIcon } from '@/components/icons/cursor-icon'
import { OpenCodeIcon } from '@/components/icons/opencode-icon'
import type { Workspace } from '@/types/workspace'

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

export function WorkspaceConnectOptions({ workspaceId, workspace }: WorkspaceConnectOptionsProps) {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  // Extract access URLs from workspace status
  const codeServerUrl = workspace.status?.accessUrls?.['code-server'] || workspace.status?.accessUrl
  const ttydUrl = workspace.status?.accessUrls?.['ttyd']
  const sshPort = workspace.status?.accessUrls?.['ssh'] || '2222'
  const workspaceName = workspace.metadata?.name || 'workspace'

  // SSH connection with jump host
  const jumpHost = `kloudlite@localhost:${sshPort}`
  const targetHost = `kl@${workspaceName}`
  const workspaceDir = `/home/kl/workspaces/${workspaceName}`
  const sshCommand = `ssh -J ${jumpHost} ${targetHost}`

  // SSH config for manual setup
  const sshConfig = `Host ${workspaceName}
  HostName ${workspaceName}
  User kl
  ProxyJump kloudlite@localhost:${sshPort}
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null`

  // SSH commands for IDEs (include jump host in the command)
  const sshJumpFlag = `-o "ProxyJump=${jumpHost}"`
  const vscodeCommand = `code --folder-uri "vscode-remote://ssh-remote+${targetHost}${workspaceDir}" --remote-ssh-command "ssh ${sshJumpFlag}"`
  const cursorCommand = `cursor --folder-uri "vscode-remote://ssh-remote+${targetHost}${workspaceDir}"`
  const zedCommand = `ssh ${sshJumpFlag} ${targetHost} -t "cd ${workspaceDir} && zed ."`
  const intellijCommand = `ssh ${sshJumpFlag} ${targetHost} -t "cd ${workspaceDir} && idea ."`

  // VS Code extension deep link
  const vsCodeExtensionUrl = workspace.metadata.namespace
    ? `vscode://kloudlite.kloudlite-workspace/connect?workspace=${workspaceName}&namespace=${workspace.metadata.namespace}`
    : ''

  const accessMethods: AccessMethod[] = [
    {
      id: 'ssh-config',
      name: 'SSH Config',
      description: 'Copy to ~/.ssh/config (required for IDEs)',
      icon: <Copy className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: sshConfig,
      category: 'Setup'
    },
    {
      id: 'vscode-extension',
      name: 'VS Code Extension',
      description: 'Open in VS Code app',
      icon: <VscVscode className="h-4 w-4 flex-shrink-0" />,
      available: !!vsCodeExtensionUrl,
      url: vsCodeExtensionUrl,
      category: 'Desktop IDEs'
    },
    {
      id: 'vscode',
      name: 'VS Code',
      description: 'Remote development via SSH',
      icon: <VscVscode className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: vscodeCommand,
      category: 'Desktop IDEs'
    },
    {
      id: 'cursor',
      name: 'Cursor',
      description: 'AI-powered editor via SSH',
      icon: <CursorIcon className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: cursorCommand,
      category: 'Desktop IDEs'
    },
    {
      id: 'intellij',
      name: 'IntelliJ IDEA',
      description: 'JetBrains IDE via SSH',
      icon: <SiIntellijidea className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: intellijCommand,
      category: 'Desktop IDEs'
    },
    {
      id: 'zed',
      name: 'Zed',
      description: 'Fast collaborative editor',
      icon: <SiZedindustries className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: zedCommand,
      category: 'Desktop IDEs'
    },
    {
      id: 'code-server',
      name: 'VS Code Web',
      description: 'Full IDE in your browser',
      icon: <VscVscode className="h-4 w-4 flex-shrink-0" />,
      available: !!codeServerUrl,
      url: codeServerUrl,
      category: 'Web-Based IDEs'
    },
    {
      id: 'ttyd-terminal',
      name: 'Web Terminal',
      description: 'Browser-based terminal',
      icon: <Terminal className="h-4 w-4 flex-shrink-0" />,
      available: !!ttydUrl,
      url: ttydUrl,
      category: 'Web Terminal & AI Assistants'
    },
    {
      id: 'claude-code-web',
      name: 'Claude Code',
      description: 'AI coding assistant',
      icon: <SiAnthropic className="h-4 w-4 flex-shrink-0" />,
      available: !!ttydUrl,
      url: ttydUrl,
      category: 'Web Terminal & AI Assistants'
    },
    {
      id: 'opencode',
      name: 'OpenCode',
      description: 'AI coding assistant',
      icon: <OpenCodeIcon className="h-4 w-4 flex-shrink-0" />,
      available: !!ttydUrl,
      url: ttydUrl,
      comingSoon: true,
      category: 'Web Terminal & AI Assistants'
    },
    {
      id: 'ssh',
      name: 'SSH Terminal',
      description: 'Direct terminal access',
      icon: <Terminal className="h-4 w-4 flex-shrink-0" />,
      available: !!workspaceName,
      command: sshCommand,
      category: 'Direct Access'
    }
  ]

  // Group methods by category
  const groupedMethods = accessMethods.reduce((acc, method) => {
    if (!acc[method.category]) {
      acc[method.category] = []
    }
    acc[method.category].push(method)
    return acc
  }, {} as Record<string, AccessMethod[]>)

  const handleCopyCommand = (command: string, methodId: string) => {
    navigator.clipboard.writeText(command)
    setCopiedCommand(methodId)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const handleConnect = (method: AccessMethod) => {
    if (method.url) {
      window.open(method.url, '_blank')
    } else if (method.command) {
      handleCopyCommand(method.command, method.id)
    }
  }

  return (
    <div className="bg-card rounded-lg border p-6">
      <h3 className="text-sm font-medium mb-4">Connect to Workspace</h3>

      <div className="space-y-6">
        {Object.entries(groupedMethods).map(([category, methods]) => (
          <div key={category}>
            <h4 className="text-xs font-medium text-muted-foreground mb-3">{category}</h4>
            <div className="flex flex-wrap gap-2">
              {methods.map((method) => (
                <button
                  key={method.id}
                  onClick={() => handleConnect(method)}
                  disabled={!method.available || method.comingSoon}
                  className={`inline-flex items-center gap-2 h-8 px-3 rounded-full border transition-all ${
                    !method.available || method.comingSoon
                      ? 'opacity-50 cursor-not-allowed bg-muted/30'
                      : 'hover:bg-muted/50 hover:border-primary/50'
                  }`}
                >
                  {method.icon}
                  <span className="text-sm font-medium whitespace-nowrap leading-none">{method.name}</span>
                  {method.comingSoon && (
                    <span className="text-[10px] text-muted-foreground leading-none">Soon</span>
                  )}
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
