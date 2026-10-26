import * as React from 'react'
import {
  Badge,
  Button,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@kloudlite/ui'

export const OpenContent = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border p-4">
    <div className="flex items-center justify-between gap-4">
      <p className="text-sm font-medium">Cluster kloudlite-prod</p>
      <CollapsibleTrigger asChild>
        <Button variant="outline" size="sm">
          Hide
        </Button>
      </CollapsibleTrigger>
    </div>
    <CollapsibleContent className="mt-4 space-y-2 text-sm">
      <div className="flex justify-between">
        <span className="text-muted-foreground">Region</span>
        <span className="font-medium">eu-west-1</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Kubernetes</span>
        <span className="font-medium">v1.31.4</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Namespaces</span>
        <span className="font-medium">18</span>
      </div>
    </CollapsibleContent>
  </Collapsible>
)

export const RichContent = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border p-4">
    <CollapsibleTrigger asChild>
      <Button variant="ghost" size="sm" className="px-0">
        Tunnel diagnostics
      </Button>
    </CollapsibleTrigger>
    <CollapsibleContent className="mt-3 space-y-3">
      <p className="text-sm text-muted-foreground">
        The wireguard peer is connected and handshaking every 25 seconds.
      </p>
      <div className="flex items-center gap-2">
        <Badge>Connected</Badge>
        <Badge variant="secondary">10.13.0.4/32</Badge>
      </div>
      <div className="rounded-md bg-muted p-4 text-xs">
        latest handshake: 12s ago{'\n'}transfer: 2.4 MiB received, 812 KiB sent
      </div>
    </CollapsibleContent>
  </Collapsible>
)

export const MultipleContents = () => (
  <div className="w-96 space-y-2">
    <Collapsible defaultOpen className="rounded-lg border p-4">
      <CollapsibleTrigger className="text-sm font-medium">
        Environment: development
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2 text-sm text-muted-foreground">
        6 deployments &middot; 2 intercepts active
      </CollapsibleContent>
    </Collapsible>
    <Collapsible defaultOpen className="rounded-lg border p-4">
      <CollapsibleTrigger className="text-sm font-medium">
        Environment: staging
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2 text-sm text-muted-foreground">
        6 deployments &middot; no intercepts
      </CollapsibleContent>
    </Collapsible>
  </div>
)
