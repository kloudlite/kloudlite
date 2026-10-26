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

export const Basic = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Invite a teammate</DialogTitle>
        <DialogDescription>
          They get read and write access to every environment in the Platform team.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Send invite</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const LongCopy = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Delete workspace kloudlite-dev</DialogTitle>
        <DialogDescription>
          Deleting this workspace removes the namespace from prod-ap-south-1,
          terminates all 6 running deployments, detaches the persistent volumes and
          revokes every tunnel currently intercepting its services. Snapshots older
          than 30 days are not recoverable.
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
