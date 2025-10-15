import { Network, Wifi } from 'lucide-react'
import type { K8sService } from '@/types/service'
import type { ServiceIntercept } from '@/types/serviceintercept'
import { Badge } from '@/components/ui/badge'

interface ServicesListProps {
  services: K8sService[]
  namespace: string
  serviceIntercepts: ServiceIntercept[]
}

export function ServicesList({ services, namespace, serviceIntercepts }: ServicesListProps) {
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
        <div>
          <h3 className="text-lg font-medium">Services</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Kubernetes services in namespace: {namespace}
          </p>
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
      <div className="mb-4">
        <h3 className="text-lg font-medium">Services</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Kubernetes services in namespace: {namespace}
        </p>
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
                return (
                  <tr key={`${service.namespace}-${service.name}`} className="hover:bg-muted/50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <Network className="h-5 w-5 text-muted-foreground mr-3" />
                        <span className="text-sm font-medium">{service.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm font-mono">
                        {service.name}.{service.namespace}.svc.cluster.local
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm font-mono">{service.clusterIP}</span>
                    </td>
                    <td className="px-6 py-4">
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
