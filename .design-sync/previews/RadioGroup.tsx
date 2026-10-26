import * as React from 'react'
import { Label, RadioGroup, RadioGroupItem } from '@kloudlite/ui'

export const ClusterSize = () => (
  <RadioGroup defaultValue="medium" className="w-80">
    <div className="flex items-center gap-3">
      <RadioGroupItem value="small" id="size-small" />
      <Label htmlFor="size-small">Small &middot; 2 vCPU, 4 GB</Label>
    </div>
    <div className="flex items-center gap-3">
      <RadioGroupItem value="medium" id="size-medium" />
      <Label htmlFor="size-medium">Medium &middot; 4 vCPU, 8 GB</Label>
    </div>
    <div className="flex items-center gap-3">
      <RadioGroupItem value="large" id="size-large" />
      <Label htmlFor="size-large">Large &middot; 8 vCPU, 16 GB</Label>
    </div>
  </RadioGroup>
)

export const WithDisabledOption = () => (
  <RadioGroup defaultValue="rolling" className="w-96">
    <div className="flex items-start gap-3">
      <RadioGroupItem value="rolling" id="strategy-rolling" className="mt-1" />
      <div className="space-y-1">
        <Label htmlFor="strategy-rolling">Rolling update</Label>
        <p className="text-xs text-muted-foreground">
          Replace pods gradually while keeping the service available.
        </p>
      </div>
    </div>
    <div className="flex items-start gap-3">
      <RadioGroupItem value="recreate" id="strategy-recreate" className="mt-1" />
      <div className="space-y-1">
        <Label htmlFor="strategy-recreate">Recreate</Label>
        <p className="text-xs text-muted-foreground">
          Terminate all pods before starting the new revision.
        </p>
      </div>
    </div>
    <div className="flex items-start gap-3">
      <RadioGroupItem value="canary" id="strategy-canary" disabled className="mt-1" />
      <div className="space-y-1">
        <Label htmlFor="strategy-canary">Canary</Label>
        <p className="text-xs text-muted-foreground">
          Requires a service mesh &mdash; not enabled on this cluster.
        </p>
      </div>
    </div>
  </RadioGroup>
)

export const Horizontal = () => (
  <RadioGroup defaultValue="spot" className="flex gap-6">
    <div className="flex items-center gap-2">
      <RadioGroupItem value="spot" id="cap-spot" />
      <Label htmlFor="cap-spot">Spot</Label>
    </div>
    <div className="flex items-center gap-2">
      <RadioGroupItem value="ondemand" id="cap-ondemand" />
      <Label htmlFor="cap-ondemand">On-demand</Label>
    </div>
  </RadioGroup>
)
