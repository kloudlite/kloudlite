import * as React from 'react'
import {
  Badge,
  Button,
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@kloudlite/ui'

export const Open = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>sample-app / development</DrawerTitle>
        <DrawerDescription>
          Running on cluster kl-dev-blr &middot; ap-south-1
        </DrawerDescription>
      </DrawerHeader>
      <div className="px-4 space-y-3 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Status</span>
          <Badge>Ready</Badge>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Workspaces</span>
          <span className="font-medium">4 active</span>
        </div>
      </div>
      <DrawerFooter>
        <Button>Open workspace</Button>
        <DrawerClose asChild>
          <Button variant="outline">Close</Button>
        </DrawerClose>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const DirectionRight = () => (
  <Drawer open direction="right">
    <DrawerTrigger asChild>
      <Button variant="outline">Log filters</Button>
    </DrawerTrigger>
    <DrawerContent className="h-full w-96">
      <DrawerHeader>
        <DrawerTitle>Log filters</DrawerTitle>
        <DrawerDescription>Applied to payments-api pods</DrawerDescription>
      </DrawerHeader>
      <div className="px-4 space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Severity</span>
          <span className="font-medium">warn and above</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Window</span>
          <span className="font-medium">Last 30 minutes</span>
        </div>
      </div>
      <DrawerFooter>
        <Button>Apply filters</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const DirectionTop = () => (
  <Drawer open direction="top">
    <DrawerTrigger asChild>
      <Button variant="outline">Connect a cluster</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Connect a cluster</DrawerTitle>
        <DrawerDescription>
          Run the install command with cluster admin access.
        </DrawerDescription>
      </DrawerHeader>
      <div className="px-4">
        <div className="rounded-md bg-muted p-4 font-mono text-xs">
          kl infra attach --cluster kl-dev-blr --token kls_9f21c
        </div>
      </div>
      <DrawerFooter>
        <Button>Copy command</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
