import * as React from 'react'
import {
  Button,
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@kloudlite/ui'

// DrawerContent wraps its children in DrawerPortal; these previews show the
// portalled content anchored to the viewport edge.
export const PortalledContent = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Recent deployments</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Recent deployments</DrawerTitle>
        <DrawerDescription>payments-api &middot; production</DrawerDescription>
      </DrawerHeader>
      <div className="px-4 pb-4 space-y-2 text-sm">
        <div className="flex items-center justify-between border-b py-2">
          <span className="font-mono text-xs">2026.8.3</span>
          <span className="text-muted-foreground">12 minutes ago</span>
        </div>
        <div className="flex items-center justify-between py-2">
          <span className="font-mono text-xs">2026.8.2</span>
          <span className="text-muted-foreground">2 days ago</span>
        </div>
      </div>
    </DrawerContent>
  </Drawer>
)

export const PortalledSidePanel = () => (
  <Drawer open direction="right">
    <DrawerTrigger asChild>
      <Button variant="outline">Log filters</Button>
    </DrawerTrigger>
    <DrawerContent className="h-full w-96">
      <DrawerHeader>
        <DrawerTitle>Log filters</DrawerTitle>
        <DrawerDescription>Applied to payments-api pods</DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Apply filters</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
