'use client'

import { Button } from '@kloudlite/ui'
import { Copy, Key } from 'lucide-react'
import { toast } from 'sonner'
import type { Installation } from '@/lib/console/storage'

interface InstallationDetailsCardProps {
  installation: Installation
}

export function InstallationDetailsCard({
  installation,
}: InstallationDetailsCardProps) {
  const copyToClipboard = () => {
    navigator.clipboard.writeText(installation.installationKey)
    toast.success('Installation key copied to clipboard')
  }

  return (
    <div className="flex items-center gap-3 rounded-lg border border-foreground/10 bg-background px-4 py-3">
      <Key className="h-4 w-4 text-muted-foreground shrink-0" />
      <span className="text-sm text-muted-foreground shrink-0">Installation Key</span>
      <code className="font-mono text-sm text-foreground truncate">
        {installation.installationKey}
      </code>
      <Button
        variant="ghost"
        size="sm"
        className="h-7 w-7 p-0 shrink-0 ml-auto"
        onClick={copyToClipboard}
      >
        <Copy className="h-3.5 w-3.5" />
      </Button>
    </div>
  )
}
