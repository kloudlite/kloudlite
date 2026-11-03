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

export const StackedActions = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Delete environment</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Delete sample-app / staging?</DrawerTitle>
        <DrawerDescription>
          All workspaces scheduled on this environment will stop immediately.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button variant="destructive">Delete environment</Button>
        <DrawerClose asChild>
          <Button variant="outline">Cancel</Button>
        </DrawerClose>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const SingleAction = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Connect a cluster</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Connect a cluster</DrawerTitle>
        <DrawerDescription>
          Paste the install command into a shell with cluster admin access.
        </DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button>Copy install command</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const RowActions = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Workspace config</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Workspace config</DrawerTitle>
        <DrawerDescription>karthik / sample-app</DrawerDescription>
      </DrawerHeader>
      <DrawerFooter className="flex-row justify-end gap-2">
        <DrawerClose asChild>
          <Button variant="outline">Discard</Button>
        </DrawerClose>
        <Button>Save changes</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
