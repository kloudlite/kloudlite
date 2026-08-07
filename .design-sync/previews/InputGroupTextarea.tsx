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

export const Default = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupTextarea rows={4} defaultValue={"Rolled api to v1.4.2 on kl-prod-ap-south-1.\nWatch p99 latency for 30 minutes."} />
    </InputGroup>
  </div>
)

export const WithBlockStartAddon = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupAddon align="block-start">
        <InputGroupText>Deployment notes</InputGroupText>
      </InputGroupAddon>
      <InputGroupTextarea rows={4} defaultValue={"Scaled console to 3 replicas.\nMigrated env dev-karthik to nodepool spot-4x."} />
    </InputGroup>
  </div>
)

export const WithFooterAction = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupTextarea rows={4} placeholder="Describe what changed in this environment..." />
      <InputGroupAddon align="block-end">
        <InputGroupText>Markdown supported</InputGroupText>
        <InputGroupButton size="sm" variant="outline">Save note</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const Invalid = () => (
  <div className="w-96 space-y-2">
    <InputGroup>
      <InputGroupAddon align="block-start">
        <InputGroupText>Pod annotations (YAML)</InputGroupText>
      </InputGroupAddon>
      <InputGroupTextarea rows={3} aria-invalid defaultValue={"kloudlite.io/env dev-karthik\n  bad: indent"} />
    </InputGroup>
    <p className="text-xs text-destructive">Invalid YAML at line 1.</p>
  </div>
)
