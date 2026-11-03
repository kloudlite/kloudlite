import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const BetweenGroups = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">payments-api</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>Open in workspace</DropdownMenuItem>
        <DropdownMenuItem>View logs</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Restart deployment</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const ManySections = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Environment</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>team-atlas-staging</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Open in workspace</DropdownMenuItem>
        <DropdownMenuItem>Copy kubeconfig</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Intercept service</DropdownMenuItem>
        <DropdownMenuItem>Restart all workloads</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Delete environment</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
