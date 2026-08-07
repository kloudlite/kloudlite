import * as React from 'react'
import { Button, Spinner } from '@kloudlite/ui'

export const Sizes = () => (
  <div className="flex items-center gap-6">
    <Spinner />
    <Spinner className="size-6" />
    <Spinner className="size-8" />
    <Spinner className="size-10" />
  </div>
)

export const WithLabel = () => (
  <div className="flex items-center gap-3 text-sm text-muted-foreground">
    <Spinner className="size-6" />
    <span>Provisioning cluster nodes&hellip;</span>
  </div>
)

export const InButton = () => (
  <div className="flex items-center gap-3">
    <Button disabled>
      <Spinner />
      Deploying
    </Button>
    <Button variant="outline" disabled>
      <Spinner />
      Syncing
    </Button>
  </div>
)

export const InPanel = () => (
  <div className="flex h-40 w-80 flex-col items-center justify-center gap-3 border">
    <Spinner className="size-8 text-primary" />
    <p className="text-sm text-muted-foreground">Loading environments</p>
  </div>
)
