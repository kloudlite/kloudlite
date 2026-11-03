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
  Input,
  Label,
} from '@kloudlite/ui'

export const Open = () => (
  <Dialog open>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Create environment</DialogTitle>
        <DialogDescription>
          A new namespace is provisioned on the Production cluster and cloned from
          the base template.
        </DialogDescription>
      </DialogHeader>
      <div className="space-y-2">
        <Label htmlFor="env-name">Environment name</Label>
        <Input id="env-name" defaultValue="karthik-dev" />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Create environment</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const ClosedWithTrigger = () => (
  <Dialog>
    <DialogTrigger asChild>
      <Button>Invite teammate</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Invite a teammate</DialogTitle>
        <DialogDescription>
          They get access to every environment in this team.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button>Send invite</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const DefaultOpen = () => (
  <Dialog defaultOpen>
    <DialogTrigger asChild>
      <Button variant="outline">Connect cluster</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Connect a cluster</DialogTitle>
        <DialogDescription>
          Run this command against the kubeconfig you want Kloudlite to manage.
        </DialogDescription>
      </DialogHeader>
      <div className="rounded-md border bg-muted p-3 font-mono text-xs">
        kl infra attach --cluster prod-ap-south-1
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Close</Button>
        </DialogClose>
        <Button>I have run it</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)

export const NonModal = () => (
  <Dialog open modal={false}>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Rotate access token</DialogTitle>
        <DialogDescription>
          The current token keeps working for 24 hours after rotation.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button>Rotate token</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
)
