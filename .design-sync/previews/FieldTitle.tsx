import * as React from 'react'
import {
  Badge,
  Checkbox,
  Field,
  FieldContent,
  FieldDescription,
  FieldTitle,
  Switch,
} from '@kloudlite/ui'

export const InCheckboxField = () => (
  <div className="w-96">
    <Field orientation="horizontal">
      <Checkbox id="ft-spot" defaultChecked />
      <FieldContent>
        <FieldTitle>Use spot instances</FieldTitle>
        <FieldDescription>
          The title labels the control when the row is not wrapped in a FieldLabel.
        </FieldDescription>
      </FieldContent>
    </Field>
  </div>
)

export const WithBadge = () => (
  <div className="w-96">
    <Field orientation="horizontal">
      <FieldContent>
        <FieldTitle>
          Sleep when idle
          <Badge variant="secondary">Beta</Badge>
        </FieldTitle>
        <FieldDescription>
          Scale the environment to zero after 30 minutes with no traffic.
        </FieldDescription>
      </FieldContent>
      <Switch id="ft-sleep" defaultChecked />
    </Field>
  </div>
)
