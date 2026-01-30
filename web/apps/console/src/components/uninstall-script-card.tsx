'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
import { Copy, Check, Terminal } from 'lucide-react'
import { toast } from 'sonner'

interface UninstallScriptCardProps {
  secretKey: string
}

export function UninstallScriptCard({ secretKey }: UninstallScriptCardProps) {
  const [copied, setCopied] = useState(false)

  const uninstallCommand = `curl -fsSL https://get.khost.dev/uninstall | bash -s -- --key ${secretKey}`

  const copyCommand = () => {
    navigator.clipboard.writeText(uninstallCommand)
    setCopied(true)
    toast.success('Uninstall command copied to clipboard')
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-4">
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Terminal className="h-4 w-4 text-red-600 dark:text-red-400" />
          <p className="text-foreground text-sm font-semibold">Uninstall Script</p>
        </div>
        <p className="text-muted-foreground text-sm">
          Run this command in your terminal to uninstall Kloudlite from your cloud infrastructure.
        </p>
      </div>

      <div className="bg-muted/50 border border-red-500/20 rounded-lg p-4">
        <div className="flex items-start justify-between gap-4">
          <code className="flex-1 font-mono text-sm leading-relaxed break-all text-red-700 dark:text-red-300">
            {uninstallCommand}
          </code>
          <Button
            variant="outline"
            size="sm"
            className="flex-shrink-0"
            onClick={copyCommand}
          >
            {copied ? (
              <>
                <Check className="mr-2 size-4 text-green-600" />
                Copied
              </>
            ) : (
              <>
                <Copy className="mr-2 size-4" />
                Copy
              </>
            )}
          </Button>
        </div>
      </div>

      <p className="text-xs text-muted-foreground">
        This will remove all Kloudlite resources from your cloud account including VMs, networking, and storage.
        Make sure to backup any important data before running this command.
      </p>
    </div>
  )
}
