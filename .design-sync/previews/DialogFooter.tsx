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
} from '@kloudlite/ui'

export const CancelAndConfirm = () => (
  <Dialog open>
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
  </Dialog>
)

export const DestructiveConfirm = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Delete cluster prod-ap-south-1</DialogTitle>
        <DialogDescription>
          All workloads are drained before the nodes are released.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button variant="destructive">Delete cluster</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const SingleAction = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Invite sent</DialogTitle>
        <DialogDescription>
          anita@kloudlite.io was invited to the Platform team as a Member.
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
