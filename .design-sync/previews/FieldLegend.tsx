import * as React from 'react'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
  Input,
} from '@kloudlite/ui'

export const LegendVariant = () => (
  <div className="w-96">
    <FieldSet>
      <FieldLegend>Cluster networking</FieldLegend>
      <FieldDescription>Applies to every node pool in kloudlite-dev.</FieldDescription>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="lg-cidr">Pod CIDR</FieldLabel>
          <Input id="lg-cidr" defaultValue="10.42.0.0/16" />
        </Field>
      </FieldGroup>
    </FieldSet>
  </div>
)

export const LabelVariant = () => (
  <div className="w-96">
    <FieldSet>
      <FieldLegend variant="label">Ingress</FieldLegend>
      <FieldDescription>
        The label variant is smaller, for nested groups inside a larger form.
      </FieldDescription>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="lg-domain">Wildcard domain</FieldLabel>
          <Input id="lg-domain" defaultValue="*.dev.kloudlite.io" />
        </Field>
      </FieldGroup>
    </FieldSet>
  </div>
)
