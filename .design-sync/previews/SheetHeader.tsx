import * as React from 'react'
import {
  Badge,
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

export const TitleAndDescription = () => (
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
        6 workspaces are scheduled on this environment.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Close</Button>
        </SheetClose>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const TitleOnly = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Filters</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Log filters</SheetTitle>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        Filtering the payments-api stream from the last 30 minutes.
      </div>
      <SheetFooter>
        <Button>Apply</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const WithStatusBadge = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Cluster status</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>eu-west-1</SheetTitle>
        <SheetDescription>Managed cluster, 18 nodes.</SheetDescription>
        <div className="flex gap-2 pt-2">
          <Badge>Healthy</Badge>
          <Badge variant="secondary">v1.31.2</Badge>
        </div>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        Last reconcile completed without drift.
      </div>
      <SheetFooter>
        <Button>View nodes</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
