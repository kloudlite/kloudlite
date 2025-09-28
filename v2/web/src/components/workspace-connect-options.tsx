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

interface WorkspaceConnectOptionsProps {
  workspaceId: string
}

interface ConnectOption {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  type: 'desktop' | 'web' | 'terminal'
  available: boolean
  command?: string
}

export function WorkspaceConnectOptions({ workspaceId }: WorkspaceConnectOptionsProps) {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  const connectOptions: ConnectOption[] = [
    {
      id: 'vscode-desktop',
      name: 'VS Code Desktop',
      description: 'Connect with Visual Studio Code on your machine',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: true,
      command: `code --folder-uri vscode-remote://kloudlite-workspace-${workspaceId}`
    },
    {
      id: 'claude-code',
      name: 'Claude Code',
      description: 'Connect with Claude Code AI assistant',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: true,
      command: `claude-code connect workspace-${workspaceId}`
    },
    {
      id: 'cursor-desktop',
      name: 'Cursor Desktop',
      description: 'Connect with Cursor IDE',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: true,
      command: `cursor --folder-uri cursor-remote://kloudlite-workspace-${workspaceId}`
    },
    {
      id: 'goland',
      name: 'GoLand',
      description: 'Connect with JetBrains GoLand IDE',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: true,
      command: `goland --remote workspace-${workspaceId}`
    },
    {
      id: 'vscode-web',
      name: 'VS Code Web',
      description: 'Open Visual Studio Code in your browser',
      icon: <Globe className="h-5 w-5" />,
      type: 'web',
      available: true
    },
    {
      id: 'zed',
      name: 'Zed',
      description: 'Connect with Zed collaborative editor',
      icon: <Code2 className="h-5 w-5" />,
      type: 'desktop',
      available: true,
      command: `zed remote://workspace-${workspaceId}`
    },
    {
      id: 'code-server',
      name: 'code-server',
      description: 'Browser-based VS Code instance',
      icon: <Globe className="h-5 w-5" />,
      type: 'web',
      available: true
    },
    {
      id: 'terminal',
      name: 'Terminal',
      description: 'Connect via SSH terminal',
      icon: <Terminal className="h-5 w-5" />,
      type: 'terminal',
      available: true,
      command: `ssh workspace-${workspaceId}@dev.kloudlite.io`
    }
  ]

  const handleCopyCommand = (command: string, optionId: string) => {
    navigator.clipboard.writeText(command)
    setCopiedCommand(optionId)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const handleConnect = (option: ConnectOption) => {
    if (option.type === 'web') {
      // For web-based options, open in new tab
      window.open(`/workspaces/${workspaceId}/connect/${option.id}`, '_blank')
    } else if (option.command) {
      // For desktop/terminal options, copy command
      handleCopyCommand(option.command, option.id)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-medium text-gray-900">Connect to Workspace</h2>
        <p className="text-sm text-gray-600 mt-1">
          Choose your preferred development environment to connect to this workspace
        </p>
      </div>

      {/* Desktop Applications */}
      <div>
        <h3 className="text-sm font-medium text-gray-700 mb-3">Desktop Applications</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          {connectOptions
            .filter(option => option.type === 'desktop')
            .map((option) => (
              <div
                key={option.id}
                className="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition-colors"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium text-gray-900">{option.name}</h4>
                    <p className="text-xs text-gray-600 mt-0.5">{option.description}</p>
                    {option.command && (
                      <div className="mt-2">
                        <button
                          onClick={() => handleCopyCommand(option.command!, option.id)}
                          className="text-xs text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1"
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
        <h3 className="text-sm font-medium text-gray-700 mb-3">Web Applications</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          {connectOptions
            .filter(option => option.type === 'web')
            .map((option) => (
              <div
                key={option.id}
                className="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition-colors"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium text-gray-900">{option.name}</h4>
                    <p className="text-xs text-gray-600 mt-0.5">{option.description}</p>
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
        <h3 className="text-sm font-medium text-gray-700 mb-3">Terminal</h3>
        <div className="grid gap-3">
          {connectOptions
            .filter(option => option.type === 'terminal')
            .map((option) => (
              <div
                key={option.id}
                className="bg-gray-900 text-white rounded-lg border border-gray-800 p-4"
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-10 h-10 bg-gray-800 rounded-lg flex items-center justify-center">
                    {option.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-medium">{option.name}</h4>
                    <p className="text-xs text-gray-400 mt-0.5">{option.description}</p>
                    {option.command && (
                      <div className="mt-2 flex items-center gap-2">
                        <code className="text-xs bg-gray-800 px-2 py-1 rounded font-mono">
                          {option.command}
                        </code>
                        <button
                          onClick={() => handleCopyCommand(option.command!, option.id)}
                          className="text-xs text-blue-400 hover:text-blue-300 font-medium flex items-center gap-1"
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