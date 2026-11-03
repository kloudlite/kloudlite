import * as React from 'react'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@kloudlite/ui'

export const NoMatches = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search commands&hellip;" value="kafka" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Switch workspace</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const DescriptiveEmpty = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search clusters&hellip;" value="edge-eu" />
    <CommandList>
      <CommandEmpty>
        <p className="font-medium">No clusters match &ldquo;edge-eu&rdquo;</p>
        <p className="text-muted-foreground">
          Try a region name such as eu-west-1.
        </p>
      </CommandEmpty>
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
        <CommandItem>kloudlite-dev &middot; ap-south-1</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
