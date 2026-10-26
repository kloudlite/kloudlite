import * as React from 'react'
import {
  Avatar,
  AvatarFallback,
  Badge,
  Button,
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@kloudlite/ui'

export const Default = () => (
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
              Platform engineer &middot; joined March 2024
            </p>
          </div>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
)

export const AlignStart = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <HoverCard open>
      <HoverCardTrigger asChild>
        <Button variant="outline">api-gateway</Button>
      </HoverCardTrigger>
      <HoverCardContent align="start">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-semibold">api-gateway</h4>
            <Badge variant="secondary">3 replicas</Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            Exposed on port 8080 in the staging-eu environment.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
)

export const WideContent = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <HoverCard open>
      <HoverCardTrigger asChild>
        <Button variant="link">staging-eu</Button>
      </HoverCardTrigger>
      <HoverCardContent className="w-80">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">staging-eu</h4>
          <p className="text-sm text-muted-foreground">
            Shared preview environment for the billing squad. Auto-sleeps after
            two hours of inactivity.
          </p>
          <p className="text-xs text-muted-foreground">
            Cluster prod-ap-south-1 &middot; 4 services running
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
)
