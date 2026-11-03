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

export const Palette = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Type a command or search&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>
          Create environment
          <CommandShortcut>&#8984;N</CommandShortcut>
        </CommandItem>
        <CommandItem>Clone from production</CommandItem>
        <CommandItem>Switch workspace</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Development">
        <CommandItem>
          Open terminal
          <CommandShortcut>&#8984;J</CommandShortcut>
        </CommandItem>
        <CommandItem>Intercept service&hellip;</CommandItem>
        <CommandItem>Restart tunnel</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const Flat = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search clusters&hellip;" />
    <CommandList>
      <CommandEmpty>No clusters match.</CommandEmpty>
      <CommandGroup heading="Clusters">
        <CommandItem>kloudlite-prod &middot; eu-west-1</CommandItem>
        <CommandItem>kloudlite-dev &middot; ap-south-1</CommandItem>
        <CommandItem>edge-sandbox &middot; us-east-1</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const NoResults = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search commands&hellip;" value="postgres" />
    <CommandList>
      <CommandEmpty>No commands found for &ldquo;postgres&rdquo;.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Switch workspace</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
