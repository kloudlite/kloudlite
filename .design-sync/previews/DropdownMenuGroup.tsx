import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const TwoGroups = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">payments-api</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuGroup>
          <DropdownMenuItem>Open in workspace</DropdownMenuItem>
          <DropdownMenuItem>Intercept service</DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem>Restart deployment</DropdownMenuItem>
          <DropdownMenuItem>Roll back</DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const LabelledGroups = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Environment</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-72">
        <DropdownMenuLabel>Develop</DropdownMenuLabel>
        <DropdownMenuGroup>
          <DropdownMenuItem>
            Open in workspace
            <DropdownMenuShortcut>&#8984;O</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem>Copy kubeconfig</DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuLabel>Manage</DropdownMenuLabel>
        <DropdownMenuGroup>
          <DropdownMenuItem>Restart all workloads</DropdownMenuItem>
          <DropdownMenuItem>Delete environment</DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
