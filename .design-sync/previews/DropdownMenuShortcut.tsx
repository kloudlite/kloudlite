import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const WithShortcuts = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Workspace</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuItem>
          Open in workspace
          <DropdownMenuShortcut>&#8984;O</DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem>
          Copy kubeconfig
          <DropdownMenuShortcut>&#8984;K</DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem>
          View logs
          <DropdownMenuShortcut>&#8984;L</DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          Restart deployment
          <DropdownMenuShortcut>&#8679;&#8984;R</DropdownMenuShortcut>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const MixedItems = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Account</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuItem>Team settings</DropdownMenuItem>
        <DropdownMenuItem>Access tokens</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          Sign out
          <DropdownMenuShortcut>&#8679;&#8984;Q</DropdownMenuShortcut>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
