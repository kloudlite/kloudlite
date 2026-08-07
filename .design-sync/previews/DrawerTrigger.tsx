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

export const ButtonTrigger = () => (
  <Drawer>
    <DrawerTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>sample-app / development</DrawerTitle>
        <DrawerDescription>Cluster kl-dev-blr &middot; ap-south-1</DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Open workspace</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const TriggerWithOpenDrawer = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button>Connect a cluster</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Connect a cluster</DrawerTitle>
        <DrawerDescription>
          Run the install command with cluster admin access.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Copy install command</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
