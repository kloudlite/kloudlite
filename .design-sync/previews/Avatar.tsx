import * as React from 'react'
import { Avatar, AvatarFallback } from '@kloudlite/ui'

export const Basic = () => (
  <Avatar>
    <AvatarFallback>KT</AvatarFallback>
  </Avatar>
)

export const Sizes = () => (
  <div className="flex items-center gap-4">
    <Avatar className="size-6">
      <AvatarFallback className="text-xs">AN</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarFallback className="text-sm">AN</AvatarFallback>
    </Avatar>
    <Avatar className="size-16">
      <AvatarFallback className="text-lg">AN</AvatarFallback>
    </Avatar>
  </div>
)

export const MemberList = () => (
  <div className="flex w-80 flex-col gap-3">
    {[
      { initials: 'KT', name: 'Karthik Thirumalasetti', role: 'Owner' },
      { initials: 'PS', name: 'Priya Sharma', role: 'Admin' },
      { initials: 'MR', name: 'Marco Rossi', role: 'Developer' },
    ].map((m) => (
      <div key={m.initials} className="flex items-center gap-3">
        <Avatar>
          <AvatarFallback className="text-sm font-medium">{m.initials}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col">
          <span className="text-sm font-medium">{m.name}</span>
          <span className="text-xs text-muted-foreground">{m.role}</span>
        </div>
      </div>
    ))}
  </div>
)

export const Stack = () => (
  <div className="flex items-center gap-2">
    <div className="flex items-center">
      <Avatar className="size-8 border-2 border-background">
        <AvatarFallback className="text-xs">KT</AvatarFallback>
      </Avatar>
      <Avatar className="size-8 border-2 border-background">
        <AvatarFallback className="text-xs">PS</AvatarFallback>
      </Avatar>
      <Avatar className="size-8 border-2 border-background">
        <AvatarFallback className="text-xs">MR</AvatarFallback>
      </Avatar>
      <Avatar className="size-8 border-2 border-background">
        <AvatarFallback className="text-xs">+4</AvatarFallback>
      </Avatar>
    </div>
    <span className="text-sm text-muted-foreground">7 members in kloudlite-dev</span>
  </div>
)
