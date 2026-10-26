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

// SheetPortal is applied for you by SheetContent — these previews show its effect:
// the panel escapes clipping/stacking ancestors and pins to the viewport edge.

export const EscapesClippedParent = () => (
  <div className="h-64 w-96 overflow-hidden rounded-md border p-4">
    <p className="pb-4 text-sm text-muted-foreground">
      This card clips its overflow, yet the sheet still pins to the viewport.
    </p>
    <Sheet open>
      <SheetTrigger asChild>
        <Button variant="outline">Environment details</Button>
      </SheetTrigger>
      <SheetContent side="right">
        <SheetHeader>
          <SheetTitle>staging-team-a</SheetTitle>
          <SheetDescription>Environment on cluster eu-west-1.</SheetDescription>
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
  </div>
)

export const InsideScrollArea = () => (
  <div className="h-64 w-96 overflow-hidden rounded-md border">
    <div className="border-b px-4 py-2 text-sm font-medium">Clusters</div>
    <div className="space-y-2 p-4 text-sm">
      <div className="flex items-center justify-between">
        <span>eu-west-1</span>
        <Sheet open>
          <SheetTrigger asChild>
            <Button variant="link">Details</Button>
          </SheetTrigger>
          <SheetContent side="right">
            <SheetHeader>
              <SheetTitle>eu-west-1</SheetTitle>
              <SheetDescription>Managed cluster, 18 nodes.</SheetDescription>
            </SheetHeader>
            <div className="py-6 text-sm text-muted-foreground">
              Last reconcile completed without drift.
            </div>
            <SheetFooter>
              <Button>View nodes</Button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      </div>
      <div className="flex items-center justify-between text-muted-foreground">
        <span>ap-south-1</span>
        <span>6 nodes</span>
      </div>
    </div>
  </div>
)

export const ClosedRendersNothing = () => (
  <div className="h-64 w-96 overflow-hidden rounded-md border p-4">
    <p className="pb-4 text-sm text-muted-foreground">
      Closed: the portal mounts nothing into the body.
    </p>
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline">Environment details</Button>
      </SheetTrigger>
      <SheetContent side="right">
        <SheetHeader>
          <SheetTitle>staging-team-a</SheetTitle>
          <SheetDescription>Environment on cluster eu-west-1.</SheetDescription>
        </SheetHeader>
        <SheetFooter>
          <SheetClose asChild>
            <Button variant="outline">Close</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  </div>
)
