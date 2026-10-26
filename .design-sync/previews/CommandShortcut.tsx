import * as React from 'react'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from '@kloudlite/ui'

export const InItems = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search commands&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Shortcuts">
        <CommandItem>
          Create environment
          <CommandShortcut>&#8984;N</CommandShortcut>
        </CommandItem>
        <CommandItem>
          Open terminal
          <CommandShortcut>&#8984;J</CommandShortcut>
        </CommandItem>
        <CommandItem>
          Search everything
          <CommandShortcut>&#8984;K</CommandShortcut>
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const MultiKey = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search commands&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Tunnels">
        <CommandItem>
          Restart tunnel
          <CommandShortcut>&#8679;&#8984;R</CommandShortcut>
        </CommandItem>
        <CommandItem>
          Stop all intercepts
          <CommandShortcut>&#8679;&#8984;I</CommandShortcut>
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
