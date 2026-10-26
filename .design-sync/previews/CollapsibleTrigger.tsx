import * as React from 'react'
import {
  Button,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@kloudlite/ui'

export const AsButton = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border p-4">
    <CollapsibleTrigger asChild>
      <Button variant="outline" size="sm">
        Toggle intercept details
      </Button>
    </CollapsibleTrigger>
    <CollapsibleContent className="mt-4 text-sm text-muted-foreground">
      Traffic for <span className="font-medium">auth-api.platform.svc</span> is routed
      to your workmachine over the tunnel.
    </CollapsibleContent>
  </Collapsible>
)

export const FullWidthRow = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border">
    <CollapsibleTrigger asChild>
      <Button variant="ghost" className="w-full justify-between p-4">
        <span className="font-medium">Deployment logs</span>
        <span className="text-xs text-muted-foreground">142 lines</span>
      </Button>
    </CollapsibleTrigger>
    <CollapsibleContent className="border-t p-4 text-sm text-muted-foreground">
      pulling image ghcr.io/kloudlite/api:v1.4.2 &middot; created container &middot;
      started container
    </CollapsibleContent>
  </Collapsible>
)

export const PlainTextTrigger = () => (
  <Collapsible defaultOpen className="w-96 rounded-lg border p-4">
    <CollapsibleTrigger className="text-sm font-medium underline">
      What happens during installation?
    </CollapsibleTrigger>
    <CollapsibleContent className="mt-2 text-sm text-muted-foreground">
      Kloudlite installs the operator, a wireguard gateway and the environment
      controller into the <span className="font-medium">kloudlite</span> namespace.
    </CollapsibleContent>
  </Collapsible>
)
