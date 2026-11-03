import * as React from 'react'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Switch,
} from '@kloudlite/ui'

export const TextInput = () => (
  <div className="w-96">
    <Field>
      <FieldLabel htmlFor="env-name">Environment name</FieldLabel>
      <Input id="env-name" defaultValue="staging-eu" />
      <FieldDescription>
        Used as the Kubernetes namespace. Lowercase letters, digits and hyphens only.
      </FieldDescription>
    </Field>
  </div>
)

export const Invalid = () => (
  <div className="w-96">
    <Field data-invalid="true">
      <FieldLabel htmlFor="env-name-invalid">Environment name</FieldLabel>
      <Input id="env-name-invalid" defaultValue="Staging EU" aria-invalid="true" />
      <FieldError>Name must be lowercase and cannot contain spaces.</FieldError>
    </Field>
  </div>
)

export const WithSelect = () => (
  <div className="w-96">
    <Field>
      <FieldLabel htmlFor="region">Region</FieldLabel>
      <Select defaultValue="ap-south-1">
        <SelectTrigger id="region">
          <SelectValue placeholder="Select a region" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
          <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
        </SelectContent>
      </Select>
      <FieldDescription>Region cannot be changed after the cluster is created.</FieldDescription>
    </Field>
  </div>
)

export const Horizontal = () => (
  <div className="w-96">
    <Field orientation="horizontal">
      <FieldLabel htmlFor="autoscale">Node pool autoscaling</FieldLabel>
      <Switch id="autoscale" defaultChecked />
    </Field>
  </div>
)
