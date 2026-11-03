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
        <DialogTitle>Create environment</DialogTitle>
        <DialogDescription>
          Environments are isolated namespaces you can share with the team.
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

export const LongTitle = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          Delete workspace kloudlite-dev — this cannot be undone
        </DialogTitle>
        <DialogDescription>
          Type the workspace name to confirm before continuing.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button variant="destructive">Delete</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const WithoutDescription = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Connect a cluster</DialogTitle>
      </DialogHeader>
      <div className="rounded-md border bg-muted p-3 font-mono text-xs">
        kl infra attach --cluster prod-ap-south-1
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button>Done</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
