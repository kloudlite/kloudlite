import * as React from 'react'
import {
  Badge,
  Button,
  Input,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Separator,
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'

export const DetailPanel = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Inspect workspace</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>karthik-dev</SheetTitle>
        <SheetDescription>Workspace in staging-team-a.</SheetDescription>
      </SheetHeader>
      <div className="space-y-4 py-6 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Machine</span>
          <span className="font-medium">4 vCPU &middot; 8 GiB</span>
        </div>
        <Separator />
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Uptime</span>
          <span className="font-medium">3h 12m</span>
        </div>
        <Separator />
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">State</span>
          <Badge variant="secondary">Idle</Badge>
        </div>
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Dismiss</Button>
        </SheetClose>
        <Button>Resume</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const FormPanel = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button>New environment</Button>
    </SheetTrigger>
    <SheetContent side="right">
      <SheetHeader>
        <SheetTitle>New environment</SheetTitle>
        <SheetDescription>
          Environments are namespaced per cluster and can be cloned later.
        </SheetDescription>
      </SheetHeader>
      <div className="space-y-4 py-6">
        <div className="space-y-2">
          <Label htmlFor="env-name">Name</Label>
          <Input id="env-name" defaultValue="staging-team-b" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="env-cluster">Cluster</Label>
          <Select defaultValue="eu-west-1">
            <SelectTrigger id="env-cluster">
              <SelectValue placeholder="Select a cluster" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="eu-west-1">eu-west-1</SelectItem>
              <SelectItem value="ap-south-1">ap-south-1</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Cancel</Button>
        </SheetClose>
        <Button>Create</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SideLeftFilters = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Filters</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Filter clusters</SheetTitle>
        <SheetDescription>3 of 11 clusters match.</SheetDescription>
      </SheetHeader>
      <div className="space-y-2 py-6 text-sm">
        <div className="rounded-md border px-3 py-2">eu-west-1 &middot; Ireland</div>
        <div className="rounded-md border px-3 py-2">ap-south-1 &middot; Mumbai</div>
        <div className="rounded-md border px-3 py-2">us-east-1 &middot; N. Virginia</div>
      </div>
      <SheetFooter>
        <Button variant="outline">Reset</Button>
        <Button>Apply</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)

export const SideBottomCommands = () => (
  <Sheet open>
    <SheetTrigger asChild>
      <Button variant="outline">Setup steps</Button>
    </SheetTrigger>
    <SheetContent side="bottom">
      <SheetHeader>
        <SheetTitle>Install the kl CLI</SheetTitle>
        <SheetDescription>Two commands, then you&rsquo;re connected.</SheetDescription>
      </SheetHeader>
      <div className="space-y-2 py-4 font-mono text-xs">
        <div className="rounded-md border bg-muted px-3 py-2">
          curl -fsSL https://kloudlite.io/install | sh
        </div>
        <div className="rounded-md border bg-muted px-3 py-2">
          kl auth login --team acme
        </div>
      </div>
      <SheetFooter>
        <Button>Done</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
)
