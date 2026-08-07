import * as React from 'react'
import { Button, Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@kloudlite/ui'

export const Basic = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Restart deployment
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Rolls pods one at a time</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const OnIconAction = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Copy kubeconfig
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          Scoped to the production environment
        </TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const Sides = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center gap-3 px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Scale up
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Adds one replica</TooltipContent>
      </Tooltip>
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Drain node
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Evicts all pods</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const Controlled = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip open>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Delete workspace
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">This cannot be undone</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)
