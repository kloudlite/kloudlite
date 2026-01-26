'use client'

import { Badge, Separator, Button } from '@kloudlite/ui'
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
        <h2 className="text-xl font-semibold">Installation Details</h2>
        <p className="text-muted-foreground mt-1 text-base">View and manage your installation configuration</p>
      </div>
      <div className="space-y-6">
        {/* Status */}
        <div>
          <label className="text-foreground text-base font-medium">Status</label>
          <div className="mt-2 flex items-center gap-3">
            <Badge className={status.color}>{status.label}</Badge>
            <span className="text-muted-foreground text-base">{status.description}</span>
          </div>
        </div>

        <Separator />

        {/* Installation Key */}
        <div>
          <label className="text-foreground text-base font-medium">Installation Key</label>
          <div className="mt-2 flex items-center gap-2">
            <code className="bg-muted flex-1 px-3 py-2 font-mono text-base">
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
          <p className="text-muted-foreground mt-1 text-base">
            Use this key during installation deployment
          </p>
        </div>

        {/* Domain */}
        {installation.subdomain && installationUrl && (
          <>
            <Separator />
            <div>
              <label className="text-foreground text-base font-medium">Installation URL</label>
              <div className="mt-2 flex items-center gap-2">
                <a
                  href={installationUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary flex-1 font-mono text-base hover:underline"
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
                <p className="text-muted-foreground mt-1 text-sm">
                  Reserved on {new Date(installation.reservedAt).toLocaleDateString()}
                </p>
              )}
            </div>
          </>
        )}

        {/* Timestamps */}
        <Separator />
        <div className="grid grid-cols-2 gap-4 text-base">
          <div>
            <label className="text-foreground font-medium">Created</label>
            <p className="text-muted-foreground mt-1">
              {new Date(installation.createdAt).toLocaleString()}
            </p>
          </div>
          {installation.lastHealthCheck && (
            <div>
              <label className="text-foreground font-medium">Last Health Check</label>
              <p className="text-muted-foreground mt-1">
                {new Date(installation.lastHealthCheck).toLocaleString()}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
