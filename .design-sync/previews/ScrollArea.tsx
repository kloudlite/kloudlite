import * as React from 'react'
import { ScrollArea } from '@kloudlite/ui'

const namespaces = [
  'kl-core',
  'kl-tenant',
  'kl-account-kloudlite',
  'kube-system',
  'ingress-nginx',
  'cert-manager',
  'observability',
  'vector',
  'operators',
  'kl-gateway',
  'kl-workspaces',
  'default',
]

const logLines = [
  'reconciling deployment/api-gateway',
  'pulling ghcr.io/kloudlite/api:v1.2.9',
  'created pod api-gateway-7d9f4c',
  'waiting for readiness probe',
  'readiness probe succeeded',
  'scaled replicaset to 3',
  'reconciling service/api-gateway',
  'updated endpoints slice',
  'reconciling ingress/api',
  'certificate issued by letsencrypt',
  'rollout complete',
  'watch reconnected after 12s',
]

export const NamespaceList = () => (
  <ScrollArea type="always" className="h-64 w-80 border">
    <div className="p-3 text-sm">
      {namespaces.map((ns) => (
        <div key={ns} className="border-b py-2">
          {ns}
        </div>
      ))}
    </div>
  </ScrollArea>
)

export const LogTail = () => (
  <ScrollArea type="always" className="h-64 w-96 border bg-muted">
    <div className="p-3 font-mono text-xs leading-relaxed">
      {logLines.map((line, i) => (
        <p key={i}>
          <span className="text-muted-foreground">10:2{i % 10}:14</span> {line}
        </p>
      ))}
    </div>
  </ScrollArea>
)

export const ShortContent = () => (
  <ScrollArea type="always" className="h-64 w-80 border">
    <div className="p-3 text-sm">
      <p className="font-medium">Attached clusters</p>
      <p className="text-muted-foreground">kloudlite-prod</p>
      <p className="text-muted-foreground">kloudlite-staging</p>
    </div>
  </ScrollArea>
)
