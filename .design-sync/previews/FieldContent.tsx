import * as React from 'react'
import {
  Checkbox,
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldTitle,
  Switch,
} from '@kloudlite/ui'

export const WithCheckbox = () => (
  <div className="w-96">
    <Field orientation="horizontal">
      <Checkbox id="fc-spot" defaultChecked />
      <FieldContent>
        <FieldTitle>Use spot instances</FieldTitle>
        <FieldDescription>
          Cheaper nodes that can be reclaimed at any time. Recommended for non-production
          environments only.
        </FieldDescription>
      </FieldContent>
    </Field>
  </div>
)

export const WithSwitch = () => (
  <div className="w-96">
    <FieldGroup>
      <Field orientation="horizontal">
        <FieldContent>
          <FieldTitle>Node pool autoscaling</FieldTitle>
          <FieldDescription>Grow the pool when pods stay pending.</FieldDescription>
        </FieldContent>
        <Switch id="fc-autoscale" defaultChecked />
      </Field>
      <Field orientation="horizontal">
        <FieldContent>
          <FieldTitle>Sleep when idle</FieldTitle>
          <FieldDescription>Scale to zero after 30 minutes of no traffic.</FieldDescription>
        </FieldContent>
        <Switch id="fc-sleep" />
      </Field>
    </FieldGroup>
  </div>
)
