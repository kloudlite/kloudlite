import * as React from 'react'
import {
  Button,
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@kloudlite/ui'

export const InEmptyState = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No environments yet</EmptyTitle>
      <EmptyDescription>
        The title is the one-line statement of what is missing. Keep it short and concrete.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm">Create environment</Button>
    </EmptyContent>
  </Empty>
)

export const LongerTitle = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No deployments matched your filters</EmptyTitle>
      <EmptyDescription>
        Titles wrap and stay balanced and centred inside the header.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm" variant="outline">
        Clear filters
      </Button>
    </EmptyContent>
  </Empty>
)
