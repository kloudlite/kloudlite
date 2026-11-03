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

export const AccountHeader = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Account</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>karthik@kloudlite.io</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Team settings</DropdownMenuItem>
        <DropdownMenuItem>Access tokens</DropdownMenuItem>
        <DropdownMenuItem>Sign out</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const SectionLabels = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Switch context</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>Clusters</DropdownMenuLabel>
        <DropdownMenuItem>prod-eu-west-1</DropdownMenuItem>
        <DropdownMenuItem>staging-ap-south-1</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuLabel>Environments</DropdownMenuLabel>
        <DropdownMenuItem>karthik-dev</DropdownMenuItem>
        <DropdownMenuItem>team-atlas-staging</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const InsetLabel = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Namespace</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel inset>Visible namespaces</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem inset>kl-team-atlas</DropdownMenuItem>
        <DropdownMenuItem inset>kl-platform-core</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
