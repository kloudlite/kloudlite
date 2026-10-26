import * as React from 'react'
import {
  Badge,
  Button,
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@kloudlite/ui'

export const TitleAndDescription = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Environment details</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>sample-app / staging</DrawerTitle>
        <DrawerDescription>
          Synced from git ref refs/heads/staging &middot; 6 minutes ago
        </DrawerDescription>
      </DrawerHeader>
      <div className="px-4 pb-4 text-sm text-muted-foreground">
        Workloads reconcile automatically when the environment spec changes.
      </div>
    </DrawerContent>
  </Drawer>
)

export const TitleOnly = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Log filters</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Log filters</DrawerTitle>
      </DrawerHeader>
      <div className="px-4 pb-4 text-sm text-muted-foreground">
        Showing warn and above for the last 30 minutes.
      </div>
    </DrawerContent>
  </Drawer>
)

export const WithStatusBadge = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Cluster status</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <div className="flex items-center justify-between gap-3">
          <DrawerTitle>kl-dev-blr</DrawerTitle>
          <Badge>Ready</Badge>
        </div>
        <DrawerDescription>12 nodes &middot; Kubernetes v1.31.4</DrawerDescription>
      </DrawerHeader>
      <DrawerFooter>
        <Button variant="outline">View nodes</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)
