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

export const NoEnvironments = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyMedia variant="icon">
        <CubeIcon />
      </EmptyMedia>
      <EmptyTitle>No environments yet</EmptyTitle>
      <EmptyDescription>
        Environments give every developer an isolated namespace on the cluster. Create one to
        start running workloads.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm">Create environment</Button>
    </EmptyContent>
  </Empty>
)

export const NoLogs = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No logs for the selected window</EmptyTitle>
      <EmptyDescription>
        Nothing was written by <span className="font-medium">api-gateway</span> between 14:00 and
        14:15. Widen the time range or pick another pod.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button size="sm" variant="outline">
        Reset time range
      </Button>
    </EmptyContent>
  </Empty>
)

export const NoWorkspaces = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyMedia variant="icon">
        <CubeIcon />
      </EmptyMedia>
      <EmptyTitle>No workspaces in this cluster</EmptyTitle>
      <EmptyDescription>
        Invite a teammate or connect your laptop with the kl CLI to spin up the first workspace.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <div className="flex gap-2">
        <Button size="sm">New workspace</Button>
        <Button size="sm" variant="outline">
          Invite teammate
        </Button>
      </div>
    </EmptyContent>
  </Empty>
)
