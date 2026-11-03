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

export const CancelAndConfirm = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button>New environment</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>New environment</SheetTitle>
        <SheetDescription>Clones config from staging-team-a.</SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        The environment will be scheduled on eu-west-1.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Cancel</Button>
        </SheetClose>
        <Button>Create environment</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SingleAction = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Connect cluster</Button>
    </SheetTrigger>
    <SheetContent side="bottom">
      <SheetHeader>
        <SheetTitle>Connect a cluster</SheetTitle>
        <SheetDescription>Run the attach command, then continue.</SheetDescription>
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

export const DestructiveAction = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Manage workspace</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>karthik-dev</SheetTitle>
        <SheetDescription>
          Deleting a workspace removes its volumes permanently.
        </SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        Idle for 3 days &middot; 40 GiB volume attached.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Keep it</Button>
        </SheetClose>
        <Button variant="destructive">Delete workspace</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
