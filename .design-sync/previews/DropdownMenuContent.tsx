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

export const AlignStart = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Workspace</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>Open in workspace</DropdownMenuItem>
        <DropdownMenuItem>Copy kubeconfig</DropdownMenuItem>
        <DropdownMenuItem>View logs</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const WideWithSections = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">payments-api</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-72">
        <DropdownMenuLabel>Running in staging &middot; 3 replicas</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Intercept service on :8080</DropdownMenuItem>
        <DropdownMenuItem>Restart deployment</DropdownMenuItem>
        <DropdownMenuItem>Stream logs from all pods</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Roll back to previous image</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const SideRight = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Node pool</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent side="right" align="start" className="w-56">
        <DropdownMenuItem>Cordon nodes</DropdownMenuItem>
        <DropdownMenuItem>Drain nodes</DropdownMenuItem>
        <DropdownMenuItem>Resize pool</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
