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

export const Basic = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">payments-api</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>Open in workspace</DropdownMenuItem>
        <DropdownMenuItem>Intercept service</DropdownMenuItem>
        <DropdownMenuItem>Restart deployment</DropdownMenuItem>
        <DropdownMenuItem>View logs</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const Disabled = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Cluster</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>View nodes</DropdownMenuItem>
        <DropdownMenuItem disabled>Upgrade to v1.31 (in progress)</DropdownMenuItem>
        <DropdownMenuItem disabled>Delete cluster</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const Inset = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">View options</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel inset>Namespace</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem inset>kl-team-atlas</DropdownMenuItem>
        <DropdownMenuItem inset>kl-platform-core</DropdownMenuItem>
        <DropdownMenuItem inset>kube-system</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
