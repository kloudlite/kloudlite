import * as React from 'react'
import {
  Avatar,
  AvatarFallback,
  Button,
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@kloudlite/ui'

export const ButtonTrigger = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <HoverCard open>
      <HoverCardTrigger asChild>
        <Button variant="link">@priya</Button>
      </HoverCardTrigger>
      <HoverCardContent>
        <div className="flex gap-4">
          <Avatar>
            <AvatarFallback>PS</AvatarFallback>
          </Avatar>
          <div className="space-y-1">
            <h4 className="text-sm font-semibold">Priya Sharma</h4>
            <p className="text-sm text-muted-foreground">
              Platform engineer at Kloudlite
            </p>
          </div>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
)

export const InlineTextTrigger = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <p className="text-sm text-muted-foreground">
      Deployed by{' '}
      <HoverCard open>
        <HoverCardTrigger className="font-medium text-foreground underline">
          arjun@kloudlite.io
        </HoverCardTrigger>
        <HoverCardContent>
          <div className="space-y-1">
            <h4 className="text-sm font-semibold">Arjun Nair</h4>
            <p className="text-sm text-muted-foreground">
              Admin &middot; last active 12 minutes ago
            </p>
          </div>
        </HoverCardContent>
      </HoverCard>{' '}
      2 hours ago.
    </p>
  </div>
)
