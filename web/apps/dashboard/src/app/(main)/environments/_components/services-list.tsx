'use client'

import { Network, Copy, Check, AlertCircle, AlertTriangle, Terminal } from 'lucide-react'
import { useState } from 'react'
import type { K8sService, CompositionSpec, CompositionStatus } from '@kloudlite/types'
import { Alert, AlertDescription, AlertTitle, Badge, Button } from '@kloudlite/ui'
import { ServiceLogsViewer } from './service-logs-viewer'
import { useResourceWatch } from '@/lib/hooks/use-resource-watch'

interface ServicesListProps {
  services: K8sService[]
  namespace: string
  environmentName: string // Environment name for compose updates
  compose: CompositionSpec | null
  composeStatus: CompositionStatus | null
  envHash: string // Hash from environment.status.hash
  subdomain: string // Subdomain from environment.status.subdomain
  isEnvActive?: boolean // Whether the environment is active
}

export function ServicesList({
  services,
  namespace,
  environmentName: _environmentName,
  compose: _compose,
  composeStatus,
  envHash,
  subdomain,
  isEnvActive = true,
}: ServicesListProps) {
  const [copiedDns, setCopiedDns] = useState<string | null>(null)
  const [logsService, setLogsService] = useState<string | null>(null)
  const [announcement, setAnnouncement] = useState('')

  // Watch for service changes in this environment's namespace
  useResourceWatch('services', namespace)

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
      setAnnouncement(`Copied DNS hostname ${hostname}.`)
      setTimeout(() => setCopiedDns(null), 2000)
    } catch (err) {
      console.error('Failed to copy DNS:', err)
      setAnnouncement('Failed to copy DNS hostname.')
    }
  }
  if (services.length === 0) {
    return (
      <>
        {isEnvActive && composeStatus?.state === 'failed' && (
          <Alert variant="destructive" className="mb-4">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Composition Failed</AlertTitle>
            <AlertDescription>
              {composeStatus.message || 'The composition failed to deploy'}
            </AlertDescription>
          </Alert>
        )}

        {isEnvActive && composeStatus?.state === 'degraded' && (
          <Alert variant="destructive" className="mb-4">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Composition Degraded</AlertTitle>
            <AlertDescription>
              {composeStatus.message || 'Some services are not running properly'}
            </AlertDescription>
          </Alert>
        )}

        <div className="bg-muted/50 rounded-lg border py-12 text-center">
          <Network className="text-muted-foreground mx-auto h-12 w-12" />
          <h3 className="mt-2 text-sm font-medium">No services found</h3>
          <p className="text-muted-foreground mt-1 text-sm">
            No services exist in this environment yet.
          </p>
        </div>
      </>
    )
  }

  return (
    <>
      <p className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {announcement || `${services.length} services loaded.`}
      </p>
      {isEnvActive && composeStatus?.state === 'failed' && (
        <Alert variant="destructive" className="mb-4">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Composition Failed</AlertTitle>
          <AlertDescription>
            {composeStatus.message || 'The composition failed to deploy'}
          </AlertDescription>
        </Alert>
      )}

      {isEnvActive && composeStatus?.state === 'degraded' && (
        <Alert variant="destructive" className="mb-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Composition Degraded</AlertTitle>
          <AlertDescription>
            {composeStatus.message || 'Some services are not running properly'}
          </AlertDescription>
        </Alert>
      )}

      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-muted/50 border-b">
                <th className="text-muted-foreground w-48 px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Name
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  DNS
                </th>
                <th className="text-muted-foreground w-36 px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  IP
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Ports
                </th>
                <th className="text-muted-foreground w-24 px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {services.map((service) => {
                const isHeadless = service.clusterIP === 'None'
                return (
                  <tr key={`${service.namespace}-${service.name}`} className="hover:bg-muted/50">
                    <td className="w-48 px-6 py-4 whitespace-nowrap">
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
                    <td className="w-36 px-6 py-4 whitespace-nowrap">
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
                          {service.ports.map((port) => (
                            <div key={port.port} className="text-sm">
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
                    <td className="w-24 px-6 py-4 text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setLogsService(service.name)}
                        className="gap-1.5"
                      >
                        <Terminal className="h-4 w-4" />
                        Logs
                      </Button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Service Logs Viewer - key forces remount when service changes */}
      <ServiceLogsViewer
        key={`${namespace}-${logsService}`}
        serviceName={logsService || ''}
        namespace={namespace}
        isOpen={!!logsService}
        onClose={() => setLogsService(null)}
      />
    </>
  )
}
