import * as React from 'react'
import {
  Checkbox,
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
  FieldTitle,
  Input,
} from '@kloudlite/ui'

export const NodePoolSettings = () => (
  <div className="w-96">
    <FieldSet>
      <FieldLegend>Node pool</FieldLegend>
      <FieldDescription>
        Controls how the autoscaler grows and shrinks this pool.
      </FieldDescription>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="fs-min">Minimum nodes</FieldLabel>
          <Input id="fs-min" type="number" defaultValue={1} />
        </Field>
        <Field>
          <FieldLabel htmlFor="fs-max">Maximum nodes</FieldLabel>
          <Input id="fs-max" type="number" defaultValue={6} />
        </Field>
      </FieldGroup>
    </FieldSet>
  </div>
)

export const CheckboxSet = () => (
  <div className="w-96">
    <FieldSet>
      <FieldLegend variant="label">Environment features</FieldLegend>
      <FieldDescription>Enable the add-ons this environment needs.</FieldDescription>
      <FieldGroup>
        <Field orientation="horizontal">
          <Checkbox id="fs-tunnel" defaultChecked />
          <FieldContent>
            <FieldTitle>Developer tunnel</FieldTitle>
            <FieldDescription>Route local traffic into the cluster.</FieldDescription>
          </FieldContent>
        </Field>
        <Field orientation="horizontal">
          <Checkbox id="fs-logs" />
          <FieldContent>
            <FieldTitle>Log retention</FieldTitle>
            <FieldDescription>Keep pod logs for 7 days.</FieldDescription>
          </FieldContent>
        </Field>
      </FieldGroup>
    </FieldSet>
  </div>
)
