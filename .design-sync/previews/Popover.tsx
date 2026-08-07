import * as React from 'react'
import {
  Button,
  Input,
  Label,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@kloudlite/ui'

export const Basic = () => (
  <div className="flex h-80 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Cluster details</Button>
      </PopoverTrigger>
      <PopoverContent align="start">
        <div className="space-y-2">
          <p className="text-sm font-medium">kloudlite-prod</p>
          <p className="text-sm text-muted-foreground">
            ap-south-1 &middot; 12 nodes &middot; Kubernetes v1.31.4
          </p>
        </div>
      </PopoverContent>
    </Popover>
  </div>
)

export const WithForm = () => (
  <div className="flex h-96 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Scale nodepool</Button>
      </PopoverTrigger>
      <PopoverContent align="start">
        <div className="space-y-4">
          <div className="space-y-1">
            <p className="text-sm font-medium">Nodepool capacity</p>
            <p className="text-sm text-muted-foreground">
              Applies to the spot pool in ap-south-1.
            </p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="min-nodes">Min nodes</Label>
            <Input id="min-nodes" defaultValue="2" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="max-nodes">Max nodes</Label>
            <Input id="max-nodes" defaultValue="10" />
          </div>
          <Button size="sm">Apply</Button>
        </div>
      </PopoverContent>
    </Popover>
  </div>
)

export const SideRight = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Environment actions</Button>
      </PopoverTrigger>
      <PopoverContent side="right" align="start">
        <div className="space-y-2 text-sm">
          <p className="font-medium">staging</p>
          <p className="text-muted-foreground">Clone environment</p>
          <p className="text-muted-foreground">Suspend workspaces</p>
        </div>
      </PopoverContent>
    </Popover>
  </div>
)
