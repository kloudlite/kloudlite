import * as React from 'react'
import {
  Button,
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'

export const Default = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>staging-team-a</SheetTitle>
        <SheetDescription>
          Environment on cluster eu-west-1 &middot; synced 4 minutes ago.
        </SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        6 workspaces are scheduled here.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Close</Button>
        </SheetClose>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const MultiLine = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Connect cluster</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>Connect a cluster</SheetTitle>
        <SheetDescription>
          Kloudlite installs an agent into the kloudlite-system namespace. The agent
          needs outbound access on 443 and creates no inbound load balancers unless you
          enable ingress later.
        </SheetDescription>
      </SheetHeader>
      <div className="rounded-md border bg-muted p-4 font-mono text-xs">
        kl cluster attach --name eu-west-1
      </div>
      <SheetFooter>
        <Button>Continue</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const OnBottomSheet = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Rollout banner</Button>
    </SheetTrigger>
    <SheetContent side="bottom">
      <SheetHeader>
        <SheetTitle>Rollout in progress</SheetTitle>
        <SheetDescription>
          payments-api is updating 12 of 18 nodes in production.
        </SheetDescription>
      </SheetHeader>
      <SheetFooter>
        <Button variant="outline">View events</Button>
        <Button>Pause rollout</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
