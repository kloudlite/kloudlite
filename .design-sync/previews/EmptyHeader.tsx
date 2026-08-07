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

const PlugIcon = () => (
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
    <path d="M9 3v5M15 3v5M6 8h12v4a6 6 0 0 1-12 0zM12 18v3" />
  </svg>
)

export const InEmptyState = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyMedia variant="icon">
        <PlugIcon />
      </EmptyMedia>
      <EmptyTitle>No clusters connected</EmptyTitle>
      <EmptyDescription>
        The header groups the media, title and description into one centred block at the top of
        the empty state.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm">Connect a cluster</Button>
    </EmptyContent>
  </Empty>
)

export const TitleOnlyHeader = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No node pools configured</EmptyTitle>
      <EmptyDescription>
        A header can hold just a title and description when no media is needed.
      </EmptyDescription>
    </EmptyHeader>
  </Empty>
)
