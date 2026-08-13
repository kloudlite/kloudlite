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
        <SheetDescription>Environment on cluster eu-west-1.</SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        6 workspaces, 2 running right now.
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Close</Button>
        </SheetClose>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const WithoutDescription = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Workspace settings</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>Workspace settings</SheetTitle>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        Machine size, idle timeout and dotfiles for karthik-dev.
      </div>
      <SheetFooter>
        <Button>Save</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const LongTitle = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Rollout details</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>
          payments-api rollout to production, eu-west-1
        </SheetTitle>
        <SheetDescription>Started 6 minutes ago by ops-bot.</SheetDescription>
      </SheetHeader>
      <div className="py-6 text-sm text-muted-foreground">
        12 of 18 nodes updated to sha256:9f2c41ab.
      </div>
      <SheetFooter>
        <Button variant="outline">Pause</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
