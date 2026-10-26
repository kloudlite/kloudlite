import * as React from 'react'
import {
  Badge,
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from '@kloudlite/ui'

export const Items = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search actions&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Actions">
        <CommandItem>Create environment</CommandItem>
        <CommandItem>Switch workspace</CommandItem>
        <CommandItem>Open terminal</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const WithShortcuts = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search actions&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Development">
        <CommandItem>
          Open terminal
          <CommandShortcut>&#8984;J</CommandShortcut>
        </CommandItem>
        <CommandItem>
          Intercept service&hellip;
          <CommandShortcut>&#8984;I</CommandShortcut>
        </CommandItem>
        <CommandItem>
          Restart tunnel
          <CommandShortcut>&#8984;R</CommandShortcut>
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const DisabledAndSelected = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search environments&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Environments">
        <CommandItem>development</CommandItem>
        <CommandItem>staging</CommandItem>
        <CommandItem disabled>production &middot; locked by admin</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const WithTrailingBadge = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search deployments&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Deployments">
        <CommandItem>
          auth-api
          <Badge className="ml-auto">v1.4.2</Badge>
        </CommandItem>
        <CommandItem>
          console-web
          <Badge variant="secondary" className="ml-auto">
            v3.0.1
          </Badge>
        </CommandItem>
        <CommandItem>
          gateway
          <Badge variant="outline" className="ml-auto">
            v0.9.7
          </Badge>
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
