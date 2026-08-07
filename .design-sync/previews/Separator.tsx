import * as React from 'react'
import { Separator } from '@kloudlite/ui'

export const Horizontal = () => (
  <div className="w-80">
    <div className="space-y-1">
      <h4 className="text-sm font-medium">kl-prod</h4>
      <p className="text-sm text-muted-foreground">
        Kubernetes 1.31 &middot; ap-south-1
      </p>
    </div>
    <Separator className="my-4" />
    <div className="space-y-1">
      <h4 className="text-sm font-medium">Node pools</h4>
      <p className="text-sm text-muted-foreground">3 pools &middot; 12 nodes</p>
    </div>
  </div>
)

export const Vertical = () => (
  <div className="flex h-5 items-center gap-4 text-sm">
    <span>Clusters</span>
    <Separator orientation="vertical" />
    <span>Environments</span>
    <Separator orientation="vertical" />
    <span>Workspaces</span>
  </div>
)

export const InCard = () => (
  <div className="w-80 border border-border p-6">
    <h4 className="text-sm font-semibold">Workspace usage</h4>
    <Separator className="my-4" />
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">CPU</span>
      <span>2.4 / 4 vCPU</span>
    </div>
    <Separator className="my-2" />
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">Memory</span>
      <span>5.1 / 8 GiB</span>
    </div>
  </div>
)

export const VerticalStats = () => (
  <div className="flex h-16 items-center gap-6">
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">Clusters</p>
      <p className="text-lg font-semibold">4</p>
    </div>
    <Separator orientation="vertical" />
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">Environments</p>
      <p className="text-lg font-semibold">11</p>
    </div>
    <Separator orientation="vertical" />
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">Running workspaces</p>
      <p className="text-lg font-semibold">27</p>
    </div>
  </div>
)
