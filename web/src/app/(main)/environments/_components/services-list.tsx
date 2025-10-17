'use client'

import { Network, Wifi, Pencil } from 'lucide-react'
import { useState } from 'react'
import type { K8sService } from '@/types/service'
import type { ServiceIntercept } from '@/types/serviceintercept'
import type { Composition } from '@/types/composition'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { CompositionEditor } from './composition-editor'

interface ServicesListProps {
  services: K8sService[]
  namespace: string
  serviceIntercepts: ServiceIntercept[]
  composition: Composition | null
  user: string
}

export function ServicesList({ services, namespace, serviceIntercepts, composition, user }: ServicesListProps) {
  const [open, setOpen] = useState(false)

  // Helper function to find active intercept for a service
  const getActiveIntercept = (serviceName: string) => {
    return serviceIntercepts.find(
      (intercept) =>
        intercept.spec.serviceRef.name === serviceName &&
        intercept.spec.status === 'active' &&
        intercept.status?.phase === 'Active'
    )
  }
  if (services.length === 0) {
    return (
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium">Services</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Kubernetes services in namespace: {namespace}
            </p>
          </div>
          <CompositionEditor
            composition={composition}
            namespace={namespace}
            user={user}
            open={open}
            onOpenChange={setOpen}
          />
        </div>
        <div className="mt-8 text-center py-12 bg-muted/50 rounded-lg border">
          <Network className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium">No services found</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            No Kubernetes services exist in this namespace.
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
          <p className="text-sm text-muted-foreground mt-1">
            Kubernetes services in namespace: {namespace}
          </p>
        </div>
        <CompositionEditor
          composition={composition}
          namespace={namespace}
          user={user}
          open={open}
          onOpenChange={setOpen}
        />
      </div>

      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  DNS
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  IP
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Ports
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Intercept Status
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {services.map((service) => {
                const activeIntercept = getActiveIntercept(service.name)
                const isHeadless = service.clusterIP === 'None'
                return (
                  <tr key={`${service.namespace}-${service.name}`} className="hover:bg-muted/50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Network className="h-5 w-5 text-muted-foreground" />
                        <span className="text-sm font-medium">{service.name}</span>
                        {isHeadless && (
                          <Badge variant="outline" className="text-xs">
                            Headless
                          </Badge>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm font-mono">
                        {service.name}.{service.namespace}.svc.cluster.local
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm font-mono">
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
                        <span className="text-xs text-muted-foreground italic">No ports exposed</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {activeIntercept ? (
                        <div className="flex items-center gap-2">
                          <Badge variant="secondary" className="flex items-center gap-1">
                            <Wifi className="h-3 w-3" />
                            Intercepted
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            → {activeIntercept.spec.workspaceRef.name}
                          </span>
                        </div>
                      ) : (
                        <span className="text-xs text-muted-foreground">Not intercepted</span>
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
