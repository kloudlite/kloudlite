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
      <EmptyTitle>No workspaces in this cluster</EmptyTitle>
      <EmptyDescription>
        Workspaces are created from the kl CLI on a developer machine. Once one is running it
        shows up here with its live sync status.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm">Install the kl CLI</Button>
    </EmptyContent>
  </Empty>
)

export const WithLink = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No image registries connected</EmptyTitle>
      <EmptyDescription>
        Connect a registry so builds can push images. Read the{' '}
        <a href="#registry-docs">registry setup guide</a> for the required credentials.
      </EmptyDescription>
    </EmptyHeader>
  </Empty>
)
