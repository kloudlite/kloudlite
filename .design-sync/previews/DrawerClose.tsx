import * as React from 'react'
import {
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

export const CloseInFooter = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Workspace config</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Workspace config</DrawerTitle>
        <DrawerDescription>karthik / sample-app</DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Save changes</Button>
        <DrawerClose asChild>
          <Button variant="outline">Cancel</Button>
        </DrawerClose>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const CloseInHeader = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Log filters</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <div className="flex items-center justify-between gap-3">
          <DrawerTitle>Log filters</DrawerTitle>
          <DrawerClose asChild>
            <Button variant="ghost" size="sm">
              Done
            </Button>
          </DrawerClose>
        </div>
        <DrawerDescription>payments-api &middot; production</DrawerDescription>
      </DrawerHeader>
      <div className="px-4 pb-4 text-sm text-muted-foreground">
        Showing warn and above for the last 30 minutes.
      </div>
    </DrawerContent>
  </Drawer>
)
