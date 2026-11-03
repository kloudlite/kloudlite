import * as React from 'react'
import { Badge } from '@kloudlite/ui'

export const Variants = () => (
  <div className="flex flex-wrap items-center gap-2">
    <Badge>Default</Badge>
    <Badge variant="secondary">Secondary</Badge>
    <Badge variant="destructive">Destructive</Badge>
    <Badge variant="success">Success</Badge>
    <Badge variant="warning">Warning</Badge>
    <Badge variant="info">Info</Badge>
    <Badge variant="outline">Outline</Badge>
  </div>
)

export const InstallationStatuses = () => (
  <div className="flex flex-wrap items-center gap-2">
    <Badge variant="success">Active</Badge>
    <Badge variant="info">Installing</Badge>
    <Badge variant="warning">Degraded</Badge>
    <Badge variant="destructive">Failed</Badge>
    <Badge variant="secondary">Suspended</Badge>
  </div>
)

export const InContext = () => (
  <div className="flex w-96 flex-col gap-3">
    <div className="flex items-center justify-between gap-3 border border-border p-3">
      <div className="flex flex-col">
        <span className="text-sm font-medium">kloudlite-dev</span>
        <span className="text-xs text-muted-foreground">ap-south-1 &middot; v1.31.4</span>
      </div>
      <Badge variant="success">Active</Badge>
    </div>
    <div className="flex items-center justify-between gap-3 border border-border p-3">
      <div className="flex flex-col">
        <span className="text-sm font-medium">staging-eu</span>
        <span className="text-xs text-muted-foreground">eu-west-1 &middot; v1.30.8</span>
      </div>
      <Badge variant="warning">Degraded</Badge>
    </div>
  </div>
)

export const WithCounts = () => (
  <div className="flex flex-wrap items-center gap-2">
    <Badge variant="outline">3 intercepts</Badge>
    <Badge variant="outline">12 nodes</Badge>
    <Badge variant="secondary">v2.14.0</Badge>
  </div>
)
