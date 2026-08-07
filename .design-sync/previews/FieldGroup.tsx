import * as React from 'react'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const StackedFields = () => (
  <div className="w-96">
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="grp-env-name">Environment name</FieldLabel>
        <Input id="grp-env-name" defaultValue="staging-eu" />
        <FieldDescription>Used as the Kubernetes namespace.</FieldDescription>
      </Field>
      <Field>
        <FieldLabel htmlFor="grp-region">Region</FieldLabel>
        <Select defaultValue="eu-west-1">
          <SelectTrigger id="grp-region">
            <SelectValue placeholder="Select a region" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
            <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
          </SelectContent>
        </Select>
      </Field>
    </FieldGroup>
  </div>
)

export const WithInvalidField = () => (
  <div className="w-96">
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="grp-cluster">Cluster</FieldLabel>
        <Input id="grp-cluster" defaultValue="kloudlite-dev" />
      </Field>
      <Field data-invalid="true">
        <FieldLabel htmlFor="grp-nodes">Maximum nodes</FieldLabel>
        <Input id="grp-nodes" type="number" defaultValue={0} aria-invalid="true" />
        <FieldError errors={[{ message: 'Must be at least 1 node.' }]} />
      </Field>
    </FieldGroup>
  </div>
)
