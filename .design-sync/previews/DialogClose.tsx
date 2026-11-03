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

export const AsSecondaryButton = () => (
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

export const AsOnlyAction = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Cluster connected</DialogTitle>
        <DialogDescription>
          prod-ap-south-1 is reporting healthy. The agent is running in namespace
          kloudlite.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button>Got it</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const AsGhostLink = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Rotate access token</DialogTitle>
        <DialogDescription>
          The current token keeps working for 24 hours after rotation.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Not now</Button>
        </DialogClose>
        <Button>Rotate token</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
