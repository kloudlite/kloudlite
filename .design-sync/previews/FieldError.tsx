import * as React from 'react'
import { Field, FieldDescription, FieldError, FieldLabel, Input } from '@kloudlite/ui'

export const SingleError = () => (
  <div className="w-96">
    <Field data-invalid="true">
      <FieldLabel htmlFor="cluster-name">Cluster name</FieldLabel>
      <Input id="cluster-name" defaultValue="prod cluster" aria-invalid="true" />
      <FieldError>Cluster name cannot contain spaces.</FieldError>
    </Field>
  </div>
)

export const MultipleErrors = () => (
  <div className="w-96">
    <Field data-invalid="true">
      <FieldLabel htmlFor="registry-url">Registry URL</FieldLabel>
      <Input id="registry-url" defaultValue="http://registry" aria-invalid="true" />
      <FieldError
        errors={[
          { message: 'Registry URL must use https.' },
          { message: 'Host must be a fully qualified domain name.' },
        ]}
      />
    </Field>
  </div>
)

export const AlongsideDescription = () => (
  <div className="w-96">
    <Field data-invalid="true">
      <FieldLabel htmlFor="max-nodes-err">Maximum nodes</FieldLabel>
      <Input id="max-nodes-err" type="number" defaultValue={0} aria-invalid="true" />
      <FieldDescription>The autoscaler upper bound for this node pool.</FieldDescription>
      <FieldError errors={[{ message: 'Must be at least 1 node.' }]} />
    </Field>
  </div>
)
