import * as React from 'react'
import {
  Badge,
  Button,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@kloudlite/ui'

export const Open = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border p-4">
    <div className="flex items-center justify-between gap-4">
      <div>
        <p className="text-sm font-medium">Node pool &middot; worker-spot</p>
        <p className="text-xs text-muted-foreground">ap-south-1 &middot; 6 nodes</p>
      </div>
      <CollapsibleTrigger asChild>
        <Button variant="outline" size="sm">
          Hide details
        </Button>
      </CollapsibleTrigger>
    </div>
    <CollapsibleContent className="mt-4 space-y-2 text-sm">
      <div className="flex justify-between">
        <span className="text-muted-foreground">Instance type</span>
        <span className="font-medium">c6a.2xlarge</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Autoscaling</span>
        <span className="font-medium">3 &ndash; 12 nodes</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Taints</span>
        <span className="font-medium">kloudlite.io/spot=true</span>
      </div>
    </CollapsibleContent>
  </Collapsible>
)

export const Closed = () => (
  <Collapsible className="w-96 rounded-lg border p-4">
    <div className="flex items-center justify-between gap-4">
      <div className="flex items-center gap-2">
        <p className="text-sm font-medium">Advanced install options</p>
        <Badge variant="secondary">optional</Badge>
      </div>
      <CollapsibleTrigger asChild>
        <Button variant="ghost" size="sm">
          Show
        </Button>
      </CollapsibleTrigger>
    </div>
    <CollapsibleContent className="mt-4 text-sm text-muted-foreground">
      Override the wireguard subnet, ingress class and storage class used by the
      Kloudlite operator during installation.
    </CollapsibleContent>
  </Collapsible>
)

export const NestedList = () => (
  <div className="w-96 rounded-lg border">
    <Collapsible defaultOpen>
      <CollapsibleTrigger asChild>
        <Button variant="ghost" className="w-full justify-between p-4">
          <span className="font-medium">Environments (3)</span>
          <span className="text-xs text-muted-foreground">workspace: platform</span>
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="border-t p-4 space-y-2 text-sm">
        <div className="flex items-center justify-between">
          <span>development</span>
          <Badge>Active</Badge>
        </div>
        <div className="flex items-center justify-between">
          <span>staging</span>
          <Badge variant="secondary">Idle</Badge>
        </div>
        <div className="flex items-center justify-between">
          <span>production</span>
          <Badge variant="outline">Locked</Badge>
        </div>
      </CollapsibleContent>
    </Collapsible>
  </div>
)
