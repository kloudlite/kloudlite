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

// DrawerContent renders DrawerOverlay internally; these previews show the overlay
// dimming the page behind an open drawer.
export const OverlayBehindDrawer = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>sample-app / development</DrawerTitle>
        <DrawerDescription>
          The page behind is dimmed by the drawer overlay.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Open workspace</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const OverlayWithConfigDrawer = () => (
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
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
