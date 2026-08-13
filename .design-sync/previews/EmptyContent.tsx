import * as React from 'react'
import {
  Button,
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
  Input,
} from '@kloudlite/ui'

export const WithActions = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No environments yet</EmptyTitle>
      <EmptyDescription>
        The content slot holds whatever the user should do next.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <div className="flex gap-2">
        <Button size="sm">Create environment</Button>
        <Button size="sm" variant="outline">
          Import from spec
        </Button>
      </div>
    </EmptyContent>
  </Empty>
)

export const WithInlineForm = () => (
  <Empty className="w-96 border">
    <EmptyHeader>
      <EmptyTitle>No team members yet</EmptyTitle>
      <EmptyDescription>
        Invite someone to the kloudlite-dev organisation to get started.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Input placeholder="teammate@kloudlite.io" />
      <Button size="sm" className="w-full">
        Send invite
      </Button>
    </EmptyContent>
  </Empty>
)
