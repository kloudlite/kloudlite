'use client'

import { Network, Copy, Check } from 'lucide-react'
import { useState } from 'react'
import type { K8sService } from '@kloudlite/types'
import type { Composition } from '@kloudlite/types'
import { Badge } from '@kloudlite/ui'
import { CompositionEditor } from './composition-editor'

interface ServicesListProps {
  services: K8sService[]
  namespace: string
  composition: Composition | null
  envHash: string // Hash from environment.status.hash
  subdomain: string // Subdomain from environment.status.subdomain
}

export function ServicesList({
  services,
  namespace,
  composition,
  envHash,
  subdomain,
}: ServicesListProps) {
  const [open, setOpen] = useState(false)
  const [copiedDns, setCopiedDns] = useState<string | null>(null)

  // Generate VPN-accessible DNS hostname: {service}-{hash}.{subdomain}
  const getDnsHostname = (serviceName: string) => {
    if (!envHash || !subdomain) {
      return `${serviceName}.${namespace}.svc.cluster.local`
    }
    return `${serviceName}-${envHash}.${subdomain}`
  }

  const copyDns = async (hostname: string) => {
    try {
      await navigator.clipboard.writeText(hostname)
      setCopiedDns(hostname)
      setTimeout(() => setCopiedDns(null), 2000)
    } catch (err) {
      console.error('Failed to copy DNS:', err)
    }
  }
  if (services.length === 0) {
    return (
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium">Services</h3>
            <p className="text-muted-foreground mt-1 text-sm">
              Manage and view services in this environment
            </p>
          </div>
          <CompositionEditor
            composition={composition}
            namespace={namespace}
            open={open}
            onOpenChange={setOpen}
          />
        </div>
        <div className="bg-muted/50 mt-8 rounded-lg border py-12 text-center">
          <Network className="text-muted-foreground mx-auto h-12 w-12" />
          <h3 className="mt-2 text-sm font-medium">No services found</h3>
          <p className="text-muted-foreground mt-1 text-sm">
            No services exist in this environment yet.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Services</h3>
          <p className="text-muted-foreground mt-1 text-sm">
            Manage and view services in this environment
          </p>
        </div>
        <CompositionEditor
          composition={composition}
          namespace={namespace}
          open={open}
          onOpenChange={setOpen}
        />
      </div>

      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-muted/50 border-b">
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Name
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  DNS
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  IP
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Ports
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {services.map((service) => {
                const isHeadless = service.clusterIP === 'None'
                return (
                  <tr key={`${service.namespace}-${service.name}`} className="hover:bg-muted/50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Network className="text-muted-foreground h-5 w-5" />
                        <span className="text-sm font-medium">{service.name}</span>
                        {isHeadless && (
                          <Badge variant="outline" className="text-xs">
                            Headless
                          </Badge>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-sm">
                          {getDnsHostname(service.name)}
                        </span>
                        <button
                          onClick={() => copyDns(getDnsHostname(service.name))}
                          className="text-muted-foreground hover:text-foreground p-1 rounded transition-colors"
                          title="Copy DNS hostname"
                        >
                          {copiedDns === getDnsHostname(service.name) ? (
                            <Check className="h-4 w-4 text-green-500" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="font-mono text-sm">
                        {isHeadless ? (
                          <span className="text-muted-foreground italic">None (Headless)</span>
                        ) : (
                          service.clusterIP
                        )}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      {service.ports.length > 0 ? (
                        <div className="space-y-1">
                          {service.ports.map((port, idx) => (
                            <div key={idx} className="text-sm">
                              {port.port}
                              {port.targetPort && port.targetPort !== String(port.port) && (
                                <span className="text-muted-foreground"> → {port.targetPort}</span>
                              )}
                              <span className="text-muted-foreground ml-1">/{port.protocol}</span>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-xs italic">
                          No ports exposed
                        </span>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
