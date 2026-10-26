import * as React from 'react'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@kloudlite/ui'

export const WithPlaceholder = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Type a command or search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Suggestions">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Open terminal</CommandItem>
        <CommandItem>Intercept service&hellip;</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const WithValue = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search&hellip;" value="env" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Delete environment</CommandItem>
        <CommandItem>Environment settings</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const Disabled = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search is unavailable while syncing" disabled />
    <CommandList>
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
        <CommandItem>kloudlite-dev &middot; ap-south-1</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
