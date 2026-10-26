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

const Panel = () => (
  <SheetContent side="right">
    <SheetHeader>
      <SheetTitle>staging-team-a</SheetTitle>
      <SheetDescription>Environment on cluster eu-west-1.</SheetDescription>
    </SheetHeader>
    <div className="space-y-4 py-6 text-sm">
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground">Workspaces</span>
        <span className="font-medium">6 active</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground">Region</span>
        <span className="font-medium">Ireland</span>
      </div>
    </div>
    <SheetFooter>
      <SheetClose asChild>
        <Button variant="outline">Close</Button>
      </SheetClose>
      <Button>Open workspace</Button>
    </SheetFooter>
  </SheetContent>
)

const Row = ({ children }: { children: React.ReactNode }) => (
  <div className="w-96 rounded-md border p-4">
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium">staging-team-a</p>
        <p className="text-xs text-muted-foreground">eu-west-1 &middot; 6 workspaces</p>
      </div>
      {children}
    </div>
  </div>
)

export const Closed = () => (
  <Row>
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline">Environment details</Button>
      </SheetTrigger>
      <Panel />
    </Sheet>
  </Row>
)

export const ClosedAsLink = () => (
  <Row>
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="link">View 6 workspaces</Button>
      </SheetTrigger>
      <Panel />
    </Sheet>
  </Row>
)

export const Opened = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </SheetTrigger>
    <Panel />
  </Sheet>
)

export const Disabled = () => (
  <Row>
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" disabled>
          Environment details
        </Button>
      </SheetTrigger>
      <Panel />
    </Sheet>
  </Row>
)
