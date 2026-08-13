import * as React from 'react'
import { Button, Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@kloudlite/ui'

export const WrappingOneTooltip = () => (
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

export const WrappingAToolbar = () => (
  <TooltipProvider delayDuration={150}>
    <div className="flex items-center justify-center gap-3 px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Logs
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Stream container output</TooltipContent>
      </Tooltip>
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Shell
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Exec into the pod</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const CustomDelay = () => (
  <TooltipProvider delayDuration={0} skipDelayDuration={0}>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            Copy kubeconfig
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Opens instantly &middot; no delay</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)
