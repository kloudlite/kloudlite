import * as React from 'react'
import { Button, Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@kloudlite/ui'

export const ButtonTrigger = () => (
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

export const TextTrigger = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <span className="text-sm font-medium text-foreground">CrashLoopBackOff</span>
        </TooltipTrigger>
        <TooltipContent side="bottom">Container exited 7 times in 5 minutes</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)

export const DisabledTrigger = () => (
  <TooltipProvider>
    <div className="flex items-center justify-center px-4 py-8">
      <Tooltip defaultOpen>
        <TooltipTrigger asChild>
          <span className="text-sm text-muted-foreground">Delete cluster</span>
        </TooltipTrigger>
        <TooltipContent side="bottom">You need owner access</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
)
