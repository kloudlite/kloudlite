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
      <Button variant="outline">Log filters</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Log filters</DrawerTitle>
        <DrawerDescription>
          Applied to every pod in payments-api &middot; production.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Apply filters</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const Multiline = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Stop intercept</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Stop traffic intercept?</DrawerTitle>
        <DrawerDescription>
          Requests for payments-api currently route to your laptop over the Kloudlite
          tunnel. Stopping the intercept sends traffic back to the in-cluster
          deployment, and any request in flight is retried.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button variant="destructive">Stop intercept</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
