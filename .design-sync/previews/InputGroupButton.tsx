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
  <div className="w-96 space-y-2">
    <Label htmlFor="igb-1">Kubeconfig path</Label>
    <InputGroup>
      <InputGroupInput id="igb-1" defaultValue="~/.kube/kl-prod.yaml" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton>Copy</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const Sizes = () => (
  <div className="w-96 space-y-4">
    <InputGroup>
      <InputGroupInput defaultValue="dev-karthik" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="xs">Rename</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
    <InputGroup>
      <InputGroupInput defaultValue="ghcr.io/kloudlite/api:v1.4.2" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="sm">Pull</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const Variants = () => (
  <div className="w-96 space-y-4">
    <InputGroup>
      <InputGroupInput defaultValue="release/v1.4" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="sm" variant="outline">Checkout</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
    <InputGroup>
      <InputGroupInput defaultValue="kl-prod-ap-south-1" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="sm" variant="default">Deploy</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const Disabled = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igb-2">Cluster endpoint</Label>
    <InputGroup>
      <InputGroupInput id="igb-2" disabled defaultValue="https://api.kl-prod.kloudlite.io" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="sm" disabled>Copy</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)
