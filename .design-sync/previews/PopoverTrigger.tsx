import * as React from 'react'
import { Button, Popover, PopoverContent, PopoverTrigger } from '@kloudlite/ui'

export const ButtonTrigger = () => (
  <div className="flex h-80 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">Copy kubeconfig</Button>
      </PopoverTrigger>
      <PopoverContent align="start">
        <p className="text-sm text-muted-foreground">
          The trigger above owns the open state of this panel.
        </p>
      </PopoverContent>
    </Popover>
  </div>
)

export const TextTrigger = () => (
  <div className="flex h-72 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverTrigger className="text-sm font-medium underline">
        Why is my workspace pending?
      </PopoverTrigger>
      <PopoverContent align="start">
        <p className="text-sm text-muted-foreground">
          Waiting for a spot node to join the cluster. This usually clears within a
          minute.
        </p>
      </PopoverContent>
    </Popover>
  </div>
)
