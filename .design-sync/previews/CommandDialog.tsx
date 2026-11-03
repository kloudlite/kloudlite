import * as React from 'react'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@kloudlite/ui'

export const OpenPalette = () => (
  <CommandDialog open>
    <CommandInput placeholder="Type a command or search&hellip;" />
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
        <CommandItem>
          Open terminal
          <CommandShortcut>&#8984;J</CommandShortcut>
        </CommandItem>
        <CommandItem>Intercept service&hellip;</CommandItem>
      </CommandGroup>
    </CommandList>
  </CommandDialog>
)

export const SearchScoped = () => (
  <CommandDialog open>
    <CommandInput placeholder="Search clusters and node pools&hellip;" />
    <CommandList>
      <CommandEmpty>Nothing found.</CommandEmpty>
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
        <CommandItem>kloudlite-dev &middot; ap-south-1</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Node pools">
        <CommandItem>worker-spot &middot; 6 nodes</CommandItem>
        <CommandItem>gpu-a10 &middot; 1 node</CommandItem>
      </CommandGroup>
    </CommandList>
  </CommandDialog>
)
