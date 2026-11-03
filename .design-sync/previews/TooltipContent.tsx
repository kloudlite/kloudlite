import * as React from 'react'
import { Button, Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@kloudlite/ui'

export const Basic = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Copy kubeconfig
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Scoped to production</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const LongCopy = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Reconcile now
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" className="w-64">
          Re-applies the environment spec to the cluster. Safe to run at any time.
        </TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const WithSideOffset = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Open in VS Code
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom" sideOffset={12}>
          Connects over the kl tunnel
        </TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const SideRight = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Node pool
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">3 / 5 nodes ready</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)
