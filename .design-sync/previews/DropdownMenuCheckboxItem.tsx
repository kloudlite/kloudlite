import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const Toggles = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Workload filters</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuLabel>Show</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuCheckboxItem checked>Running workloads</DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked>Failed workloads</DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked={false}>
          System namespaces
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const AllChecked = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Log streams</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuCheckboxItem checked>payments-api</DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked>auth-gateway</DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked>ingress-nginx</DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const WithDisabled = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Sync options</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-64">
        <DropdownMenuCheckboxItem checked>Auto-sync secrets</DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked={false}>
          Mount host SSH agent
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem checked disabled>
          Managed by team policy
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
