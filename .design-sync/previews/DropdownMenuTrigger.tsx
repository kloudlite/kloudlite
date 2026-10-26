import * as React from 'react'
import {
  Avatar,
  AvatarFallback,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const ButtonTrigger = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Deployment actions</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>Restart deployment</DropdownMenuItem>
        <DropdownMenuItem>Scale replicas</DropdownMenuItem>
        <DropdownMenuItem>View logs</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const AvatarTrigger = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Avatar className="size-8">
          <AvatarFallback>KT</AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel>karthik@kloudlite.io</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Team settings</DropdownMenuItem>
        <DropdownMenuItem>Access tokens</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Sign out</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const PlainTrigger = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger className="text-sm font-medium">
        team-atlas
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>team-atlas</DropdownMenuItem>
        <DropdownMenuItem>platform-core</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Create team</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
