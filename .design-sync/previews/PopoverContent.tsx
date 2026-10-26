import * as React from 'react'
import {
  Badge,
  Button,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@kloudlite/ui'

export const AlignedStart = () => (
  <div className="flex h-80 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Deployment status</Button>
      </PopoverTrigger>
      <PopoverContent align="start">
        <div className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <p className="text-sm font-medium">api-gateway</p>
            <Badge>Healthy</Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            3/3 replicas ready &middot; rolled out 12 minutes ago.
          </p>
        </div>
      </PopoverContent>
    </Popover>
  </div>
)

export const WithActions = () => (
  <div className="flex h-96 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Delete environment</Button>
      </PopoverTrigger>
      <PopoverContent align="start">
        <div className="space-y-4">
          <div className="space-y-1">
            <p className="text-sm font-medium">Delete &ldquo;staging&rdquo;?</p>
            <p className="text-sm text-muted-foreground">
              All workspaces in this environment are torn down immediately.
            </p>
          </div>
          <div className="flex gap-2">
            <Button size="sm" variant="destructive">
              Delete
            </Button>
            <Button size="sm" variant="outline">
              Cancel
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  </div>
)
