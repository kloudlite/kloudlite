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

export const TeammateProfile = () => (
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

export const ClusterSummary = () => (
  <div className="flex h-64 items-start justify-center pt-4">
    <HoverCard defaultOpen>
      <HoverCardTrigger asChild>
        <Button variant="link">prod-ap-south-1</Button>
      </HoverCardTrigger>
      <HoverCardContent>
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-semibold">prod-ap-south-1</h4>
            <Badge>Healthy</Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            6 nodes &middot; Mumbai &middot; Kubernetes v1.30
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
)
