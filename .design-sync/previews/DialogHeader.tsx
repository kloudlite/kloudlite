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

export const TitleAndDescription = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Delete workspace kloudlite-dev</DialogTitle>
        <DialogDescription>
          Every deployment, secret and persistent volume in this workspace is
          removed. This cannot be undone.
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

export const TitleOnly = () => (
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
          <Button variant="outline">Close</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const LongDescription = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Uninstall Kloudlite from Production</DialogTitle>
        <DialogDescription>
          Uninstalling removes the Kloudlite agent from the cluster, tears down all
          12 managed namespaces, revokes tunnel access for every team member and
          deletes the stored kubeconfig. Billing stops at the end of the current
          cycle.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Keep installed</Button>
        </DialogClose>
        <Button variant="destructive">Uninstall</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
