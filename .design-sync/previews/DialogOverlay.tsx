import * as React from 'react'
import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
} from '@kloudlite/ui'

// DialogContent renders a DialogOverlay of its own; declaring one explicitly is
// how you restyle the scrim behind the dialog.
export const Default = () => (
  <Dialog open>
    <DialogPortal>
      <DialogOverlay />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete workspace kloudlite-dev</DialogTitle>
          <DialogDescription>
            The scrim behind this dialog dims the console and blocks interaction.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button variant="destructive">Delete workspace</Button>
        </DialogFooter>
      </DialogContent>
    </DialogPortal>
  </Dialog>
)

export const CustomScrim = () => (
  <Dialog open>
    <DialogPortal>
      <DialogOverlay className="bg-muted" />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect a cluster</DialogTitle>
          <DialogDescription>
            Point Kloudlite at an existing kubeconfig context to start managing it.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button>Done</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </DialogPortal>
  </Dialog>
)
