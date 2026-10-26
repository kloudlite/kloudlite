import * as React from 'react'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'

// SheetOverlay is rendered for you by SheetContent — these previews show it doing
// its job: dimming the page behind an open sheet.

const PageBehind = () => (
  <div className="grid grid-cols-2 gap-4">
    <Card>
      <CardHeader>
        <CardTitle>eu-west-1</CardTitle>
        <CardDescription>18 nodes &middot; healthy</CardDescription>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        6 environments scheduled.
      </CardContent>
    </Card>
    <Card>
      <CardHeader>
        <CardTitle>ap-south-1</CardTitle>
        <CardDescription>6 nodes &middot; healthy</CardDescription>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        2 environments scheduled.
      </CardContent>
    </Card>
  </div>
)

export const DimmingPageBehind = () => (
  <div className="space-y-4">
    <PageBehind />
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

export const NoOverlayWhenClosed = () => (
  <div className="space-y-4">
    <PageBehind />
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

export const BehindBottomSheet = () => (
  <div className="space-y-4">
    <PageBehind />
    <Sheet open>
      <SheetTrigger asChild>
        <Button variant="outline">Connect cluster</Button>
      </SheetTrigger>
      <SheetContent side="bottom">
        <SheetHeader>
          <SheetTitle>Connect a cluster</SheetTitle>
          <SheetDescription>Run the attach command to continue.</SheetDescription>
        </SheetHeader>
        <div className="rounded-md border bg-muted p-4 font-mono text-xs">
          kl cluster attach --name eu-west-1
        </div>
        <SheetFooter>
          <Button>Continue</Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  </div>
)
