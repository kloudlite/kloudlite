import * as React from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@kloudlite/ui'

export const Initials = () => (
  <Avatar>
    <AvatarFallback>KT</AvatarFallback>
  </Avatar>
)

export const AfterFailedImage = () => (
  <div className="flex items-center gap-4">
    <Avatar>
      <AvatarImage src="/avatars/priya.png" alt="Priya Sharma" />
      <AvatarFallback>PS</AvatarFallback>
    </Avatar>
    <span className="text-sm text-muted-foreground">
      Fallback renders when the avatar image is missing
    </span>
  </div>
)

export const Tinted = () => (
  <div className="flex items-center gap-4">
    <Avatar>
      <AvatarFallback className="bg-primary text-primary-foreground text-sm font-medium">
        KT
      </AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarFallback className="bg-success text-success-foreground text-sm font-medium">
        OK
      </AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarFallback className="bg-destructive text-destructive-foreground text-sm font-medium">
        !
      </AvatarFallback>
    </Avatar>
  </div>
)

export const NonTextContent = () => (
  <div className="flex items-center gap-4">
    <Avatar className="size-12">
      <AvatarFallback className="text-xs font-medium">K8s</AvatarFallback>
    </Avatar>
    <Avatar className="size-12">
      <AvatarFallback className="text-xs font-medium">+9</AvatarFallback>
    </Avatar>
    <span className="text-sm text-muted-foreground">Cluster and overflow badges</span>
  </div>
)
