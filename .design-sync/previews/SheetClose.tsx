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

export const AsFooterButton = () => (
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
        <Button>Open workspace</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const CancelInForm = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button>New environment</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>New environment</SheetTitle>
        <SheetDescription>Scheduled on eu-west-1.</SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        Config is cloned from staging-team-a.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="ghost">Cancel</Button>
        </SheetClose>
        <Button>Create</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const InlineTextLink = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Setup steps</Button>
    </SheetTrigger>
    <SheetContent side="bottom">
      <SheetHeader>
        <SheetTitle>Install the kl CLI</SheetTitle>
        <SheetDescription>Two commands and you&rsquo;re connected.</SheetDescription>
      </SheetHeader>
      <div className="rounded-md border bg-muted p-4 font-mono text-xs">
        curl -fsSL https://kloudlite.io/install | sh
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="link">Skip for now</Button>
        </SheetClose>
        <Button>Done</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
