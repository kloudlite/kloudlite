import * as React from 'react'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@kloudlite/ui'

export const SingleGroup = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Clone from production</CommandItem>
        <CommandItem>Delete environment</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const MultipleGroups = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>
          Create environment
          <CommandShortcut>&#8984;N</CommandShortcut>
        </CommandItem>
        <CommandItem>Switch workspace</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Development">
        <CommandItem>Open terminal</CommandItem>
        <CommandItem>Intercept service&hellip;</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
        <CommandItem>kloudlite-dev &middot; ap-south-1</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const NoHeading = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search node pools&hellip;" />
    <CommandList>
      <CommandEmpty>No node pools found.</CommandEmpty>
      <CommandGroup>
        <CommandItem>worker-spot &middot; 6 nodes</CommandItem>
        <CommandItem>worker-ondemand &middot; 3 nodes</CommandItem>
        <CommandItem>gpu-a10 &middot; 1 node</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
