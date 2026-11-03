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

export const Basic = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search actions&hellip;" />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Workspace">
        <CommandItem>Switch workspace</CommandItem>
        <CommandItem>Invite teammate</CommandItem>
        <CommandItem>Manage installations</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const Scrolling = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search deployments&hellip;" />
    <CommandList>
      <CommandEmpty>No deployments found.</CommandEmpty>
      <CommandGroup heading="development">
        <CommandItem>auth-api &middot; v1.4.2</CommandItem>
        <CommandItem>console-web &middot; v3.0.1</CommandItem>
        <CommandItem>gateway &middot; v0.9.7</CommandItem>
        <CommandItem>message-office &middot; v1.1.0</CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="staging">
        <CommandItem>auth-api &middot; v1.4.1</CommandItem>
        <CommandItem>console-web &middot; v3.0.0</CommandItem>
        <CommandItem>gateway &middot; v0.9.7</CommandItem>
        <CommandItem>message-office &middot; v1.0.9</CommandItem>
        <CommandItem>webhooks &middot; v0.4.3</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)

export const SingleGroup = () => (
  <Command className="w-96 rounded-lg border shadow-md">
    <CommandInput placeholder="Search regions&hellip;" />
    <CommandList>
      <CommandEmpty>No regions found.</CommandEmpty>
      <CommandGroup heading="Regions">
        <CommandItem>ap-south-1 &middot; Mumbai</CommandItem>
        <CommandItem>eu-west-1 &middot; Ireland</CommandItem>
        <CommandItem>us-east-1 &middot; N. Virginia</CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
)
