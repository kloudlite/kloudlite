import * as React from 'react'
import { Progress } from '@kloudlite/ui'

export const ValueSweep = () => (
  <div className="w-96 space-y-4">
    {[0, 35, 72, 100].map((v) => (
      <div key={v} className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Pulling layers</span>
          <span className="font-medium">{v}%</span>
        </div>
        <Progress value={v} />
      </div>
    ))}
  </div>
)

export const InstallStep = () => (
  <div className="w-96 space-y-2">
    <div className="flex items-center justify-between text-sm">
      <span className="font-medium">Installing kloudlite-agent</span>
      <span className="text-muted-foreground">Step 3 of 5</span>
    </div>
    <Progress value={60} />
    <p className="text-xs text-muted-foreground">
      Applying CRDs to cluster kloudlite-prod &middot; ap-south-1
    </p>
  </div>
)

export const Thick = () => (
  <div className="w-80 space-y-2">
    <Progress value={45} className="h-4" />
    <p className="text-xs text-muted-foreground">
      Image pull &middot; 1.4 GB of 3.1 GB
    </p>
  </div>
)
