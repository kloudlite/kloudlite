import * as React from 'react'
import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@kloudlite/ui'

const Body = () => (
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Create environment</DialogTitle>
      <DialogDescription>
        A namespace is provisioned on the Production cluster.
      </DialogDescription>
    </DialogHeader>
    <DialogFooter>
      <DialogClose asChild>
        <Button variant="outline">Cancel</Button>
      </DialogClose>
      <Button>Create environment</Button>
    </DialogFooter>
  </DialogContent>
)

export const AsButton = () => (
  <Dialog>
    <DialogTrigger asChild>
      <Button>New environment</Button>
    </DialogTrigger>
    <Body />
  </Dialog>
)

export const OutlineTrigger = () => (
  <Dialog>
    <DialogTrigger asChild>
      <Button variant="outline">Connect a cluster</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Connect a cluster</DialogTitle>
        <DialogDescription>
          Point Kloudlite at an existing kubeconfig context.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button>Done</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const TriggerWithOpenDialog = () => (
  <Dialog defaultOpen>
    <DialogTrigger asChild>
      <Button variant="destructive">Delete workspace</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Delete workspace kloudlite-dev?</DialogTitle>
        <DialogDescription>
          This cannot be undone. All deployments and volumes are removed.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button variant="destructive">Delete workspace</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
