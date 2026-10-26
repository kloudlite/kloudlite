import * as React from 'react'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
  Input,
} from '@kloudlite/ui'

export const WithLabel = () => (
  <div className="w-96">
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="sep-name">Environment name</FieldLabel>
        <Input id="sep-name" defaultValue="staging-eu" />
      </Field>
      <FieldSeparator>Advanced</FieldSeparator>
      <Field>
        <FieldLabel htmlFor="sep-cidr">Pod CIDR</FieldLabel>
        <Input id="sep-cidr" defaultValue="10.42.0.0/16" />
        <FieldDescription>Leave as-is unless it clashes with your VPC.</FieldDescription>
      </Field>
    </FieldGroup>
  </div>
)

export const Plain = () => (
  <div className="w-96">
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="sep-registry">Registry</FieldLabel>
        <Input id="sep-registry" defaultValue="registry.kloudlite.io" />
      </Field>
      <FieldSeparator />
      <Field>
        <FieldLabel htmlFor="sep-tag">Image tag</FieldLabel>
        <Input id="sep-tag" defaultValue="api-gateway:1.14.2" />
      </Field>
    </FieldGroup>
  </div>
)
