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
  Input,
  Label,
} from '@kloudlite/ui'

export const Basic = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Create environment</DialogTitle>
        <DialogDescription>
          Environments are isolated namespaces on the Production cluster.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Create</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const WithForm = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Invite a teammate</DialogTitle>
        <DialogDescription>
          Invites expire after 7 days if they are not accepted.
        </DialogDescription>
      </DialogHeader>
      <div className="space-y-2">
        <Label htmlFor="invite-email">Email address</Label>
        <Input id="invite-email" type="email" defaultValue="anita@kloudlite.io" />
      </div>
      <div className="space-y-2">
        <Label htmlFor="invite-role">Role</Label>
        <Input id="invite-role" defaultValue="Member" />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Send invite</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const NarrowWidth = () => (
  <Dialog open>
    <DialogContent className="w-80">
      <DialogHeader>
        <DialogTitle>Stop workmachine?</DialogTitle>
        <DialogDescription>
          Unsaved changes outside /workspace are lost.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Stop</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
