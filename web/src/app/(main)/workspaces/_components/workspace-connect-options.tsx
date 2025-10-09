'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Terminal,
  Code2,
  Globe,
  ExternalLink,
  Copy,
  Check
} from 'lucide-react'
import type { Workspace } from '@/types/workspace'

interface WorkspaceConnectOptionsProps {
  workspaceId: string
  workspace: Workspace
}

interface ConnectOption {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  type: 'desktop' | 'web' | 'terminal'
  available: boolean
  command?: string
  url?: string
}

export function WorkspaceConnectOptions({ workspaceId, workspace }: WorkspaceConnectOptionsProps) {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  // Extract access URLs from workspace status
  const codeServerUrl = workspace.status?.accessUrls?.['code-server'] || workspace.status?.accessUrl
  const ttydUrl = workspace.status?.accessUrls?.['ttyd']
  const sshPort = workspace.status?.accessUrls?.['ssh']

  const connectOptions: ConnectOption[] = [
    {
      id: 'code-server',
      name: 'code-server',
      description: 'Browser-based VS Code instance running in the workspace',
      icon: <Globe className="h-5 w-5" />,
      type: 'web',
      available: !!codeServerUrl,
      url: codeServerUrl
    },
    {
      id: 'terminal',
      name: 'Web Terminal (ttyd)',
      description: 'Terminal access in your browser with Fish shell',
      icon: <Globe className="h-5 w-5" />,
      type: 'web',
      available: !!ttydUrl,
      url: ttydUrl
    },
    {
      id: 'ssh',
      name: 'SSH',
      description: 'Connect via SSH terminal',
      icon: <Terminal className="h-5 w-5" />,
      type: 'terminal',
      available: !!workspace.status?.podIP,
      command: sshPort
        ? `ssh -p ${sshPort} kl@${workspace.status?.podIP || 'workspace'}`
        : `ssh kl@${workspace.status?.podIP || 'workspace'}`
    },
    {
      id: 'vscode-desktop',
      name: 'VS Code Desktop',
      description: 'Connect with Visual Studio Code via SSH Remote',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: !!workspace.status?.podIP,
      command: `code --remote ssh-remote+kl@${workspace.status?.podIP || 'workspace'} /workspace`
    },
    {
      id: 'claude-code',
      name: 'Claude Code',
      description: 'Connect with Claude Code AI assistant via SSH',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: !!workspace.status?.podIP,
      command: `claude-code --ssh kl@${workspace.status?.podIP || 'workspace'}`
    },
    {
      id: 'cursor-desktop',
      name: 'Cursor Desktop',
      description: 'Connect with Cursor IDE via SSH Remote',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: !!workspace.status?.podIP,
      command: `cursor --remote ssh-remote+kl@${workspace.status?.podIP || 'workspace'} /workspace`
    }
  ]

  const handleCopyCommand = (command: string, optionId: string) => {
    navigator.clipboard.writeText(command)
    setCopiedCommand(optionId)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const handleConnect = (option: ConnectOption) => {
    if (option.type === 'web' && option.url) {
      // For web-based options with URLs, open in new tab
      window.open(option.url, '_blank')
    } else if (option.command) {
      // For desktop/terminal options, copy command
      handleCopyCommand(option.command, option.id)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-medium">Connect to Workspace</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Choose your preferred development environment to connect to this workspace
        </p>
      </div>

      {/* Desktop Applications */}
      <div>
        <h3 className="text-sm font-medium mb-3">Desktop Applications</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          {connectOptions
            .filter(option => option.type === 'desktop')
            .map((option) => (
              <div
                key={option.id}
                className="bg-card rounded-lg border p-4 hover:border-border/80 transition-colors"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-muted rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium">{option.name}</h4>
                    <p className="text-xs text-muted-foreground mt-0.5">{option.description}</p>
                    {option.command && (
                      <div className="mt-2">
                        <button
                          onClick={() => handleCopyCommand(option.command!, option.id)}
                          className="text-xs text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                        >
                          {copiedCommand === option.id ? (
                            <>
                              <Check className="h-3 w-3" />
                              Copied!
                            </>
                          ) : (
                            <>
                              <Copy className="h-3 w-3" />
                              Copy command
                            </>
                          )}
                        </button>
                      </div>
                    )}
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleConnect(option)}
                    disabled={!option.available}
                  >
                    Connect
                  </Button>
                </div>
              </div>
            ))}
        </div>
      </div>

      {/* Web Applications */}
      <div>
        <h3 className="text-sm font-medium mb-3">Web Applications</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          {connectOptions
            .filter(option => option.type === 'web')
            .map((option) => (
              <div
                key={option.id}
                className="bg-card rounded-lg border p-4 hover:border-border/80 transition-colors"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-muted rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium">{option.name}</h4>
                    <p className="text-xs text-muted-foreground mt-0.5">{option.description}</p>
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleConnect(option)}
                    disabled={!option.available}
                    className="gap-1"
                  >
                    Open
                    <ExternalLink className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            ))}
        </div>
      </div>

      {/* Terminal */}
      <div>
        <h3 className="text-sm font-medium mb-3">Terminal</h3>
        <div className="grid gap-3">
          {connectOptions
            .filter(option => option.type === 'terminal')
            .map((option) => (
              <div
                key={option.id}
                className="bg-card rounded-lg border p-4"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-muted rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium">{option.name}</h4>
                    <p className="text-xs text-muted-foreground mt-0.5">{option.description}</p>
                    {option.command && (
                      <div className="mt-2 flex items-center gap-2">
                        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                          {option.command}
                        </code>
                        <button
                          onClick={() => handleCopyCommand(option.command!, option.id)}
                          className="text-xs text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                        >
                          {copiedCommand === option.id ? (
                            <>
                              <Check className="h-3 w-3" />
                              Copied!
                            </>
                          ) : (
                            <>
                              <Copy className="h-3 w-3" />
                              Copy
                            </>
                          )}
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
        </div>
      </div>
    </div>
  )
}