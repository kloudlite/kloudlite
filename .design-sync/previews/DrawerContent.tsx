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

export const Bottom = () => (
  <Drawer open>
    <DrawerTrigger asChild>
      <Button variant="outline">Connect a cluster</Button>
    </DrawerTrigger>
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle>Connect a cluster</DrawerTitle>
        <DrawerDescription>
          Run the install command on any Kubernetes cluster you control.
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

export const SidePanel = () => (
  <Drawer open direction="right">
    <DrawerTrigger asChild>
      <Button variant="outline">Workspace config</Button>
    </DrawerTrigger>
    <DrawerContent className="h-full w-96">
      <DrawerHeader>
        <DrawerTitle>Workspace config</DrawerTitle>
        <DrawerDescription>karthik / sample-app</DrawerDescription>
      </DrawerHeader>
      <div className="px-4 space-y-3 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Machine</span>
          <span className="font-medium">4 vCPU &middot; 8 GiB</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Image</span>
          <span className="font-medium">ghcr.io/kloudlite/base:2026.8</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Idle timeout</span>
          <span className="font-medium">30 minutes</span>
        </div>
      </div>
      <DrawerFooter>
        <Button>Save changes</Button>
      </DrawerFooter>
    </DrawerContent>
  </Drawer>
)

export const ScrollableBody = () => (
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
        {[
          ['2026.8.3', 'Succeeded', '12 minutes ago'],
          ['2026.8.2', 'Succeeded', '2 days ago'],
          ['2026.8.1', 'Rolled back', '3 days ago'],
          ['2026.7.28', 'Succeeded', '9 days ago'],
        ].map(([tag, state, when]) => (
          <div key={tag} className="flex items-center justify-between border-b py-2">
            <span className="font-mono text-xs">{tag}</span>
            <span className="font-medium">{state}</span>
            <span className="text-muted-foreground">{when}</span>
          </div>
        ))}
      </div>
    </DrawerContent>
  </Drawer>
)
