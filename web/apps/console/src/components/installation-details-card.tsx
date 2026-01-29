'use client'

import { Badge, Button } from '@kloudlite/ui'
import { Copy, ExternalLink } from 'lucide-react'
import { toast } from 'sonner'
import type { Installation } from '@/lib/console/supabase-storage-service'

interface InstallationDetailsCardProps {
  installation: Installation
  status: {
    label: string
    color: string
    description: string
  }
  domain: string
  installationUrl: string | null
}

export function InstallationDetailsCard({
  installation,
  status,
  domain,
  installationUrl,
}: InstallationDetailsCardProps) {
  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-foreground">Installation Details</h2>
        <p className="text-muted-foreground mt-1 text-sm">View and manage your installation configuration</p>
      </div>
      <div className="space-y-5">
        {/* Status */}
        <div>
          <label className="text-foreground text-sm font-medium">Status</label>
          <div className="mt-2 flex items-center gap-3">
            <span className={`inline-flex px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider rounded-md border ${status.color}`}>
              {status.label}
            </span>
            <span className="text-muted-foreground text-sm">{status.description}</span>
          </div>
        </div>

        <div className="h-px bg-foreground/10" />

        {/* Installation Key */}
        <div>
          <label className="text-foreground text-sm font-medium">Installation Key</label>
          <div className="mt-2 flex items-center gap-2">
            <code className="bg-muted border border-foreground/10 flex-1 px-3 py-2 rounded-md font-mono text-sm">
              {installation.installationKey}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={() => copyToClipboard(installation.installationKey, 'Installation key')}
            >
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <p className="text-muted-foreground mt-2 text-xs">
            Use this key during installation deployment
          </p>
        </div>

        {/* Domain */}
        {installation.subdomain && installationUrl && (
          <>
            <div className="h-px bg-foreground/10" />
            <div>
              <label className="text-foreground text-sm font-medium">Installation URL</label>
              <div className="mt-2 flex items-center gap-2">
                <a
                  href={installationUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary flex-1 font-mono text-sm hover:underline"
                >
                  {installation.subdomain}.{domain}
                </a>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(installationUrl, '_blank')}
                >
                  <ExternalLink className="h-4 w-4" />
                </Button>
              </div>
              {installation.reservedAt && (
                <p className="text-muted-foreground mt-2 text-xs">
                  Reserved on {new Date(installation.reservedAt).toLocaleDateString()}
                </p>
              )}
            </div>
          </>
        )}

        {/* Timestamps */}
        <div className="h-px bg-foreground/10" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="text-foreground text-sm font-medium">Created</label>
            <p className="text-muted-foreground mt-2 text-sm">
              {new Date(installation.createdAt).toLocaleString()}
            </p>
          </div>
          {installation.lastHealthCheck && (
            <div>
              <label className="text-foreground text-sm font-medium">Last Health Check</label>
              <p className="text-muted-foreground mt-2 text-sm">
                {new Date(installation.lastHealthCheck).toLocaleString()}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
