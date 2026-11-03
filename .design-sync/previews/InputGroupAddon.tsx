import * as React from 'react'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
  Label,
} from '@kloudlite/ui'

export const InlineStart = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="iga-1">Registry image</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>ghcr.io/</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="iga-1" defaultValue="kloudlite/console:v0.9.1" />
    </InputGroup>
  </div>
)

export const InlineEnd = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="iga-2">Workspace subdomain</Label>
    <InputGroup>
      <InputGroupInput id="iga-2" defaultValue="dev-karthik" />
      <InputGroupAddon align="inline-end">
        <InputGroupText>.kloudlite.io</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const BlockStart = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupAddon align="block-start">
        <InputGroupText>Git branch</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput defaultValue="release/v1.4" />
    </InputGroup>
  </div>
)

export const BlockEnd = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupInput defaultValue="ghcr.io/kloudlite/api:v1.4.2" />
      <InputGroupAddon align="block-end">
        <InputGroupText>Last pushed 12 minutes ago</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)
