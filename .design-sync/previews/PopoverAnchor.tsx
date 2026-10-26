import * as React from 'react'
import {
  Popover,
  PopoverAnchor,
  PopoverContent,
  PopoverTrigger,
} from '@kloudlite/ui'

export const AnchoredToRow = () => (
  <div className="flex h-96 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverAnchor asChild>
        <div className="flex w-80 items-center justify-between border p-3">
          <span className="text-sm font-medium">nodepool / spot-ap-south-1</span>
          <PopoverTrigger className="text-sm text-muted-foreground">
            Details
          </PopoverTrigger>
        </div>
      </PopoverAnchor>
      <PopoverContent align="start">
        <p className="text-sm text-muted-foreground">
          Positioned against the row, not the trigger link.
        </p>
      </PopoverContent>
    </Popover>
  </div>
)

export const AnchoredToField = () => (
  <div className="flex h-96 items-start justify-center pt-4">
    <Popover defaultOpen>
      <PopoverAnchor asChild>
        <div className="w-80 space-y-2">
          <div className="border-b py-1 font-mono text-sm">
            kl workspace create --env staging
          </div>
          <PopoverTrigger className="text-xs text-muted-foreground">
            What does this do?
          </PopoverTrigger>
        </div>
      </PopoverAnchor>
      <PopoverContent side="bottom" align="start" sideOffset={8}>
        <p className="text-sm">
          Creates a workspace bound to the staging environment and attaches it to
          your local <span className="font-mono">kl</span> session.
        </p>
      </PopoverContent>
    </Popover>
  </div>
)
