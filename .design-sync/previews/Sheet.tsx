import * as React from 'react'
import {
  Badge,
  Button,
  Label,
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  Switch,
} from '@kloudlite/ui'

export const SideRight = () => (
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
      <div className="space-y-4 py-6 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Status</span>
          <Badge>Running</Badge>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Workspaces</span>
          <span className="font-medium">6 active</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Image digest</span>
          <span className="font-mono text-xs">sha256:9f2c41ab</span>
        </div>
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Close</Button>
        </SheetClose>
        <Button>Open workspace</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SideLeft = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Filters</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Log filters</SheetTitle>
        <SheetDescription>Narrow the stream for payments-api.</SheetDescription>
      </SheetHeader>
      <div className="space-y-4 py-6">
        <div className="flex items-center justify-between">
          <Label htmlFor="only-errors">Errors only</Label>
          <Switch id="only-errors" defaultChecked />
        </div>
        <div className="flex items-center justify-between">
          <Label htmlFor="include-sidecars">Include sidecars</Label>
          <Switch id="include-sidecars" />
        </div>
      </div>
      <SheetFooter>
        <Button>Apply filters</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SideTop = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Deployment banner</Button>
    </SheetTrigger>
    <SheetContent side="top">
      <SheetHeader>
        <SheetTitle>Rollout in progress</SheetTitle>
        <SheetDescription>
          payments-api is rolling out to 12 of 18 nodes in production.
        </SheetDescription>
      </SheetHeader>
      <SheetFooter>
        <Button variant="outline">View events</Button>
        <Button>Pause rollout</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SideBottom = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Connect cluster</Button>
    </SheetTrigger>
    <SheetContent side="bottom">
      <SheetHeader>
        <SheetTitle>Connect a cluster</SheetTitle>
        <SheetDescription>
          Run this from a shell with kubectl access to the target cluster.
        </SheetDescription>
      </SheetHeader>
      <div className="rounded-md border bg-muted p-4 font-mono text-xs">
        kl cluster attach --name eu-west-1 --token kl_9f2c41ab
      </div>
      <SheetFooter>
        <Button>I&rsquo;ve run it</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
