import * as React from 'react'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@kloudlite/ui'

export const BetweenGroups = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Switch workspace</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Development">
        <CommandItem>Open terminal</CommandItem>
        <CommandItem>Intercept service&hellip;</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const ManySections = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Node pools">
        <CommandItem>worker-spot &middot; 6 nodes</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Account">
        <CommandItem>Billing &amp; usage</CommandItem>
        <CommandItem>Sign out</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
