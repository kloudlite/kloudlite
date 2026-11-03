import * as React from 'react'
import {
  Button,
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@kloudlite/ui'

const CubeIcon = () => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth={1.5}
    strokeLinecap="round"
    strokeLinejoin="round"
    className="size-6"
    aria-hidden="true"
  >
    <path d="M12 2.5 20.5 7v10L12 21.5 3.5 17V7z" />
    <path d="M3.5 7 12 11.5 20.5 7M12 11.5v10" />
  </svg>
)

export const IconVariant = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyMedia variant="icon">
        <CubeIcon />
      </EmptyMedia>
      <EmptyTitle>No environments yet</EmptyTitle>
      <EmptyDescription>
        The icon variant puts the glyph in a muted square above the title.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm">Create environment</Button>
    </EmptyContent>
  </Empty>
)

export const DefaultVariant = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyMedia>
        <div className="flex size-16 items-center justify-center rounded-full bg-muted text-lg font-medium text-muted-foreground">
          KL
        </div>
      </EmptyMedia>
      <EmptyTitle>No clusters connected</EmptyTitle>
      <EmptyDescription>
        The default variant is transparent, so any illustration or avatar can be dropped in.
      </EmptyDescription>
    </EmptyHeader>
  </Empty>
)
