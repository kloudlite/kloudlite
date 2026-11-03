import * as React from 'react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@kloudlite/ui'

export const OpenSubmenu = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Environment</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>Open in workspace</DropdownMenuItem>
        <DropdownMenuSub open>
          <DropdownMenuSubTrigger>Intercept service</DropdownMenuSubTrigger>
          <DropdownMenuPortal>
            <DropdownMenuSubContent className="w-56">
              <DropdownMenuItem>payments-api :8080</DropdownMenuItem>
              <DropdownMenuItem>auth-gateway :4000</DropdownMenuItem>
              <DropdownMenuItem>search-indexer :9200</DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuPortal>
        </DropdownMenuSub>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Copy kubeconfig</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)

export const ClosedSubmenu = () => (
  <div className="flex h-96 w-96 flex-col items-start p-4">
    <DropdownMenu open modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">Cluster</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuItem>View nodes</DropdownMenuItem>
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>Node pools</DropdownMenuSubTrigger>
          <DropdownMenuPortal>
            <DropdownMenuSubContent className="w-56">
              <DropdownMenuItem>default &middot; 3 nodes</DropdownMenuItem>
              <DropdownMenuItem>gpu-a10 &middot; 1 node</DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuPortal>
        </DropdownMenuSub>
        <DropdownMenuItem>Copy kubeconfig</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
)
