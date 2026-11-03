import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const Basic = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Sort environments</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuRadioGroup value="updated">
          <DropdownMenuRadioItem value="updated">Last updated</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="name">Name</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="cost">Monthly cost</DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const TwoGroups = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Log viewer</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>Level</DropdownMenuLabel>
        <DropdownMenuRadioGroup value="warn">
          <DropdownMenuRadioItem value="info">Info</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="warn">Warn</DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
        <DropdownMenuSeparator />
        <DropdownMenuLabel>Window</DropdownMenuLabel>
        <DropdownMenuRadioGroup value="1h">
          <DropdownMenuRadioItem value="15m">Last 15 minutes</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="1h">Last hour</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="24h">Last 24 hours</DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
