import * as React from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@kloudlite/ui'

// Inline SVG data URIs so the image always resolves in a sandboxed render.
const avatarSrc = (bg: string, fg: string, initials: string) =>
  'data:image/svg+xml;utf8,' +
  encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96"><rect width="96" height="96" fill="${bg}"/><text x="48" y="60" font-family="sans-serif" font-size="36" font-weight="600" text-anchor="middle" fill="${fg}">${initials}</text></svg>`
  )

export const Basic = () => (
  <Avatar>
    <AvatarImage src={avatarSrc('rgb(38,38,38)', 'rgb(250,250,250)', 'KT')} alt="Karthik Thirumalasetti" />
    <AvatarFallback>KT</AvatarFallback>
  </Avatar>
)

export const Large = () => (
  <Avatar className="size-24">
    <AvatarImage
      src={avatarSrc('rgb(30,64,175)', 'rgb(255,255,255)', 'PS')}
      alt="Priya Sharma"
      className="object-cover"
    />
    <AvatarFallback>PS</AvatarFallback>
  </Avatar>
)

export const BrokenSourceFallsBack = () => (
  <div className="flex items-center gap-4">
    <Avatar>
      <AvatarImage src="/avatars/does-not-exist.png" alt="Marco Rossi" />
      <AvatarFallback>MR</AvatarFallback>
    </Avatar>
    <span className="text-sm text-muted-foreground">
      Image failed to load &mdash; fallback initials shown
    </span>
  </div>
)

export const InRow = () => (
  <div className="flex w-80 items-center gap-3">
    <Avatar>
      <AvatarImage src={avatarSrc('rgb(22,101,52)', 'rgb(255,255,255)', 'DV')} alt="Deploy bot" />
      <AvatarFallback>DV</AvatarFallback>
    </Avatar>
    <div className="flex flex-col">
      <span className="text-sm font-medium">deploy-bot</span>
      <span className="text-xs text-muted-foreground">
        Rolled out payments-api to production
      </span>
    </div>
  </div>
)
