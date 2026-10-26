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

export const Basic = () => (
  <Drawer open>
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

export const LongTitle = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Installation</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>
          Finish connecting kloudlite-dev to your Kubernetes cluster
        </DrawerTitle>
        <DrawerDescription>Step 2 of 3 &middot; install the agent</DrawerDescription>
      </DrawerHeader>
      <div className="px-4 pb-4 text-sm text-muted-foreground">
        The agent registers the cluster and starts reporting node capacity.
      </div>
    </DrawerContent>
  </Drawer>
)
