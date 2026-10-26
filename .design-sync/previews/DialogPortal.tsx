import * as React from 'react'
import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogPortal,
  DialogTitle,
} from '@kloudlite/ui'

// DialogContent already renders inside a DialogPortal; declaring the portal
// explicitly is how you opt a dialog into a custom container.
export const ExplicitPortal = () => (
  <Dialog open>
    <DialogPortal>
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
    </DialogPortal>
  </Dialog>
)

export const ForceMounted = () => (
  <Dialog open>
    <DialogPortal forceMount>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite a teammate</DialogTitle>
          <DialogDescription>
            Invites expire after 7 days if they are not accepted.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button>Send invite</Button>
        </DialogFooter>
      </DialogContent>
    </DialogPortal>
  </Dialog>
)
