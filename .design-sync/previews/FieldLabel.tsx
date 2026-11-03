import * as React from 'react'
import {
  Checkbox,
  Field,
  FieldContent,
  FieldDescription,
  FieldLabel,
  FieldTitle,
  Input,
} from '@kloudlite/ui'

export const InField = () => (
  <div className="w-96">
    <Field>
      <FieldLabel htmlFor="image-tag">Image tag</FieldLabel>
      <Input id="image-tag" defaultValue="api-gateway:1.14.2" />
      <FieldDescription>The tag pushed to your connected registry.</FieldDescription>
    </Field>
  </div>
)

export const WrappingAControl = () => (
  <div className="w-96">
    <FieldLabel htmlFor="spot-nodes">
      <Field orientation="horizontal">
        <Checkbox id="spot-nodes" defaultChecked />
        <FieldContent>
          <FieldTitle>Use spot instances</FieldTitle>
          <FieldDescription>
            Cheaper nodes that can be reclaimed. Recommended for non-production environments.
          </FieldDescription>
        </FieldContent>
      </Field>
    </FieldLabel>
  </div>
)
