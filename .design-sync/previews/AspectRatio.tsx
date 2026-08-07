import * as React from 'react'
import { AspectRatio, Badge } from '@kloudlite/ui'

export const Widescreen = () => (
  <div className="w-96">
    <AspectRatio ratio={16 / 9}>
      <div className="flex h-full w-full flex-col items-center justify-center gap-2 bg-muted text-muted-foreground">
        <span className="text-sm font-medium">Cluster topology &middot; 16:9</span>
        <span className="text-xs">ap-south-1 &middot; 12 nodes</span>
      </div>
    </AspectRatio>
  </div>
)

export const Square = () => (
  <div className="w-64">
    <AspectRatio ratio={1}>
      <div className="flex h-full w-full items-center justify-center bg-accent text-accent-foreground">
        <span className="text-lg font-semibold">kloudlite-dev</span>
      </div>
    </AspectRatio>
  </div>
)

export const WithOverlayContent = () => (
  <div className="w-96">
    <AspectRatio ratio={21 / 9}>
      <div className="flex h-full w-full items-end justify-between bg-secondary p-4 text-secondary-foreground">
        <div className="flex flex-col gap-1">
          <span className="text-sm font-semibold">payments-api</span>
          <span className="text-xs text-muted-foreground">
            ghcr.io/kloudlite/payments-api:v2.14.0
          </span>
        </div>
        <Badge variant="success">Running</Badge>
      </div>
    </AspectRatio>
  </div>
)
